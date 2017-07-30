package main

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"sync"
	"time"
)

const (
	CV_LOAD_REFRESH_MS = 500
	CV_COLUMN_NUM      = 3
	CV_DATE_FORMAT     = "2006-01-02 15:04"
)

type CommitViewHandler func(*CommitView, Action) error

type LoadingCommitsRefreshTask struct {
	refreshRate time.Duration
	ticker      *time.Ticker
	channels    *Channels
	cancelCh    chan<- bool
}

type CommitListener interface {
	OnCommitSelect(*Commit) error
}

type RefViewData struct {
	viewPos        *ViewPos
	tableFormatter *TableFormatter
}

type CommitView struct {
	channels        *Channels
	repoData        RepoData
	activeRef       *Oid
	activeRefName   string
	active          bool
	refViewData     map[*Oid]*RefViewData
	handlers        map[ActionType]CommitViewHandler
	refreshTask     *LoadingCommitsRefreshTask
	commitListeners []CommitListener
	viewDimension   ViewDimension
	viewSearch      *ViewSearch
	lock            sync.Mutex
}

func NewCommitView(repoData RepoData, channels *Channels) *CommitView {
	commitView := &CommitView{
		channels:    channels,
		repoData:    repoData,
		refViewData: make(map[*Oid]*RefViewData),
		handlers: map[ActionType]CommitViewHandler{
			ACTION_PREV_LINE:     MoveUpCommit,
			ACTION_NEXT_LINE:     MoveDownCommit,
			ACTION_PREV_PAGE:     MoveUpCommitPage,
			ACTION_NEXT_PAGE:     MoveDownCommitPage,
			ACTION_SCROLL_RIGHT:  ScrollCommitViewRight,
			ACTION_SCROLL_LEFT:   ScrollCommitViewLeft,
			ACTION_FIRST_LINE:    MoveToFirstCommit,
			ACTION_LAST_LINE:     MoveToLastCommit,
			ACTION_ADD_FILTER:    AddCommitFilter,
			ACTION_REMOVE_FILTER: RemoveCommitFilter,
		},
	}

	commitView.viewSearch = NewViewSearch(commitView, channels)

	return commitView
}

func (commitView *CommitView) Initialise() (err error) {
	log.Info("Initialising CommitView")
	return
}

func (commitView *CommitView) Render(win RenderWindow) (err error) {
	log.Debug("Rendering CommitView")
	commitView.lock.Lock()
	defer commitView.lock.Unlock()

	commitView.viewDimension = win.ViewDimensions()

	refViewData, ok := commitView.refViewData[commitView.activeRef]
	if !ok {
		return fmt.Errorf("No RefViewData exists for oid %v", commitView.activeRef)
	}

	commitSetState := commitView.repoData.CommitSetState(commitView.activeRef)

	rows := win.Rows() - 2
	viewPos := refViewData.viewPos
	viewPos.DetermineViewStartRow(rows, commitSetState.commitNum)

	commitCh, err := commitView.repoData.Commits(commitView.activeRef, viewPos.viewStartRowIndex, rows)
	if err != nil {
		return err
	}

	tableFormatter := refViewData.tableFormatter
	tableFormatter.Resize(rows)
	tableFormatter.Clear()

	rowIndex := uint(0)

	for commit := range commitCh {
		if err = commitView.renderCommit(tableFormatter, rowIndex, commit); err != nil {
			return
		}

		rowIndex++
	}

	if err = tableFormatter.Render(win, viewPos.viewStartColumn, true); err != nil {
		return
	}

	if commitSetState.commitNum > 0 {
		if err = win.SetSelectedRow((viewPos.activeRowIndex-viewPos.viewStartRowIndex)+1, commitView.active); err != nil {
			return
		}
	}

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

	if searchActive, searchPattern, lastSearchFoundMatch := commitView.viewSearch.SearchActive(); searchActive && lastSearchFoundMatch {
		if err = win.Highlight(searchPattern, CMP_ALLVIEW_SEARCH_MATCH); err != nil {
			return
		}
	}

	return err
}

