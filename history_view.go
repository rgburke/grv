package main

import (
	log "github.com/Sirupsen/logrus"
	"sync"
)

const (
	HV_BRANCH_VIEW_WIDTH = 40
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

	return map[WindowView]ViewLayout{
		historyView.refView: ViewLayout{
			viewDimension: refViewDim,
			startRow:      0,
			startCol:      0,
		},
		historyView.commitView: ViewLayout{
			viewDimension: commitViewDim,
			startRow:      0,
			startCol:      refViewDim.cols,
		},
		historyView.diffView: ViewLayout{
			viewDimension: diffViewDim,
			startRow:      commitViewDim.rows,
			startCol:      refViewDim.cols,
		},
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

func (historyView *HistoryView) RenderStatusBar(RenderWindow) (err error) {
	return
}

func (historyView *HistoryView) RenderHelpBar(lineBuilder *LineBuilder) (err error) {
	historyView.lock.Lock()
	defer historyView.lock.Unlock()

	RenderKeyBindingHelp(historyView.ViewId(), lineBuilder, []ActionMessage{
		ActionMessage{action: ACTION_NEXT_VIEW, message: "Next View"},
		ActionMessage{action: ACTION_PREV_VIEW, message: "Previous View"},
		ActionMessage{action: ACTION_FULL_SCREEN_VIEW, message: "Toggle Full Screen"},
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

	switch action {
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
