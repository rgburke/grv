package main

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"
	"sync"

	log "github.com/Sirupsen/logrus"
)

type diffViewHandler func(*DiffView, Action) error

type diffLineType int

const (
	dltUnset diffLineType = iota
	dltNormal
	dltDiffCommitAuthor
	dltDiffCommitAuthorDate
	dltDiffCommitCommitter
	dltDiffCommitCommitterDate
	dltDiffCommitSummary
	dltDiffStatsFile
	dltGitDiffHeader
	dltGitDiffExtendedHeader
	dltUnifiedDiffHeader
	dltHunkStart
	dltLineAdded
	dltLineRemoved
)

const (
	dvDateFormat = "Mon Jan 2 15:04:05 2006 -0700"
)

var diffLineThemeComponentID = map[diffLineType]ThemeComponentID{
	dltNormal:                  CmpDiffviewDifflineNormal,
	dltDiffCommitAuthor:        CmpDiffviewDifflineDiffCommitAuthor,
	dltDiffCommitAuthorDate:    CmpDiffviewDifflineDiffCommitAuthorDate,
	dltDiffCommitCommitter:     CmpDiffviewDifflineDiffCommitCommitter,
	dltDiffCommitCommitterDate: CmpDiffviewDifflineDiffCommitCommitterDate,
	dltDiffCommitSummary:       CmpDiffviewDifflineDiffCommitSummary,
	dltDiffStatsFile:           CmpDiffviewDifflineDiffStatsFile,
	dltGitDiffHeader:           CmpDiffviewDifflineGitDiffHeader,
	dltGitDiffExtendedHeader:   CmpDiffviewDifflineGitDiffExtendedHeader,
	dltUnifiedDiffHeader:       CmpDiffviewDifflineUnifiedDiffHeader,
	dltHunkStart:               CmpDiffviewDifflineHunkStart,
	dltLineAdded:               CmpDiffviewDifflineLineAdded,
	dltLineRemoved:             CmpDiffviewDifflineLineRemoved,
}

type diffLineData struct {
	line     string
	lineType diffLineType
}

func (diffLine *diffLineData) getThemeComponentID() ThemeComponentID {
	diffLine.determineDiffLineType()
	return diffLineThemeComponentID[diffLine.lineType]
}

func (diffLine *diffLineData) determineDiffLineType() {
	if diffLine.lineType != dltUnset {
		return
	}

	var lineType diffLineType
	line := diffLine.line

	switch {
	case strings.HasPrefix(line, "diff --git"):
		lineType = dltGitDiffHeader
	case strings.HasPrefix(line, "index"):
		lineType = dltGitDiffExtendedHeader
	case strings.HasPrefix(line, "---") || strings.HasPrefix(line, "+++"):
		lineType = dltUnifiedDiffHeader
	case strings.HasPrefix(line, "@@"):
		lineType = dltHunkStart
	case strings.HasPrefix(line, "+"):
		lineType = dltLineAdded
	case strings.HasPrefix(line, "-"):
		lineType = dltLineRemoved
	default:
		lineType = dltNormal
	}

	diffLine.lineType = lineType
}

type diffLines struct {
	lines   []*diffLineData
	viewPos ViewPos
}

// DiffView contains all state for the diff view
type DiffView struct {
	channels      *Channels
	repoData      RepoData
	activeCommit  *Commit
	commitDiffs   map[*Commit]*diffLines
	viewPos       ViewPos
	viewDimension ViewDimension
	handlers      map[ActionType]diffViewHandler
	active        bool
	viewSearch    *ViewSearch
	lock          sync.Mutex
}

// NewDiffView creates a new diff view instance
func NewDiffView(repoData RepoData, channels *Channels) *DiffView {
	diffView := &DiffView{
		repoData:    repoData,
		channels:    channels,
		viewPos:     NewViewPosition(),
		commitDiffs: make(map[*Commit]*diffLines),
		handlers: map[ActionType]diffViewHandler{
			ActionPrevLine:    moveUpDiffLine,
			ActionNextLine:    moveDownDiffLine,
			ActionPrevPage:    moveUpDiffPage,
			ActionNextPage:    moveDownDiffPage,
			ActionScrollRight: scrollDiffViewRight,
			ActionScrollLeft:  scrollDiffViewLeft,
			ActionFirstLine:   moveToFirstDiffLine,
			ActionLastLine:    moveToLastDiffLine,
			ActionCenterView:  centerDiffView,
		},
	}

	diffView.viewSearch = NewViewSearch(diffView, channels)

	return diffView
}

