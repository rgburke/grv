package main

import (
	"sync"

	log "github.com/Sirupsen/logrus"
)

// StatusView manages the status bar view and help view displayed at the bottom of grv
type StatusView struct {
	statusBarView WindowView
	helpBarView   WindowView
	statusBarWin  *Window
	helpBarWin    *Window
	active        bool
	lock          sync.Mutex
}

// NewStatusView creates a new instance
func NewStatusView(rootView RootView, repoData RepoData, channels *Channels, config ConfigSetter) *StatusView {
	return &StatusView{
		statusBarView: NewStatusBarView(rootView, repoData, channels, config),
		helpBarView:   NewHelpBarView(rootView),
		statusBarWin:  NewWindow("statusBarView", config),
		helpBarWin:    NewWindow("helpBarView", config),
	}
}

// Initialise does nothing
func (statusView *StatusView) Initialise() (err error) {
	return
}

// HandleKeyPress does nothing
func (statusView *StatusView) HandleKeyPress(keystring string) (err error) {
	return
}

// HandleAction passes on the action to its child views for them to habdle
func (statusView *StatusView) HandleAction(action Action) (err error) {
	if err = statusView.statusBarView.HandleAction(action); err != nil {
		return
	}

	err = statusView.helpBarView.HandleAction(action)

	return
}

// OnActiveChange updates the active state of this view and its child views
func (statusView *StatusView) OnActiveChange(active bool) {
	statusView.lock.Lock()
	defer statusView.lock.Unlock()

	log.Debugf("StatusView active: %v", active)

	statusView.active = active
	statusView.statusBarView.OnActiveChange(active)
	statusView.helpBarView.OnActiveChange(active)
}

// ViewID returns the view ID of the status view
func (statusView *StatusView) ViewID() ViewID {
	return ViewStatus
}

// Render generates its child views and returns the windows that constitute the status view as a whole
func (statusView *StatusView) Render(viewDimension ViewDimension) (wins []*Window, err error) {
	statusView.lock.Lock()
	defer statusView.lock.Unlock()

	viewDimension.rows--

	statusView.statusBarWin.Resize(viewDimension)
	statusView.helpBarWin.Resize(viewDimension)

	statusView.statusBarWin.Clear()
	statusView.helpBarWin.Clear()

	if err = statusView.statusBarView.Render(statusView.statusBarWin); err != nil {
		return
	}

	if err = statusView.helpBarView.Render(statusView.helpBarWin); err != nil {
		return
	}

	statusView.statusBarWin.SetPosition(0, 0)
	statusView.helpBarWin.SetPosition(1, 0)

	wins = []*Window{statusView.statusBarWin, statusView.helpBarWin}

	return
}

// RenderStatusBar does nothing
func (statusView *StatusView) RenderStatusBar(lineBuilder *LineBuilder) (err error) {
	return
}

// RenderHelpBar does nothing
func (statusView *StatusView) RenderHelpBar(lineBuilder *LineBuilder) (err error) {
	return
}

// ActiveView returns the status bar view
// The help bar view is display only
func (statusView *StatusView) ActiveView() (childView AbstractView) {
	return statusView.statusBarView
}
