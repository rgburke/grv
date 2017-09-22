package main

import (
	"sync"

	log "github.com/Sirupsen/logrus"
)

const (
	hvBranchViewWidth = 35
)

type viewOrientation int

const (
	voDefault viewOrientation = iota
	voColumn
	voCount
)

// HistoryView manages the history view and it's child views
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
	orientation          viewOrientation
	lock                 sync.Mutex
}

type viewLayout struct {
	viewDimension ViewDimension
	startRow      uint
	startCol      uint
}

// NewHistoryView creates a new instance of the history view
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
		channels:    channels,
		refView:     refView,
		commitView:  commitView,
		diffView:    diffView,
		views:       []WindowView{refView, commitView, diffView},
		orientation: voDefault,
		viewWins: map[WindowView]*Window{
			refView:    refViewWin,
			commitView: commitViewWin,
			diffView:   diffViewWin,
		},
		activeViewPos: 1,
	}
}

// Initialise sets up the history view and calls initialise on its child views
func (historyView *HistoryView) Initialise() (err error) {
	for _, childView := range historyView.views {
		if err = childView.Initialise(); err != nil {
			break
		}
	}

	return
}

// Render generates the history view and returns windows (one for each child view) representing the view as a whole
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

func (historyView *HistoryView) determineViewDimensions(viewDimension ViewDimension) map[WindowView]viewLayout {
	refViewLayout := viewLayout{viewDimension: viewDimension}
	commitViewLayout := viewLayout{viewDimension: viewDimension}
	diffViewLayout := viewLayout{viewDimension: viewDimension}

	refViewLayout.viewDimension.cols = Min(hvBranchViewWidth, viewDimension.cols/2)

	if historyView.orientation == voColumn {
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

	return map[WindowView]viewLayout{
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

// RenderHelpBar renders key binding help info for the history view
func (historyView *HistoryView) RenderHelpBar(lineBuilder *LineBuilder) (err error) {
	RenderKeyBindingHelp(historyView.ViewID(), lineBuilder, []ActionMessage{
		{action: ActionNextView, message: "Next View"},
		{action: ActionPrevView, message: "Previous View"},
		{action: ActionFullScreenView, message: "Toggle Full Screen"},
		{action: ActionToggleViewLayout, message: "Toggle Layout"},
	})

	return
}

// HandleKeyPress passes the keypress onto the active child view
func (historyView *HistoryView) HandleKeyPress(keystring string) (err error) {
	log.Debugf("HistoryView handling keys %v", keystring)
	activeChildView := historyView.ActiveView()
	return activeChildView.HandleKeyPress(keystring)
}

// HandleAction handles the provided action if the history view supports it or passes it down to the active child view
func (historyView *HistoryView) HandleAction(action Action) (err error) {
	log.Debugf("HistoryView handling action %v", action)

	switch action.ActionType {
	case ActionNextView:
		historyView.lock.Lock()
		historyView.activeViewPos++
		historyView.activeViewPos %= uint(len(historyView.views))
		historyView.lock.Unlock()
		historyView.OnActiveChange(true)
		historyView.channels.UpdateDisplay()
		return
	case ActionPrevView:
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
	case ActionFullScreenView:
		historyView.lock.Lock()
		defer historyView.lock.Unlock()

		historyView.fullScreenActiveView = !historyView.fullScreenActiveView
		historyView.channels.UpdateDisplay()
		return
	case ActionToggleViewLayout:
		historyView.lock.Lock()
		defer historyView.lock.Unlock()

		historyView.orientation = (historyView.orientation + 1) % voCount
		historyView.channels.UpdateDisplay()
		return
	}

	activeChildView := historyView.ActiveView()
	return activeChildView.HandleAction(action)
}

// OnActiveChange updates whether this view (and it's active child view) are active
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

// ViewID returns the view ID for the history view
func (historyView *HistoryView) ViewID() ViewID {
	return ViewHistory
}

// ActiveView returns the active child view
func (historyView *HistoryView) ActiveView() AbstractView {
	historyView.lock.Lock()
	defer historyView.lock.Unlock()

	return historyView.views[historyView.activeViewPos]
}
