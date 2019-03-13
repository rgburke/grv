package main

import (
	"sync"
)

// GRVStatusView manages the status bar view and help view displayed at the bottom of grv
type GRVStatusView struct {
	statusBarView WindowView
	helpBarView   WindowView
	statusBarWin  *Window
	helpBarWin    *Window
	viewState     ViewState
	lock          sync.Mutex
}

// NewGRVStatusView creates a new instance
func NewGRVStatusView(helpRenderer HelpRenderer, repoData RepoData, channels Channels, config ConfigSetter) *GRVStatusView {
	return &GRVStatusView{
		statusBarView: NewStatusBarView(repoData, channels, config),
		helpBarView:   NewHelpBarView(helpRenderer),
		statusBarWin:  NewWindow("statusBarView", config),
		helpBarWin:    NewWindow("helpBarView", config),
	}
}

// Initialise does nothing
func (grvStatusView *GRVStatusView) Initialise() (err error) {
	return
}

// Dispose of any resources held by the view
func (grvStatusView *GRVStatusView) Dispose() {

}

// HandleEvent does nothing
func (grvStatusView *GRVStatusView) HandleEvent(event Event) (err error) {
	if err = grvStatusView.statusBarView.HandleEvent(event); err != nil {
		return
	}

	err = grvStatusView.helpBarView.HandleEvent(event)

	return
}

// HandleAction passes on the action to its child views for them to habdle
func (grvStatusView *GRVStatusView) HandleAction(action Action) (err error) {
	if err = grvStatusView.statusBarView.HandleAction(action); err != nil {
		return
	}

	err = grvStatusView.helpBarView.HandleAction(action)

	return
}

// OnStateChange updates the active state of this view and its child views
func (grvStatusView *GRVStatusView) OnStateChange(viewState ViewState) {
	grvStatusView.lock.Lock()
	defer grvStatusView.lock.Unlock()

	grvStatusView.viewState = viewState
	grvStatusView.statusBarView.OnStateChange(viewState)
	grvStatusView.helpBarView.OnStateChange(viewState)
}

// ViewID returns the view ID of the status view
func (grvStatusView *GRVStatusView) ViewID() ViewID {
	return ViewGRVStatus
}

// Render generates its child views and returns the windows that constitute the status view as a whole
func (grvStatusView *GRVStatusView) Render(viewDimension ViewDimension) (wins []*Window, err error) {
	grvStatusView.lock.Lock()
	defer grvStatusView.lock.Unlock()

	viewDimension.rows--

	grvStatusView.statusBarWin.Resize(viewDimension)
	grvStatusView.helpBarWin.Resize(viewDimension)

	grvStatusView.statusBarWin.Clear()
	grvStatusView.helpBarWin.Clear()

	if err = grvStatusView.statusBarView.Render(grvStatusView.statusBarWin); err != nil {
		return
	}

	if err = grvStatusView.helpBarView.Render(grvStatusView.helpBarWin); err != nil {
		return
	}

	grvStatusView.statusBarWin.SetPosition(0, 0)
	grvStatusView.helpBarWin.SetPosition(1, 0)

	wins = []*Window{grvStatusView.statusBarWin, grvStatusView.helpBarWin}

	return
}

// RenderHelpBar does nothing
func (grvStatusView *GRVStatusView) RenderHelpBar(lineBuilder *LineBuilder) (err error) {
	return grvStatusView.statusBarView.RenderHelpBar(lineBuilder)
}

// ActiveView returns the status bar view
// The help bar view is display only
func (grvStatusView *GRVStatusView) ActiveView() (childView BaseView) {
	return grvStatusView.statusBarView
}

// Title returns the title of the status view
func (grvStatusView *GRVStatusView) Title() string {
	return "Status View"
}
