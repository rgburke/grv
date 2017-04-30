package main

import (
	"sync"
)

type StatusBarView struct {
	rootView RootView
	repoData RepoData
	lock     sync.Mutex
}

func NewStatusBarView(rootView RootView, repoData RepoData) *StatusBarView {
	return &StatusBarView{
		rootView: rootView,
		repoData: repoData,
	}
}

func (statusBarView *StatusBarView) Initialise() (err error) {
	return
}

func (statusBarView *StatusBarView) HandleKeyPress(keystring string) (err error) {
	return
}

func (statusBarView *StatusBarView) HandleAction(Action) (err error) {
	return
}

func (statusBarView *StatusBarView) OnActiveChange(active bool) {
	return
}

func (statusBarView *StatusBarView) ViewId() ViewId {
	return VIEW_STATUS_BAR
}

func (statusBarView *StatusBarView) Render(win RenderWindow) (err error) {
	statusBarView.lock.Lock()
	defer statusBarView.lock.Unlock()

	lineBuilder, err := win.LineBuilder(0, 1)
	if err != nil {
		return
	}

	lineBuilder.Append(" %v", statusBarView.repoData.Path())
	win.ApplyStyle(CMP_STATUSBARVIEW_INFO)

	return
}

func (statusBarView *StatusBarView) RenderStatusBar(RenderWindow) (err error) {
	return
}

func (statusBarView *StatusBarView) RenderHelpBar(RenderWindow) (err error) {
	return
}
