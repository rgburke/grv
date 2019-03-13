package main

import (
	"fmt"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
)

// ChildWindowView provides an interface to the child view
// that extends the AbstractWindowView
type ChildWindowView interface {
	viewPos() ViewPos
	rows() uint
	viewDimension() ViewDimension
	onRowSelected(rowIndex uint) error
	line(rowIndex uint) string
}

// Lock represents a lock that can be locked and unlocked
type Lock interface {
	Lock()
	Unlock()
}

type abstractWindowViewHandler func(*AbstractWindowView, Action) error

// AbstractWindowView handles behaviour common to all window views
type AbstractWindowView struct {
	child         ChildWindowView
	channels      Channels
	config        Config
	variables     GRVVariableSetter
	viewSearch    *ViewSearch
	viewState     ViewState
	borderWidth   uint
	rowDescriptor string
	handlers      map[ActionType]abstractWindowViewHandler
	lock          Lock
}

// NewAbstractWindowView create a new instance
func NewAbstractWindowView(child ChildWindowView, channels Channels, config Config, variables GRVVariableSetter, lock Lock, rowDescriptor string) *AbstractWindowView {
	abstractWindowView := &AbstractWindowView{
		child:         child,
		channels:      channels,
		config:        config,
		variables:     variables,
		lock:          lock,
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

	abstractWindowView.viewSearch = NewViewSearch(abstractWindowView, channels)

	return abstractWindowView
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

// OnStateChange updates the active state of the view
func (abstractWindowView *AbstractWindowView) OnStateChange(viewState ViewState) {
	abstractWindowView.lock.Lock()
	defer abstractWindowView.lock.Unlock()

	abstractWindowView.viewState = viewState

	if viewState == ViewStateActive {
		abstractWindowView.setVariables()
	}
}

// Line returns the content of the line with the provided index
func (abstractWindowView *AbstractWindowView) Line(lineIndex uint) string {
	abstractWindowView.lock.Lock()
	defer abstractWindowView.lock.Unlock()

	return abstractWindowView.child.line(lineIndex)
}

// LineNumber returns the number of rows in this view
func (abstractWindowView *AbstractWindowView) LineNumber() uint {
	abstractWindowView.lock.Lock()
	defer abstractWindowView.lock.Unlock()

	return abstractWindowView.child.rows()
}

// ViewPos returns the current view position for this view
func (abstractWindowView *AbstractWindowView) ViewPos() ViewPos {
	abstractWindowView.lock.Lock()
	defer abstractWindowView.lock.Unlock()

	return abstractWindowView.child.viewPos()
}

// OnSearchMatch sets the active row to the search match row
// unless the position has been modified since the search started
func (abstractWindowView *AbstractWindowView) OnSearchMatch(startPos ViewPos, matchLineIndex uint) {
	abstractWindowView.lock.Lock()
	defer abstractWindowView.lock.Unlock()

	viewPos := abstractWindowView.child.viewPos()

	if viewPos != startPos {
		log.Debugf("Selected ref has changed since search started")
		return
	}

	if rows := abstractWindowView.child.rows(); matchLineIndex < rows {
		viewPos.SetActiveRowIndex(matchLineIndex)
		abstractWindowView.notifyChildRowSelected(matchLineIndex)
	} else {
		log.Warnf("Search match line index is greater than number of rows: %v > %v", matchLineIndex, rows)
	}
}

// HandleEvent does nothing
func (abstractWindowView *AbstractWindowView) HandleEvent(event Event) (err error) {
	return
}

func (abstractWindowView *AbstractWindowView) renderEmptyView(win RenderWindow, msg string) (err error) {
	viewPos := abstractWindowView.child.viewPos()
	startColumn := viewPos.ViewStartColumn()

	if err = win.SetRow(2, startColumn, CmpNone, "   %v", msg); err != nil {
		return
	}

	win.DrawBorder()

	return
}

func (abstractWindowView *AbstractWindowView) runReportingTask(message string, operation func(chan bool)) {
	abstractWindowView.channels.ReportStatus(message)

	go func() {
		quit := make(chan bool)

		operation(quit)

		ticker := time.NewTicker(time.Millisecond * 250)
		dots := 0

		for {
			select {
			case <-ticker.C:
				dots = (dots + 1) % 4
				abstractWindowView.channels.ReportStatus("%v%v", message, strings.Repeat(".", dots))
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()
}

func (abstractWindowView *AbstractWindowView) notifyChildRowSelected(rowIndex uint) (err error) {
	err = abstractWindowView.child.onRowSelected(rowIndex)
	abstractWindowView.setVariables()

	return
}

func (abstractWindowView *AbstractWindowView) setVariables() {
	rowIndex := abstractWindowView.child.viewPos().ActiveRowIndex()
	rows := abstractWindowView.child.rows()
	viewState := abstractWindowView.viewState

	abstractWindowView.variables.SetViewVariable(VarLineNumer, fmt.Sprintf("%v", rowIndex+1), viewState)
	abstractWindowView.variables.SetViewVariable(VarLineCount, fmt.Sprintf("%v", rows), viewState)

	line := abstractWindowView.child.line(rowIndex)
	abstractWindowView.variables.SetViewVariable(VarLineText, line, viewState)
}

// HandleAction checks if this action is supported by the AbstractWindowView
// and if so handles it
func (abstractWindowView *AbstractWindowView) HandleAction(action Action) (handled bool, err error) {
	if handled, err = abstractWindowView.viewSearch.HandleAction(action); handled {
		log.Debugf("Action handled by ViewSearch")
	} else if handler, ok := abstractWindowView.handlers[action.ActionType]; ok {
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
		err = abstractWindowView.notifyChildRowSelected(viewPos.ActiveRowIndex())
		abstractWindowView.channels.UpdateDisplay()
	}

	return
}

func moveDownRow(abstractWindowView *AbstractWindowView, action Action) (err error) {
	rows := abstractWindowView.child.rows()
	viewPos := abstractWindowView.child.viewPos()

	if viewPos.MoveLineDown(rows) {
		log.Debugf("Moving cursor down one %v", abstractWindowView.rowDescriptor)
		err = abstractWindowView.notifyChildRowSelected(viewPos.ActiveRowIndex())
		abstractWindowView.channels.UpdateDisplay()
	}

	return
}

func moveUpPage(abstractWindowView *AbstractWindowView, action Action) (err error) {
	viewPos := abstractWindowView.child.viewPos()

	if viewPos.MovePageUp(abstractWindowView.child.viewDimension().rows - abstractWindowView.borderWidth) {
		log.Debugf("Moving cursor up one page of %vs", abstractWindowView.rowDescriptor)
		err = abstractWindowView.notifyChildRowSelected(viewPos.ActiveRowIndex())
		abstractWindowView.channels.UpdateDisplay()
	}

	return
}

func moveDownPage(abstractWindowView *AbstractWindowView, action Action) (err error) {
	rows := abstractWindowView.child.rows()
	viewPos := abstractWindowView.child.viewPos()

	if viewPos.MovePageDown(abstractWindowView.child.viewDimension().rows-abstractWindowView.borderWidth, rows) {
		log.Debugf("Moving cursor down one page of %vs", abstractWindowView.rowDescriptor)
		err = abstractWindowView.notifyChildRowSelected(viewPos.ActiveRowIndex())
		abstractWindowView.channels.UpdateDisplay()
	}

	return
}

func moveUpHalfPage(abstractWindowView *AbstractWindowView, action Action) (err error) {
	viewPos := abstractWindowView.child.viewPos()

	if viewPos.MovePageUp(abstractWindowView.child.viewDimension().rows/2 - abstractWindowView.borderWidth) {
		log.Debugf("Moving cursor up half page of %vs", abstractWindowView.rowDescriptor)
		err = abstractWindowView.notifyChildRowSelected(viewPos.ActiveRowIndex())
		abstractWindowView.channels.UpdateDisplay()
	}

	return
}

func moveDownHalfPage(abstractWindowView *AbstractWindowView, action Action) (err error) {
	rows := abstractWindowView.child.rows()
	viewPos := abstractWindowView.child.viewPos()

	if viewPos.MovePageDown(abstractWindowView.child.viewDimension().rows/2-abstractWindowView.borderWidth, rows) {
		log.Debugf("Moving cursor down half page of %vs", abstractWindowView.rowDescriptor)
		err = abstractWindowView.notifyChildRowSelected(viewPos.ActiveRowIndex())
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
		err = abstractWindowView.notifyChildRowSelected(viewPos.ActiveRowIndex())
		abstractWindowView.channels.UpdateDisplay()
	}

	return
}

func moveToLastRow(abstractWindowView *AbstractWindowView, action Action) (err error) {
	rows := abstractWindowView.child.rows()
	viewPos := abstractWindowView.child.viewPos()

	if viewPos.MoveToLastLine(rows) {
		log.Debugf("Moving cursor to last %v", abstractWindowView.rowDescriptor)
		err = abstractWindowView.notifyChildRowSelected(viewPos.ActiveRowIndex())
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
		err = abstractWindowView.notifyChildRowSelected(viewPos.ActiveRowIndex())
		abstractWindowView.channels.UpdateDisplay()
	}

	return
}

func moveCursorMiddleOfView(abstractWindowView *AbstractWindowView, action Action) (err error) {
	rows := abstractWindowView.child.rows()
	viewPos := abstractWindowView.child.viewPos()

	if viewPos.MoveCursorMiddlePage(abstractWindowView.child.viewDimension().rows-abstractWindowView.borderWidth, rows) {
		log.Debugf("Moving cursor to %v in middle of view", abstractWindowView.rowDescriptor)
		err = abstractWindowView.notifyChildRowSelected(viewPos.ActiveRowIndex())
		abstractWindowView.channels.UpdateDisplay()
	}

	return
}

func moveCursorBottomOfView(abstractWindowView *AbstractWindowView, action Action) (err error) {
	rows := abstractWindowView.child.rows()
	viewPos := abstractWindowView.child.viewPos()

	if viewPos.MoveCursorBottomPage(abstractWindowView.child.viewDimension().rows-abstractWindowView.borderWidth, rows) {
		log.Debugf("Moving cursor to %v at bottom of view", abstractWindowView.rowDescriptor)
		err = abstractWindowView.notifyChildRowSelected(viewPos.ActiveRowIndex())
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

	err = abstractWindowView.notifyChildRowSelected(selectedIndex)
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
		err = abstractWindowView.notifyChildRowSelected(viewPos.ActiveRowIndex())
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
		err = abstractWindowView.notifyChildRowSelected(viewPos.ActiveRowIndex())
		abstractWindowView.channels.UpdateDisplay()
	}

	return
}
