package main

import (
	"fmt"
	"sync"

	log "github.com/Sirupsen/logrus"
)

type summaryViewHandler func(*SummaryView, Action) error

type summaryViewLine interface {
	render(*LineBuilder)
	renderString() string
}

type emptyLineRenderer struct{}

func (emptyLineRenderer *emptyLineRenderer) render(lineBuilder *LineBuilder) {}
func (emptyLineRenderer *emptyLineRenderer) renderString() string {
	return ""
}

var emptyLine = &emptyLineRenderer{}

type headerRenderer struct {
	header string
}

func (headerRenderer *headerRenderer) render(lineBuilder *LineBuilder) {
	lineBuilder.AppendWithStyle(CmpNone, " %v", headerRenderer.header)
}

func (headerRenderer *headerRenderer) renderString() string {
	return headerRenderer.header
}

type branchRenderer struct {
	branchName  string
	aheadBehind string
}

func (branchRenderer *branchRenderer) render(lineBuilder *LineBuilder) {
	lineBuilder.AppendWithStyle(CmpNone, " %v", branchRenderer.branchName)
}

func (branchRenderer *branchRenderer) renderString() string {
	return branchRenderer.branchName
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

	summaryView.generateRows()

	return
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
		if lineBuilder, err = win.LineBuilder(rowIndex, startColumn); err != nil {
			return
		}

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

	return
}

func (summaryView *SummaryView) rows() uint {
	return uint(len(summaryView.lines))
}

func (summaryView *SummaryView) viewDimension() ViewDimension {
	return summaryView.lastViewDimension
}

func (summaryView *SummaryView) onRowSelected(rowIndex uint) (err error) {
	return
}

func (summaryView *SummaryView) isSelectableRow(rowIndex uint) (isSelectable bool) {
	if rowIndex >= summaryView.rows() {
		return
	}

	switch summaryView.lines[rowIndex].(type) {
	case *emptyLineRenderer:
		isSelectable = false
	case *headerRenderer:
		isSelectable = false
	default:
		isSelectable = true
	}

	return
}

func (summaryView *SummaryView) generateRows() {
	lines := summaryView.generateBranchRows()
	summaryView.lines = lines
	summaryView.channels.UpdateDisplay()
}

func (summaryView *SummaryView) generateBranchRows() (rows []summaryViewLine) {
	ref := summaryView.repoData.Head()
	var branchName string

	if _, isDetached := ref.(*HEAD); isDetached {
		GetDetachedHeadDisplayValue(ref.Oid())
	} else {
		branchName = ref.Shorthand()
	}

	var aheadBehind string

	if branch, isLocalBranch := ref.(*LocalBranch); isLocalBranch && branch.IsTrackingBranch() {
		aheadBehind = fmt.Sprintf(" (ahead: %v, behind: %v)", branch.ahead, branch.behind)
	}

	rows = append(rows,
		emptyLine,
		&headerRenderer{
			header: "Branch",
		},
		&branchRenderer{
			branchName:  branchName,
			aheadBehind: aheadBehind,
		},
		emptyLine,
	)

	return
}

// HandleAction checks if the summary view supports the provided action and executes it if so
func (summaryView *SummaryView) HandleAction(action Action) (err error) {
	summaryView.lock.Lock()
	defer summaryView.lock.Unlock()

	var handled bool
	if handler, ok := summaryView.handlers[action.ActionType]; ok {
		log.Debugf("Action handled by SummaryView")
		err = handler(summaryView, action)
	} else if handled, err = summaryView.AbstractWindowView.HandleAction(action); handled {
		log.Debugf("Action handled by AbstractWindowView")
	} else {
		log.Debugf("Action not handled")
	}

	return
}
