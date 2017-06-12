package main

import (
	log "github.com/Sirupsen/logrus"
	"sync"
)

const (
	HV_BRANCH_VIEW_WIDTH = 35
)

type ViewOrientation int

const (
	VO_DEFAULT ViewOrientation = iota
	VO_COLUMN
	VO_COUNT
)

type HistoryView struct {
	channels             *Channels
	refView              WindowView
	commitView           WindowView
	diffView             WindowView
	views                []WindowView
	viewWins             map[WindowView]*Window
	activeViewPos        uint
	active               bool
	fullScreenActiveView bool
	viewOrientation      ViewOrientation
	lock                 sync.Mutex
}

type ViewLayout struct {
	viewDimension ViewDimension
	startRow      uint
	startCol      uint
}

func NewHistoryView(repoData RepoData, channels *Channels, config Config) *HistoryView {
	refView := NewRefView(repoData, channels)
	commitView := NewCommitView(repoData, channels)
	diffView := NewDiffView(repoData, channels)

	refViewWin := NewWindow("refView", config)
	commitViewWin := NewWindow("commitView", config)
	diffViewWin := NewWindow("diffView", config)

	refView.RegisterRefListener(commitView)
	commitView.RegisterCommitListner(diffView)

	return &HistoryView{
		channels:   channels,
		refView:    refView,
		commitView: commitView,
		diffView:   diffView,
		views:      []WindowView{refView, commitView, diffView},
		viewWins: map[WindowView]*Window{
			refView:    refViewWin,
			commitView: commitViewWin,
			diffView:   diffViewWin,
		},
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
	historyView.lock.Lock()
	defer historyView.lock.Unlock()

	if historyView.fullScreenActiveView {
		return historyView.renderActiveViewFullScreen(viewDimension)
	}

	viewLayouts := historyView.determineViewDimensions(viewDimension)

	for view, viewLayout := range viewLayouts {
		win := historyView.viewWins[view]
		win.Resize(viewLayout.viewDimension)
		win.Clear()
		win.SetPosition(viewLayout.startRow, viewLayout.startCol)

		if err = view.Render(win); err != nil {
			return
		}

		wins = append(wins, win)
	}

	return
}

func (historyView *HistoryView) determineViewDimensions(viewDimension ViewDimension) map[WindowView]ViewLayout {
	refViewLayout := ViewLayout{viewDimension: viewDimension}
	commitViewLayout := ViewLayout{viewDimension: viewDimension}
	diffViewLayout := ViewLayout{viewDimension: viewDimension}

	refViewLayout.viewDimension.cols = Min(HV_BRANCH_VIEW_WIDTH, viewDimension.cols/2)

	if historyView.viewOrientation == VO_COLUMN {
		remainingCols := viewDimension.cols - refViewLayout.viewDimension.cols

		commitViewLayout.viewDimension.cols = remainingCols / 2
		commitViewLayout.startCol = refViewLayout.viewDimension.cols

		diffViewLayout.viewDimension.cols = remainingCols - commitViewLayout.viewDimension.cols
		diffViewLayout.startCol = viewDimension.cols - diffViewLayout.viewDimension.cols
	} else {
		commitViewLayout.viewDimension.cols = viewDimension.cols - refViewLayout.viewDimension.cols
		commitViewLayout.viewDimension.rows = viewDimension.rows / 2
		commitViewLayout.startCol = refViewLayout.viewDimension.cols

		diffViewLayout.viewDimension.cols = viewDimension.cols - refViewLayout.viewDimension.cols
		diffViewLayout.viewDimension.rows = viewDimension.rows - commitViewLayout.viewDimension.rows
		diffViewLayout.startRow = commitViewLayout.viewDimension.rows
		diffViewLayout.startCol = refViewLayout.viewDimension.cols
	}

	log.Debugf("RefView layout: %v", refViewLayout)
	log.Debugf("CommitView layout: %v", commitViewLayout)
	log.Debugf("DiffView layout: %v", diffViewLayout)

	return map[WindowView]ViewLayout{
		historyView.refView:    refViewLayout,
		historyView.commitView: commitViewLayout,
		historyView.diffView:   diffViewLayout,
	}
}

func (historyView *HistoryView) renderActiveViewFullScreen(viewDimension ViewDimension) (wins []*Window, err error) {
	view := historyView.views[historyView.activeViewPos]
	win := historyView.viewWins[view]

	win.Resize(viewDimension)
	win.Clear()
	win.SetPosition(0, 0)

	if err = view.Render(win); err != nil {
		return
	}

	wins = append(wins, win)

	return
}

func (historyView *HistoryView) RenderStatusBar(lineBuilder *LineBuilder) (err error) {
	return
}

func (historyView *HistoryView) RenderHelpBar(lineBuilder *LineBuilder) (err error) {
	RenderKeyBindingHelp(historyView.ViewId(), lineBuilder, []ActionMessage{
		ActionMessage{action: ACTION_NEXT_VIEW, message: "Next View"},
		ActionMessage{action: ACTION_PREV_VIEW, message: "Previous View"},
		ActionMessage{action: ACTION_FULL_SCREEN_VIEW, message: "Toggle Full Screen"},
		ActionMessage{action: ACTION_TOGGLE_VIEW_LAYOUT, message: "Toggle Layout"},
	})

	return
}

func (historyView *HistoryView) HandleKeyPress(keystring string) (err error) {
	log.Debugf("HistoryView handling keys %v", keystring)
	activeChildView := historyView.ActiveView()
	return activeChildView.HandleKeyPress(keystring)
}

func (historyView *HistoryView) HandleAction(action Action) (err error) {
	log.Debugf("HistoryView handling action %v", action)

	switch action.ActionType {
	case ACTION_NEXT_VIEW:
		historyView.lock.Lock()
		historyView.activeViewPos++
		historyView.activeViewPos %= uint(len(historyView.views))
		historyView.lock.Unlock()
		historyView.OnActiveChange(true)
		historyView.channels.UpdateDisplay()
		return
	case ACTION_PREV_VIEW:
		historyView.lock.Lock()

		if historyView.activeViewPos == 0 {
			historyView.activeViewPos = uint(len(historyView.views)) - 1
		} else {
			historyView.activeViewPos--
		}

		historyView.lock.Unlock()
		historyView.OnActiveChange(true)
		historyView.channels.UpdateDisplay()
		return
	case ACTION_FULL_SCREEN_VIEW:
		historyView.lock.Lock()
		defer historyView.lock.Unlock()

		historyView.fullScreenActiveView = !historyView.fullScreenActiveView
		historyView.channels.UpdateDisplay()
	case ACTION_TOGGLE_VIEW_LAYOUT:
		historyView.lock.Lock()
		defer historyView.lock.Unlock()

		historyView.viewOrientation = (historyView.viewOrientation + 1) % VO_COUNT
		historyView.channels.UpdateDisplay()
		return
	}

	activeChildView := historyView.ActiveView()
	return activeChildView.HandleAction(action)
}

func (historyView *HistoryView) OnActiveChange(active bool) {
	log.Debugf("History active set to %v", active)
	historyView.lock.Lock()
	defer historyView.lock.Unlock()

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

func (historyView *HistoryView) ActiveView() AbstractView {
	historyView.lock.Lock()
	defer historyView.lock.Unlock()

	return historyView.views[historyView.activeViewPos]
}
