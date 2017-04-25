package main

// Link against ncurses with wide character support in case goncurses doesn't

// #cgo pkg-config: ncursesw
// #include <stdlib.h>
// #include <locale.h>
import "C"

import (
	"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
	gc "github.com/rthornton128/goncurses"
	"sync"
	"time"
	"unsafe"
)

const (
	INPUT_NO_WIN_SLEEP_MS = 50 * time.Millisecond
)

type InputUI interface {
	GetInput() (KeyPressEvent, error)
}

type UI interface {
	Initialise() error
	ViewDimension() ViewDimension
	Update([]*Window) error
	ShowError(error)
	Free()
}

type NCursesUI struct {
	windows     map[*Window]*gc.Window
	windowsLock sync.RWMutex
	stdscr      *gc.Window
	config      Config
	colors      map[ThemeColor]int16
}

type KeyPressEvent struct {
	key gc.Key
}

func (keyPressEvent KeyPressEvent) String() string {
	return fmt.Sprintf("%c:%v", keyPressEvent.key, keyPressEvent.key)
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

	for _, win := range wins {
		if nwin, ok := ui.windows[win]; ok {
			drawWindow(win, nwin)
		} else {
			err = errors.New("Algorithm error")
			break
		}
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

func (ui *NCursesUI) GetInput() (keyPressEvent KeyPressEvent, err error) {
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
		keyPressEvent = KeyPressEvent{key: activeWin.GetChar()}
	} else {
		time.Sleep(INPUT_NO_WIN_SLEEP_MS)
	}

	return
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
