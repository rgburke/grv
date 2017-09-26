package main

// Link against ncurses with wide character support in case goncurses doesn't

// #cgo !darwin pkg-config: ncursesw
// #cgo darwin openbsd LDFLAGS: -lncurses
// #include <stdlib.h>
// #include <locale.h>
// #include <sys/select.h>
// #include <sys/ioctl.h>
//
// static void grv_FD_ZERO(void *set) {
// 	FD_ZERO((fd_set *)set);
// }
//
// static void grv_FD_SET(int fd, void *set) {
// 	FD_SET(fd, (fd_set *)set);
// }
//
// static int grv_FD_ISSET(int fd, void *set) {
// 	return FD_ISSET(fd, (fd_set *)set);
// }
//
import "C"

import (
	"errors"
	"os"
	"sync"
	"syscall"
	"time"
	"unsafe"

	log "github.com/Sirupsen/logrus"
	gc "github.com/rgburke/goncurses"
)

const (
	// UINoKey is the value returned when there was no user input available
	UINoKey         = -1
	inputNoWinSleep = 50 * time.Millisecond
)

var systemColors = map[SystemColorValue]int16{
	ColorNone:    -1,
	ColorBlack:   gc.C_BLACK,
	ColorRed:     gc.C_RED,
	ColorGreen:   gc.C_GREEN,
	ColorYellow:  gc.C_YELLOW,
	ColorBlue:    gc.C_BLUE,
	ColorMagenta: gc.C_MAGENTA,
	ColorCyan:    gc.C_CYAN,
	ColorWhite:   gc.C_WHITE,
}

var convert256To16Color = []int16{
	0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
	0, 4, 4, 4, 12, 12, 2, 6, 4, 4, 12, 12, 2, 2, 6, 4,
	12, 12, 2, 2, 2, 6, 12, 12, 10, 10, 10, 10, 14, 12, 10, 10,
	10, 10, 10, 14, 1, 5, 4, 4, 12, 12, 3, 8, 4, 4, 12, 12,
	2, 2, 6, 4, 12, 12, 2, 2, 2, 6, 12, 12, 10, 10, 10, 10,
	14, 12, 10, 10, 10, 10, 10, 14, 1, 1, 5, 4, 12, 12, 1, 1,
	5, 4, 12, 12, 3, 3, 8, 4, 12, 12, 2, 2, 2, 6, 12, 12,
	10, 10, 10, 10, 14, 12, 10, 10, 10, 10, 10, 14, 1, 1, 1, 5,
	12, 12, 1, 1, 1, 5, 12, 12, 1, 1, 1, 5, 12, 12, 3, 3,
	3, 7, 12, 12, 10, 10, 10, 10, 14, 12, 10, 10, 10, 10, 10, 14,
	9, 9, 9, 9, 13, 12, 9, 9, 9, 9, 13, 12, 9, 9, 9, 9,
	13, 12, 9, 9, 9, 9, 13, 12, 11, 11, 11, 11, 7, 12, 10, 10,
	10, 10, 10, 14, 9, 9, 9, 9, 9, 13, 9, 9, 9, 9, 9, 13,
	9, 9, 9, 9, 9, 13, 9, 9, 9, 9, 9, 13, 9, 9, 9, 9,
	9, 13, 11, 11, 11, 11, 11, 15, 0, 0, 0, 0, 0, 0, 8, 8,
	8, 8, 8, 8, 7, 7, 7, 7, 7, 7, 15, 15, 15, 15, 15, 15,
}

var color256Components = []byte{0x00, 0x5f, 0x87, 0xaf, 0xd7, 0xff}

var color256GreyComponents = []byte{
	0x08, 0x12, 0x1c, 0x26, 0x30, 0x3a, 0x44, 0x4e,
	0x58, 0x62, 0x6c, 0x76, 0x80, 0x8a, 0x94, 0x9e,
	0xa8, 0xb2, 0xbc, 0xc6, 0xd0, 0xda, 0xe4, 0xee,
}

// Key is a raw code received from ncurses
type Key int