// Initialise does nothing
func (diffView *DiffView) Initialise() (err error) {
	return
}

// Render generates and writes the diff view to the provided window
func (diffView *DiffView) Render(win RenderWindow) (err error) {
	diffView.lock.Lock()
	defer diffView.lock.Unlock()

	diffView.viewDimension = win.ViewDimensions()

	if diffView.activeCommit == nil {
		return
	}

	rows := win.Rows() - 2
	viewPos := diffView.viewPos
	diffLines := diffView.commitDiffs[diffView.activeCommit]
	lineNum := uint(len(diffLines.lines))
	viewPos.DetermineViewStartRow(rows, lineNum)

	lineIndex := viewPos.ViewStartRowIndex()
	startColumn := viewPos.ViewStartColumn()

	for rowIndex := uint(0); rowIndex < rows && lineIndex < lineNum; rowIndex++ {
		diffLine := diffLines.lines[lineIndex]
		themeComponentID := diffLine.getThemeComponentID()

		if diffLine.lineType == dltHunkStart {
			lineParts := strings.SplitAfter(diffLine.line, "@@")

			if len(lineParts) != 3 {
				return fmt.Errorf("Unable to display hunk header line: %v", diffLine.line)
			}

			var lineBuilder *LineBuilder
			if lineBuilder, err = win.LineBuilder(rowIndex+1, startColumn); err != nil {
				return
			}

			lineBuilder.
				AppendWithStyle(themeComponentID, " %v", strings.Join(lineParts[:2], "")).
				AppendWithStyle(CmpDiffviewDifflineHunkHeader, "%v", lineParts[2])

		} else if diffLine.lineType == dltDiffStatsFile {
			sepIndex := strings.LastIndex(diffLine.line, "|")

			if sepIndex == -1 || sepIndex >= len(diffLine.line)-1 {
				return fmt.Errorf("Unable to display diff stats file line: %v", diffLine.line)
			}

			filePart := diffLine.line[0:sepIndex]
			changePart := diffLine.line[sepIndex+1:]

			var lineBuilder *LineBuilder
			if lineBuilder, err = win.LineBuilder(rowIndex+1, startColumn); err != nil {
				return
			}

			lineBuilder.AppendWithStyle(CmpDiffviewDifflineDiffStatsFile, " %v |", filePart)

			for _, char := range changePart {
				switch char {
				case '+':
					lineBuilder.AppendWithStyle(CmpDiffviewDifflineLineAdded, "%c", char)
				case '-':
					lineBuilder.AppendWithStyle(CmpDiffviewDifflineLineRemoved, "%c", char)
				default:
					lineBuilder.Append("%c", char)
				}
			}
		} else if err = win.SetRow(rowIndex+1, startColumn, themeComponentID, " %v", diffLines.lines[lineIndex].line); err != nil {
			return
		}

		lineIndex++
	}

	if err = win.SetSelectedRow(viewPos.SelectedRowIndex()+1, diffView.active); err != nil {
		return
	}

	win.DrawBorder()

	if err = win.SetTitle(CmpCommitviewTitle, "Diff for commit %v", diffView.activeCommit.commit.Id().String()); err != nil {
		return
	}

	if err = win.SetFooter(CmpCommitviewFooter, "Line %v of %v", viewPos.ActiveRowIndex()+1, lineNum); err != nil {
		return
	}

	if searchActive, searchPattern, lastSearchFoundMatch := diffView.viewSearch.SearchActive(); searchActive && lastSearchFoundMatch {
		if err = win.Highlight(searchPattern, CmpAllviewSearchMatch); err != nil {
			return
		}
	}

	return
}

// RenderStatusBar does nothing
func (diffView *DiffView) RenderStatusBar(lineBuilder *LineBuilder) (err error) {
	return
}

// RenderHelpBar does nothing
func (diffView *DiffView) RenderHelpBar(lineBuilder *LineBuilder) (err error) {
	return
}

