package main

import (
	"fmt"
	"sync"

	log "github.com/Sirupsen/logrus"
)

const (
	viewMinActiveViewRows = 6
)

// ViewID is an ID assigned to each view in grv
type ViewID int

// The set of view IDs
const (
	ViewAll ViewID = iota
	ViewMain
	ViewHistory
	ViewStatus
	ViewRef
	ViewCommit
	ViewDiff
	ViewStatusBar
	ViewHelpBar
	ViewError
)

// AbstractView exposes common functionality amongst all views
type AbstractView interface {
	Initialise() error
	HandleKeyPress(keystring string) error
	HandleAction(Action) error
	OnActiveChange(bool)
	ViewID() ViewID
	RenderStatusBar(*LineBuilder) error
	RenderHelpBar(*LineBuilder) error
}

// WindowView is a single window view
type WindowView interface {
	AbstractView
	Render(RenderWindow) error
}

// WindowViewCollection is a view that contains multiple child views
type WindowViewCollection interface {
	AbstractView
	Render(ViewDimension) ([]*Window, error)
	ActiveView() AbstractView
	Title() string
}

// RootView exposes functionality of the view at the top of the hierarchy
type RootView interface {
	ActiveViewHierarchy() []AbstractView
	ActiveViewIDHierarchy() []ViewID
}

// ViewDimension describes the size of a view
type ViewDimension struct {
	rows uint
	cols uint
}

// String returns a string representation of the view dimensions
func (viewDimension ViewDimension) String() string {
	return fmt.Sprintf("rows:%v,cols:%v", viewDimension.rows, viewDimension.cols)
}

// View is the top level view in grv
// All views in grv are children of this view
type View struct {
	views         []WindowViewCollection
	activeViewPos uint
	statusView    WindowViewCollection
	channels      *Channels
	promptActive  bool
	errorView     *ErrorView
	errorViewWin  *Window
	activeViewWin *Window
	errors        []error
	lock          sync.Mutex
}

// NewView creates a new instance
func NewView(repoData RepoData, channels *Channels, config ConfigSetter) (view *View) {
	view = &View{
		views: []WindowViewCollection{
			NewHistoryView(repoData, channels, config),
		},
		channels: channels,
	}

	view.statusView = NewStatusView(view, repoData, channels, config)
	view.errorView = NewErrorView()
	view.errorViewWin = NewWindow("errorView", config)
	view.activeViewWin = NewWindow("activeView", config)

	return
}

// Initialise sets up all child views
func (view *View) Initialise() (err error) {
	for _, childView := range view.views {
		if err = childView.Initialise(); err != nil {
			break
		}
	}

	view.OnActiveChange(true)

	return
}

// Render generates all windows to be drawn to the UI
func (view *View) Render(viewDimension ViewDimension) (wins []*Window, err error) {
	log.Debug("Rendering View")

	if viewDimension.rows < 4 {
		log.Errorf("Terminal is not large enough to render GRV")
		return
	}

	activeViewDim := viewDimension
	activeViewDim.rows -= 3

	statusViewDim := viewDimension
	statusViewDim.rows = 2

	errorViewDim := viewDimension
	errorViewDim.rows = 0

	if len(view.errors) > 0 {
		view.determineErrorViewDimensions(&errorViewDim, &activeViewDim)
	}

	view.lock.Lock()
	childView := view.views[view.activeViewPos]
	view.lock.Unlock()

	startRow := uint(0)
	if err = view.renderActiveView(activeViewDim.cols); err != nil {
		return
	}

	wins = append(wins, view.activeViewWin)
	startRow++

	activeViewWins, err := childView.Render(activeViewDim)
	if err != nil {
		return
	}

	for _, win := range activeViewWins {
		win.OffsetPosition(int(startRow), 0)
	}

	wins = append(wins, activeViewWins...)
	startRow += activeViewDim.rows

	if errorViewDim.rows > 0 {
		if wins, err = view.renderErrorView(wins, errorViewDim, activeViewDim); err != nil {
			return
		}

		view.errorViewWin.OffsetPosition(int(startRow), 0)
		startRow += errorViewDim.rows
	}

	statusViewWins, err := view.statusView.Render(statusViewDim)
	if err != nil {
		return
	}

	for _, win := range statusViewWins {
		win.OffsetPosition(int(startRow), 0)
	}

	wins = append(wins, statusViewWins...)

	return wins, err
}

func (view *View) determineErrorViewDimensions(errorViewDim, activeViewDim *ViewDimension) {
	view.errorView.SetErrors(view.errors)
	view.errors = nil

	errorRowsRequired := view.errorView.DisplayRowsRequired()

	if activeViewDim.rows > errorRowsRequired+viewMinActiveViewRows {
		errorViewDim.rows = errorRowsRequired
		activeViewDim.rows -= errorRowsRequired
	} else {
		log.Errorf("Unable to display all %v errors, not enough space", errorRowsRequired)

		if activeViewDim.rows > viewMinActiveViewRows {
			errorViewDim.rows = activeViewDim.rows - viewMinActiveViewRows
			activeViewDim.rows = viewMinActiveViewRows
		} else {
			log.Error("Unable to display any errors")
		}
	}
}

