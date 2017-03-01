package main

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
)

type WindowView interface {
	Initialise() error
	Render(RenderWindow) error
	Handle(KeyPressEvent) error
	OnActiveChange(bool)
}

type WindowViewCollection interface {
	Initialise() error
	Render(ViewDimension) ([]*Window, error)
	Handle(KeyPressEvent) error
	OnActiveChange(bool)
}

type ViewDimension struct {
	rows uint
	cols uint
}

type View struct {
	views           []WindowViewCollection
	activeViewIndex uint
}

func (viewDimension ViewDimension) String() string {
	return fmt.Sprintf("rows:%v,cols:%v", viewDimension.rows, viewDimension.cols)
}

func NewView(repoData RepoData, channels *Channels) (view *View) {
	view = &View{}
	view.views = []WindowViewCollection{
		NewHistoryView(repoData, channels),
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
	return view.views[view.activeViewIndex].Render(viewDimension)
}

func (view *View) Handle(keyPressEvent KeyPressEvent) error {
	log.Debugf("View handling key %v", keyPressEvent)
	return view.views[view.activeViewIndex].Handle(keyPressEvent)
}

func (view *View) OnActiveChange(active bool) {
	log.Debugf("View active %v", active)
	view.views[view.activeViewIndex].OnActiveChange(active)
}
