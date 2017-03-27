package main

import (
	"bufio"
	"bytes"
	log "github.com/Sirupsen/logrus"
	gc "github.com/rthornton128/goncurses"
	"sync"
)

type DiffViewHandler func(*DiffView) error

type DiffLine struct {
	line string
}

type Diff struct {
	lines []*DiffLine
}

type DiffView struct {
	channels       *Channels
	repoData       RepoData
	activeCommit   *Commit
	commitDiffs    map[*Commit]*Diff
	activeIndex    uint
	viewStartIndex uint
	handlers       map[gc.Key]DiffViewHandler
	active         bool
	lock           sync.Mutex
}

func NewDiffView(repoData RepoData, channels *Channels) *DiffView {
	return &DiffView{
		repoData:    repoData,
		channels:    channels,
		commitDiffs: make(map[*Commit]*Diff),
		handlers: map[gc.Key]DiffViewHandler{
			gc.KEY_UP:   MoveUpLine,
			gc.KEY_DOWN: MoveDownLine,
		},
	}
}

func (diffView *DiffView) Initialise() (err error) {
	return
}

func (diffView *DiffView) Render(win RenderWindow) (err error) {
	diffView.lock.Lock()
	defer diffView.lock.Unlock()

	if diffView.activeCommit == nil {
		return
	}

	rows := win.Rows() - 2

	if diffView.viewStartIndex > diffView.activeIndex {
		diffView.viewStartIndex = diffView.activeIndex
	} else if rowDiff := diffView.activeIndex - diffView.viewStartIndex; rowDiff >= rows {
		diffView.viewStartIndex += (rowDiff - rows) + 1
	}

	diff := diffView.commitDiffs[diffView.activeCommit]
	lineNum := uint(len(diff.lines))
	lineIndex := diffView.viewStartIndex

	for rowIndex := uint(0); rowIndex < rows && lineIndex < lineNum; rowIndex++ {
		if err = win.SetRow(rowIndex+1, " %v", diff.lines[lineIndex].line); err != nil {
			return
		}

		lineIndex++
	}

	if err = win.SetSelectedRow((diffView.activeIndex-diffView.viewStartIndex)+1, diffView.active); err != nil {
		return
	}

	win.DrawBorder()

	if err = win.SetTitle(CMP_COMMITVIEW_TITLE, "Diff for commit %v", diffView.activeCommit.commit.Id().String()); err != nil {
		return
	}

	if err = win.SetFooter(CMP_COMMITVIEW_FOOTER, "Line %v of %v", lineIndex+1, lineNum); err != nil {
		return
	}

	return
}

func (diffView *DiffView) OnActiveChange(active bool) {
	log.Debugf("DiffView active: %v", active)
	diffView.lock.Lock()
	defer diffView.lock.Unlock()

	diffView.active = active
}

func (diffView *DiffView) OnCommitSelect(commit *Commit) (err error) {
	diffView.lock.Lock()
	defer diffView.lock.Unlock()

	if _, ok := diffView.commitDiffs[commit]; ok {
		diffView.activeCommit = commit
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
	diffView.channels.UpdateDisplay()

	return
}

func (diffView *DiffView) Handle(keyPressEvent KeyPressEvent) (err error) {
	log.Debugf("DiffView handling key %v", keyPressEvent)
	diffView.lock.Lock()
	defer diffView.lock.Unlock()

	if handler, ok := diffView.handlers[keyPressEvent.key]; ok {
		err = handler(diffView)
	}

	return
}

func MoveDownLine(diffView *DiffView) (err error) {
	diff := diffView.commitDiffs[diffView.activeCommit]
	lineNum := len(diff.lines)

	if lineNum > 0 && diffView.activeIndex < uint(lineNum-1) {
		diffView.activeIndex++
		diffView.channels.UpdateDisplay()
	}

	return
}

func MoveUpLine(diffView *DiffView) (err error) {
	if diffView.activeIndex > 0 {
		diffView.activeIndex--
		diffView.channels.UpdateDisplay()
	}

	return
}
