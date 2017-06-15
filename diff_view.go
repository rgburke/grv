package main

import (
	"bufio"
	"bytes"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"strings"
	"sync"
)

type DiffViewHandler func(*DiffView, Action) error

type DiffLineType int

const (
	DLT_UNSET DiffLineType = iota
	DLT_NORMAL
	DLT_DIFF_COMMIT_AUTHOR
	DLT_DIFF_COMMIT_AUTHOR_DATE
	DLT_DIFF_COMMIT_COMMITTER
	DLT_DIFF_COMMIT_COMMITTER_DATE
	DLT_DIFF_COMMIT_SUMMARY
	DLT_DIFF_STATS_FILE
	DLT_GIT_DIFF_HEADER
	DLT_GIT_DIFF_EXTENDED_HEADER
	DLT_UNIFIED_DIFF_HEADER
	DLT_HUNK_START
	DLT_LINE_ADDED
	DLT_LINE_REMOVED
)

const (
	DV_DATE_FORMAT = "Mon Jan 2 15:04:05 2006 -0700"
)

var diffLineThemeComponentId = map[DiffLineType]ThemeComponentId{
	DLT_NORMAL:                     CMP_DIFFVIEW_DIFFLINE_NORMAL,
	DLT_DIFF_COMMIT_AUTHOR:         CMP_DIFFVIEW_DIFFLINE_DIFF_COMMIT_AUTHOR,
	DLT_DIFF_COMMIT_AUTHOR_DATE:    CMP_DIFFVIEW_DIFFLINE_DIFF_COMMIT_AUTHOR_DATE,
	DLT_DIFF_COMMIT_COMMITTER:      CMP_DIFFVIEW_DIFFLINE_DIFF_COMMIT_COMMITTER,
	DLT_DIFF_COMMIT_COMMITTER_DATE: CMP_DIFFVIEW_DIFFLINE_DIFF_COMMIT_COMMITTER_DATE,
	DLT_DIFF_COMMIT_SUMMARY:        CMP_DIFFVIEW_DIFFLINE_DIFF_COMMIT_SUMMARY,
	DLT_DIFF_STATS_FILE:            CMP_DIFFVIEW_DIFFLINE_DIFF_STATS_FILE,
	DLT_GIT_DIFF_HEADER:            CMP_DIFFVIEW_DIFFLINE_GIT_DIFF_HEADER,
	DLT_GIT_DIFF_EXTENDED_HEADER:   CMP_DIFFVIEW_DIFFLINE_GIT_DIFF_EXTENDED_HEADER,
	DLT_UNIFIED_DIFF_HEADER:        CMP_DIFFVIEW_DIFFLINE_UNIFIED_DIFF_HEADER,
	DLT_HUNK_START:                 CMP_DIFFVIEW_DIFFLINE_HUNK_START,
	DLT_LINE_ADDED:                 CMP_DIFFVIEW_DIFFLINE_LINE_ADDED,
	DLT_LINE_REMOVED:               CMP_DIFFVIEW_DIFFLINE_LINE_REMOVED,
}

type DiffLine struct {
	line         string
	diffLineType DiffLineType
}

func (diffLine *DiffLine) GetThemeComponentId() ThemeComponentId {
	diffLine.DetermineDiffLineType()
	return diffLineThemeComponentId[diffLine.diffLineType]
}

func (diffLine *DiffLine) DetermineDiffLineType() {
	if diffLine.diffLineType != DLT_UNSET {
		return
	}

	var diffLineType DiffLineType
	line := diffLine.line

	switch {
	case strings.HasPrefix(line, "diff --git"):
		diffLineType = DLT_GIT_DIFF_HEADER
	case strings.HasPrefix(line, "index"):
		diffLineType = DLT_GIT_DIFF_EXTENDED_HEADER
	case strings.HasPrefix(line, "---") || strings.HasPrefix(line, "+++"):
		diffLineType = DLT_UNIFIED_DIFF_HEADER
	case strings.HasPrefix(line, "@@"):
		diffLineType = DLT_HUNK_START
	case strings.HasPrefix(line, "+"):
		diffLineType = DLT_LINE_ADDED
	case strings.HasPrefix(line, "-"):
		diffLineType = DLT_LINE_REMOVED
	default:
		diffLineType = DLT_NORMAL
	}

	diffLine.diffLineType = diffLineType
}