// InputUI is capable of providing input from the UI
type InputUI interface {
	GetInput(force bool) (Key, error)
	CancelGetInput() error
}

// UI exposes methods for updaing the display
type UI interface {
	InputUI
	Initialise() error
	Resize() error
	ViewDimension() ViewDimension
	Update([]*Window) error
	Suspend()
	Resume() error
	Free()
}

type signalPipe struct {
	read  *os.File
	write *os.File
}

func (signalPipe signalPipe) ReadFd() int {
	return int(signalPipe.read.Fd())
}

type nCursesWindow struct {
	*gc.Window
	isHidden bool
}

func (nwin *nCursesWindow) hidden() bool {
	return nwin.isHidden
}

func (nwin *nCursesWindow) setHidden(isHidden bool) {
	nwin.isHidden = isHidden
}

// NCursesUI implements the UI and InputUI interfaces
// It manages displaying grv in the terminal and receiving input
type NCursesUI struct {
	windows       map[*Window]*nCursesWindow
	lock          sync.Mutex
	stdscr        *nCursesWindow
	config        Config
	pipe          signalPipe
	maxColors     int
	maxColorPairs int
}

// NewNCursesDisplay creates a new NCursesUI instance
func NewNCursesDisplay(config Config) *NCursesUI {
	return &NCursesUI{
		windows: make(map[*Window]*nCursesWindow),
		config:  config,
	}
}

// Free releases ncurses resourese used
func (ui *NCursesUI) Free() {
	ui.lock.Lock()
	defer ui.lock.Unlock()

	ui.free()
}

func (ui *NCursesUI) free() {
	log.Info("Deleting NCurses windows")

	for _, nwin := range ui.windows {
		if err := nwin.Delete(); err != nil {
			log.Errorf("Error when deleting ncurses window: %v", err)
		}
	}

	ui.windows = make(map[*Window]*nCursesWindow)

	log.Info("Ending NCurses")
	gc.End()
}

// Initialise sets up NCurses
func (ui *NCursesUI) Initialise() (err error) {
	ui.lock.Lock()
	defer ui.lock.Unlock()

	log.Info("Initialising NCurses")

	emptyCString := C.CString("")
	C.setlocale(C.LC_ALL, emptyCString)
	C.free(unsafe.Pointer(emptyCString))

	if err = ui.initialiseNCurses(); err != nil {
		return
	}

	ui.config.AddOnChangeListener(CfTheme, ui)

	read, write, err := os.Pipe()
	if err != nil {
		return
	}

	ui.pipe = signalPipe{
		read:  read,
		write: write,
	}

	return
}

func (ui *NCursesUI) initialiseNCurses() (err error) {
	stdscr, err := gc.Init()
	if err != nil {
		return
	}

	ui.stdscr = &nCursesWindow{Window: stdscr}

	if gc.HasColors() {
		if e := gc.StartColor(); e != nil {
			log.Errorf("Error calling StartColor: %v", e)
		}

		if e := gc.UseDefaultColors(); e != nil {
			log.Errorf("Error calling UseDefaultColors: %v", e)
		}

		ui.maxColors = gc.Colors()
		ui.maxColorPairs = gc.ColorPairs()

		log.Infof("COLORS: %v, COLOR_PAIRS: %v", ui.maxColors, ui.maxColorPairs)

		theme := ui.config.GetTheme()
		ui.initialiseColorPairsFromTheme(theme)
	}

	gc.Echo(false)
	gc.Raw(true)

	if err = gc.Cursor(0); err != nil {
		return
	}

	if err = ui.stdscr.Keypad(true); err != nil {
		return
	}

	return
}

// Suspend ends ncurses to leave the terminal in the correct state when
// GRV is suspended
func (ui *NCursesUI) Suspend() {
	ui.lock.Lock()
	defer ui.lock.Unlock()

	gc.End()
}

// Resume reinitialises ncurses
func (ui *NCursesUI) Resume() (err error) {
	ui.lock.Lock()
	defer ui.lock.Unlock()

	ui.stdscr.Refresh()
	return ui.resize()
}