func (commitView *CommitView) renderCommit(tableFormatter *TableFormatter, rowIndex uint, commit *Commit) (err error) {
	author := commit.commit.Author()
	commitRefs := commitView.repoData.RefsForCommit(commit)

	if err = tableFormatter.SetCellWithStyle(rowIndex, 0, CMP_COMMITVIEW_DATE, "%v", author.When.Format(CV_DATE_FORMAT)); err != nil {
		return
	}

	if err = tableFormatter.SetCellWithStyle(rowIndex, 1, CMP_COMMITVIEW_AUTHOR, "%v", author.Name); err != nil {
		return
	}

	if len(commitRefs.tags) > 0 {
		for _, tag := range commitRefs.tags {
			if err = tableFormatter.AppendToCellWithStyle(rowIndex, 2, CMP_COMMITVIEW_TAG, "<%v>", tag.name); err != nil {
				return
			}

			if err = tableFormatter.AppendToCell(rowIndex, 2, " "); err != nil {
				return
			}
		}
	}

	if len(commitRefs.branches) > 0 {
		for _, branch := range commitRefs.branches {
			if branch.isRemote {
				if err = tableFormatter.AppendToCellWithStyle(rowIndex, 2, CMP_COMMITVIEW_LOCAL_BRANCH, "{%v}", branch.name); err != nil {
					return
				}
			} else {
				if err = tableFormatter.AppendToCellWithStyle(rowIndex, 2, CMP_COMMITVIEW_REMOTE_BRANCH, "[%v]", branch.name); err != nil {
					return
				}
			}

			if err = tableFormatter.AppendToCell(rowIndex, 2, " "); err != nil {
				return
			}
		}
	}

	if err = tableFormatter.AppendToCellWithStyle(rowIndex, 2, CMP_COMMITVIEW_SUMMARY, "%v", commit.commit.Summary()); err != nil {
		return
	}

	return
}

func (commitView *CommitView) RenderStatusBar(lineBuilder *LineBuilder) (err error) {
	return
}

func (commitView *CommitView) RenderHelpBar(lineBuilder *LineBuilder) (err error) {
	RenderKeyBindingHelp(commitView.ViewId(), lineBuilder, []ActionMessage{
		ActionMessage{action: ACTION_FILTER_PROMPT, message: "Add Filter"},
		ActionMessage{action: ACTION_REMOVE_FILTER, message: "Remove Filter"},
	})

	return
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

		commitSetState := commitView.repoData.CommitSetState(oid)
		commitView.channels.ReportStatus("Loaded %v commits for ref %v", commitSetState.commitNum, refName)

		return nil
	}); err != nil {
		return
	}

	commitView.activeRef = oid
	commitView.activeRefName = refName

	refViewData, refViewDataExists := commitView.refViewData[oid]
	if !refViewDataExists {
		refViewData = &RefViewData{
			viewPos:        NewViewPos(),
			tableFormatter: NewTableFormatter(CV_COLUMN_NUM),
		}

		commitView.refViewData[oid] = refViewData
	}

	commitSetState := commitView.repoData.CommitSetState(oid)

	if commitSetState.loading {
		commitView.refreshTask.Start()
		commitView.channels.ReportStatus("Loading commits for ref %v", refName)
	} else {
		commitView.refreshTask.Stop()
	}

	var commit *Commit

	if refViewDataExists {
		commit, err = commitView.repoData.CommitByIndex(commitView.activeRef, refViewData.viewPos.activeRowIndex)
	} else {
		commit, err = commitView.repoData.Commit(commitView.activeRef)
	}

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