// OnActiveChange sets whether the diff view is the active view or not
func (diffView *DiffView) OnActiveChange(active bool) {
	log.Debugf("DiffView active: %v", active)
	diffView.lock.Lock()
	defer diffView.lock.Unlock()

	diffView.active = active
}

// ViewID returns the diff views ID
func (diffView *DiffView) ViewID() ViewID {
	return ViewDiff
}

// OnCommitSelect loads/fetches the diff for the selected commit and refreshes the display
func (diffView *DiffView) OnCommitSelect(commit *Commit) (err error) {
	log.Debugf("DiffView loading diff for selected commit %v", commit.commit.Id())

	diffView.lock.Lock()
	defer diffView.lock.Unlock()

	if diffLines, ok := diffView.commitDiffs[diffView.activeCommit]; ok {
		diffLines.viewPos = diffView.viewPos
	}

	if diffLines, ok := diffView.commitDiffs[commit]; ok {
		diffView.activeCommit = commit
		diffView.viewPos = diffLines.viewPos
		diffView.channels.UpdateDisplay()
		return
	}

	if err = diffView.generateDiffLines(commit); err != nil {
		return
	}

	diffView.activeCommit = commit
	diffView.viewPos = NewViewPosition()
	diffView.channels.UpdateDisplay()

	return
}

func (diffView *DiffView) generateDiffLines(commit *Commit) (err error) {
	var lines []*diffLineData

	author := commit.commit.Author()
	committer := commit.commit.Committer()

	lines = append(lines,
		&diffLineData{
			line:     fmt.Sprintf("Author:\t%v <%v>", author.Name, author.Email),
			lineType: dltDiffCommitAuthor,
		},
		&diffLineData{
			line:     fmt.Sprintf("AuthorDate:\t%v", author.When.Format(dvDateFormat)),
			lineType: dltDiffCommitAuthorDate,
		},
		&diffLineData{
			line:     fmt.Sprintf("Committer:\t%v <%v>", committer.Name, committer.Email),
			lineType: dltDiffCommitCommitter,
		},
		&diffLineData{
			line:     fmt.Sprintf("CommitterDate:\t%v", committer.When.Format(dvDateFormat)),
			lineType: dltDiffCommitCommitterDate,
		},
		&diffLineData{
			lineType: dltNormal,
		},
		&diffLineData{
			line:     commit.commit.Summary(),
			lineType: dltDiffCommitSummary,
		},
		&diffLineData{
			lineType: dltNormal,
		},
	)

	diff, err := diffView.repoData.Diff(commit)
	if err != nil {
		return
	}

	scanner := bufio.NewScanner(bytes.NewReader(diff.stats.Bytes()))

	for scanner.Scan() {
		lines = append(lines, &diffLineData{
			line:     strings.TrimPrefix(scanner.Text(), " "),
			lineType: dltDiffStatsFile,
		})
	}

	if len(lines) > 0 {
		prevLine := lines[len(lines)-1]

		if prevLine.lineType == dltDiffStatsFile {
			prevLine.lineType = dltNormal
		}
	}

	lines = append(lines, &diffLineData{
		lineType: dltNormal,
	})

	scanner = bufio.NewScanner(bytes.NewReader(diff.diffText.Bytes()))

	for scanner.Scan() {
		lines = append(lines, &diffLineData{
			line: scanner.Text(),
		})
	}

	diffView.commitDiffs[commit] = &diffLines{
		lines: lines,
	}

	return
}

// HandleKeyPress does nothing
func (diffView *DiffView) HandleKeyPress(keystring string) (err error) {
	log.Debugf("DiffView handling key %v - NOP", keystring)
	return
}

// ViewPos returns the current view position
func (diffView *DiffView) ViewPos() ViewPos {
	return diffView.viewPos
}

// OnSearchMatch sets the current view position to the search match position
func (diffView *DiffView) OnSearchMatch(startPos ViewPos, matchLineIndex uint) {
	diffView.lock.Lock()
	defer diffView.lock.Unlock()

	viewPos := diffView.ViewPos()

	if viewPos != startPos {
		log.Debugf("Selected ref has changed since search started")
		return
	}

	viewPos.SetActiveRowIndex(matchLineIndex)
}

