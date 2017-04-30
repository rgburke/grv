package main

import (
	"sync"
)

type StatusView struct {
	statusBarView WindowView
	helpBarView   WindowView
	statusBarWin  *Window
	helpBarWin    *Window
	active        bool
	lock          sync.Mutex
}

func NewStatusView(rootView RootView, repoData RepoData, config Config) *StatusView {
	return &StatusView{
		statusBarView: NewStatusBarView(rootView, repoData),
		helpBarView:   NewHelpBarView(rootView),
		statusBarWin:  NewWindow("statusBarView", config),
		helpBarWin:    NewWindow("helpBarView", config),
	}
}

func (statusView *StatusView) Initialise() (err error) {
	return
}

func (statusView *StatusView) HandleKeyPress(keystring string) (err error) {
	return
}

func (statusView *StatusView) HandleAction(Action) (err error) {
	return
}

func (statusView *StatusView) OnActiveChange(active bool) {
	statusView.lock.Lock()
	defer statusView.lock.Unlock()

	statusView.active = active
	statusView.statusBarView.OnActiveChange(active)
	statusView.helpBarView.OnActiveChange(active)
}

func (statusView *StatusView) ViewId() ViewId {
	return VIEW_STATUS
}

func (statusView *StatusView) Render(viewDimension ViewDimension) (wins []*Window, err error) {
	statusView.lock.Lock()
	defer statusView.lock.Unlock()

	viewDimension.rows -= 1

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

func (statusView *StatusView) RenderStatusBar(RenderWindow) (err error) {
	return
}

func (statusView *StatusView) RenderHelpBar(RenderWindow) (err error) {
	return
}

func (statusView *StatusView) ActiveViewHierarchy() (viewIds []ViewId) {
	return
}

func (statusView *StatusView) ActiveView() (windowView WindowView) {
	return
}