// Resize determines the current terminal dimensions reinitialises NCurses
func (ui *NCursesUI) Resize() (err error) {
	ui.lock.Lock()
	defer ui.lock.Unlock()

	return ui.resize()
}

func (ui *NCursesUI) resize() (err error) {
	log.Info("Resizing display")

	ui.free()

	var winSize C.struct_winsize

	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, os.Stdin.Fd(), C.TIOCGWINSZ, uintptr(unsafe.Pointer(&winSize)))
	if errno != 0 {
		return errno
	}

	if err = gc.ResizeTerm(int(winSize.ws_row), int(winSize.ws_col)); err != nil {
		return
	}

	return ui.initialiseNCurses()
}

// ViewDimension returns the dimensions of the terminal
func (ui *NCursesUI) ViewDimension() ViewDimension {
	ui.lock.Lock()
	defer ui.lock.Unlock()

	y, x := ui.stdscr.MaxYX()
	viewDimension := ViewDimension{rows: uint(y), cols: uint(x)}

	log.Debugf("Determining ViewDimension: %v", viewDimension)

	return viewDimension
}

// Update draws the provided windows to the terminal display
func (ui *NCursesUI) Update(wins []*Window) (err error) {
	ui.lock.Lock()
	defer ui.lock.Unlock()

	log.Debug("Updating display")

	if err = ui.createAndUpdateWindows(wins); err != nil {
		return
	}

	if err = ui.drawWindows(wins); err != nil {
		return
	}

	err = gc.Update()

	return
}

func (ui *NCursesUI) createAndUpdateWindows(wins []*Window) (err error) {
	log.Debug("Creating and updating NCurses windows")

	winMap := make(map[*Window]bool)

	for _, win := range wins {
		winMap[win] = true
	}

	for win, nwin := range ui.windows {
		if _, ok := winMap[win]; ok {
			nwin.Resize(int(win.rows), int(win.cols))
			nwin.MoveWindow(int(win.startRow), int(win.startCol))
			nwin.setHidden(false)
			log.Debugf("Moving NCurses window %v to row:%v,col:%v", win.ID(), win.startRow, win.startCol)
		} else if !nwin.hidden() {
			nwin.Erase()
			nwin.Resize(0, 0)
			nwin.NoutRefresh()
			nwin.setHidden(true)
			log.Debugf("Hiding NCurses window %v - %v:%v", win.ID())
		}
	}

	newWins := make([]*Window, 0)

	for _, win := range wins {
		if _, ok := ui.windows[win]; !ok {
			newWins = append(newWins, win)
		}
	}

	if len(newWins) > 0 {
		var nwinRaw *gc.Window
		var nwin *nCursesWindow

		for _, win := range newWins {
			log.Debugf("Creating new NCurses window %v with position row:%v,col:%v and dimensions rows:%v,cols:%v",
				win.ID(), win.startRow, win.startCol, win.rows, win.cols)
			if nwinRaw, err = gc.NewWindow(int(win.rows), int(win.cols), int(win.startRow), int(win.startCol)); err != nil {
				return
			}

			nwin = &nCursesWindow{Window: nwinRaw}

			if err = nwin.Keypad(true); err != nil {
				return
			}

			ui.windows[win] = nwin
		}
	}

	return
}

func (ui *NCursesUI) drawWindows(wins []*Window) (err error) {
	var cursorWin *Window

	for _, win := range wins {
		if nwin, ok := ui.windows[win]; ok {
			drawWindow(win, nwin)

			if win.IsCursorSet() {
				cursorWin = win
			}
		} else {
			err = errors.New("Algorithm error")
			return
		}
	}

	if cursorWin == nil {
		err = gc.Cursor(0)
	} else {
		if err = gc.Cursor(1); err != nil {
			return
		}

		nwin := ui.windows[cursorWin]
		nwin.Move(int(cursorWin.cursor.row), int(cursorWin.cursor.col))
		nwin.NoutRefresh()
	}

	return
}