type DiffLines struct {
	lines   []*DiffLine
	viewPos *ViewPos
}

type DiffView struct {
	channels      *Channels
	repoData      RepoData
	activeCommit  *Commit
	commitDiffs   map[*Commit]*DiffLines
	viewPos       *ViewPos
	viewDimension ViewDimension
	handlers      map[ActionType]DiffViewHandler
	active        bool
	search        *Search
	lock          sync.Mutex
}

func NewDiffView(repoData RepoData, channels *Channels) *DiffView {
	return &DiffView{
		repoData:    repoData,
		channels:    channels,
		viewPos:     NewViewPos(),
		commitDiffs: make(map[*Commit]*DiffLines),
		handlers: map[ActionType]DiffViewHandler{
			ACTION_PREV_LINE:    MoveUpDiffLine,
			ACTION_NEXT_LINE:    MoveDownDiffLine,
			ACTION_PREV_PAGE:    MoveUpDiffPage,
			ACTION_NEXT_PAGE:    MoveDownDiffPage,
			ACTION_SCROLL_RIGHT: ScrollDiffViewRight,
			ACTION_SCROLL_LEFT:  ScrollDiffViewLeft,
			ACTION_FIRST_LINE:   MoveToFirstDiffLine,
			ACTION_LAST_LINE:    MoveToLastDiffLine,
			ACTION_SEARCH:       DoDiffSearch,
		},
	}
}

func (diffView *DiffView) Initialise() (err error) {
	return
}

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

	lineIndex := viewPos.viewStartRowIndex
	startColumn := viewPos.viewStartColumn

	for rowIndex := uint(0); rowIndex < rows && lineIndex < lineNum; rowIndex++ {
		diffLine := diffLines.lines[lineIndex]
		themeComponentId := diffLine.GetThemeComponentId()

		if diffLine.diffLineType == DLT_HUNK_START {
			lineParts := strings.SplitAfter(diffLine.line, "@@")

			if len(lineParts) != 3 {
				return fmt.Errorf("Unable to display hunk header line: %v", diffLine.line)
			}

			var lineBuilder *LineBuilder
			if lineBuilder, err = win.LineBuilder(rowIndex+1, startColumn); err != nil {
				return
			}

			lineBuilder.
				AppendWithStyle(themeComponentId, " %v", strings.Join(lineParts[:2], "")).
				AppendWithStyle(CMP_DIFFVIEW_DIFFLINE_HUNK_HEADER, "%v", lineParts[2])

		} else if diffLine.diffLineType == DLT_DIFF_STATS_FILE {
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

			lineBuilder.AppendWithStyle(CMP_DIFFVIEW_DIFFLINE_DIFF_STATS_FILE, " %v |", filePart)

			for _, char := range changePart {
				switch char {
				case '+':
					lineBuilder.AppendWithStyle(CMP_DIFFVIEW_DIFFLINE_LINE_ADDED, "%c", char)
				case '-':
					lineBuilder.AppendWithStyle(CMP_DIFFVIEW_DIFFLINE_LINE_REMOVED, "%c", char)
				default:
					lineBuilder.Append("%c", char)
				}
			}
		} else if err = win.SetRow(rowIndex+1, startColumn, themeComponentId, " %v", diffLines.lines[lineIndex].line); err != nil {
			return
		}

		lineIndex++
	}

	if err = win.SetSelectedRow((viewPos.activeRowIndex-viewPos.viewStartRowIndex)+1, diffView.active); err != nil {
		return
	}

	win.DrawBorder()

	if err = win.SetTitle(CMP_COMMITVIEW_TITLE, "Diff for commit %v", diffView.activeCommit.commit.Id().String()); err != nil {
		return
	}

	if err = win.SetFooter(CMP_COMMITVIEW_FOOTER, "Line %v of %v", viewPos.activeRowIndex+1, lineNum); err != nil {
		return
	}

	return
}

