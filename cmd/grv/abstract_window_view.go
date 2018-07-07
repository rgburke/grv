package main

import (
	log "github.com/Sirupsen/logrus"
)

// ChildWindowView provides an interface to the child view
// that extends the AbstractWindowView
type ChildWindowView interface {
	viewPos() ViewPos
	rows() uint
	viewDimension() ViewDimension
	onRowSelected(rowIndex uint) error
}

type abstractWindowViewHandler func(*AbstractWindowView, Action) error

// AbstractWindowView handles behaviour common to all window views
type AbstractWindowView struct {
	child         ChildWindowView
	channels      Channels
	config        Config
	borderWidth   uint
	rowDescriptor string
	handlers      map[ActionType]abstractWindowViewHandler
}

// NewAbstractWindowView create a new instance
func NewAbstractWindowView(child ChildWindowView, channels Channels, config Config, rowDescriptor string) *AbstractWindowView {
	return &AbstractWindowView{
		child:         child,
		channels:      channels,
		config:        config,
		borderWidth:   2,
		rowDescriptor: rowDescriptor,
		handlers: map[ActionType]abstractWindowViewHandler{
			ActionPrevLine:           moveUpRow,
			ActionNextLine:           moveDownRow,
			ActionPrevPage:           moveUpPage,
			ActionNextPage:           moveDownPage,
			ActionPrevHalfPage:       moveUpHalfPage,
			ActionNextHalfPage:       moveDownHalfPage,
			ActionScrollRight:        scrollRight,
			ActionScrollLeft:         scrollLeft,
			ActionFirstLine:          moveToFirstRow,
			ActionLastLine:           moveToLastRow,
			ActionCenterView:         centerView,
			ActionScrollCursorTop:    scrollToViewTop,
			ActionScrollCursorBottom: scrollToViewBottom,
			ActionCursorTopView:      moveCursorTopOfView,
			ActionCursorMiddleView:   moveCursorMiddleOfView,
			ActionCursorBottomView:   moveCursorBottomOfView,
			ActionMouseSelect:        mouseSelectRow,
			ActionMouseScrollDown:    mouseScrollDown,
			ActionMouseScrollUp:      mouseScrollUp,
		},
	}
}

// Initialise does nothing
func (abstractWindowView *AbstractWindowView) Initialise() (err error) {
	return
}

// Dispose does nothing
func (abstractWindowView *AbstractWindowView) Dispose() {

}

// Render does nothing
func (abstractWindowView *AbstractWindowView) Render(win RenderWindow) (err error) {
	return
}

// RenderHelpBar does nothing
func (abstractWindowView *AbstractWindowView) RenderHelpBar(lineBuilder *LineBuilder) (err error) {
	return
}

// OnActiveChange does nothing
func (abstractWindowView *AbstractWindowView) OnActiveChange(active bool) {

}

// HandleEvent does nothing
func (abstractWindowView *AbstractWindowView) HandleEvent(event Event) (err error) {
	return
}

// RenderEmptyView renders an empty view displaying the provided message
func (abstractWindowView *AbstractWindowView) RenderEmptyView(win RenderWindow, msg string) (err error) {
	viewPos := abstractWindowView.child.viewPos()
	startColumn := viewPos.ViewStartColumn()

	if err = win.SetRow(2, startColumn, CmpNone, "   %v", msg); err != nil {
		return
	}

	win.DrawBorder()

	return
}

// HandleAction checks if this action is supported by the AbstractWindowView
// and if so handles it
func (abstractWindowView *AbstractWindowView) HandleAction(action Action) (handled bool, err error) {
	if handler, ok := abstractWindowView.handlers[action.ActionType]; ok {
		abstractWindowView.logViewPos("before")

		err = handler(abstractWindowView, action)
		handled = true

		abstractWindowView.logViewPos("after")
	}

	return
}

func (abstractWindowView *AbstractWindowView) logViewPos(stage string) {
	viewPos := abstractWindowView.child.viewPos()
	log.Debugf("ViewPos %v action: ActiveRowIndex:%v, ViewStartRowIndex:%v, ViewStartColumn:%v",
		stage, viewPos.ActiveRowIndex(), viewPos.ViewStartRowIndex(), viewPos.ViewStartColumn())
}

func moveUpRow(abstractWindowView *AbstractWindowView, action Action) (err error) {
	viewPos := abstractWindowView.child.viewPos()

	if viewPos.MoveLineUp() {
		log.Debugf("Moving cursor up one %v", abstractWindowView.rowDescriptor)
		err = abstractWindowView.child.onRowSelected(viewPos.ActiveRowIndex())
		abstractWindowView.channels.UpdateDisplay()
	}

	return
}

