package main

import (
	"sync"

	log "github.com/Sirupsen/logrus"
)

// HelpView displays help information
type HelpView struct {
	*AbstractWindowView
	activeViewPos     ViewPos
	lastViewDimension ViewDimension
	lock              sync.Mutex
}

// NewHelpView creates a new instance
func NewHelpView(channels Channels, config Config) *HelpView {
	helpView := &HelpView{
		activeViewPos: NewViewPosition(),
	}

	helpView.AbstractWindowView = NewAbstractWindowView(helpView, channels, config, "help line")

	return helpView
}

// ViewID returns the ViewID of the help view
func (helpView *HelpView) ViewID() ViewID {
	return ViewHelp
}

// Render generates help information and writes it to the provided window
func (helpView *HelpView) Render(win RenderWindow) (err error) {
	helpView.lock.Lock()
	defer helpView.lock.Unlock()

	return
}

func (helpView *HelpView) viewPos() ViewPos {
	return helpView.activeViewPos
}

func (helpView *HelpView) rows() uint {
	return 0
}

func (helpView *HelpView) viewDimension() ViewDimension {
	return helpView.lastViewDimension
}

func (helpView *HelpView) onRowSelected(rowIndex uint) (err error) {
	return
}

// HandleAction handles the action if supported
func (helpView *HelpView) HandleAction(action Action) (err error) {
	helpView.lock.Lock()
	defer helpView.lock.Unlock()

	var handled bool
	if handled, err = helpView.AbstractWindowView.HandleAction(action); handled {
		log.Debugf("Action handled by AbstractWindowView")
	} else {
		log.Debugf("Action not handled")
	}

	return
}
