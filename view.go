package main

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"sync"
)

type ViewId int

const (
	VIEW_ALL ViewId = iota
	VIEW_MAIN
	VIEW_HISTORY
	VIEW_STATUS
	VIEW_REF
	VIEW_COMMIT
	VIEW_DIFF
	VIEW_STATUS_BAR
	VIEW_HELP_BAR
	VIEW_ERROR
)

type AbstractView interface {
	Initialise() error
	HandleKeyPress(keystring string) error
	HandleAction(Action) error
	OnActiveChange(bool)
	ViewId() ViewId
	RenderStatusBar(*LineBuilder) error
	RenderHelpBar(*LineBuilder) error
}

type WindowView interface {
	AbstractView
	Render(RenderWindow) error
}

type WindowViewCollection interface {
	AbstractView
	Render(ViewDimension) ([]*Window, error)
	ActiveView() AbstractView
}

type RootView interface {
	ActiveViewHierarchy() []AbstractView
	ActiveViewIdHierarchy() []ViewId
}

type ViewDimension struct {
	rows uint
	cols uint
}

type View struct {
	views         []WindowViewCollection
	activeViewPos uint
	statusView    WindowViewCollection
	channels      *Channels
	promptActive  bool
	errorView     *ErrorView
	errorViewWin  *Window
	errors        []error
	lock          sync.Mutex
}

func (viewDimension ViewDimension) String() string {
	return fmt.Sprintf("rows:%v,cols:%v", viewDimension.rows, viewDimension.cols)
}

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

	return
}

func (view *View) Initialise() (err error) {
	for _, childView := range view.views {
		if err = childView.Initialise(); err != nil {
			break
		}
	}

	view.OnActiveChange(true)

	return
}

func (view *View) Render(viewDimension ViewDimension) (wins []*Window, err error) {
	log.Debug("Rendering View")

	if viewDimension.rows < 3 {
		log.Errorf("Terminal is not large enough to render GRV")
		return
	}

	activeViewDim := viewDimension
	activeViewDim.rows -= 2

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

	activeViewWins, err := childView.Render(activeViewDim)
	if err != nil {
		return
	}

	statusViewWins, err := view.statusView.Render(statusViewDim)
	if err != nil {
		return
	}

	for _, win := range statusViewWins {
		win.OffsetPosition(int(activeViewDim.rows+errorViewDim.rows), 0)
	}

	wins = append(activeViewWins, statusViewWins...)

	if errorViewDim.rows > 0 {
		wins, err = view.renderErrorView(wins, errorViewDim, activeViewDim)
	}

	return wins, err
}

func (view *View) determineErrorViewDimensions(errorViewDim, activeViewDim *ViewDimension) {
	view.errorView.SetErrors(view.errors)
	view.errors = nil

	errorRowsRequired := view.errorView.DisplayRowsRequired()

	if activeViewDim.rows > errorRowsRequired {
		errorViewDim.rows = errorRowsRequired
		activeViewDim.rows -= errorRowsRequired
	} else {
		log.Errorf("Unable to display %v errors, not enough space", errorRowsRequired)
	}
}

func (view *View) renderErrorView(wins []*Window, errorViewDim, activeViewDim ViewDimension) (allWins []*Window, err error) {
	view.errorViewWin.Resize(errorViewDim)
	view.errorViewWin.Clear()
	view.errorViewWin.SetPosition(activeViewDim.rows, 0)

	if err = view.errorView.Render(view.errorViewWin); err != nil {
		return
	}

	allWins = append(wins, view.errorViewWin)

	return
}

func (view *View) RenderStatusBar(lineBuilder *LineBuilder) (err error) {
	return
}

func (view *View) RenderHelpBar(lineBuilder *LineBuilder) (err error) {
	view.lock.Lock()
	promptActive := view.promptActive
	view.lock.Unlock()

	if !promptActive {
		RenderKeyBindingHelp(view.ViewId(), lineBuilder, []ActionMessage{
			ActionMessage{action: ACTION_PROMPT, message: "Command Prompt"},
		})
	}

	return
}

func (view *View) HandleKeyPress(keystring string) error {
	log.Debugf("View handling keys %v", keystring)
	return view.ActiveView().HandleKeyPress(keystring)
}

func (view *View) HandleAction(action Action) (err error) {
	log.Debugf("View handling action %v", action)

	switch action.ActionType {
	case ACTION_PROMPT, ACTION_SEARCH_PROMPT, ACTION_REVERSE_SEARCH_PROMPT:
		view.prompt(action)
		return
	}

	return view.ActiveView().HandleAction(action)
}

func (view *View) OnActiveChange(active bool) {
	log.Debugf("View active %v", active)
	view.ActiveView().OnActiveChange(active)
}

func (view *View) ViewId() ViewId {
	return VIEW_MAIN
}

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

func (view *View) ActiveViewIdHierarchy() (viewIds []ViewId) {
	viewHierarchy := view.ActiveViewHierarchy()

	for _, activeView := range viewHierarchy {
		viewIds = append(viewIds, activeView.ViewId())
	}

	return
}

func (view *View) ActiveView() AbstractView {
	view.lock.Lock()
	defer view.lock.Unlock()

	if view.promptActive {
		return view.statusView
	}

	return view.views[view.activeViewPos]
}

func (view *View) prompt(action Action) {
	view.lock.Lock()
	view.views[view.activeViewPos].OnActiveChange(false)
	view.statusView.OnActiveChange(true)
	view.promptActive = true
	view.lock.Unlock()

	view.statusView.HandleAction(action)

	view.lock.Lock()
	view.promptActive = false
	view.statusView.OnActiveChange(false)
	view.views[view.activeViewPos].OnActiveChange(true)
	view.lock.Unlock()

	view.channels.UpdateDisplay()
}

func (view *View) SetErrors(errors []error) {
	view.errors = errors
}
