package main

import (
	"bytes"
	"fmt"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
)

const (
	cvLoadRefreshMs = 500
	cvColumnNum     = 4
	cvDateFormat    = "2006-01-02 15:04"
)

type commitViewHandler func(*CommitView, Action) error

type loadingCommitsRefreshTask struct {
	refreshRate time.Duration
	ticker      *time.Ticker
	channels    *Channels
	cancelCh    chan<- bool
}

type referenceViewData struct {
	viewPos        ViewPos
	tableFormatter *TableFormatter
}

// CommitListener is notified when a commit is selected
type CommitListener interface {
	OnCommitSelect(*Commit) error
}

// CommitView is the overall instance representing the commit view
type CommitView struct {
	channels        *Channels
	repoData        RepoData
	activeRef       *Oid
	activeRefName   string
	active          bool
	refViewData     map[*Oid]*referenceViewData
	handlers        map[ActionType]commitViewHandler
	refreshTask     *loadingCommitsRefreshTask
	commitListeners []CommitListener
	viewDimension   ViewDimension
	viewSearch      *ViewSearch
	lock            sync.Mutex
}

// NewCommitView creates a new instance of the commit view
func NewCommitView(repoData RepoData, channels *Channels) *CommitView {
	commitView := &CommitView{
		channels:    channels,
		repoData:    repoData,
		refViewData: make(map[*Oid]*referenceViewData),
		handlers: map[ActionType]commitViewHandler{
			ActionPrevLine:     moveUpCommit,
			ActionNextLine:     moveDownCommit,
			ActionPrevPage:     moveUpCommitPage,
			ActionNextPage:     moveDownCommitPage,
			ActionScrollRight:  scrollCommitViewRight,
			ActionScrollLeft:   scrollCommitViewLeft,
			ActionFirstLine:    moveToFirstCommit,
			ActionLastLine:     moveToLastCommit,
			ActionAddFilter:    addCommitFilter,
			ActionRemoveFilter: removeCommitFilter,
		},
	}

	commitView.viewSearch = NewViewSearch(commitView, channels)

	return commitView
}

// Initialise currently does nothing
func (commitView *CommitView) Initialise() (err error) {
	log.Info("Initialising CommitView")
	return
}

// Render generates and draws the commit view to the provided window
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

	commitCh, err := commitView.repoData.Commits(commitView.activeRef, viewPos.ViewStartRowIndex(), rows)
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

	if err = tableFormatter.Render(win, viewPos.ViewStartColumn(), true); err != nil {
		return
	}

	if commitSetState.commitNum > 0 {
		if err = win.SetSelectedRow(viewPos.SelectedRowIndex()+1, commitView.active); err != nil {
			return
		}
	}

	if err = win.SetTitle(CmpCommitviewTitle, "Commits for %v", commitView.activeRefName); err != nil {
		return
	}

	var selectedCommit uint
	if commitSetState.commitNum == 0 {
		selectedCommit = 0
	} else {
		selectedCommit = viewPos.ActiveRowIndex() + 1
	}

	var footerText bytes.Buffer

	footerText.WriteString(fmt.Sprintf("Commit %v of %v", selectedCommit, commitSetState.commitNum))

	if commitSetState.filterState != nil {
		filtersApplied := commitSetState.filterState.filtersApplied
		filtersTextSuffix := ""

		if filtersApplied > 1 {
			filtersTextSuffix = "s"
		}

		footerText.WriteString(fmt.Sprintf(" (%v filter%v applied)", commitSetState.filterState.filtersApplied, filtersTextSuffix))
	}

	if err = win.SetFooter(CmpCommitviewFooter, "%v", footerText.String()); err != nil {
		return
	}

	if searchActive, searchPattern, lastSearchFoundMatch := commitView.viewSearch.SearchActive(); searchActive && lastSearchFoundMatch {
		if err = win.Highlight(searchPattern, CmpAllviewSearchMatch); err != nil {
			return
		}
	}

	return err
}