func (diffView *DiffView) RenderStatusBar(lineBuilder *LineBuilder) (err error) {
	return
}

func (diffView *DiffView) RenderHelpBar(lineBuilder *LineBuilder) (err error) {
	diffView.lock.Lock()
	defer diffView.lock.Unlock()

	return
}

func (diffView *DiffView) OnActiveChange(active bool) {
	log.Debugf("DiffView active: %v", active)
	diffView.lock.Lock()
	defer diffView.lock.Unlock()

	diffView.active = active
}

func (diffView *DiffView) ViewId() ViewId {
	return VIEW_DIFF
}

func (diffView *DiffView) OnCommitSelect(commit *Commit) (err error) {
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
	diffView.viewPos = NewViewPos()
	diffView.channels.UpdateDisplay()

	return
}

func (diffView *DiffView) generateDiffLines(commit *Commit) (err error) {
	var lines []*DiffLine

	author := commit.commit.Author()
	committer := commit.commit.Committer()

	lines = append(lines,
		&DiffLine{
			line:         fmt.Sprintf("Author:\t%v <%v>", author.Name, author.Email),
			diffLineType: DLT_DIFF_COMMIT_AUTHOR,
		},
		&DiffLine{
			line:         fmt.Sprintf("AuthorDate:\t%v", author.When.Format(DV_DATE_FORMAT)),
			diffLineType: DLT_DIFF_COMMIT_AUTHOR_DATE,
		},
		&DiffLine{
			line:         fmt.Sprintf("Comitter:\t%v <%v>", committer.Name, committer.Email),
			diffLineType: DLT_DIFF_COMMIT_COMMITTER,
		},
		&DiffLine{
			line:         fmt.Sprintf("ComitterDate:\t%v", committer.When.Format(DV_DATE_FORMAT)),
			diffLineType: DLT_DIFF_COMMIT_COMMITTER_DATE,
		},
		&DiffLine{
			diffLineType: DLT_NORMAL,
		},
		&DiffLine{
			line:         commit.commit.Summary(),
			diffLineType: DLT_DIFF_COMMIT_SUMMARY,
		},
		&DiffLine{
			diffLineType: DLT_NORMAL,
		},
	)

	diff, err := diffView.repoData.Diff(commit)
	if err != nil {
		return
	}

	scanner := bufio.NewScanner(bytes.NewReader(diff.stats.Bytes()))

	for scanner.Scan() {
		lines = append(lines, &DiffLine{
			line:         strings.TrimPrefix(scanner.Text(), " "),
			diffLineType: DLT_DIFF_STATS_FILE,
		})
	}

	if len(lines) > 0 {
		prevLine := lines[len(lines)-1]

		if prevLine.diffLineType == DLT_DIFF_STATS_FILE {
			prevLine.diffLineType = DLT_NORMAL
		}
	}

	lines = append(lines, &DiffLine{
		diffLineType: DLT_NORMAL,
	})

	scanner = bufio.NewScanner(bytes.NewReader(diff.diffText.Bytes()))

	for scanner.Scan() {
		lines = append(lines, &DiffLine{
			line: scanner.Text(),
		})
	}

	diffView.commitDiffs[commit] = &DiffLines{
		lines: lines,
	}

	return
}

func (diffView *DiffView) HandleKeyPress(keystring string) (err error) {
	log.Debugf("DiffView handling key %v - NOP", keystring)
	return
}

func (diffView *DiffView) HandleAction(action Action) (err error) {
	log.Debugf("DiffView handling action %v", action)
	diffView.lock.Lock()
	defer diffView.lock.Unlock()

	if handler, ok := diffView.handlers[action.ActionType]; ok {
		err = handler(diffView, action)
	}

	return
}

