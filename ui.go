package main

import (
	"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
	gc "github.com/rthornton128/goncurses"
)

type UI interface {
	Initialise() error
	ViewDimension() ViewDimension
	Update([]*Window) error
	GetInput() (KeyPressEvent, error)
	ShowError(error)
	Free()
}

type NCursesUI struct {
	windows map[*Window]*gc.Window
	stdscr  *gc.Window
}

type KeyPressEvent struct {
	key gc.Key
}

func (keyPressEvent KeyPressEvent) String() string {
	return fmt.Sprintf("%c:%v", keyPressEvent.key, keyPressEvent.key)
}

func NewNcursesDisplay() *NCursesUI {
	return &NCursesUI{
		windows: make(map[*Window]*gc.Window),
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

	ui.stdscr, err = gc.Init()
	if err != nil {
		return
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

	for _, win := range wins {
		if nwin, ok := ui.windows[win]; !ok {
			log.Debugf("Creating new NCurses window %v with position row:%v,col:%v and dimensions rows:%v,cols:%v", win.Id(), win.startRow, win.startCol, win.rows, win.cols)
			if nwin, err = gc.NewWindow(int(win.rows), int(win.cols), int(win.startRow), int(win.startCol)); err != nil {
				return
			}

			if err = nwin.Keypad(true); err != nil {
				return
			}
			nwin.Timeout(0)
			ui.windows[win] = nwin
		}

	}

	return
}

func (ui *NCursesUI) drawWindows(wins []*Window) (err error) {
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
		row := win.cells[rowIndex]
		nwin.Move(int(rowIndex), 0)

		for colIndex := uint(0); colIndex < win.cols; colIndex++ {
			cell := row[colIndex]

			if cell.style.acs_char != 0 {
				nwin.AddChar(cell.style.acs_char)
			} else {
				nwin.AttrOn(cell.style.attr)
				nwin.Print(fmt.Sprintf("%c", cell.codePoint))
				nwin.AttrOff(cell.style.attr)
			}
		}
	}

	nwin.NoutRefresh()
}

func (ui *NCursesUI) GetInput() (keyPressEvent KeyPressEvent, err error) {
	for _, nwin := range ui.windows {
		if y, x := nwin.MaxYX(); y > 0 && x > 0 {
			keyPressEvent = KeyPressEvent{key: nwin.GetChar()}
			return
		}
	}

	err = errors.New("Unable to find active window to receive input from")
	return
}

func (ui *NCursesUI) ShowError(err error) {
	// TODO
}
