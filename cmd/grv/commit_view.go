package main

import (
	"bytes"
	"fmt"
	"strings"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
)

const (
	cvLoadRefreshMs                     = 500
	cvColumnNum                         = 4
	cvDateFormat                        = "2006-01-02 15:04"
	cvCommitGraphLoadRequestChannelSize = 10
	cvCommitSelectedChannelSize         = 100
	cvCommitGraphCommitIndexOffset      = uint(200)
)

type commitViewHandler func(*CommitView, Action) error

type loadingCommitsRefreshTask struct {
	refreshRate time.Duration
	ticker      *time.Ticker
	channels    Channels
	cancelCh    chan<- bool
}

type referenceViewData struct {
	viewPos        ViewPos
	tableFormatter *TableFormatter
	commitGraph    *CommitGraph
}

type commitGraphLoadRequest struct {
	commitIndex uint
	ref         Ref
}

// CommitViewListener is notified when a commit is selected
type CommitViewListener interface {
	OnCommitSelected(*Commit) error
}

// CommitView is the overall instance representing the commit view
type CommitView struct {
	*AbstractWindowView
	channels               Channels
	repoData               RepoData
	repoController         RepoController
	config                 Config
	activeRef              Ref
	active                 bool
	refViewData            map[string]*referenceViewData
	handlers               map[ActionType]commitViewHandler
	refreshTask            *loadingCommitsRefreshTask
	commitViewListeners    []CommitViewListener
	commitViewListenerLock sync.Mutex
	lastViewDimension      ViewDimension
	viewSearch             *ViewSearch
	loadingDotCount        uint
	lastDotRenderTime      time.Time
	commitGraphLoadCh      chan commitGraphLoadRequest
	commitSelectedCh       chan *Commit
	waitGroup              sync.WaitGroup
	lock                   sync.Mutex
}

// NewCommitView creates a new instance of the commit view
func NewCommitView(repoData RepoData, repoController RepoController, channels Channels, config Config) *CommitView {
	commitView := &CommitView{
		channels:          channels,
		repoData:          repoData,
		repoController:    repoController,
		config:            config,
		commitGraphLoadCh: make(chan commitGraphLoadRequest, cvCommitGraphLoadRequestChannelSize),
		commitSelectedCh:  make(chan *Commit, cvCommitSelectedChannelSize),
		refViewData:       make(map[string]*referenceViewData),
		lastDotRenderTime: time.Now(),
		handlers: map[ActionType]commitViewHandler{
			ActionAddFilter:            addCommitFilter,
			ActionRemoveFilter:         removeCommitFilter,
			ActionSelect:               selectCommit,
			ActionCheckoutCommit:       checkoutCommit,
			ActionCreateBranch:         createBranchFromCommit,
			ActionShowAvailableActions: showActionsForCommit,
		},
	}

	commitView.AbstractWindowView = NewAbstractWindowView(commitView, channels, config, "commit")
	commitView.viewSearch = NewViewSearch(commitView, channels)

	return commitView
}

// Initialise currently does nothing
func (commitView *CommitView) Initialise() (err error) {
	log.Info("Initialising CommitView")

	commitView.lock.Lock()
	defer commitView.lock.Unlock()

	commitView.repoData.RegisterCommitSetListener(commitView)

	commitView.waitGroup.Add(2)
	go commitView.processCommitGraphLoadRequests()
	go commitView.processSelectedCommits()

	return
}

// Dispose of any resources held by the view
func (commitView *CommitView) Dispose() {
	commitView.lock.Lock()

	close(commitView.commitGraphLoadCh)
	close(commitView.commitSelectedCh)
	commitView.commitGraphLoadCh = nil
	commitView.commitSelectedCh = nil

	commitView.lock.Unlock()

	commitView.waitGroup.Wait()
}

