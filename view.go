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
	VIEW_REF
	VIEW_COMMIT
	VIEW_DIFF
)

type AbstractView interface {
	Initialise() error
	HandleKeyPress(keystring string) error
	HandleAction(Action) error
	OnActiveChange(bool)
	ViewId() ViewId
}

type WindowView interface {
	AbstractView
	Render(RenderWindow) error
}

type WindowViewCollection interface {
	AbstractView
	Render(ViewDimension) ([]*Window, error)
	ActiveViewHierarchy() []ViewId
}

type ViewDimension struct {
	rows uint
	cols uint
}

type View struct {
	views         []WindowViewCollection
	activeViewPos uint
}

func (viewDimension ViewDimension) String() string {
	return fmt.Sprintf("rows:%v,cols:%v", viewDimension.rows, viewDimension.cols)
}

func NewView(repoData RepoData, channels *Channels, config Config) (view *View) {
	view = &View{}
	view.views = []WindowViewCollection{
		NewHistoryView(repoData, channels, config),
	}

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

func (view *View) Render(viewDimension ViewDimension) ([]*Window, error) {
	log.Debug("Rendering View")
	return view.views[view.activeViewPos].Render(viewDimension)
}

func (view *View) HandleKeyPress(keystring string) error {
	log.Debugf("View handling keys %v", keystring)
	return view.views[view.activeViewPos].HandleKeyPress(keystring)
}

func (view *View) HandleAction(action Action) error {
	log.Debugf("View handling action %v", action)
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
