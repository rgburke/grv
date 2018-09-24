package main

import (
	"fmt"
	"sync"

	log "github.com/Sirupsen/logrus"
)

const (
	svIndentationSpace = "     "
)

type summaryViewHandler func(*SummaryView, Action) error

type summaryViewLine interface {
	render(*LineBuilder)
	renderString() string
	isSelectable() bool
}

type singleValueLine struct {
	value            string
	themeComponentID ThemeComponentID
	selectable       bool
}

func (singleValueLine *singleValueLine) render(lineBuilder *LineBuilder) {
	lineBuilder.AppendWithStyle(singleValueLine.themeComponentID, "%v", singleValueLine.value)
}

func (singleValueLine *singleValueLine) renderString() string {
	return singleValueLine.value
}

func (singleValueLine *singleValueLine) isSelectable() bool {
	return singleValueLine.selectable
}

var emptyLine = &singleValueLine{}

func newHeaderRenderer(header string) summaryViewLine {
	return &singleValueLine{
		value:            header,
		themeComponentID: CmpSummaryViewHeader,
	}
}

type branchLine struct {
	head Ref
}

func (branchLine *branchLine) getBranchName() string {
	if _, isDetached := branchLine.head.(*HEAD); isDetached {
		return GetDetachedHeadDisplayValue(branchLine.head.Oid())
	}

	return branchLine.head.Shorthand()
}

func (branchLine *branchLine) render(lineBuilder *LineBuilder) {
	lineBuilder.AppendWithStyle(CmpNone, "%v", branchLine.getBranchName())

	if branch, isLocalBranch := branchLine.head.(*LocalBranch); isLocalBranch && branch.IsTrackingBranch() {
		lineBuilder.
			AppendWithStyle(CmpSummaryViewNormal, " (").
			AppendACSChar(AcsUarrow, CmpSummaryViewNormal).
			AppendWithStyle(CmpSummaryViewBranchAhead, "%v ", branch.ahead).
			AppendACSChar(AcsDarrow, CmpSummaryViewNormal).
			AppendWithStyle(CmpSummaryViewBranchBehind, "%v", branch.behind).
			AppendWithStyle(CmpSummaryViewNormal, ")")
	}
}

func (branchLine *branchLine) renderString() string {
	if branch, isLocalBranch := branchLine.head.(*LocalBranch); isLocalBranch && branch.IsTrackingBranch() {
		return fmt.Sprintf("%v (^%v v%v)", branchLine.getBranchName(), branch.ahead, branch.behind)
	}

	return branchLine.getBranchName()
}

func (branchLine *branchLine) isSelectable() bool {
	return true
}

// SummaryView displays a summary view of repo state
type SummaryView struct {
	*SelectableRowView
	channels          Channels
	repoData          RepoData
	repoController    RepoController
	config            Config
	activeViewPos     ViewPos
	lastViewDimension ViewDimension
	variables         GRVVariableSetter
	handlers          map[ActionType]summaryViewHandler
	lines             []summaryViewLine
	lock              sync.Mutex
}

// NewGitSummaryView creates a new summary view instance
func NewGitSummaryView(repoData RepoData, repoController RepoController, channels Channels, config Config, variables GRVVariableSetter) *SummaryView {
	summaryView := &SummaryView{
		repoData:       repoData,
		repoController: repoController,
		channels:       channels,
		config:         config,
		activeViewPos:  NewViewPosition(),
		variables:      variables,
		handlers:       map[ActionType]summaryViewHandler{},
	}

	summaryView.SelectableRowView = NewSelectableRowView(summaryView, channels, config, variables, &summaryView.lock, "summary row")

	return summaryView
}

// Initialise the summary view
func (summaryView *SummaryView) Initialise() (err error) {
	summaryView.lock.Lock()
	defer summaryView.lock.Unlock()

	summaryView.repoData.RegisterRefStateListener(summaryView)
	summaryView.generateRows()
	return summaryView.selectNearestSelectableRow()
}