func drawWindow(win *Window, nwin *nCursesWindow) {
	log.Debugf("Drawing window %v", win.ID())

	nwin.SetBackground(gc.ColorPair(int16(CmpAllviewDefault)))

	for rowIndex := uint(0); rowIndex < win.rows; rowIndex++ {
		line := win.lines[rowIndex]
		nwin.Move(int(rowIndex), 0)

		for colIndex := uint(0); colIndex < win.cols; colIndex++ {
			cell := line.cells[colIndex]

			if cell.style.acsChar != 0 {
				nwin.AddChar(cell.style.acsChar)
			} else if cell.codePoints.Len() > 0 {
				attr := cell.style.attr | gc.ColorPair(int16(cell.style.themeComponentID))
				if err := nwin.AttrOn(attr); err != nil {
					log.Errorf("Error when attempting to set AttrOn with %v: %v", attr, err)
				}

				nwin.Print(cell.codePoints.String())

				if err := nwin.AttrOff(attr); err != nil {
					log.Errorf("Error when attempting to set AttrOff with %v: %v", attr, err)
				}
			}
		}
	}

	nwin.NoutRefresh()
}

// GetInput blocks until user input is available
// A single key code is returned on each invocation
// Setting force = true makes this function non-blocking.
func (ui *NCursesUI) GetInput(force bool) (key Key, err error) {
	key = UINoKey

	if !force {
		rfds := &syscall.FdSet{}
		stdinFd := syscall.Stdin
		pipeFd := ui.pipe.ReadFd()

	OuterLoop:
		for {
			fdZero(rfds)
			fdSet(stdinFd, rfds)
			fdSet(pipeFd, rfds)
			nullPointer := uintptr(unsafe.Pointer(nil))

			if _, _, errno := syscall.Syscall6(syscall.SYS_SELECT, uintptr(pipeFd+1), uintptr(unsafe.Pointer(rfds)),
				nullPointer, nullPointer, nullPointer, 0); errno != 0 {
				err = errno
			}

			switch {
			case err != nil:
				return
			case fdIsset(pipeFd, rfds):
				if _, err := ui.pipe.read.Read(make([]byte, 8)); err != nil {
					log.Errorf("Error when reading from pipe: %v", err)
					continue
				}

				return
			case fdIsset(stdinFd, rfds) && !ReadLineActive():
				break OuterLoop
			}
		}
	}

	var activeWin *nCursesWindow

	ui.lock.Lock()
	for _, nwin := range ui.windows {
		if y, x := nwin.MaxYX(); y > 0 && x > 0 {
			activeWin = nwin
			break
		}
	}
	ui.lock.Unlock()

	if activeWin != nil {
		if force {
			activeWin.Timeout(0)
		}

		key = Key(activeWin.GetChar())

		if force {
			activeWin.Timeout(-1)
		}
	} else {
		time.Sleep(inputNoWinSleep)
		key = UINoKey
	}

	return
}

// CancelGetInput causes an invocation of GetInput (which is blocking) to return
func (ui *NCursesUI) CancelGetInput() error {
	ui.lock.Lock()
	defer ui.lock.Unlock()

	_, err := ui.pipe.write.Write([]byte{0})
	return err
}

func (ui *NCursesUI) onConfigVariableChange(configVariable ConfigVariable) {
	theme := ui.config.GetTheme()

	ui.lock.Lock()
	defer ui.lock.Unlock()

	ui.initialiseColorPairsFromTheme(theme)
}