func (commitView *CommitView) renderCommit(tableFormatter *TableFormatter, rowIndex uint, commit *Commit) (err error) {
	author := commit.commit.Author()
	commitRefs := commitView.repoData.RefsForCommit(commit)
	colIndex := uint(0)

	if err = tableFormatter.SetCellWithStyle(rowIndex, colIndex, CmpCommitviewShortOid, "%v", commit.oid.ShortID()); err != nil {
		return
	}

	colIndex++
	if err = tableFormatter.SetCellWithStyle(rowIndex, colIndex, CmpCommitviewDate, "%v", author.When.Format(cvDateFormat)); err != nil {
		return
	}

	colIndex++
	if err = tableFormatter.SetCellWithStyle(rowIndex, colIndex, CmpCommitviewAuthor, "%v", author.Name); err != nil {
		return
	}

	colIndex++
	if len(commitRefs.tags) > 0 {
		for _, tag := range commitRefs.tags {
			if err = tableFormatter.AppendToCellWithStyle(rowIndex, colIndex, CmpCommitviewTag, "<%v>", tag.name); err != nil {
				return
			}

			if err = tableFormatter.AppendToCell(rowIndex, colIndex, " "); err != nil {
				return
			}
		}
	}

	if len(commitRefs.branches) > 0 {
		for _, branch := range commitRefs.branches {
			if branch.isRemote {
				if err = tableFormatter.AppendToCellWithStyle(rowIndex, colIndex, CmpCommitviewLocalBranch, "{%v}", branch.name); err != nil {
					return
				}
			} else {
				if err = tableFormatter.AppendToCellWithStyle(rowIndex, colIndex, CmpCommitviewRemoteBranch, "[%v]", branch.name); err != nil {
					return
				}
			}

			if err = tableFormatter.AppendToCell(rowIndex, colIndex, " "); err != nil {
				return
			}
		}
	}

	if err = tableFormatter.AppendToCellWithStyle(rowIndex, colIndex, CmpCommitviewSummary, "%v", commit.commit.Summary()); err != nil {
		return
	}

	return
}

// RenderStatusBar does nothing
func (commitView *CommitView) RenderStatusBar(lineBuilder *LineBuilder) (err error) {
	return
}

// RenderHelpBar shows key bindings custom to the commit view
func (commitView *CommitView) RenderHelpBar(lineBuilder *LineBuilder) (err error) {
	RenderKeyBindingHelp(commitView.ViewID(), lineBuilder, []ActionMessage{
		{action: ActionFilterPrompt, message: "Add Filter"},
		{action: ActionRemoveFilter, message: "Remove Filter"},
	})

	return
}

func newLoadingCommitsRefreshTask(refreshRate time.Duration, channels *Channels) *loadingCommitsRefreshTask {
	return &loadingCommitsRefreshTask{
		refreshRate: refreshRate,
		channels:    channels,
	}
}