// Render generates and writes the summary view to the provided window
func (summaryView *SummaryView) Render(win RenderWindow) (err error) {
	summaryView.lock.Lock()
	defer summaryView.lock.Unlock()

	summaryView.lastViewDimension = win.ViewDimensions()
	lineNum := summaryView.rows()

	rows := win.Rows() - 2
	viewPos := summaryView.activeViewPos
	viewPos.DetermineViewStartRow(rows, lineNum)

	lineIndex := viewPos.ViewStartRowIndex()
	startColumn := viewPos.ViewStartColumn()
	var lineBuilder *LineBuilder

	for rowIndex := uint(0); rowIndex < rows && lineIndex < lineNum; rowIndex++ {
		if lineBuilder, err = win.LineBuilder(rowIndex+1, startColumn); err != nil {
			return
		}

		lineBuilder.Append(svIndentationSpace)
		line := summaryView.lines[lineIndex]
		line.render(lineBuilder)

		lineIndex++
	}

	if err = win.SetSelectedRow(viewPos.SelectedRowIndex()+1, summaryView.active); err != nil {
		return
	}

	if searchActive, searchPattern, lastSearchFoundMatch := summaryView.viewSearch.SearchActive(); searchActive && lastSearchFoundMatch {
		if err = win.Highlight(searchPattern, CmpAllviewSearchMatch); err != nil {
			return
		}
	}

	return
}

// RenderHelpBar shows key bindings custom to the summary view
func (summaryView *SummaryView) RenderHelpBar(lineBuilder *LineBuilder) (err error) {
	return
}

// ViewID returns the diff views ID
func (summaryView *SummaryView) ViewID() ViewID {
	return ViewGitSummary
}

func (summaryView *SummaryView) viewPos() ViewPos {
	return summaryView.activeViewPos
}

func (summaryView *SummaryView) line(lineIndex uint) (line string) {
	if lineIndex >= summaryView.rows() {
		return
	}

	return summaryView.lines[lineIndex].renderString()
}

func (summaryView *SummaryView) rows() uint {
	return uint(len(summaryView.lines))
}

func (summaryView *SummaryView) viewDimension() ViewDimension {
	return summaryView.lastViewDimension
}

func (summaryView *SummaryView) onRowSelected(rowIndex uint) (err error) {
	summaryView.SelectableRowView.setVariables()
	return
}

func (summaryView *SummaryView) isSelectableRow(rowIndex uint) (isSelectable bool) {
	if rowIndex >= summaryView.rows() {
		return
	}

	return summaryView.lines[rowIndex].isSelectable()
}

func (summaryView *SummaryView) generateRows() {
	lines := summaryView.generateBranchRows()
	summaryView.lines = lines
	summaryView.channels.UpdateDisplay()
}

func (summaryView *SummaryView) generateBranchRows() (rows []summaryViewLine) {
	ref := summaryView.repoData.Head()
	rows = append(rows,
		emptyLine,
		newHeaderRenderer("Branch"),
		&branchLine{head: ref},
		emptyLine,
	)

	return
}

// OnRefsChanged regenerates the summary view
func (summaryView *SummaryView) OnRefsChanged(addedRefs, removedRefs []Ref, updatedRefs []*UpdatedRef) {
	summaryView.lock.Lock()
	defer summaryView.lock.Unlock()

	summaryView.generateRows()
}

// OnHeadChanged regenerates the summary view
func (summaryView *SummaryView) OnHeadChanged(oldHead, newHead Ref) {
	summaryView.lock.Lock()
	defer summaryView.lock.Unlock()

	summaryView.generateRows()
}

// OnTrackingBranchesUpdated regenerates the summary view
func (summaryView *SummaryView) OnTrackingBranchesUpdated(trackingBranches []*LocalBranch) {
	summaryView.lock.Lock()
	defer summaryView.lock.Unlock()

	summaryView.generateRows()
}

// HandleAction checks if the summary view supports the provided action and executes it if so
func (summaryView *SummaryView) HandleAction(action Action) (err error) {
	summaryView.lock.Lock()
	defer summaryView.lock.Unlock()

	var handled bool
	if handler, ok := summaryView.handlers[action.ActionType]; ok {
		log.Debugf("Action handled by SummaryView")
		err = handler(summaryView, action)
	} else if handled, err = summaryView.SelectableRowView.HandleAction(action); handled {
		log.Debugf("Action handled by SelectableRowChildWindowView")
	} else {
		log.Debugf("Action not handled")
	}

	return
}
