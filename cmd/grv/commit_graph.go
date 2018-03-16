package main

import (
	"math"
	"sync"
)

// CommitGraph handles building and displaying a commit graph
type CommitGraph struct {
	repoData      RepoData
	rows          []*commitGraphRow
	parentCommits []*Commit
	lock          sync.Mutex
}

type commitGraphRow struct {
	cells []commitGraphCellType
}

type commitGraphCellType int

const (
	cgtEmpty              commitGraphCellType = iota //
	cgtCommit                                        // o
	cgtMergeCommit                                   // M
	cgtParentLine                                    // │
	cgtMergeCommitLine                               // ┐
	cgtCrossLine                                     // ─
	cgtBranchOffLine                                 // ┘
	cgtMultiBranchOffLine                            // ┴
	cgtShiftIn                                       // ┘
	cgtShiftDown                                     // ┌
)

// NewCommitGraph creates a new CommitGraph instance
func NewCommitGraph(repoData RepoData) *CommitGraph {
	return &CommitGraph{
		repoData: repoData,
	}
}

func newCommitGraphRow() *commitGraphRow {
	return &commitGraphRow{}
}

func (commitGraphRow *commitGraphRow) add(cellType commitGraphCellType) {
	commitGraphRow.cells = append(commitGraphRow.cells, cellType)
}

func (commitGraphRow *commitGraphRow) isEmpty() bool {
	return len(commitGraphRow.cells) == 0
}

type commitGraphRowBuilder struct {
	parentCommits       []*Commit
	commitCellType      commitGraphCellType
	parentCommitIndexes map[int]bool
	parentsSeen         int
}

func newCommitGraphRowBuilder(parentCommits []*Commit, commitCellType commitGraphCellType, parentCommitIndexes map[int]bool) *commitGraphRowBuilder {
	return &commitGraphRowBuilder{
		parentCommits:       parentCommits,
		commitCellType:      commitCellType,
		parentCommitIndexes: parentCommitIndexes,
	}
}

func (builder *commitGraphRowBuilder) isParentCommit(parentIndex int) bool {
	_, isParentCommit := builder.parentCommitIndexes[parentIndex]
	return isParentCommit
}

func (builder *commitGraphRowBuilder) determineParentCommitCellType(parentIndex int) (cellType commitGraphCellType) {
	if builder.parentsSeen == 0 {
		cellType = builder.commitCellType
	} else {
		if builder.parentsSeen < len(builder.parentCommitIndexes)-1 {
			cellType = cgtMultiBranchOffLine
		} else {
			cellType = cgtBranchOffLine
		}
	}

	return
}

func (builder *commitGraphRowBuilder) determineSeparatorCellType(parentIndex int) (cellType commitGraphCellType, exists bool) {
	if (builder.parentsSeen > 0 && builder.commitCellType == cgtMergeCommit &&
		!(parentIndex == len(builder.parentCommits)-1 && (builder.lastCellIsBranchOff() || builder.lastCellIsShiftedIn()))) ||
		(builder.parentsSeen > 0 && builder.parentsSeen < len(builder.parentCommitIndexes) && builder.commitCellType == cgtCommit) ||
		builder.isShiftedInCell(parentIndex) {
		cellType = cgtCrossLine
		exists = true
	} else if parentIndex != len(builder.parentCommits)-1 {
		cellType = cgtEmpty
		exists = true
	}

	return
}

func (builder *commitGraphRowBuilder) isShiftedInCell(parentIndex int) (shiftedIn bool) {
	if parentIndex+1 < len(builder.parentCommits) {
		shiftedIn = builder.parentCommits[parentIndex+1] == nil
	}

	return
}

func (builder *commitGraphRowBuilder) lastCellIsBranchOff() bool {
	if len(builder.parentCommitIndexes) < 2 {
		return false
	}

	for parentIndex := range builder.parentCommitIndexes {
		if parentIndex == len(builder.parentCommits)-1 {
			return true
		}
	}

	return false
}

func (builder *commitGraphRowBuilder) lastCellIsShiftedIn() bool {
	return len(builder.parentCommits) > 0 && builder.parentCommits[len(builder.parentCommits)-1] == nil
}

func (builder *commitGraphRowBuilder) build() *commitGraphRow {
	row := newCommitGraphRow()

	for parentIndex, parentCommit := range builder.parentCommits {
		if parentCommit == nil {
			if builder.isShiftedInCell(parentIndex) {
				row.add(cgtCrossLine)
			} else {
				row.add(cgtShiftIn)
			}
		} else if builder.isParentCommit(parentIndex) {
			row.add(builder.determineParentCommitCellType(parentIndex))
			builder.parentsSeen++
		} else {
			if builder.isShiftedInCell(parentIndex) {
				row.add(cgtShiftDown)
			} else {
				row.add(cgtParentLine)
			}
		}

		if cellSeparatorType, exists := builder.determineSeparatorCellType(parentIndex); exists {
			row.add(cellSeparatorType)
		}
	}

	rowIsEmpty := row.isEmpty()
	if rowIsEmpty {
		row.add(builder.commitCellType)
	}
	if builder.commitCellType == cgtMergeCommit {
		if !builder.lastCellIsBranchOff() && !builder.lastCellIsShiftedIn() {
			if rowIsEmpty {
				row.add(cgtCrossLine)
			}

			row.add(cgtMergeCommitLine)
		}
	}

	return row
}