func (diffView *DiffView) Line(lineIndex uint) (line string, lineExists bool) {
	diffLines := diffView.commitDiffs[diffView.activeCommit]
	lineNum := uint(len(diffLines.lines))

	if lineIndex < lineNum {
		diffLine := diffLines.lines[lineIndex]
		line = diffLine.line
		lineExists = true
	}

	return
}

func MoveDownDiffLine(diffView *DiffView, action Action) (err error) {
	diffLines := diffView.commitDiffs[diffView.activeCommit]
	lineNum := uint(len(diffLines.lines))
	viewPos := diffView.viewPos

	if viewPos.MoveLineDown(lineNum) {
		log.Debugf("Moving down one line in diff view")
		diffView.channels.UpdateDisplay()
	}

	return
}

func MoveUpDiffLine(diffView *DiffView, action Action) (err error) {
	viewPos := diffView.viewPos

	if viewPos.MoveLineUp() {
		log.Debugf("Moving up one line in diff view")
		diffView.channels.UpdateDisplay()
	}

	return
}

func MoveDownDiffPage(diffView *DiffView, action Action) (err error) {
	diffLines := diffView.commitDiffs[diffView.activeCommit]
	lineNum := uint(len(diffLines.lines))
	viewPos := diffView.viewPos

	if viewPos.MovePageDown(diffView.viewDimension.rows-2, lineNum) {
		log.Debugf("Moving down one page in diff view")
		diffView.channels.UpdateDisplay()
	}

	return
}

func MoveUpDiffPage(diffView *DiffView, action Action) (err error) {
	viewPos := diffView.viewPos

	if viewPos.MovePageUp(diffView.viewDimension.rows - 2) {
		log.Debugf("Moving up one page in diff view")
		diffView.channels.UpdateDisplay()
	}

	return
}

func ScrollDiffViewRight(diffView *DiffView, action Action) (err error) {
	viewPos := diffView.viewPos
	viewPos.MovePageRight(diffView.viewDimension.cols)
	log.Debugf("Scrolling right. View starts at column %v", viewPos.viewStartColumn)
	diffView.channels.UpdateDisplay()

	return
}

func ScrollDiffViewLeft(diffView *DiffView, action Action) (err error) {
	viewPos := diffView.viewPos

	if viewPos.MovePageLeft(diffView.viewDimension.cols) {
		log.Debugf("Scrolling left. View starts at column %v", viewPos.viewStartColumn)
		diffView.channels.UpdateDisplay()
	}

	return
}

func MoveToFirstDiffLine(diffView *DiffView, action Action) (err error) {
	viewPos := diffView.viewPos

	if viewPos.MoveToFirstLine() {
		log.Debugf("Moving to first line in diff view")
		diffView.channels.UpdateDisplay()
	}

	return
}

func MoveToLastDiffLine(diffView *DiffView, action Action) (err error) {
	diffLines := diffView.commitDiffs[diffView.activeCommit]
	lineNum := uint(len(diffLines.lines))
	viewPos := diffView.viewPos

	if viewPos.MoveToLastLine(lineNum) {
		log.Debugf("Moving to last line in diff view")
		diffView.channels.UpdateDisplay()
	}

	return
}

func DoDiffSearch(diffView *DiffView, action Action) (err error) {
	if !(len(action.Args) > 0) {
		return fmt.Errorf("Expected search pattern")
	}

	pattern, ok := action.Args[0].(string)
	if !ok {
		return fmt.Errorf("Expected search pattern")
	}

	search, err := NewSearch(pattern, diffView)
	if err != nil {
		return
	}

	diffView.search = search

	return FindNextDiffMatch(diffView, action)
}

func FindNextDiffMatch(diffView *DiffView, action Action) (err error) {
	diffLines := diffView.commitDiffs[diffView.activeCommit]
	lineNum := uint(len(diffLines.lines))
	viewPos := diffView.viewPos
	lineIndex := (viewPos.activeRowIndex + 1) % lineNum

	matchLineIndex, found := diffView.search.FindNext(lineIndex)

	if found {
		viewPos.activeRowIndex = matchLineIndex
		diffView.channels.UpdateDisplay()
	}

	return
}
