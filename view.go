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
)

type AbstractView interface {
	Initialise() error
	HandleKeyPress(keystring string) error
	HandleAction(Action) error
	OnActiveChange(bool)
	ViewId() ViewId
	RenderStatusBar(RenderWindow) error
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
		win.OffsetPosition(int(activeViewDim.rows), 0)
	}

	return append(activeViewWins, statusViewWins...), err
}

func (view *View) RenderStatusBar(RenderWindow) (err error) {
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

	switch action {
	case ACTION_PROMPT:
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