func (refreshTask *loadingCommitsRefreshTask) start() {
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

func (refreshTask *loadingCommitsRefreshTask) stop() {
	if refreshTask.ticker != nil {
		refreshTask.ticker.Stop()
		refreshTask.cancelCh <- true
		close(refreshTask.cancelCh)
		refreshTask.ticker = nil
	}
}

// OnRefSelect handles a new ref being selected and fetches/loads the relevant commits to display
func (commitView *CommitView) OnRefSelect(refName string, oid *Oid) (err error) {
	log.Debugf("CommitView loading commits for selected oid %v", oid)
	commitView.lock.Lock()
	defer commitView.lock.Unlock()

	if commitView.refreshTask != nil {
		commitView.refreshTask.stop()
	}

	refreshTask := newLoadingCommitsRefreshTask(time.Millisecond*cvLoadRefreshMs, commitView.channels)
	commitView.refreshTask = refreshTask

	if err = commitView.repoData.LoadCommits(oid, func(oid *Oid) error {
		commitView.lock.Lock()
		defer commitView.lock.Unlock()

		refreshTask.stop()

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
		refViewData = &referenceViewData{
			viewPos:        NewViewPosition(),
			tableFormatter: NewTableFormatter(cvColumnNum),
		}

		commitView.refViewData[oid] = refViewData
	}

	commitSetState := commitView.repoData.CommitSetState(oid)

	if commitSetState.loading {
		commitView.refreshTask.start()
		commitView.channels.ReportStatus("Loading commits for ref %v", refName)
	} else {
		commitView.refreshTask.stop()
	}

	var commit *Commit

	if refViewDataExists {
		commit, err = commitView.repoData.CommitByIndex(commitView.activeRef, refViewData.viewPos.ActiveRowIndex())
	} else {
		commit, err = commitView.repoData.Commit(commitView.activeRef)
	}

	if err != nil {
		return
	}

	commitView.notifyCommitListeners(commit)

	return
}

// OnActiveChange updates whether this view is currently active
func (commitView *CommitView) OnActiveChange(active bool) {
	log.Debugf("CommitView active: %v", active)
	commitView.lock.Lock()
	defer commitView.lock.Unlock()

	commitView.active = active
}

// ViewID returns the ViewID for the commit view
func (commitView *CommitView) ViewID() ViewID {
	return ViewCommit
}

// RegisterCommitListner accepts a listener to be notified when a commit is selected
func (commitView *CommitView) RegisterCommitListner(commitListener CommitListener) {
	commitView.commitListeners = append(commitView.commitListeners, commitListener)
}

func (commitView *CommitView) notifyCommitListeners(commit *Commit) {
	log.Debugf("Notifying commit listeners of selected commit %v", commit.commit.Id().String())

	for _, commitListener := range commitView.commitListeners {
		if err := commitListener.OnCommitSelect(commit); err != nil {
			commitView.channels.ReportError(err)
		}
	}
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

// ViewPos returns the current view position
func (commitView *CommitView) ViewPos() ViewPos {
	refViewData := commitView.refViewData[commitView.activeRef]
	return refViewData.viewPos
}

// OnSearchMatch updates the view position when there is a search match
func (commitView *CommitView) OnSearchMatch(startPos ViewPos, matchLineIndex uint) {
	commitView.lock.Lock()
	defer commitView.lock.Unlock()

	viewPos := commitView.ViewPos()

	if viewPos != startPos {
		log.Debugf("Selected ref has changed since search started")
		return
	}

	viewPos.SetActiveRowIndex(matchLineIndex)
}

// Line returns the rendered line at the index provided
func (commitView *CommitView) Line(lineIndex uint) (line string) {
	commitView.lock.Lock()
	defer commitView.lock.Unlock()

	commitSetState := commitView.repoData.CommitSetState(commitView.activeRef)

	if lineIndex >= commitSetState.commitNum {
		log.Errorf("Invalid lineIndex: %v", lineIndex)
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

	return
}

// LineNumber returns the total number of rendered lines the commit view has
func (commitView *CommitView) LineNumber() (lineNumber uint) {
	commitView.lock.Lock()
	defer commitView.lock.Unlock()

	commitSetState := commitView.repoData.CommitSetState(commitView.activeRef)
	return commitSetState.commitNum
}

// HandleKeyPress does nothing
func (commitView *CommitView) HandleKeyPress(keystring string) (err error) {
	log.Debugf("CommitView handling key %v - NOP", keystring)
	return
}

// HandleAction checks if commit view supports this action and if it does executes it
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

func moveUpCommit(commitView *CommitView, action Action) (err error) {
	viewPos := commitView.ViewPos()

	if viewPos.MoveLineUp() {
		log.Debug("Moving up one commit")
		if err = commitView.selectCommit(viewPos.ActiveRowIndex()); err != nil {
			return
		}
		commitView.channels.UpdateDisplay()
	}

	return
}

func moveDownCommit(commitView *CommitView, action Action) (err error) {
	commitSetState := commitView.repoData.CommitSetState(commitView.activeRef)
	viewPos := commitView.ViewPos()

	if viewPos.MoveLineDown(commitSetState.commitNum) {
		log.Debug("Moving down one commit")
		if err = commitView.selectCommit(viewPos.ActiveRowIndex()); err != nil {
			return
		}
		commitView.channels.UpdateDisplay()
	}

	return
}

func moveUpCommitPage(commitView *CommitView, action Action) (err error) {
	viewPos := commitView.ViewPos()

	if viewPos.MovePageUp(commitView.viewDimension.rows - 2) {
		log.Debug("Moving up one page")
		if err = commitView.selectCommit(viewPos.ActiveRowIndex()); err != nil {
			return
		}
		commitView.channels.UpdateDisplay()
	}

	return
}

func moveDownCommitPage(commitView *CommitView, action Action) (err error) {
	commitSetState := commitView.repoData.CommitSetState(commitView.activeRef)
	viewPos := commitView.ViewPos()

	if viewPos.MovePageDown(commitView.viewDimension.rows-2, commitSetState.commitNum) {
		log.Debug("Moving down one page")
		if err = commitView.selectCommit(viewPos.ActiveRowIndex()); err != nil {
			return
		}
		commitView.channels.UpdateDisplay()
	}

	return
}

func scrollCommitViewRight(commitView *CommitView, action Action) (err error) {
	viewPos := commitView.ViewPos()
	viewPos.MovePageRight(commitView.viewDimension.cols)
	log.Debugf("Scrolling right. View starts at column %v", viewPos.ViewStartColumn())
	commitView.channels.UpdateDisplay()

	return
}

func scrollCommitViewLeft(commitView *CommitView, action Action) (err error) {
	viewPos := commitView.ViewPos()

	if viewPos.MovePageLeft(commitView.viewDimension.cols) {
		log.Debugf("Scrolling left. View starts at column %v", viewPos.ViewStartColumn())
		commitView.channels.UpdateDisplay()
	}

	return
}

func moveToFirstCommit(commitView *CommitView, action Action) (err error) {
	viewPos := commitView.ViewPos()

	if viewPos.MoveToFirstLine() {
		log.Debug("Moving up to first commit")
		if err = commitView.selectCommit(viewPos.ActiveRowIndex()); err != nil {
			return
		}
		commitView.channels.UpdateDisplay()
	}

	return
}

func moveToLastCommit(commitView *CommitView, action Action) (err error) {
	commitSetState := commitView.repoData.CommitSetState(commitView.activeRef)
	viewPos := commitView.ViewPos()

	if viewPos.MoveToLastLine(commitSetState.commitNum) {
		log.Debug("Moving to last commit")
		if err = commitView.selectCommit(viewPos.ActiveRowIndex()); err != nil {
			return
		}
		commitView.channels.UpdateDisplay()
	}

	return
}

func addCommitFilter(commitView *CommitView, action Action) (err error) {
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

	if err = commitView.repoData.AddCommitFilter(commitView.activeRef, commitFilter); err != nil {
		return
	}

	commitView.ViewPos().SetActiveRowIndex(0)

	go func() {
		// TODO: Works in practice, but there is no guarantee the filtered commit set will have
		// been populated after 250ms. Need an event based mechanism to be notified when a filtered
		// set has started to be populated so that the first commit can be selected
		time.Sleep(250 * time.Millisecond)
		commitView.lock.Lock()
		defer commitView.lock.Unlock()

		if err := commitView.selectCommit(commitView.ViewPos().ActiveRowIndex()); err != nil {
			log.Errorf("Unable to select commit after filter has been applied: %v", err)
		}
	}()

	commitView.channels.UpdateDisplay()

	return
}

func removeCommitFilter(commitView *CommitView, action Action) (err error) {
	if err = commitView.repoData.RemoveCommitFilter(commitView.activeRef); err != nil {
		return
	}

	commitView.ViewPos().SetActiveRowIndex(0)

	commit, err := commitView.repoData.CommitByIndex(commitView.activeRef, commitView.ViewPos().ActiveRowIndex())
	if err != nil {
		return
	}

	commitView.notifyCommitListeners(commit)

	commitView.channels.UpdateDisplay()

	return
}
