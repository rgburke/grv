package main

import (
	"bufio"
	"bytes"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"strings"
	"sync"
)

type DiffViewHandler func(*DiffView) error

type DiffLineType int

const (
	DLT_UNSET DiffLineType = iota
	DLT_NORMAL
	DLT_GIT_DIFF_HEADER
	DLT_GIT_DIFF_EXTENDED_HEADER
	DLT_UNIFIED_DIFF_HEADER
	DLT_HUNK_START
	DLT_LINE_ADDED
	DLT_LINE_REMOVED
)

var diffLineThemeComponentId = map[DiffLineType]ThemeComponentId{
	DLT_NORMAL:                   CMP_DIFFVIEW_DIFFLINE_NORMAL,
	DLT_GIT_DIFF_HEADER:          CMP_DIFFVIEW_DIFFLINE_GIT_DIFF_HEADER,
	DLT_GIT_DIFF_EXTENDED_HEADER: CMP_DIFFVIEW_DIFFLINE_GIT_DIFF_EXTENDED_HEADER,
	DLT_UNIFIED_DIFF_HEADER:      CMP_DIFFVIEW_DIFFLINE_UNIFIED_DIFF_HEADER,
	DLT_HUNK_START:               CMP_DIFFVIEW_DIFFLINE_HUNK_START,
	DLT_LINE_ADDED:               CMP_DIFFVIEW_DIFFLINE_LINE_ADDED,
	DLT_LINE_REMOVED:             CMP_DIFFVIEW_DIFFLINE_LINE_REMOVED,
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

type Diff struct {
	lines   []*DiffLine
	viewPos *ViewPos
}

type DiffView struct {
	channels      *Channels
	repoData      RepoData
	activeCommit  *Commit
	commitDiffs   map[*Commit]*Diff
	viewPos       *ViewPos
	viewDimension ViewDimension
	handlers      map[Action]DiffViewHandler
	active        bool
	lock          sync.Mutex
}

func NewDiffView(repoData RepoData, channels *Channels) *DiffView {
	return &DiffView{
		repoData:    repoData,
		channels:    channels,
		viewPos:     NewViewPos(),
		commitDiffs: make(map[*Commit]*Diff),
		handlers: map[Action]DiffViewHandler{
			ACTION_PREV_LINE:    MoveUpDiffLine,
			ACTION_NEXT_LINE:    MoveDownDiffLine,
			ACTION_SCROLL_RIGHT: ScrollDiffViewRight,
			ACTION_SCROLL_LEFT:  ScrollDiffViewLeft,
			ACTION_FIRST_LINE:   MoveToFirstDiffLine,
			ACTION_LAST_LINE:    MoveToLastDiffLine,
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
	diff := diffView.commitDiffs[diffView.activeCommit]
	lineNum := uint(len(diff.lines))
	viewPos.DetermineViewStartRow(rows, lineNum)

	lineIndex := viewPos.viewStartRowIndex
	startColumn := viewPos.viewStartColumn

	for rowIndex := uint(0); rowIndex < rows && lineIndex < lineNum; rowIndex++ {
		diffLine := diff.lines[lineIndex]
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
				AppendWithStyle(themeComponentId, "%v", strings.Join(lineParts[:2], "")).
				AppendWithStyle(CMP_DIFFVIEW_DIFFLINE_HUNK_HEADER, "%v", lineParts[2])
		} else if err = win.SetRow(rowIndex+1, startColumn, themeComponentId, " %v", diff.lines[lineIndex].line); err != nil {
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

	if diff, ok := diffView.commitDiffs[diffView.activeCommit]; ok {
		diff.viewPos = diffView.viewPos
	}

	if diff, ok := diffView.commitDiffs[commit]; ok {
		diffView.activeCommit = commit
		diffView.viewPos = diff.viewPos
		diffView.channels.UpdateDisplay()
		return
	}

	buf, err := diffView.repoData.Diff(commit)
	if err != nil {
		return
	}

	scanner := bufio.NewScanner(bytes.NewReader(buf.Bytes()))
	var lines []*DiffLine

	for scanner.Scan() {
		lines = append(lines, &DiffLine{
			line: scanner.Text(),
		})
	}

	diffView.commitDiffs[commit] = &Diff{
		lines: lines,
	}

	diffView.activeCommit = commit
	diffView.viewPos = NewViewPos()
	diffView.channels.UpdateDisplay()

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

	if handler, ok := diffView.handlers[action]; ok {
		err = handler(diffView)
	}

	return
}

func MoveDownDiffLine(diffView *DiffView) (err error) {
	diff := diffView.commitDiffs[diffView.activeCommit]
	lineNum := uint(len(diff.lines))
	viewPos := diffView.viewPos

	if viewPos.MoveLineDown(lineNum) {
		log.Debugf("Moving down one line in diff view")
		diffView.channels.UpdateDisplay()
	}

	return
}

func MoveUpDiffLine(diffView *DiffView) (err error) {
	viewPos := diffView.viewPos

	if viewPos.MoveLineUp() {
		log.Debugf("Moving up one line in diff view")
		diffView.channels.UpdateDisplay()
	}

	return
}

func ScrollDiffViewRight(diffView *DiffView) (err error) {
	viewPos := diffView.viewPos
	viewPos.MovePageRight(diffView.viewDimension.cols)
	log.Debugf("Scrolling right. View starts at column %v", viewPos.viewStartColumn)
	diffView.channels.UpdateDisplay()

	return
}

func ScrollDiffViewLeft(diffView *DiffView) (err error) {
	viewPos := diffView.viewPos

	if viewPos.MovePageLeft(diffView.viewDimension.cols) {
		log.Debugf("Scrolling left. View starts at column %v", viewPos.viewStartColumn)
		diffView.channels.UpdateDisplay()
	}

	return
}

func MoveToFirstDiffLine(diffView *DiffView) (err error) {
	viewPos := diffView.viewPos

	if viewPos.MoveToFirstLine() {
		log.Debugf("Moving to first line in diff view")
		diffView.channels.UpdateDisplay()
	}

	return
}

func MoveToLastDiffLine(diffView *DiffView) (err error) {
	diff := diffView.commitDiffs[diffView.activeCommit]
	lineNum := uint(len(diff.lines))
	viewPos := diffView.viewPos

	if viewPos.MoveToLastLine(lineNum) {
		log.Debugf("Moving to last line in diff view")
		diffView.channels.UpdateDisplay()
	}

	return
}
