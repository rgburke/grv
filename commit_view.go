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

type LoadingCommitsRefreshTask struct {
	refreshRate time.Duration
	ticker      *time.Ticker
	channels    *Channels
	cancelCh    chan<- bool
}

type CommitListener interface {
	OnCommitSelect(*Commit) error
}

type CommitView struct {
	channels        *Channels
	repoData        RepoData
	activeRef       *Oid
	activeRefName   string
	active          bool
	viewPos         map[*Oid]*ViewPos
	handlers        map[gc.Key]CommitViewHandler
	refreshTask     *LoadingCommitsRefreshTask
	displayCommits  []*Commit
	commitListeners []CommitListener
	lock            sync.Mutex
}

func NewCommitView(repoData RepoData, channels *Channels) *CommitView {
	return &CommitView{
		channels: channels,
		repoData: repoData,
		viewPos:  make(map[*Oid]*ViewPos),
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

	var viewPos *ViewPos
	var ok bool
	if viewPos, ok = commitView.viewPos[commitView.activeRef]; !ok {
		return fmt.Errorf("No ViewPos exists for oid %v", commitView.activeRef)
	}

	rows := win.Rows() - 2

	if viewPos.viewStartRowIndex > viewPos.activeRowIndex {
		viewPos.viewStartRowIndex = viewPos.activeRowIndex
	} else if rowDiff := viewPos.activeRowIndex - viewPos.viewStartRowIndex; rowDiff >= rows {
		viewPos.viewStartRowIndex += (rowDiff - rows) + 1
	}

	commitCh, err := commitView.repoData.Commits(commitView.activeRef, viewPos.viewStartRowIndex, rows)
	if err != nil {
		return err
	}

	var lineBuilder *LineBuilder
	rowIndex := uint(1)

	for commit := range commitCh {
		if lineBuilder, err = win.LineBuilder(rowIndex); err != nil {
			return
		}

		author := commit.commit.Author()

		lineBuilder.
			Append(" ").
			AppendWithStyle(CMP_COMMITVIEW_DATE, "%v", author.When.Format("2006-01-02 15:04")).
			Append(" ").
			AppendWithStyle(CMP_COMMITVIEW_AUTHOR, "%v", author.Name).
			Append(" ").
			AppendWithStyle(CMP_COMMITVIEW_SUMMARY, "%v", commit.commit.Summary())

		rowIndex++
	}

	commitSetState := commitView.repoData.CommitSetState(commitView.activeRef)

	if commitSetState.commitNum > 0 {
		if err = win.SetSelectedRow((viewPos.activeRowIndex-viewPos.viewStartRowIndex)+1, commitView.active); err != nil {
			return
		}
	}

	win.DrawBorder()

	if err = win.SetTitle(CMP_COMMITVIEW_TITLE, "Commits for %v", commitView.activeRefName); err != nil {
		return
	}

	var selectedCommit uint
	if commitSetState.commitNum == 0 {
		selectedCommit = 0
	} else {
		selectedCommit = viewPos.activeRowIndex + 1
	}

	if err = win.SetFooter(CMP_COMMITVIEW_FOOTER, "Commit %v of %v", selectedCommit, commitSetState.commitNum); err != nil {
		return
	}

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

func (commitView *CommitView) OnRefSelect(refName string, oid *Oid) (err error) {
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

	commitView.activeRef = oid
	commitView.activeRefName = refName

	if _, ok := commitView.viewPos[oid]; !ok {
		commitView.viewPos[oid] = &ViewPos{}
	}

	commitSetState := commitView.repoData.CommitSetState(oid)

	if commitSetState.loading {
		commitView.refreshTask.Start()
	} else {
		commitView.refreshTask.Stop()
	}

	commit, err := commitView.repoData.Commit(oid)
	if err != nil {
		return
	}

	commitView.notifyCommitListeners(commit)

	return
}

func (commitView *CommitView) OnActiveChange(active bool) {
	log.Debugf("CommitView active: %v", active)
	commitView.lock.Lock()
	defer commitView.lock.Unlock()

	commitView.active = active
}

func (commitView *CommitView) RegisterCommitListner(commitListener CommitListener) {
	commitView.commitListeners = append(commitView.commitListeners, commitListener)
}

func (commitView *CommitView) notifyCommitListeners(commit *Commit) {
	log.Debugf("Notifying commit listners of selected commit %v", commit.commit.Id().String())

	for _, commitListener := range commitView.commitListeners {
		if err := commitListener.OnCommitSelect(commit); err != nil {
			commitView.channels.ReportError(err)
		}
	}

	return
}

func (commitView *CommitView) selectCommit(commitIndex uint) (err error) {
	commitSetState := commitView.repoData.CommitSetState(commitView.activeRef)

	if commitSetState.commitNum == 0 {
		return fmt.Errorf("Cannot select commit as there are no commits for ref %v", commitView.activeRef)
	}

	if commitIndex >= commitSetState.commitNum {
		return fmt.Errorf("Invalid commitIndex: %v, only %v commits are loaded", commitIndex, commitSetState.commitNum)
	}

	selectedCommit, err := commitView.repoData.CommitByIndex(commitView.activeRef, commitIndex)
	if err != nil {
		return
	}

	commitView.notifyCommitListeners(selectedCommit)

	return
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
	viewPos := commitView.viewPos[commitView.activeRef]

	if viewPos.activeRowIndex > 0 {
		log.Debug("Moving up one commit")
		viewPos.activeRowIndex--
		commitView.selectCommit(viewPos.activeRowIndex)
		commitView.channels.UpdateDisplay()
	}

	return
}

func MoveDownCommit(commitView *CommitView) (err error) {
	commitSetState := commitView.repoData.CommitSetState(commitView.activeRef)
	viewPos := commitView.viewPos[commitView.activeRef]

	if commitSetState.commitNum > 0 && viewPos.activeRowIndex < commitSetState.commitNum-1 {
		log.Debug("Moving down one commit")
		viewPos.activeRowIndex++
		commitView.selectCommit(viewPos.activeRowIndex)
		commitView.channels.UpdateDisplay()
	}

	return
}
