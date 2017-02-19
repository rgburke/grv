package main

import (
	log "github.com/Sirupsen/logrus"
	gc "github.com/rthornton128/goncurses"
)

const (
	HV_BRANCH_VIEW_WIDTH = 40
)

type HistoryView struct {
	refView         WindowView
	commitView      WindowView
	refViewWin      *Window
	commitViewWin   *Window
	views           []WindowView
	activeViewIndex uint
	active          bool
}

func NewHistoryView(repoData RepoData) *HistoryView {
	refView := NewRefView(repoData)
	commitView := NewCommitView(repoData)
	refView.RegisterRefListener(commitView)

	return &HistoryView{
		refView:         refView,
		commitView:      commitView,
		refViewWin:      NewWindow("refView"),
		commitViewWin:   NewWindow("commitView"),
		views:           []WindowView{refView, commitView},
		activeViewIndex: 1,
	}
}

func (historyView *HistoryView) Initialise(channels HandlerChannels) (err error) {
	for _, childView := range historyView.views {
		if err = childView.Initialise(channels); err != nil {
			break
		}
	}

	return
}

func (historyView *HistoryView) Render(viewDimension ViewDimension) (wins []*Window, err error) {
	log.Debug("Rendering HistoryView")

	refViewDim := viewDimension
	refViewDim.cols = Min(HV_BRANCH_VIEW_WIDTH, refViewDim.cols/2)
	log.Debugf("RefView dimensions: %v", refViewDim)

	commitViewDim := viewDimension
	commitViewDim.cols = viewDimension.cols - refViewDim.cols
	log.Debugf("CommitView dimensions: %v", commitViewDim)

	historyView.refViewWin.Resize(refViewDim)
	historyView.commitViewWin.Resize(commitViewDim)

	historyView.refViewWin.Clear()
	historyView.commitViewWin.Clear()

	if err = historyView.refView.Render(historyView.refViewWin); err != nil {
		return
	}

	if err = historyView.commitView.Render(historyView.commitViewWin); err != nil {
		return
	}

	historyView.refViewWin.SetPosition(0, 0)
	historyView.commitViewWin.SetPosition(0, refViewDim.cols)

	wins = []*Window{historyView.refViewWin, historyView.commitViewWin}
	return
}

func (historyView *HistoryView) Handle(keyPressEvent KeyPressEvent, channels HandlerChannels) (err error) {
	log.Debugf("HistoryView handling key %v", keyPressEvent)

	switch keyPressEvent.key {
	case gc.KEY_TAB:
		historyView.activeViewIndex++
		historyView.activeViewIndex %= uint(len(historyView.views))
		historyView.OnActiveChange(true)
		channels.displayCh <- true
		return
	}

	view := historyView.views[historyView.activeViewIndex]

	err = view.Handle(keyPressEvent, channels)
	return
}

func (historyView *HistoryView) OnActiveChange(active bool) {
	log.Debugf("History active set to %v", active)

	historyView.active = active

	for viewIndex := uint(0); viewIndex < uint(len(historyView.views)); viewIndex++ {
		if viewIndex == historyView.activeViewIndex {
			historyView.views[viewIndex].OnActiveChange(active)
		} else {
			historyView.views[viewIndex].OnActiveChange(false)
		}
	}
}