// Render generates and draws the commit view to the provided window
func (commitView *CommitView) Render(win RenderWindow) (err error) {
	log.Debug("Rendering CommitView")
	commitView.lock.Lock()
	defer commitView.lock.Unlock()

	commitView.lastViewDimension = win.ViewDimensions()

	if commitView.activeRef == nil {
		return commitView.renderEmptyView(win, "No commits to display")
	}

	refViewData, ok := commitView.refViewData[commitView.activeRef.Name()]
	if !ok {
		return fmt.Errorf("No RefViewData exists for ref %v", commitView.activeRef.Name())
	}

	commitSetState := commitView.repoData.CommitSetState(commitView.activeRef)
	commitNum := commitSetState.commitNum

	viewPos := refViewData.viewPos
	rows := win.Rows() - 2
	viewPos.DetermineViewStartRow(rows, commitNum)

	commitDisplayNum := rows
	startCommitIndex := viewPos.ViewStartRowIndex()

	commitCh, err := commitView.repoData.Commits(commitView.activeRef, startCommitIndex, commitDisplayNum)
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
	} else {
		if commitSetState.loading {
			commitView.loadingDotCount %= 4
			err = commitView.renderEmptyView(win, fmt.Sprintf("Fetching commits%v", strings.Repeat(".", int(commitView.loadingDotCount))))

			if time.Since(commitView.lastDotRenderTime).Seconds() >= (cvLoadRefreshMs / 1000.0) {
				commitView.lastDotRenderTime = time.Now()
				commitView.loadingDotCount++
			}
		} else {
			err = commitView.renderEmptyView(win, "No commits to display")
		}

		if err != nil {
			return
		}
	}

	win.DrawBorder()

	if err = win.SetTitle(CmpCommitviewTitle, "Commits for %v", commitView.activeRef.Shorthand()); err != nil {
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

func (commitView *CommitView) renderEmptyView(win RenderWindow, message string) (err error) {
	if err = win.SetRow(2, 1, CmpNone, "   %v", message); err != nil {
		return
	}

	win.DrawBorder()

	return
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
			if err = tableFormatter.AppendToCellWithStyle(rowIndex, colIndex, CmpCommitviewTag, "<%v>", tag.Shorthand()); err != nil {
				return
			}

			if err = tableFormatter.AppendToCell(rowIndex, colIndex, " "); err != nil {
				return
			}
		}
	}

	if len(commitRefs.branches) > 0 {
		for _, branch := range commitRefs.branches {
			if branch.IsRemote() {
				if err = tableFormatter.AppendToCellWithStyle(rowIndex, colIndex, CmpCommitviewLocalBranch, "{%v}", branch.Shorthand()); err != nil {
					return
				}
			} else {
				if err = tableFormatter.AppendToCellWithStyle(rowIndex, colIndex, CmpCommitviewRemoteBranch, "[%v]", branch.Shorthand()); err != nil {
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

// RenderHelpBar shows key bindings custom to the commit view
func (commitView *CommitView) RenderHelpBar(lineBuilder *LineBuilder) (err error) {
	RenderKeyBindingHelp(commitView.ViewID(), lineBuilder, commitView.config, []ActionMessage{
		{action: ActionShowAvailableActions, message: "Show actions for commit"},
		{action: ActionFilterPrompt, message: "Add Filter"},
		{action: ActionRemoveFilter, message: "Remove Filter"},
	})

	return
}

func newLoadingCommitsRefreshTask(refreshRate time.Duration, channels Channels) *loadingCommitsRefreshTask {
	return &loadingCommitsRefreshTask{
		refreshRate: refreshRate,
		channels:    channels,
	}
}

func (refreshTask *loadingCommitsRefreshTask) start() {
	log.Debug("Starting commit load refresh task")

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
	log.Debug("Stopping commit load refresh task")

	if refreshTask.ticker != nil {
		refreshTask.ticker.Stop()
		refreshTask.cancelCh <- true
		close(refreshTask.cancelCh)
		refreshTask.ticker = nil
	}
}

// OnRefSelect handles a new ref being selected and fetches/loads the relevant commits to display
func (commitView *CommitView) OnRefSelect(ref Ref) (err error) {
	log.Debugf("CommitView loading commits for selected ref %v:%v", ref.Shorthand(), ref.Oid())
	commitView.lock.Lock()
	defer commitView.lock.Unlock()

	if commitView.refreshTask != nil {
		commitView.refreshTask.stop()
	}

	refreshTask := newLoadingCommitsRefreshTask(time.Millisecond*cvLoadRefreshMs, commitView.channels)
	commitView.refreshTask = refreshTask

	if err = commitView.repoData.LoadCommits(ref); err != nil {
		return
	}

	refViewData, refViewDataExists := commitView.refViewData[ref.Name()]
	if !refViewDataExists {
		refViewData = &referenceViewData{
			viewPos:        NewViewPosition(),
			tableFormatter: NewTableFormatter(cvColumnNum, commitView.config),
			commitGraph:    NewCommitGraph(commitView.repoData),
		}

		if err = refViewData.tableFormatter.SetCellRendererListener(3, commitView); err != nil {
			return
		}

		commitView.refViewData[ref.Name()] = refViewData
	}

	commitView.activeRef = ref
	commitSetState := commitView.repoData.CommitSetState(ref)

	if commitSetState.loading {
		commitView.refreshTask.start()
		commitView.channels.ReportStatus("Loading commits for ref %v", ref.Shorthand())
	} else {
		commitView.refreshTask.stop()
	}

	var commit *Commit

	if refViewDataExists {
		commitIndex := refViewData.viewPos.ActiveRowIndex()
		commit, err = commitView.repoData.CommitByIndex(commitView.activeRef, commitIndex)
	} else {
		commit, err = commitView.repoData.Commit(commitView.activeRef.Oid())
	}

	if err != nil {
		return
	}

	commitView.notifyCommitViewListeners(commit)

	return
}

// OnCommitsLoaded stops the refresh task if it's still running
func (commitView *CommitView) OnCommitsLoaded(ref Ref) {
	commitView.lock.Lock()
	defer commitView.lock.Unlock()

	if commitView.refreshTask != nil && commitView.activeRef.Name() == ref.Name() {
		log.Debugf("Commits for ref %v loaded. Stopping display refresh task", ref.Name())
		commitView.refreshTask.stop()
	}

	commitSetState := commitView.repoData.CommitSetState(ref)
	commitView.channels.ReportStatus("Loaded %v commits for ref %v", commitSetState.commitNum, ref.Shorthand())
}

// OnCommitsUpdated adjusts the active row index to take account of the newly loaded commits
func (commitView *CommitView) OnCommitsUpdated(ref Ref) {
	commitView.lock.Lock()
	defer commitView.lock.Unlock()

	if commitView.activeRef.Name() == ref.Name() {
		commitSetState := commitView.repoData.CommitSetState(ref)
		if commitSetState.filterState != nil {
			log.Debugf("Filters applied - leaving active row index unchanged")
			return
		}

		viewPos := commitView.viewPos()
		if viewPos.ActiveRowIndex() > commitSetState.commitNum {
			viewPos.SetActiveRowIndex(uint(MaxInt(0, int(commitSetState.commitNum)-1)))
		}

		if err := commitView.selectCommit(viewPos.ActiveRowIndex()); err != nil {
			commitView.channels.ReportError(err)
		}

		if commitView.config.GetBool(CfCommitGraph) {
			refViewData := commitView.refViewData[ref.Name()]
			refViewData.commitGraph.Clear()
		}

		commitView.channels.UpdateDisplay()
	}
}

func (commitView *CommitView) preRenderCell(rowIndex, colIndex uint, lineBuilder *LineBuilder, tableCell *TableCell) (err error) {
	commitSetState := commitView.repoData.CommitSetState(commitView.activeRef)

	if commitSetState.filterState == nil && commitView.config.GetBool(CfCommitGraph) && commitView.commitGraphLoadCh != nil {
		refViewData := commitView.refViewData[commitView.activeRef.Name()]
		commitIndex := refViewData.viewPos.ViewStartRowIndex() + rowIndex

		if commitIndex >= refViewData.commitGraph.Rows() {
			request := commitGraphLoadRequest{
				commitIndex: commitIndex,
				ref:         commitView.activeRef,
			}

			select {
			case commitView.commitGraphLoadCh <- request:
			default:
			}
		} else {
			refViewData.commitGraph.Render(lineBuilder, commitIndex)
		}
	}

	return
}

func (commitView *CommitView) postRenderCell(rowIndex, colIndex uint, lineBuilder *LineBuilder, tableCell *TableCell) (err error) {
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

// RegisterCommitViewListener accepts a listener to be notified when a commit is selected
func (commitView *CommitView) RegisterCommitViewListener(commitViewListener CommitViewListener) {
	if commitViewListener == nil {
		return
	}

	log.Debugf("Registering CommitViewListener %T", commitViewListener)

	commitView.commitViewListenerLock.Lock()
	defer commitView.commitViewListenerLock.Unlock()

	commitView.commitViewListeners = append(commitView.commitViewListeners, commitViewListener)
}

func (commitView *CommitView) commitViewListenerCount() uint {
	commitView.commitViewListenerLock.Lock()
	defer commitView.commitViewListenerLock.Unlock()

	return uint(len(commitView.commitViewListeners))
}

func (commitView *CommitView) notifyCommitViewListeners(commit *Commit) {
	if commitView.commitSelectedCh != nil {
		commitView.commitSelectedCh <- commit
	}
}

func (commitView *CommitView) processSelectedCommits() {
	defer commitView.waitGroup.Done()

	for commit := range commitView.commitSelectedCh {
		log.Debugf("Notifying commit listeners of selected commit %v", commit.oid)
		commitViewListeners := commitView.commitViewListenersCopy()

		for _, commitViewListener := range commitViewListeners {
			if err := commitViewListener.OnCommitSelected(commit); err != nil {
				commitView.channels.ReportError(err)
			}
		}
	}
}

func (commitView *CommitView) commitViewListenersCopy() []CommitViewListener {
	commitView.commitViewListenerLock.Lock()
	defer commitView.commitViewListenerLock.Unlock()

	return append([]CommitViewListener(nil), commitView.commitViewListeners...)
}

func (commitView *CommitView) selectCommit(lineIndex uint) (err error) {
	commitIndex := lineIndex
	commitSetState := commitView.repoData.CommitSetState(commitView.activeRef)

	if commitSetState.commitNum == 0 {
		return fmt.Errorf("Cannot select commit as there are no commits for ref %v", commitView.activeRef.Name())
	}

	if commitIndex >= commitSetState.commitNum {
		return fmt.Errorf("Invalid commitIndex: %v, only %v commits are loaded", commitIndex, commitSetState.commitNum)
	}

	selectedCommit, err := commitView.repoData.CommitByIndex(commitView.activeRef, commitIndex)
	if err != nil {
		return
	}

	commitView.viewPos().SetActiveRowIndex(lineIndex)
	commitView.notifyCommitViewListeners(selectedCommit)

	return
}

func (commitView *CommitView) createCommitViewListenerView(commit *Commit) {
	createViewArgs := CreateViewArgs{
		viewID:   ViewDiff,
		viewArgs: []interface{}{commit.oid.String()},
		registerViewListener: func(observer interface{}) (err error) {
			if observer == nil {
				return fmt.Errorf("Invalid CommitViewListener: %v", observer)
			}

			if commitViewListener, ok := observer.(CommitViewListener); ok {
				commitView.RegisterCommitViewListener(commitViewListener)
			} else {
				err = fmt.Errorf("Observer is not a CommitViewListener but has type %T", observer)
			}

			return
		},
	}

	commitView.channels.DoAction(Action{
		ActionType: ActionSplitView,
		Args: []interface{}{
			ActionSplitViewArgs{
				CreateViewArgs: createViewArgs,
				orientation:    CoDynamic,
			},
		},
	})
}

// ViewPos returns the current view position
func (commitView *CommitView) ViewPos() ViewPos {
	commitView.lock.Lock()
	defer commitView.lock.Unlock()

	return commitView.viewPos()
}

func (commitView *CommitView) viewPos() ViewPos {
	refViewData := commitView.refViewData[commitView.activeRef.Name()]
	return refViewData.viewPos
}

// OnSearchMatch updates the view position when there is a search match
func (commitView *CommitView) OnSearchMatch(startPos ViewPos, matchLineIndex uint) {
	commitView.lock.Lock()
	defer commitView.lock.Unlock()

	viewPos := commitView.viewPos()

	if viewPos != startPos {
		log.Debugf("Selected ref has changed since search started")
		return
	}

	commitView.selectCommit(matchLineIndex)
}

// Line returns the rendered line at the index provided
func (commitView *CommitView) Line(lineIndex uint) (line string) {
	commitView.lock.Lock()
	defer commitView.lock.Unlock()

	rows := commitView.rows()

	if lineIndex >= rows {
		log.Errorf("Invalid lineIndex: %v", lineIndex)
		return
	}

	refViewData, ok := commitView.refViewData[commitView.activeRef.Name()]
	if !ok {
		log.Errorf("Not refViewData for ref %v", commitView.activeRef.Name())
		return
	}

	tableFormatter := refViewData.tableFormatter
	tableFormatter.Clear()

	commitIndex := lineIndex
	commit, err := commitView.repoData.CommitByIndex(commitView.activeRef, commitIndex)
	if err != nil {
		log.Errorf("Error when retrieving commit during search: %v", err)
		return
	}

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
func (commitView *CommitView) LineNumber() (rows uint) {
	commitView.lock.Lock()
	defer commitView.lock.Unlock()

	return commitView.rows()
}

func (commitView *CommitView) rows() (rows uint) {
	commitSetState := commitView.repoData.CommitSetState(commitView.activeRef)
	return commitSetState.commitNum
}

func (commitView *CommitView) viewDimension() ViewDimension {
	return commitView.lastViewDimension
}

func (commitView *CommitView) onRowSelected(rowIndex uint) error {
	return commitView.selectCommit(rowIndex)
}

// HandleEvent reacts to an event
func (commitView *CommitView) HandleEvent(event Event) (err error) {
	commitView.lock.Lock()
	defer commitView.lock.Unlock()

	switch event.EventType {
	case ViewRemovedEvent:
		commitView.removeCommitViewListeners(event.Args)
	}

	return
}

func (commitView *CommitView) removeCommitViewListeners(views []interface{}) {
	for _, view := range views {
		if commitViewListener, ok := view.(CommitViewListener); ok {
			commitView.removeCommitViewListener(commitViewListener)
		}
	}
}

func (commitView *CommitView) removeCommitViewListener(commitViewListener CommitViewListener) {
	commitView.commitViewListenerLock.Lock()
	defer commitView.commitViewListenerLock.Unlock()

	for index, listener := range commitView.commitViewListeners {
		if commitViewListener == listener {
			log.Debugf("Removing CommitViewListener %T", commitViewListener)
			commitView.commitViewListeners = append(commitView.commitViewListeners[:index], commitView.commitViewListeners[index+1:]...)
			break
		}
	}
}

func (commitView *CommitView) processCommitGraphLoadRequests() {
	defer commitView.waitGroup.Done()
	log.Info("Started processing commit graph load requests")

	for request := range commitView.commitGraphLoadCh {
		request = commitView.retrieveLatestCommitGraphLoadRequest(request)
		log.Debugf("Processing commit graph load request: %v:%v", request.ref.Name(), request.commitIndex)
		activeRef, commitGraph := commitView.retriveDataForCommitGraphLoadRequest(request)

		if commitGraph == nil {
			continue
		}

		if err := commitView.processCommitGraphLoadRequest(request, commitGraph, activeRef); err != nil {
			log.Errorf("Failed to process CommitGraph load request: %v", err)
		}
	}

	log.Info("Finished processing commit graph load requests")
}

func (commitView *CommitView) retrieveLatestCommitGraphLoadRequest(request commitGraphLoadRequest) commitGraphLoadRequest {
	requestFound := true

	for requestFound {
		select {
		case request = <-commitView.commitGraphLoadCh:
		default:
			requestFound = false
		}
	}

	return request
}

func (commitView *CommitView) retriveDataForCommitGraphLoadRequest(request commitGraphLoadRequest) (activeRef Ref, commitGraph *CommitGraph) {
	commitView.lock.Lock()
	defer commitView.lock.Unlock()

	activeRef = commitView.activeRef

	if activeRef != nil && request.ref == activeRef {
		refViewData := commitView.refViewData[commitView.activeRef.Name()]
		commitGraph = refViewData.commitGraph
	}

	return
}

func (commitView *CommitView) processCommitGraphLoadRequest(request commitGraphLoadRequest, commitGraph *CommitGraph, ref Ref) (err error) {
	commitGraphRows := commitGraph.Rows()
	commitIndex := request.commitIndex + cvCommitGraphCommitIndexOffset
	commitSetState := commitView.repoData.CommitSetState(ref)

	if commitIndex+1 > commitSetState.commitNum {
		commitIndex = commitSetState.commitNum - 1
	}

	if commitIndex < commitGraphRows {
		return
	}

	commitNum := (commitIndex + 1) - commitGraphRows
	commitCh, err := commitView.repoData.Commits(ref, commitGraphRows, commitNum)
	if err != nil {
		return
	}

	for commit := range commitCh {
		if err = commitGraph.AddCommit(commit); err != nil {
			return
		} else if commitView.channels.Exit() {
			return
		}
	}

	log.Debugf("Added %v commits to commit graph starting at index %v for ref %v", commitNum, commitGraphRows, ref.Name())
	commitView.channels.UpdateDisplay()

	return
}

// HandleAction checks if commit view supports this action and if it does executes it
func (commitView *CommitView) HandleAction(action Action) (err error) {
	log.Debugf("CommitView handling action %v", action)
	commitView.lock.Lock()
	defer commitView.lock.Unlock()

	if commitView.activeRef == nil {
		return
	}

	var handled bool
	if handler, ok := commitView.handlers[action.ActionType]; ok {
		log.Debugf("Action handled by CommitView")
		err = handler(commitView, action)
	} else if handled, err = commitView.viewSearch.HandleAction(action); handled {
		log.Debugf("Action handled by ViewSearch")
	} else if handled, err = commitView.AbstractWindowView.HandleAction(action); handled {
		log.Debugf("Action handled by AbstractWindowView")
	} else {
		log.Debugf("Action not handled")
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
	} else if commitFilter == nil {
		log.Debugf("Query string does not define commit filter: \"%v\"", query)
		return
	}

	if err = commitView.repoData.AddCommitFilter(commitView.activeRef, commitFilter); err != nil {
		return
	}

	commitView.viewPos().SetActiveRowIndex(0)

	go func() {
		// TODO: Works in practice, but there is no guarantee the filtered commit set will have
		// been populated after 250ms. Need an event based mechanism to be notified when a filtered
		// set has started to be populated so that the first commit can be selected
		time.Sleep(250 * time.Millisecond)
		commitView.lock.Lock()
		defer commitView.lock.Unlock()

		if err := commitView.selectCommit(commitView.viewPos().ActiveRowIndex()); err != nil {
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

	if err = commitView.selectCommit(0); err != nil {
		return
	}

	commitView.channels.UpdateDisplay()

	return
}

func selectCommit(commitView *CommitView, action Action) (err error) {
	viewPos := commitView.viewPos()

	if commitView.commitViewListenerCount() == 0 {
		var commit *Commit
		if commit, err = commitView.repoData.CommitByIndex(commitView.activeRef, viewPos.ActiveRowIndex()); err != nil {
			return
		}

		commitView.createCommitViewListenerView(commit)
	}

	return commitView.selectCommit(viewPos.ActiveRowIndex())
}

func checkoutCommit(commitView *CommitView, action Action) (err error) {
	viewPos := commitView.viewPos()

	commit, err := commitView.repoData.CommitByIndex(commitView.activeRef, viewPos.ActiveRowIndex())
	if err != nil {
		return
	}

	if commitView.config.GetBool(CfConfirmCheckout) {
		question := fmt.Sprintf("Are you sure you want to checkout commit %v?", commit.oid.ShortID())

		commitView.channels.DoAction(YesNoQuestion(question, func(response QuestionResponse) {
			if response == ResponseYes {
				commitView.checkoutCommit(commit)
			}
		}))
	} else {
		commitView.checkoutCommit(commit)
	}

	return
}

func (commitView *CommitView) checkoutCommit(commit *Commit) {
	commitView.repoController.CheckoutCommit(commit, func(err error) {
		if err != nil {
			commitView.channels.ReportError(err)
			return
		}

		commitView.channels.ReportStatus("Checked out commit %v", commit.oid.ShortID())
	})
}

func createBranchFromCommit(commitView *CommitView, action Action) (err error) {
	if len(action.Args) == 0 {
		return fmt.Errorf("Expected branch name argument")
	}

	branchName, isString := action.Args[0].(string)
	if !isString {
		return fmt.Errorf("Expected first argument to be branch name but found %T", action.Args[0])
	}

	viewPos := commitView.viewPos()

	commit, err := commitView.repoData.CommitByIndex(commitView.activeRef, viewPos.ActiveRowIndex())
	if err != nil {
		return
	}

	if err = commitView.repoController.CreateBranch(branchName, commit.oid); err != nil {
		return
	}

	commitView.channels.ReportStatus("Created branch %v at %v", branchName, commit.oid.ShortID())

	return
}

func showActionsForCommit(commitView *CommitView, action Action) (err error) {
	if commitView.rows() == 0 {
		return
	}

	viewPos := commitView.viewPos()
	commit, err := commitView.repoData.CommitByIndex(commitView.activeRef, viewPos.ActiveRowIndex())
	if err != nil {
		return
	}

	commitAuthor := commit.commit.Author().Name

	commitView.channels.DoAction(Action{
		ActionType: ActionCreateContextMenu,
		Args: []interface{}{
			ActionCreateContextMenuArgs{
				viewDimension: ViewDimension{
					rows: 10,
					cols: 60,
				},
				config: ContextMenuConfig{
					Entries: []ContextMenuEntry{
						{
							DisplayName: "Checkout commit",
							Value:       Action{ActionType: ActionCheckoutCommit},
						},
						{
							DisplayName: "Create branch from commit",
							Value:       Action{ActionType: ActionBranchNamePrompt},
						},
						{
							DisplayName: fmt.Sprintf(`Filter commits by author "%v"`, commitAuthor),
							Value: Action{
								ActionType: ActionAddFilter,
								Args: []interface{}{
									fmt.Sprintf(`authorname = "%v"`, strings.Replace(commitAuthor, `"`, `\"`, -1)),
								},
							},
						},
					},
					OnSelect: func(entry ContextMenuEntry, entryIndex uint) {
						if selectedAction, ok := entry.Value.(Action); ok {
							commitView.channels.DoAction(selectedAction)
						} else {
							log.Errorf("Expected Action instance but found: %v", entry.Value)
						}
					},
				},
			},
		},
	})

	return
}
