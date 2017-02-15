package main

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
)

type WindowView interface {
	Initialise(HandlerChannels) error
	Render(RenderWindow) error
	Handle(KeyPressEvent, HandlerChannels) error
	OnActiveChange(bool)
}

type WindowViewCollection interface {
	Initialise(HandlerChannels) error
	Render(ViewDimension) ([]*Window, error)
	Handle(KeyPressEvent, HandlerChannels) error
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

func NewView(repoData RepoData) (view *View) {
	view = &View{}
	view.views = []WindowViewCollection{
		NewHistoryView(repoData),
	}

	return
}

func (view *View) Initialise(channels HandlerChannels) (err error) {
	for _, childView := range view.views {
		if err = childView.Initialise(channels); err != nil {
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

func (view *View) Handle(keyPressEvent KeyPressEvent, channels HandlerChannels) error {
	log.Debugf("View handling key %v", keyPressEvent)
	return view.views[view.activeViewIndex].Handle(keyPressEvent, channels)
}

func (view *View) OnActiveChange(active bool) {
	log.Debugf("View active %v", active)
	view.views[view.activeViewIndex].OnActiveChange(active)
}
