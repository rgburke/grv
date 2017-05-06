package main

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
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
	RenderHelpBar(RenderWindow) error
}

type WindowView interface {
	AbstractView
	Render(RenderWindow) error
}

type WindowViewCollection interface {
	AbstractView
	Render(ViewDimension) ([]*Window, error)
	ActiveViewHierarchy() []ViewId
	ActiveView() WindowView
}

type RootView interface {
	ActiveView() WindowView
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

	activeViewWins, err := view.views[view.activeViewPos].Render(activeViewDim)
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

func (view *View) RenderHelpBar(RenderWindow) (err error) {
	return
}

func (view *View) HandleKeyPress(keystring string) error {
	log.Debugf("View handling keys %v", keystring)
	return view.views[view.activeViewPos].HandleKeyPress(keystring)
}

func (view *View) HandleAction(action Action) error {
	log.Debugf("View handling action %v", action)

	switch action {
	case ACTION_PROMPT:
		view.prompt(action)
		return nil
	}

	return view.views[view.activeViewPos].HandleAction(action)
}

func (view *View) OnActiveChange(active bool) {
	log.Debugf("View active %v", active)
	view.views[view.activeViewPos].OnActiveChange(active)
}

func (view *View) ViewId() ViewId {
	return VIEW_MAIN
}

func (view *View) ActiveViewHierarchy() []ViewId {
	return append([]ViewId{view.ViewId()}, view.views[view.activeViewPos].ActiveViewHierarchy()...)
}

func (view *View) ActiveView() WindowView {
	return view.views[view.activeViewPos].ActiveView()
}

func (view *View) prompt(action Action) {
	view.views[view.activeViewPos].OnActiveChange(false)
	view.statusView.OnActiveChange(true)
	view.statusView.HandleAction(action)
	view.statusView.OnActiveChange(false)
	view.views[view.activeViewPos].OnActiveChange(true)
	log.Debug("view.go: Updating display")
	view.channels.UpdateDisplay()
}
