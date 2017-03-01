package main

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	gc "github.com/rthornton128/goncurses"
	"sync"
	"time"
)

const (
	CV_LOAD_REFRESH_MS = 500
)

type CommitViewHandler func(*CommitView) error

type ViewIndex struct {
	activeIndex    uint
	viewStartIndex uint
}

type LoadingCommitsRefreshTask struct {
	refreshRate time.Duration
	ticker      *time.Ticker
	channels    *Channels
	cancelCh    chan<- bool
}

type CommitView struct {
	channels     *Channels
	repoData     RepoData
	activeBranch *Oid
	active       bool
	viewIndex    map[*Oid]*ViewIndex
	handlers     map[gc.Key]CommitViewHandler
	refreshTask  *LoadingCommitsRefreshTask
	lock         sync.Mutex
}

func NewCommitView(repoData RepoData, channels *Channels) *CommitView {
	return &CommitView{
		channels:  channels,
		repoData:  repoData,
		viewIndex: make(map[*Oid]*ViewIndex),
		handlers: map[gc.Key]CommitViewHandler{
			gc.KEY_UP:   MoveUpCommit,
			gc.KEY_DOWN: MoveDownCommit,
		},
	}
}

func (commitView *CommitView) Initialise() (err error) {
	log.Info("Initialising CommitView")
	return
}

func (commitView *CommitView) Render(win RenderWindow) (err error) {
	log.Debug("Rendering CommitView")
	commitView.lock.Lock()
	defer commitView.lock.Unlock()

	var viewIndex *ViewIndex
	var ok bool
	if viewIndex, ok = commitView.viewIndex[commitView.activeBranch]; !ok {
		return fmt.Errorf("No ViewIndex exists for oid %v", commitView.activeBranch)
	}

	rows := win.Rows() - 2

	if viewIndex.viewStartIndex > viewIndex.activeIndex {
		viewIndex.viewStartIndex = viewIndex.activeIndex
	} else if rowDiff := viewIndex.activeIndex - viewIndex.viewStartIndex; rowDiff >= rows {
		viewIndex.viewStartIndex += (rowDiff - rows) + 1
	}

	commitCh, err := commitView.repoData.Commits(commitView.activeBranch, viewIndex.viewStartIndex, rows)
	if err != nil {
		return err
	}

	rowIndex := uint(1)

	for commit := range commitCh {
		author := commit.commit.Author()

		if err = win.SetRow(rowIndex, " %v %s %s", author.When.Format("2006-01-02 15:04"), author.Name, commit.commit.Summary()); err != nil {
			break
		}

		rowIndex++
	}

	if err = win.SetSelectedRow((viewIndex.activeIndex-viewIndex.viewStartIndex)+1, commitView.active); err != nil {
		return
	}

	win.DrawBorder()

	return err
}

func NewLoadingCommitsRefreshTask(refreshRate time.Duration, channels *Channels) *LoadingCommitsRefreshTask {
	return &LoadingCommitsRefreshTask{
		refreshRate: refreshRate,
		channels:    channels,
	}
}

func (refreshTask *LoadingCommitsRefreshTask) Start() {
	refreshTask.ticker = time.NewTicker(refreshTask.refreshRate)
	cancelCh := make(chan bool)
	refreshTask.cancelCh = cancelCh

	go func(cancelCh <-chan bool) {
		for {
			select {
			case <-refreshTask.ticker.C:
				log.Debug("Updating display with newly loaded commits")
				refreshTask.channels.UpdateDisplay()
			case <-cancelCh:
				refreshTask.channels.UpdateDisplay()
				return
			}
		}
	}(cancelCh)
}

func (refreshTask *LoadingCommitsRefreshTask) Stop() {
	if refreshTask.ticker != nil {
		refreshTask.ticker.Stop()
		refreshTask.cancelCh <- true
		close(refreshTask.cancelCh)
		refreshTask.ticker = nil
	}
}

func (commitView *CommitView) OnRefSelect(oid *Oid) (err error) {
	log.Debugf("CommitView loading commits for selected oid %v", oid)
	commitView.lock.Lock()
	defer commitView.lock.Unlock()

	if commitView.refreshTask != nil {
		commitView.refreshTask.Stop()
	}

	refreshTask := NewLoadingCommitsRefreshTask(time.Millisecond*CV_LOAD_REFRESH_MS, commitView.channels)
	commitView.refreshTask = refreshTask

	if err = commitView.repoData.LoadCommits(oid, func(oid *Oid) error {
		commitView.lock.Lock()
		defer commitView.lock.Unlock()

		refreshTask.Stop()

		return nil
	}); err != nil {
		return
	}

	commitView.activeBranch = oid

	if _, ok := commitView.viewIndex[oid]; !ok {
		commitView.viewIndex[oid] = &ViewIndex{}
	}

	commitSetState := commitView.repoData.CommitSetState(oid)

	if commitSetState.loading {
		commitView.refreshTask.Start()
	} else {
		commitView.refreshTask.Stop()
	}

	return
}

func (commitView *CommitView) OnActiveChange(active bool) {
	log.Debugf("CommitView active %v", active)
	commitView.lock.Lock()
	defer commitView.lock.Unlock()

	commitView.active = active
}

func (commitView *CommitView) Handle(keyPressEvent KeyPressEvent) (err error) {
	log.Debugf("CommitView handling key %v", keyPressEvent)
	commitView.lock.Lock()
	defer commitView.lock.Unlock()

	if handler, ok := commitView.handlers[keyPressEvent.key]; ok {
		err = handler(commitView)
	}

	return
}

func MoveUpCommit(commitView *CommitView) (err error) {
	viewIndex := commitView.viewIndex[commitView.activeBranch]

	if viewIndex.activeIndex > 0 {
		log.Debug("Moving up one commit")
		viewIndex.activeIndex--
		commitView.channels.UpdateDisplay()
	}

	return
}

func MoveDownCommit(commitView *CommitView) (err error) {
	commitSetState := commitView.repoData.CommitSetState(commitView.activeBranch)
	viewIndex := commitView.viewIndex[commitView.activeBranch]

	if viewIndex.activeIndex < commitSetState.commitNum-1 {
		log.Debug("Moving down one commit")
		viewIndex.activeIndex++
		commitView.channels.UpdateDisplay()
	}

	return
}