func (view *View) renderErrorView(wins []*Window, errorViewDim, activeViewDim ViewDimension) (allWins []*Window, err error) {
	view.errorViewWin.Resize(errorViewDim)
	view.errorViewWin.Clear()

	if err = view.errorView.Render(view.errorViewWin); err != nil {
		return
	}

	allWins = append(wins, view.errorViewWin)

	return
}

func (view *View) renderActiveView(availableCols uint) (err error) {
	viewTitles := make([]string, len(view.views))
	cols := uint(0)

	for index, childView := range view.views {
		viewTitles[index] = fmt.Sprintf(" [%v] %v ", index+1, childView.Title())
		cols += uint(len(viewTitles)) + 1
	}

	if cols > availableCols {
		maxColsPerView := availableCols / uint(len(viewTitles))

		for index, viewTitle := range viewTitles {
			if uint(len(viewTitle)) > maxColsPerView {
				viewTitles[index] = fmt.Sprintf("%*s ", maxColsPerView-1, viewTitles[index])
			}
		}
	}

	win := view.activeViewWin
	win.Resize(ViewDimension{rows: 1, cols: availableCols})
	win.Clear()
	win.SetPosition(0, 0)
	win.ApplyStyle(CmpMainviewNormalView)

	lineBuilder, err := view.activeViewWin.LineBuilder(0, 1)
	if err != nil {
		return
	}

	for index, viewTitle := range viewTitles {
		var themeComponentID ThemeComponentID

		if uint(index) == view.activeViewPos {
			themeComponentID = CmpMainviewActiveView
		} else {
			themeComponentID = CmpMainviewNormalView
		}

		lineBuilder.AppendWithStyle(themeComponentID, "%v", viewTitle)
	}

	return
}

// RenderStatusBar does nothing
func (view *View) RenderStatusBar(lineBuilder *LineBuilder) (err error) {
	return
}

// RenderHelpBar renders key binding help to the help bar for this view
func (view *View) RenderHelpBar(lineBuilder *LineBuilder) (err error) {
	view.lock.Lock()
	promptActive := view.promptActive
	view.lock.Unlock()

	if !promptActive {
		RenderKeyBindingHelp(view.ViewID(), lineBuilder, []ActionMessage{
			{action: ActionPrompt, message: "Command Prompt"},
		})
	}

	return
}

// HandleKeyPress passes the key press on to child view to handle
func (view *View) HandleKeyPress(keystring string) error {
	log.Debugf("View handling keys %v", keystring)
	return view.ActiveView().HandleKeyPress(keystring)
}

// HandleAction checks if this view can handle the action
// If not the action is passed down to child views to handle
func (view *View) HandleAction(action Action) (err error) {
	log.Debugf("View handling action %v", action)

	switch action.ActionType {
	case ActionPrompt, ActionSearchPrompt, ActionReverseSearchPrompt, ActionFilterPrompt:
		err = view.prompt(action)
		return
	case ActionShowStatus:
		err = view.statusView.HandleAction(action)
		return
	}

	return view.ActiveView().HandleAction(action)
}

// OnActiveChange updates the active state of the currently active child view
func (view *View) OnActiveChange(active bool) {
	log.Debugf("View active %v", active)
	view.ActiveView().OnActiveChange(active)
}

// ViewID returns the view ID of this view
func (view *View) ViewID() ViewID {
	return ViewMain
}

// ActiveViewHierarchy generates the currently active view hierarchy and returns the views that define it
func (view *View) ActiveViewHierarchy() []AbstractView {
	viewHierarchy := []AbstractView{view}
	var parentView WindowViewCollection = view
	var ok bool

	for {
		childView := parentView.ActiveView()
		viewHierarchy = append(viewHierarchy, childView)

		if parentView, ok = childView.(WindowViewCollection); !ok {
			break
		}
	}

	return viewHierarchy
}

// ActiveViewIDHierarchy generates the currently active view hierarchy and returns the view ID's that define it
func (view *View) ActiveViewIDHierarchy() (viewIds []ViewID) {
	viewHierarchy := view.ActiveViewHierarchy()

	for _, activeView := range viewHierarchy {
		viewIds = append(viewIds, activeView.ViewID())
	}

	return
}

// ActiveView returns the currently active child view
func (view *View) ActiveView() AbstractView {
	view.lock.Lock()
	defer view.lock.Unlock()

	if view.promptActive {
		return view.statusView
	}

	return view.views[view.activeViewPos]
}

func (view *View) prompt(action Action) (err error) {
	view.lock.Lock()
	view.views[view.activeViewPos].OnActiveChange(false)
	view.statusView.OnActiveChange(true)
	view.promptActive = true
	view.lock.Unlock()

	err = view.statusView.HandleAction(action)

	view.lock.Lock()
	view.promptActive = false
	view.statusView.OnActiveChange(false)
	view.views[view.activeViewPos].OnActiveChange(true)
	view.lock.Unlock()

	view.channels.UpdateDisplay()

	return
}

// SetErrors sets errors to be displayed in the error view
func (view *View) SetErrors(errors []error) {
	view.errors = errors
}

// Title returns the title of this view
func (view *View) Title() string {
	return "Main View"
}