func moveDownRow(abstractWindowView *AbstractWindowView, action Action) (err error) {
	rows := abstractWindowView.child.rows()
	viewPos := abstractWindowView.child.viewPos()

	if viewPos.MoveLineDown(rows) {
		log.Debugf("Moving cursor down one %v", abstractWindowView.rowDescriptor)
		err = abstractWindowView.child.onRowSelected(viewPos.ActiveRowIndex())
		abstractWindowView.channels.UpdateDisplay()
	}

	return
}

func moveUpPage(abstractWindowView *AbstractWindowView, action Action) (err error) {
	viewPos := abstractWindowView.child.viewPos()

	if viewPos.MovePageUp(abstractWindowView.child.viewDimension().rows - abstractWindowView.borderWidth) {
		log.Debugf("Moving cursor up one page of %vs", abstractWindowView.rowDescriptor)
		err = abstractWindowView.child.onRowSelected(viewPos.ActiveRowIndex())
		abstractWindowView.channels.UpdateDisplay()
	}

	return
}

func moveDownPage(abstractWindowView *AbstractWindowView, action Action) (err error) {
	rows := abstractWindowView.child.rows()
	viewPos := abstractWindowView.child.viewPos()

	if viewPos.MovePageDown(abstractWindowView.child.viewDimension().rows-abstractWindowView.borderWidth, rows) {
		log.Debugf("Moving cursor down one page of %vs", abstractWindowView.rowDescriptor)
		err = abstractWindowView.child.onRowSelected(viewPos.ActiveRowIndex())
		abstractWindowView.channels.UpdateDisplay()
	}

	return
}

func moveUpHalfPage(abstractWindowView *AbstractWindowView, action Action) (err error) {
	viewPos := abstractWindowView.child.viewPos()

	if viewPos.MovePageUp(abstractWindowView.child.viewDimension().rows/2 - abstractWindowView.borderWidth) {
		log.Debugf("Moving cursor up half page of %vs", abstractWindowView.rowDescriptor)
		err = abstractWindowView.child.onRowSelected(viewPos.ActiveRowIndex())
		abstractWindowView.channels.UpdateDisplay()
	}

	return
}

func moveDownHalfPage(abstractWindowView *AbstractWindowView, action Action) (err error) {
	rows := abstractWindowView.child.rows()
	viewPos := abstractWindowView.child.viewPos()

	if viewPos.MovePageDown(abstractWindowView.child.viewDimension().rows/2-abstractWindowView.borderWidth, rows) {
		log.Debugf("Moving cursor down half page of %vs", abstractWindowView.rowDescriptor)
		err = abstractWindowView.child.onRowSelected(viewPos.ActiveRowIndex())
		abstractWindowView.channels.UpdateDisplay()
	}

	return
}

func scrollRight(abstractWindowView *AbstractWindowView, action Action) (err error) {
	log.Debug("Scrolling right")
	viewPos := abstractWindowView.child.viewPos()
	viewPos.MovePageRight(abstractWindowView.child.viewDimension().cols)
	abstractWindowView.channels.UpdateDisplay()

	return
}

func scrollLeft(abstractWindowView *AbstractWindowView, action Action) (err error) {
	viewPos := abstractWindowView.child.viewPos()

	if viewPos.MovePageLeft(abstractWindowView.child.viewDimension().cols) {
		log.Debug("Scrolling left")
		abstractWindowView.channels.UpdateDisplay()
	}

	return
}

func moveToFirstRow(abstractWindowView *AbstractWindowView, action Action) (err error) {
	viewPos := abstractWindowView.child.viewPos()

	if viewPos.MoveToFirstLine() {
		log.Debugf("Moving cursor to first %v", abstractWindowView.rowDescriptor)
		err = abstractWindowView.child.onRowSelected(viewPos.ActiveRowIndex())
		abstractWindowView.channels.UpdateDisplay()
	}

	return
}

func moveToLastRow(abstractWindowView *AbstractWindowView, action Action) (err error) {
	rows := abstractWindowView.child.rows()
	viewPos := abstractWindowView.child.viewPos()

	if viewPos.MoveToLastLine(rows) {
		log.Debugf("Moving cursor to last %v", abstractWindowView.rowDescriptor)
		err = abstractWindowView.child.onRowSelected(viewPos.ActiveRowIndex())
		abstractWindowView.channels.UpdateDisplay()
	}

	return
}

func centerView(abstractWindowView *AbstractWindowView, action Action) (err error) {
	viewPos := abstractWindowView.child.viewPos()

	if viewPos.CenterActiveRow(abstractWindowView.child.viewDimension().rows - abstractWindowView.borderWidth) {
		log.Debugf("Centering view on selected %v", abstractWindowView.rowDescriptor)
		abstractWindowView.channels.UpdateDisplay()
	}

	return
}

