package main

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
	child       ChildWindowView
	channels    Channels
	config      Config
	borderWidth uint
	handlers    map[ActionType]abstractWindowViewHandler
}

// NewAbstractWindowView create a new instance
func NewAbstractWindowView(child ChildWindowView, channels Channels, config Config) *AbstractWindowView {
	return &AbstractWindowView{
		child:       child,
		channels:    channels,
		config:      config,
		borderWidth: 2,
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

// HandleAction checks if this action is supported by the AbstractWindowView
// and if so handles it
func (abstractWindowView *AbstractWindowView) HandleAction(action Action) (handled bool, err error) {
	if handler, ok := abstractWindowView.handlers[action.ActionType]; ok {
		err = handler(abstractWindowView, action)
		handled = true
	}

	return
}

func moveUpRow(abstractWindowView *AbstractWindowView, action Action) (err error) {
	viewPos := abstractWindowView.child.viewPos()

	if viewPos.MoveLineUp() {
		err = abstractWindowView.child.onRowSelected(viewPos.ActiveRowIndex())
		abstractWindowView.channels.UpdateDisplay()
	}

	return
}

func moveDownRow(abstractWindowView *AbstractWindowView, action Action) (err error) {
	rows := abstractWindowView.child.rows()
	viewPos := abstractWindowView.child.viewPos()

	if viewPos.MoveLineDown(rows) {
		err = abstractWindowView.child.onRowSelected(viewPos.ActiveRowIndex())
		abstractWindowView.channels.UpdateDisplay()
	}

	return
}

func moveUpPage(abstractWindowView *AbstractWindowView, action Action) (err error) {
	viewPos := abstractWindowView.child.viewPos()

	if viewPos.MovePageUp(abstractWindowView.child.viewDimension().rows - abstractWindowView.borderWidth) {
		err = abstractWindowView.child.onRowSelected(viewPos.ActiveRowIndex())
		abstractWindowView.channels.UpdateDisplay()
	}

	return
}

func moveDownPage(abstractWindowView *AbstractWindowView, action Action) (err error) {
	rows := abstractWindowView.child.rows()
	viewPos := abstractWindowView.child.viewPos()

	if viewPos.MovePageDown(abstractWindowView.child.viewDimension().rows-abstractWindowView.borderWidth, rows) {
		err = abstractWindowView.child.onRowSelected(viewPos.ActiveRowIndex())
		abstractWindowView.channels.UpdateDisplay()
	}

	return
}

func moveUpHalfPage(abstractWindowView *AbstractWindowView, action Action) (err error) {
	viewPos := abstractWindowView.child.viewPos()

	if viewPos.MovePageUp(abstractWindowView.child.viewDimension().rows/2 - abstractWindowView.borderWidth) {
		err = abstractWindowView.child.onRowSelected(viewPos.ActiveRowIndex())
		abstractWindowView.channels.UpdateDisplay()
	}

	return
}

func moveDownHalfPage(abstractWindowView *AbstractWindowView, action Action) (err error) {
	rows := abstractWindowView.child.rows()
	viewPos := abstractWindowView.child.viewPos()

	if viewPos.MovePageDown(abstractWindowView.child.viewDimension().rows/2-abstractWindowView.borderWidth, rows) {
		err = abstractWindowView.child.onRowSelected(viewPos.ActiveRowIndex())
		abstractWindowView.channels.UpdateDisplay()
	}

	return
}

func scrollRight(abstractWindowView *AbstractWindowView, action Action) (err error) {
	viewPos := abstractWindowView.child.viewPos()
	viewPos.MovePageRight(abstractWindowView.child.viewDimension().cols)
	abstractWindowView.channels.UpdateDisplay()

	return
}

func scrollLeft(abstractWindowView *AbstractWindowView, action Action) (err error) {
	viewPos := abstractWindowView.child.viewPos()

	if viewPos.MovePageLeft(abstractWindowView.child.viewDimension().cols) {
		abstractWindowView.channels.UpdateDisplay()
	}

	return
}

func moveToFirstRow(abstractWindowView *AbstractWindowView, action Action) (err error) {
	viewPos := abstractWindowView.child.viewPos()

	if viewPos.MoveToFirstLine() {
		err = abstractWindowView.child.onRowSelected(viewPos.ActiveRowIndex())
		abstractWindowView.channels.UpdateDisplay()
	}

	return
}

func moveToLastRow(abstractWindowView *AbstractWindowView, action Action) (err error) {
	rows := abstractWindowView.child.rows()
	viewPos := abstractWindowView.child.viewPos()

	if viewPos.MoveToLastLine(rows) {
		err = abstractWindowView.child.onRowSelected(viewPos.ActiveRowIndex())
		abstractWindowView.channels.UpdateDisplay()
	}

	return
}

func centerView(abstractWindowView *AbstractWindowView, action Action) (err error) {
	viewPos := abstractWindowView.child.viewPos()

	if viewPos.CenterActiveRow(abstractWindowView.child.viewDimension().rows - abstractWindowView.borderWidth) {
		abstractWindowView.channels.UpdateDisplay()
	}

	return
}

func scrollToViewTop(abstractWindowView *AbstractWindowView, action Action) (err error) {
	viewPos := abstractWindowView.child.viewPos()

	if viewPos.ScrollActiveRowTop() {
		abstractWindowView.channels.UpdateDisplay()
	}

	return
}

func scrollToViewBottom(abstractWindowView *AbstractWindowView, action Action) (err error) {
	viewPos := abstractWindowView.child.viewPos()

	if viewPos.ScrollActiveRowBottom(abstractWindowView.child.viewDimension().rows - abstractWindowView.borderWidth) {
		abstractWindowView.channels.UpdateDisplay()
	}

	return
}

func moveCursorTopOfView(abstractWindowView *AbstractWindowView, action Action) (err error) {
	viewPos := abstractWindowView.child.viewPos()

	if viewPos.MoveCursorTopPage() {
		err = abstractWindowView.child.onRowSelected(viewPos.ActiveRowIndex())
		abstractWindowView.channels.UpdateDisplay()
	}

	return
}

func moveCursorMiddleOfView(abstractWindowView *AbstractWindowView, action Action) (err error) {
	rows := abstractWindowView.child.rows()
	viewPos := abstractWindowView.child.viewPos()

	if viewPos.MoveCursorMiddlePage(abstractWindowView.child.viewDimension().rows-abstractWindowView.borderWidth, rows) {
		err = abstractWindowView.child.onRowSelected(viewPos.ActiveRowIndex())
		abstractWindowView.channels.UpdateDisplay()
	}

	return
}

func moveCursorBottomOfView(abstractWindowView *AbstractWindowView, action Action) (err error) {
	rows := abstractWindowView.child.rows()
	viewPos := abstractWindowView.child.viewPos()

	if viewPos.MoveCursorBottomPage(abstractWindowView.child.viewDimension().rows-abstractWindowView.borderWidth, rows) {
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
		err = abstractWindowView.child.onRowSelected(viewPos.ActiveRowIndex())
		abstractWindowView.channels.UpdateDisplay()
	}

	return
}