// HandleAction checks if the diff view supports the provided action and executes it if so
func (diffView *DiffView) HandleAction(action Action) (err error) {
	log.Debugf("DiffView handling action %v", action)
	diffView.lock.Lock()
	defer diffView.lock.Unlock()

	if handler, ok := diffView.handlers[action.ActionType]; ok {
		err = handler(diffView, action)
	} else {
		_, err = diffView.viewSearch.HandleAction(action)
	}

	return
}

// Line returns the rendered line from the diff view at the specified line index
func (diffView *DiffView) Line(lineIndex uint) (line string) {
	diffView.lock.Lock()
	defer diffView.lock.Unlock()

	diffLines := diffView.commitDiffs[diffView.activeCommit]
	lineNum := uint(len(diffLines.lines))

	if lineIndex >= lineNum {
		log.Errorf("Invalid lineIndex: %v", lineIndex)
		return
	}

	diffLine := diffLines.lines[lineIndex]
	line = diffLine.line

	return
}

// LineNumber returns the number of lines the diff view currently has
func (diffView *DiffView) LineNumber() (lineNumber uint) {
	diffView.lock.Lock()
	defer diffView.lock.Unlock()

	diffLines := diffView.commitDiffs[diffView.activeCommit]
	lineNum := uint(len(diffLines.lines))

	return lineNum
}

func moveDownDiffLine(diffView *DiffView, action Action) (err error) {
	diffLines := diffView.commitDiffs[diffView.activeCommit]
	lineNum := uint(len(diffLines.lines))
	viewPos := diffView.viewPos

	if viewPos.MoveLineDown(lineNum) {
		log.Debugf("Moving down one line in diff view")
		diffView.channels.UpdateDisplay()
	}

	return
}

func moveUpDiffLine(diffView *DiffView, action Action) (err error) {
	viewPos := diffView.viewPos

	if viewPos.MoveLineUp() {
		log.Debugf("Moving up one line in diff view")
		diffView.channels.UpdateDisplay()
	}

	return
}

func moveDownDiffPage(diffView *DiffView, action Action) (err error) {
	diffLines := diffView.commitDiffs[diffView.activeCommit]
	lineNum := uint(len(diffLines.lines))
	viewPos := diffView.viewPos

	if viewPos.MovePageDown(diffView.viewDimension.rows-2, lineNum) {
		log.Debugf("Moving down one page in diff view")
		diffView.channels.UpdateDisplay()
	}

	return
}

func moveUpDiffPage(diffView *DiffView, action Action) (err error) {
	viewPos := diffView.viewPos

	if viewPos.MovePageUp(diffView.viewDimension.rows - 2) {
		log.Debugf("Moving up one page in diff view")
		diffView.channels.UpdateDisplay()
	}

	return
}

func scrollDiffViewRight(diffView *DiffView, action Action) (err error) {
	viewPos := diffView.viewPos
	viewPos.MovePageRight(diffView.viewDimension.cols)
	log.Debugf("Scrolling right. View starts at column %v", viewPos.ViewStartColumn())
	diffView.channels.UpdateDisplay()

	return
}

func scrollDiffViewLeft(diffView *DiffView, action Action) (err error) {
	viewPos := diffView.viewPos

	if viewPos.MovePageLeft(diffView.viewDimension.cols) {
		log.Debugf("Scrolling left. View starts at column %v", viewPos.ViewStartColumn())
		diffView.channels.UpdateDisplay()
	}

	return
}

func moveToFirstDiffLine(diffView *DiffView, action Action) (err error) {
	viewPos := diffView.viewPos

	if viewPos.MoveToFirstLine() {
		log.Debugf("Moving to first line in diff view")
		diffView.channels.UpdateDisplay()
	}

	return
}

func moveToLastDiffLine(diffView *DiffView, action Action) (err error) {
	diffLines := diffView.commitDiffs[diffView.activeCommit]
	lineNum := uint(len(diffLines.lines))
	viewPos := diffView.viewPos

	if viewPos.MoveToLastLine(lineNum) {
		log.Debugf("Moving to last line in diff view")
		diffView.channels.UpdateDisplay()
	}

	return
}

func centerDiffView(diffView *DiffView, action Action) (err error) {
	viewPos := diffView.viewPos

	if viewPos.CenterActiveRow(diffView.viewDimension.rows - 2) {
		log.Debug("Centering DiffView")
		diffView.channels.UpdateDisplay()
	}

	return
}
