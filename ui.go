package main

// Link against ncurses with wide character support in case goncurses doesn't

// #cgo pkg-config: ncursesw
// #include <stdlib.h>
// #include <locale.h>
// #include <sys/select.h>
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
	log "github.com/Sirupsen/logrus"
	gc "github.com/rthornton128/goncurses"
	"os"
	"sync"
	"syscall"
	"time"
	"unsafe"
)

const (
	INPUT_NO_WIN_SLEEP_MS = 50 * time.Millisecond
	UI_NO_KEY             = -1
)

type Key int

type InputUI interface {
	GetInput(force bool) (Key, error)
	CancelGetInput() error
}

type UI interface {
	InputUI
	Initialise() error
	ViewDimension() ViewDimension
	Update([]*Window) error
	ShowError(error)
	Free()
}

type SignalPipe struct {
	read  *os.File
	write *os.File
}

func (signalPipe SignalPipe) ReadFd() int {
	return int(signalPipe.read.Fd())
}

type NCursesUI struct {
	windows     map[*Window]*gc.Window
	windowsLock sync.RWMutex
	stdscr      *gc.Window
	config      Config
	colors      map[ThemeColor]int16
	pipe        SignalPipe
}

func NewNcursesDisplay(config Config) *NCursesUI {
	return &NCursesUI{
		windows: make(map[*Window]*gc.Window),
		config:  config,
		colors: map[ThemeColor]int16{
			COLOR_NONE:    -1,
			COLOR_BLACK:   gc.C_BLACK,
			COLOR_RED:     gc.C_RED,
			COLOR_GREEN:   gc.C_GREEN,
			COLOR_YELLOW:  gc.C_YELLOW,
			COLOR_BLUE:    gc.C_BLUE,
			COLOR_MAGENTA: gc.C_MAGENTA,
			COLOR_CYAN:    gc.C_CYAN,
			COLOR_WHITE:   gc.C_WHITE,
		},
	}
}

func (ui *NCursesUI) Free() {
	log.Info("Deleting NCurses windows")

	for _, nwin := range ui.windows {
		nwin.Delete()
	}

	log.Info("Ending NCurses")
	gc.End()
}

func (ui *NCursesUI) Initialise() (err error) {
	log.Info("Initialising NCurses")

	emptyCString := C.CString("")
	C.setlocale(C.LC_ALL, emptyCString)
	C.free(unsafe.Pointer(emptyCString))

	ui.stdscr, err = gc.Init()
	if err != nil {
		return
	}

	if gc.HasColors() {
		gc.StartColor()
		gc.UseDefaultColors()
		ui.onConfigVariableChange(CV_THEME)
	}

	gc.Echo(false)
	gc.Raw(true)

	if err = gc.Cursor(0); err != nil {
		return
	}

	if err = ui.stdscr.Keypad(true); err != nil {
		return
	}

	ui.config.AddOnChangeListener(CV_THEME, ui)

	read, write, err := os.Pipe()
	if err != nil {
		return
	}

	ui.pipe = SignalPipe{
		read:  read,
		write: write,
	}

	return
}

func (ui *NCursesUI) ViewDimension() ViewDimension {
	y, x := ui.stdscr.MaxYX()
	viewDimension := ViewDimension{rows: uint(y), cols: uint(x)}

	log.Debugf("Determining ViewDimension: %v", viewDimension)

	return viewDimension
}

func (ui *NCursesUI) Update(wins []*Window) (err error) {
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

	ui.windowsLock.RLock()
	for _, win := range wins {
		winMap[win] = true
	}

	for win, nwin := range ui.windows {
		if _, ok := winMap[win]; ok {
			nwin.Resize(int(win.rows), int(win.cols))
			nwin.MoveWindow(int(win.startRow), int(win.startCol))
			log.Debugf("Moving NCurses window %v to row:%v,col:%v", win.Id(), win.startRow, win.startCol)
		} else {
			nwin.Resize(0, 0)
			nwin.MoveWindow(0, 0)
			nwin.NoutRefresh()
			log.Debugf("Hiding NCurses window %v", win.Id())
		}
	}

	newWins := make([]*Window, 0)

	for _, win := range wins {
		if _, ok := ui.windows[win]; !ok {
			newWins = append(newWins, win)
		}
	}
	ui.windowsLock.RUnlock()

	if len(newWins) > 0 {
		ui.windowsLock.Lock()
		defer ui.windowsLock.Unlock()
		var nwin *gc.Window

		for _, win := range newWins {
			log.Debugf("Creating new NCurses window %v with position row:%v,col:%v and dimensions rows:%v,cols:%v", win.Id(), win.startRow, win.startCol, win.rows, win.cols)
			if nwin, err = gc.NewWindow(int(win.rows), int(win.cols), int(win.startRow), int(win.startCol)); err != nil {
				return
			}

			if err = nwin.Keypad(true); err != nil {
				return
			}

			ui.windows[win] = nwin
		}
	}

	return
}

