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
	gc "github.com/rthornton128/goncurses"
)

const (
	// UINoKey is the value returned when there was no user input available
	UINoKey         = -1
	inputNoWinSleep = 50 * time.Millisecond
)

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
	windows map[*Window]*nCursesWindow
	lock    sync.Mutex
	stdscr  *nCursesWindow
	config  Config
	colors  map[ThemeColor]int16
	pipe    signalPipe
}

// NewNCursesDisplay creates a new NCursesUI instance
func NewNCursesDisplay(config Config) *NCursesUI {
	return &NCursesUI{
		windows: make(map[*Window]*nCursesWindow),
		config:  config,
		colors: map[ThemeColor]int16{
			ColorNone:    -1,
			ColorBlack:   gc.C_BLACK,
			ColorRed:     gc.C_RED,
			ColorGreen:   gc.C_GREEN,
			ColorYellow:  gc.C_YELLOW,
			ColorBlue:    gc.C_BLUE,
			ColorMagenta: gc.C_MAGENTA,
			ColorCyan:    gc.C_CYAN,
			ColorWhite:   gc.C_WHITE,
		},
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
			/* TODO: Find way not to avoid checking return value */
			syscall.Select(pipeFd+1, rfds, nil, nil, nil)

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
	for themeComponentID, themeComponent := range theme.GetAllComponents() {
		fgcolor := ui.colors[themeComponent.fgcolor]
		bgcolor := ui.colors[themeComponent.bgcolor]

		if err := gc.InitPair(int16(themeComponentID), fgcolor, bgcolor); err != nil {
			log.Errorf("Error when seting color pair %v:%v - %v", fgcolor, bgcolor, err)
		}
	}
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