func scrollToViewTop(abstractWindowView *AbstractWindowView, action Action) (err error) {
	viewPos := abstractWindowView.child.viewPos()

	if viewPos.ScrollActiveRowTop() {
		log.Debugf("Scrolling view start to selected %v", abstractWindowView.rowDescriptor)
		abstractWindowView.channels.UpdateDisplay()
	}

	return
}

func scrollToViewBottom(abstractWindowView *AbstractWindowView, action Action) (err error) {
	viewPos := abstractWindowView.child.viewPos()

	if viewPos.ScrollActiveRowBottom(abstractWindowView.child.viewDimension().rows - abstractWindowView.borderWidth) {
		log.Debugf("Scrolling view end to selected %v", abstractWindowView.rowDescriptor)
		abstractWindowView.channels.UpdateDisplay()
	}

	return
}

func moveCursorTopOfView(abstractWindowView *AbstractWindowView, action Action) (err error) {
	viewPos := abstractWindowView.child.viewPos()

	if viewPos.MoveCursorTopPage() {
		log.Debugf("Moving cursor to %v at top of view", abstractWindowView.rowDescriptor)
		err = abstractWindowView.child.onRowSelected(viewPos.ActiveRowIndex())
		abstractWindowView.channels.UpdateDisplay()
	}

	return
}

func moveCursorMiddleOfView(abstractWindowView *AbstractWindowView, action Action) (err error) {
	rows := abstractWindowView.child.rows()
	viewPos := abstractWindowView.child.viewPos()

	if viewPos.MoveCursorMiddlePage(abstractWindowView.child.viewDimension().rows-abstractWindowView.borderWidth, rows) {
		log.Debugf("Moving cursor to %v in middle of view", abstractWindowView.rowDescriptor)
		err = abstractWindowView.child.onRowSelected(viewPos.ActiveRowIndex())
		abstractWindowView.channels.UpdateDisplay()
	}

	return
}

func moveCursorBottomOfView(abstractWindowView *AbstractWindowView, action Action) (err error) {
	rows := abstractWindowView.child.rows()
	viewPos := abstractWindowView.child.viewPos()

	if viewPos.MoveCursorBottomPage(abstractWindowView.child.viewDimension().rows-abstractWindowView.borderWidth, rows) {
		log.Debugf("Moving cursor to %v at bottom of view", abstractWindowView.rowDescriptor)
		err = abstractWindowView.child.onRowSelected(viewPos.ActiveRowIndex())
		abstractWindowView.channels.UpdateDisplay()
	}

	return
}

func mouseSelectRow(abstractWindowView *AbstractWindowView, action Action) (err error) {
	mouseEvent, err := GetMouseEventFromAction(action)
	if err != nil {
		return
	}

	if mouseEvent.row == 0 || mouseEvent.row == abstractWindowView.child.viewDimension().rows-1 {
		return
	}

	viewPos := abstractWindowView.child.viewPos()
	selectedIndex := viewPos.ViewStartRowIndex() + mouseEvent.row - 1

	if selectedIndex >= abstractWindowView.child.rows() {
		return
	}

	log.Debugf("Mouse selected %v at index: %v", abstractWindowView.rowDescriptor, selectedIndex)
	viewPos.SetActiveRowIndex(selectedIndex)

	err = abstractWindowView.child.onRowSelected(selectedIndex)
	abstractWindowView.channels.UpdateDisplay()

	return
}

func mouseScrollDown(abstractWindowView *AbstractWindowView, action Action) (err error) {
	viewPos := abstractWindowView.child.viewPos()
	rows := abstractWindowView.child.rows()
	pageRows := abstractWindowView.child.viewDimension().rows - abstractWindowView.borderWidth
	scrollRows := uint(abstractWindowView.config.GetInt(CfMouseScrollRows))

	if viewPos.ScrollDown(rows, pageRows, scrollRows) {
		log.Debugf("Mouse scrolled down %v %vs", scrollRows, abstractWindowView.rowDescriptor)
		err = abstractWindowView.child.onRowSelected(viewPos.ActiveRowIndex())
		abstractWindowView.channels.UpdateDisplay()
	}

	return
}

func mouseScrollUp(abstractWindowView *AbstractWindowView, action Action) (err error) {
	viewPos := abstractWindowView.child.viewPos()
	pageRows := abstractWindowView.child.viewDimension().rows - abstractWindowView.borderWidth
	scrollRows := uint(abstractWindowView.config.GetInt(CfMouseScrollRows))

	if viewPos.ScrollUp(pageRows, scrollRows) {
		log.Debugf("Mouse scrolled up %v %vs", scrollRows, abstractWindowView.rowDescriptor)
		err = abstractWindowView.child.onRowSelected(viewPos.ActiveRowIndex())
		abstractWindowView.channels.UpdateDisplay()
	}

	return
}