func (ui *NCursesUI) initialiseColorPairsFromTheme(theme Theme) {
	defaultComponent := theme.GetComponent(CmpAllviewDefault)
	fgDefault := ui.getNCursesColor(defaultComponent.fgcolor)
	bgDefault := ui.getNCursesColor(defaultComponent.bgcolor)

	for themeComponentID, themeComponent := range theme.GetAllComponents() {
		if int(themeComponentID) >= ui.maxColorPairs {
			log.Errorf("Not enough color pairs for theme. Required: %v, Actual: %v",
				len(theme.GetAllComponents()), ui.maxColorPairs)
			break
		}

		fgcolor := ui.getNCursesColor(themeComponent.fgcolor)
		bgcolor := ui.getNCursesColor(themeComponent.bgcolor)

		if fgcolor == -1 {
			fgcolor = fgDefault
		}
		if bgcolor == -1 {
			bgcolor = bgDefault
		}

		log.Debugf("Initialising color pair for ThemeComponentID %v - %v:%v", themeComponentID, fgcolor, bgcolor)

		if err := gc.InitPair(int16(themeComponentID), fgcolor, bgcolor); err != nil {
			log.Errorf("Error when seting color pair %v:%v - %v", fgcolor, bgcolor, err)
		}
	}
}

func (ui *NCursesUI) getNCursesColor(themeColor ThemeColor) (colorNumber int16) {
	switch themeColor := themeColor.(type) {
	case *SystemColor:
		if systemColorNumber, ok := systemColors[themeColor.systemColorValue]; ok {
			colorNumber = systemColorNumber
		} else {
			log.Errorf("Invalid SystemColorValue: %v", themeColor.systemColorValue)
		}
	case *ColorNumber:
		colorNumber = themeColor.number
	case *RGBColor:
		redIndex := getColorComponentIndex(themeColor.red, color256Components)
		greenIndex := getColorComponentIndex(themeColor.green, color256Components)
		blueIndex := getColorComponentIndex(themeColor.blue, color256Components)

		greyRedIndex := getColorComponentIndex(themeColor.red, color256GreyComponents)
		greyGreenIndex := getColorComponentIndex(themeColor.green, color256GreyComponents)
		greyBlueIndex := getColorComponentIndex(themeColor.blue, color256GreyComponents)
		greyIndex := (greyRedIndex + greyGreenIndex + greyBlueIndex) / 3
		greyValue := color256GreyComponents[greyIndex]

		colorDistance := colorDistanceSquared(themeColor.red, themeColor.green, themeColor.blue,
			color256Components[redIndex], color256Components[greenIndex], color256Components[blueIndex])

		greyColorDistance := colorDistanceSquared(themeColor.red, themeColor.green, themeColor.blue,
			greyValue, greyValue, greyValue)

		if colorDistance < greyColorDistance {
			colorNumber = int16(16 + (36 * redIndex) + (6 * greenIndex) + blueIndex)
		} else {
			colorNumber = int16(232 + greyIndex)
		}
	default:
		log.Errorf("Unsupported ThemeColor type: %T", themeColor)
	}

	if colorNumber != -1 && ui.maxColors < 256 {
		colorNumber = convert256To16Color[colorNumber]

		if ui.maxColors < 16 && colorNumber > 8 {
			colorNumber -= 8
		}
	}

	return
}

func getColorComponentIndex(value byte, components []byte) int {
	low := 0
	high := len(components) - 1

	for low <= high {
		mid := (low + high) / 2

		if value < components[mid] {
			high = mid - 1
		} else if value > components[mid] {
			low = mid + 1
		} else {
			return mid
		}
	}

	if low > len(components)-1 {
		return high
	} else if high < 0 {
		return low
	} else if (components[low] - value) < (value - components[high]) {
		return low
	}

	return high
}

func colorDistanceSquared(r1, g1, b1, r2, g2, b2 byte) int {
	return (int(r1) - int(r2)) * (int(r1) - int(r2)) *
		(int(g1) - int(g2)) * (int(g1) - int(g2)) *
		(int(b1) - int(b2)) * (int(b1) - int(b2))
}

func fdZero(set *syscall.FdSet) {
	C.grv_FD_ZERO(unsafe.Pointer(set))
}

func fdSet(fd int, set *syscall.FdSet) {
	C.grv_FD_SET(C.int(fd), unsafe.Pointer(set))
}

func fdIsset(fd int, set *syscall.FdSet) bool {
	return C.grv_FD_ISSET(C.int(fd), unsafe.Pointer(set)) != 0
}