// AddCommit adds the next commit to the graph
func (commitGraph *CommitGraph) AddCommit(commit *Commit) (err error) {
	commitGraph.lock.Lock()
	defer commitGraph.lock.Unlock()

	parentCommits, err := commitGraph.repoData.CommitParents(commit.oid)
	if err != nil {
		return
	}

	commitGraph.beforeAddCommitUpdateParentCommits()
	commitCellType := commitGraph.determineCommitCellType(parentCommits)
	parentCommitIndexes := commitGraph.determineParentIndexes(commit)

	row := newCommitGraphRowBuilder(commitGraph.parentCommits, commitCellType, parentCommitIndexes).build()
	commitGraph.addRow(row)

	commitGraph.afterAddCommitUpdateParentCommits(parentCommitIndexes, parentCommits)

	return
}

func (commitGraph *CommitGraph) determineCommitCellType(parentCommits []*Commit) commitGraphCellType {
	if len(parentCommits) > 1 {
		return cgtMergeCommit
	}

	return cgtCommit
}

func (commitGraph *CommitGraph) determineParentIndexes(commit *Commit) (parentCommitIndexes map[int]bool) {
	parentCommitIndexes = make(map[int]bool)

	for parentIndex, parentCommit := range commitGraph.parentCommits {
		if parentCommit != nil && parentCommit.oid.Equal(commit.oid) {
			parentCommitIndexes[parentIndex] = true
		}
	}

	return
}

func (commitGraph *CommitGraph) beforeAddCommitUpdateParentCommits() {
	if len(commitGraph.parentCommits) < 2 {
		return
	}

	for i := len(commitGraph.parentCommits) - 2; i > -1; i-- {
		if commitGraph.parentCommits[i] == nil {
			commitGraph.parentCommits[i] = commitGraph.parentCommits[i+1]
			commitGraph.parentCommits[i+1] = nil
		}
	}
}

func (commitGraph *CommitGraph) afterAddCommitUpdateParentCommits(parentCommitIndexes map[int]bool, parentCommits []*Commit) {
	if len(parentCommits) == 0 {
		return
	} else if len(commitGraph.parentCommits) == 0 {
		commitGraph.parentCommits = append(commitGraph.parentCommits, parentCommits...)
		return
	}

	minParentIndex := commitGraph.minParentIndex(parentCommitIndexes)
	if len(parentCommitIndexes) > 1 {
		for parentIndex := range parentCommitIndexes {
			if parentIndex > minParentIndex {
				commitGraph.parentCommits[parentIndex] = nil
			}
		}
	}

	var nilsIndex int
	for nilsIndex = len(commitGraph.parentCommits) - 1; commitGraph.parentCommits[nilsIndex] == nil; nilsIndex-- {
	}

	if nilsIndex != len(commitGraph.parentCommits)-1 {
		commitGraph.parentCommits = commitGraph.parentCommits[:nilsIndex+1]
	}

	commitGraph.parentCommits[minParentIndex] = parentCommits[0]

	for i := 1; i < len(parentCommits); i++ {
		commitGraph.parentCommits = append(commitGraph.parentCommits, parentCommits[i])
	}
}

func (commitGraph *CommitGraph) minParentIndex(parentCommitIndexes map[int]bool) (minParentIndex int) {
	if len(parentCommitIndexes) == 0 {
		return -1
	}

	minParentIndex = math.MaxInt32

	for parentIndex := range parentCommitIndexes {
		if parentIndex < minParentIndex {
			minParentIndex = parentIndex
		}
	}

	return
}

func (commitGraph *CommitGraph) addRow(row *commitGraphRow) {
	commitGraph.rows = append(commitGraph.rows, row)
}

func (commitGraph *CommitGraph) rowCount() uint {
	return uint(len(commitGraph.rows))
}

// Rows returns the number of rows in the commit graph
func (commitGraph *CommitGraph) Rows() uint {
	commitGraph.lock.Lock()
	defer commitGraph.lock.Unlock()

	return commitGraph.rowCount()
}

// Render the graph row for the specified commit
func (commitGraph *CommitGraph) Render(lineBuilder *LineBuilder, commitIndex uint) {
	commitGraph.lock.Lock()
	defer commitGraph.lock.Unlock()

	if commitIndex >= commitGraph.rowCount() {
		return
	}

	row := commitGraph.rows[commitIndex]
	themeComponentID := CmpNone

	for _, cellType := range row.cells {
		switch cellType {
		case cgtEmpty:
			lineBuilder.AppendWithStyle(themeComponentID, " ")
		case cgtCommit:
			lineBuilder.AppendACSChar(AcsDiamond, themeComponentID)
		case cgtMergeCommit:
			lineBuilder.AppendWithStyle(themeComponentID, "M")
		case cgtParentLine:
			lineBuilder.AppendACSChar(AcsVline, themeComponentID)
		case cgtMergeCommitLine:
			lineBuilder.AppendACSChar(AcsUrcorner, themeComponentID)
		case cgtCrossLine:
			lineBuilder.AppendACSChar(AcsHline, themeComponentID)
		case cgtBranchOffLine, cgtShiftIn:
			lineBuilder.AppendACSChar(AcsLrcorner, themeComponentID)
		case cgtMultiBranchOffLine:
			lineBuilder.AppendACSChar(AcsBtee, themeComponentID)
		case cgtShiftDown:
			lineBuilder.AppendACSChar(AcsUlcorner, themeComponentID)
		}
	}

	lineBuilder.Append(" ")

	return
}