func (commitView *CommitView) ViewId() ViewId {
	return VIEW_COMMIT
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

func (commitView *CommitView) ViewPos() *ViewPos {
	refViewData := commitView.refViewData[commitView.activeRef]
	return refViewData.viewPos
}

func (commitView *CommitView) OnSearchMatch(startPos *ViewPos, matchLineIndex uint) {
	commitView.lock.Lock()
	defer commitView.lock.Unlock()

	viewPos := commitView.ViewPos()

	if viewPos != startPos {
		log.Debugf("Selected ref has changed since search started")
		return
	}

	viewPos.activeRowIndex = matchLineIndex
}

func (commitView *CommitView) Line(lineIndex uint) (line string, lineExists bool) {
	commitView.lock.Lock()
	defer commitView.lock.Unlock()

	commitSetState := commitView.repoData.CommitSetState(commitView.activeRef)

	if lineIndex >= commitSetState.commitNum {
		return
	}

	commit, err := commitView.repoData.CommitByIndex(commitView.activeRef, lineIndex)

	if err != nil {
		log.Errorf("Error when retrieving commit during search: %v", err)
		return
	}

	refViewData, ok := commitView.refViewData[commitView.activeRef]
	if !ok {
		log.Errorf("Not refViewData for ref %v", commitView.activeRef)
		return
	}

	tableFormatter := refViewData.tableFormatter
	tableFormatter.Clear()

	if err = commitView.renderCommit(tableFormatter, 0, commit); err != nil {
		log.Errorf("Error when rendering commit: %v", err)
		return
	}

	line, err = tableFormatter.RowString(0)
	if err != nil {
		log.Errorf("Error when retrieving row string: %v", err)
		return
	}

	lineExists = true

	return
}

func (commitView *CommitView) LineNumber() (lineNumber uint) {
	commitView.lock.Lock()
	defer commitView.lock.Unlock()

	commitSetState := commitView.repoData.CommitSetState(commitView.activeRef)
	return commitSetState.commitNum
}

func (commitView *CommitView) HandleKeyPress(keystring string) (err error) {
	log.Debugf("CommitView handling key %v - NOP", keystring)
	return
}

func (commitView *CommitView) HandleAction(action Action) (err error) {
	log.Debugf("CommitView handling action %v", action)
	commitView.lock.Lock()
	defer commitView.lock.Unlock()

	if handler, ok := commitView.handlers[action.ActionType]; ok {
		err = handler(commitView, action)
	} else {
		_, err = commitView.viewSearch.HandleAction(action)
	}

	return
}

func MoveUpCommit(commitView *CommitView, action Action) (err error) {
	viewPos := commitView.ViewPos()

	if viewPos.MoveLineUp() {
		log.Debug("Moving up one commit")
		commitView.selectCommit(viewPos.activeRowIndex)
		commitView.channels.UpdateDisplay()
	}

	return
}

func MoveDownCommit(commitView *CommitView, action Action) (err error) {
	commitSetState := commitView.repoData.CommitSetState(commitView.activeRef)
	viewPos := commitView.ViewPos()

	if viewPos.MoveLineDown(commitSetState.commitNum) {
		log.Debug("Moving down one commit")
		commitView.selectCommit(viewPos.activeRowIndex)
		commitView.channels.UpdateDisplay()
	}

	return
}

func MoveUpCommitPage(commitView *CommitView, action Action) (err error) {
	viewPos := commitView.ViewPos()

	if viewPos.MovePageUp(commitView.viewDimension.rows - 2) {
		log.Debug("Moving up one page")
		commitView.selectCommit(viewPos.activeRowIndex)
		commitView.channels.UpdateDisplay()
	}

	return
}

func MoveDownCommitPage(commitView *CommitView, action Action) (err error) {
	commitSetState := commitView.repoData.CommitSetState(commitView.activeRef)
	viewPos := commitView.ViewPos()

	if viewPos.MovePageDown(commitView.viewDimension.rows-2, commitSetState.commitNum) {
		log.Debug("Moving down one page")
		commitView.selectCommit(viewPos.activeRowIndex)
		commitView.channels.UpdateDisplay()
	}

	return
}

func ScrollCommitViewRight(commitView *CommitView, action Action) (err error) {
	viewPos := commitView.ViewPos()
	viewPos.MovePageRight(commitView.viewDimension.cols)
	log.Debugf("Scrolling right. View starts at column %v", viewPos.viewStartColumn)
	commitView.channels.UpdateDisplay()

	return
}

func ScrollCommitViewLeft(commitView *CommitView, action Action) (err error) {
	viewPos := commitView.ViewPos()

	if viewPos.MovePageLeft(commitView.viewDimension.cols) {
		log.Debugf("Scrolling left. View starts at column %v", viewPos.viewStartColumn)
		commitView.channels.UpdateDisplay()
	}

	return
}

func MoveToFirstCommit(commitView *CommitView, action Action) (err error) {
	viewPos := commitView.ViewPos()

	if viewPos.MoveToFirstLine() {
		log.Debug("Moving up to first commit")
		commitView.selectCommit(viewPos.activeRowIndex)
		commitView.channels.UpdateDisplay()
	}

	return
}

func MoveToLastCommit(commitView *CommitView, action Action) (err error) {
	commitSetState := commitView.repoData.CommitSetState(commitView.activeRef)
	viewPos := commitView.ViewPos()

	if viewPos.MoveToLastLine(commitSetState.commitNum) {
		log.Debug("Moving to last commit")
		commitView.selectCommit(viewPos.activeRowIndex)
		commitView.channels.UpdateDisplay()
	}

	return
}

func AddCommitFilter(commitView *CommitView, action Action) (err error) {
	if !(len(action.Args) > 0) {
		return fmt.Errorf("Expected filter query argument")
	}

	query, ok := action.Args[0].(string)
	if !ok {
		return fmt.Errorf("Expected filter query argument to have type string")
	}

	commitFilter, errors := CreateCommitFilter(query)
	if len(errors) > 0 {
		commitView.channels.ReportErrors(errors)
		return
	}

	err = commitView.repoData.AddCommitFilter(commitView.activeRef, commitFilter)
	commitView.channels.UpdateDisplay()

	return
}

func RemoveCommitFilter(commitView *CommitView, action Action) (err error) {
	err = commitView.repoData.RemoveCommitFilter(commitView.activeRef)
	commitView.channels.UpdateDisplay()

	return
}