func (ui *NCursesUI) drawWindows(wins []*Window) (err error) {
	ui.windowsLock.RLock()
	defer ui.windowsLock.RUnlock()
	var cursorWin *Window

	for _, win := range wins {
		if nwin, ok := ui.windows[win]; ok {
			drawWindow(win, nwin)

			if win.IsCursorSet() {
				cursorWin = win
			}
		} else {
			err = errors.New("Algorithm error")
			break
		}
	}

	if cursorWin == nil {
		gc.Cursor(0)
	} else {
		gc.Cursor(1)
		nwin := ui.windows[cursorWin]
		nwin.Move(int(cursorWin.cursor.row), int(cursorWin.cursor.col))
		nwin.NoutRefresh()
	}

	return
}

func drawWindow(win *Window, nwin *gc.Window) {
	log.Debugf("Drawing window %v", win.Id())

	for rowIndex := uint(0); rowIndex < win.rows; rowIndex++ {
		line := win.lines[rowIndex]
		nwin.Move(int(rowIndex), 0)

		for colIndex := uint(0); colIndex < win.cols; colIndex++ {
			cell := line.cells[colIndex]

			if cell.style.acs_char != 0 {
				nwin.AddChar(cell.style.acs_char)
			} else if cell.codePoints.Len() > 0 {
				attr := cell.style.attr | gc.ColorPair(int16(cell.style.componentId))
				nwin.AttrOn(attr)
				nwin.Print(cell.codePoints.String())
				nwin.AttrOff(attr)
			}
		}
	}

	nwin.NoutRefresh()
}

func (ui *NCursesUI) GetInput(force bool) (key Key, err error) {
	key = UI_NO_KEY

	if !force {
		rfds := &syscall.FdSet{}
		stdinFd := syscall.Stdin
		pipeFd := ui.pipe.ReadFd()

	OuterLoop:
		for {
			FD_ZERO(rfds)
			FD_SET(stdinFd, rfds)
			FD_SET(pipeFd, rfds)
			_, err = syscall.Select(pipeFd+1, rfds, nil, nil, nil)

			switch {
			case err != nil:
				return
			case FD_ISSET(pipeFd, rfds):
				ui.pipe.read.Read(make([]byte, 8))
				return
			case FD_ISSET(stdinFd, rfds) && !ReadLineActive():
				break OuterLoop
			}
		}
	}

	var activeWin *gc.Window

	ui.windowsLock.RLock()
	for _, nwin := range ui.windows {
		if y, x := nwin.MaxYX(); y > 0 && x > 0 {
			activeWin = nwin
			break
		}
	}
	ui.windowsLock.RUnlock()

	if activeWin != nil {
		if force {
			activeWin.Timeout(0)
		}

		key = Key(activeWin.GetChar())

		if force {
			activeWin.Timeout(-1)
		}
	} else {
		time.Sleep(INPUT_NO_WIN_SLEEP_MS)
		key = UI_NO_KEY
	}

	return
}

func (ui *NCursesUI) CancelGetInput() error {
	_, err := ui.pipe.write.Write([]byte{0})
	return err
}

func (ui *NCursesUI) ShowError(err error) {
	// TODO
}

func (ui *NCursesUI) onConfigVariableChange(configVariable ConfigVariable) {
	theme := ui.config.GetTheme()

	for themeComponentId, themeComponent := range theme.GetAllComponents() {
		fgcolor := ui.colors[themeComponent.fgcolor]
		bgcolor := ui.colors[themeComponent.bgcolor]
		gc.InitPair(int16(themeComponentId), fgcolor, bgcolor)
	}
}

func FD_ZERO(set *syscall.FdSet) {
	C.grv_FD_ZERO(unsafe.Pointer(set))
}

func FD_SET(fd int, set *syscall.FdSet) {
	C.grv_FD_SET(C.int(fd), unsafe.Pointer(set))
}

func FD_ISSET(fd int, set *syscall.FdSet) bool {
	return C.grv_FD_ISSET(C.int(fd), unsafe.Pointer(set)) != 0
}
