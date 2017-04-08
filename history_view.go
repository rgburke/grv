package main

import (
	log "github.com/Sirupsen/logrus"
)

const (
	HV_BRANCH_VIEW_WIDTH = 40
)

type HistoryView struct {
	channels      *Channels
	refView       WindowView
	commitView    WindowView
	diffView      WindowView
	refViewWin    *Window
	commitViewWin *Window
	diffViewWin   *Window
	views         []WindowView
	activeViewPos uint
	active        bool
}

func NewHistoryView(repoData RepoData, channels *Channels, config Config) *HistoryView {
	refView := NewRefView(repoData, channels)
	commitView := NewCommitView(repoData, channels)
	diffView := NewDiffView(repoData, channels)

	refView.RegisterRefListener(commitView)
	commitView.RegisterCommitListner(diffView)

	return &HistoryView{
		channels:      channels,
		refView:       refView,
		commitView:    commitView,
		diffView:      diffView,
		refViewWin:    NewWindow("refView", config),
		commitViewWin: NewWindow("commitView", config),
		diffViewWin:   NewWindow("diffView", config),
		views:         []WindowView{refView, commitView, diffView},
		activeViewPos: 1,
	}
}

func (historyView *HistoryView) Initialise() (err error) {
	for _, childView := range historyView.views {
		if err = childView.Initialise(); err != nil {
			break
		}
	}

	return
}

func (historyView *HistoryView) Render(viewDimension ViewDimension) (wins []*Window, err error) {
	log.Debug("Rendering HistoryView")

	refViewDim := viewDimension
	refViewDim.cols = Min(HV_BRANCH_VIEW_WIDTH, viewDimension.cols/2)

	commitViewDim := viewDimension
	commitViewDim.cols = viewDimension.cols - refViewDim.cols

	diffViewDim := commitViewDim
	diffViewDim.rows = viewDimension.rows / 2
	commitViewDim.rows = viewDimension.rows - diffViewDim.rows

	log.Debugf("RefView dimensions: %v", refViewDim)
	log.Debugf("CommitView dimensions: %v", commitViewDim)
	log.Debugf("DiffView dimensions: %v", diffViewDim)

	historyView.refViewWin.Resize(refViewDim)
	historyView.commitViewWin.Resize(commitViewDim)
	historyView.diffViewWin.Resize(diffViewDim)

	historyView.refViewWin.Clear()
	historyView.commitViewWin.Clear()
	historyView.diffViewWin.Clear()

	if err = historyView.refView.Render(historyView.refViewWin); err != nil {
		return
	}

	if err = historyView.commitView.Render(historyView.commitViewWin); err != nil {
		return
	}

	if err = historyView.diffView.Render(historyView.diffViewWin); err != nil {
		return
	}

	historyView.refViewWin.SetPosition(0, 0)
	historyView.commitViewWin.SetPosition(0, refViewDim.cols)
	historyView.diffViewWin.SetPosition(commitViewDim.rows, refViewDim.cols)

	wins = []*Window{historyView.refViewWin, historyView.commitViewWin, historyView.diffViewWin}
	return
}

func (historyView *HistoryView) HandleKeyPress(keyPressEvent KeyPressEvent) (err error) {
	log.Debugf("HistoryView handling key %v", keyPressEvent)
	activeChildView := historyView.views[historyView.activeViewPos]
	return activeChildView.HandleKeyPress(keyPressEvent)
}

func (historyView *HistoryView) HandleAction(action Action) (err error) {
	log.Debugf("HistoryView handling action %v", action)

	switch action {
	case ACTION_HISTORY_VIEW_NEXT_VIEW:
		historyView.activeViewPos++
		historyView.activeViewPos %= uint(len(historyView.views))
		historyView.OnActiveChange(true)
		historyView.channels.UpdateDisplay()
		return
	}

	activeChildView := historyView.views[historyView.activeViewPos]
	return activeChildView.HandleAction(action)
}

func (historyView *HistoryView) OnActiveChange(active bool) {
	log.Debugf("History active set to %v", active)

	historyView.active = active

	for viewPos := uint(0); viewPos < uint(len(historyView.views)); viewPos++ {
		if viewPos == historyView.activeViewPos {
			historyView.views[viewPos].OnActiveChange(active)
		} else {
			historyView.views[viewPos].OnActiveChange(false)
		}
	}
}

func (historyView *HistoryView) ViewId() ViewId {
	return VIEW_HISTORY
}

func (historyView *HistoryView) ActiveViewHierarchy() []ViewId {
	if historyView.active {
		return []ViewId{historyView.ViewId(), historyView.views[historyView.activeViewPos].ViewId()}
	}

	return []ViewId{}
}
