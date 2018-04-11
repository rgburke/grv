package main

import (
	"math"
	"sync"

	log "github.com/Sirupsen/logrus"
)

// TODO: Rewrite this mess

var branchThemeComponentIDs = []ThemeComponentID{
	CmpCommitviewGraphBranch1,
	CmpCommitviewGraphBranch2,
	CmpCommitviewGraphBranch3,
	CmpCommitviewGraphBranch4,
	CmpCommitviewGraphBranch5,
	CmpCommitviewGraphBranch6,
	CmpCommitviewGraphBranch7,
}

// CommitGraph handles building and displaying a commit graph
type CommitGraph struct {
	repoData      RepoData
	rows          []*commitGraphRow
	parentCommits []*Commit
	branchIndexes []int
	lock          sync.Mutex
}

type commitGraphRow struct {
	cells []*commitGraphCell
}

type commitGraphCell struct {
	cellType    commitGraphCellType
	branchIndex int
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

func newCommitGraphCell(cellType commitGraphCellType, branchIndex int) *commitGraphCell {
	return &commitGraphCell{
		cellType:    cellType,
		branchIndex: branchIndex,
	}
}

func (commitGraphRow *commitGraphRow) add(cellType commitGraphCellType, branchIndex int) {
	commitGraphRow.cells = append(commitGraphRow.cells, newCommitGraphCell(cellType, branchIndex))
}

func (commitGraphRow *commitGraphRow) isEmpty() bool {
	return len(commitGraphRow.cells) == 0
}

type commitGraphRowBuilder struct {
	parentCommits       []*Commit
	commitCellType      commitGraphCellType
	parentCommitIndexes map[int]bool
	branchIndexes       []int
	parentsSeen         int
}

func newCommitGraphRowBuilder(parentCommits []*Commit, commitCellType commitGraphCellType, parentCommitIndexes map[int]bool, branchIndexes []int) *commitGraphRowBuilder {
	return &commitGraphRowBuilder{
		parentCommits:       parentCommits,
		commitCellType:      commitCellType,
		parentCommitIndexes: parentCommitIndexes,
		branchIndexes:       branchIndexes,
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
		if builder.isShiftedInCell(parentIndex) {
			cellType = cgtCrossLine
		} else if builder.parentsSeen < len(builder.parentCommitIndexes)-1 {
			cellType = cgtMultiBranchOffLine
		} else {
			cellType = cgtBranchOffLine
		}
	}

	return
}

func (builder *commitGraphRowBuilder) determineSeparatorCellType(parentIndex int) (cellType commitGraphCellType, branchIndex int, exists bool) {
	isMergeCrossLine := builder.parentsSeen > 0 && builder.commitCellType == cgtMergeCommit &&
		!(parentIndex == len(builder.parentCommits)-1 && (builder.lastCellIsBranchOff() || builder.lastCellIsShiftedIn()))

	isBranchOff := len(builder.parentCommitIndexes) > 1 && builder.parentsSeen > 0 && builder.parentsSeen < len(builder.parentCommitIndexes)

	isShiftedInCell := builder.isShiftedInCell(parentIndex)

	notLastParentCommit := parentIndex != len(builder.parentCommits)-1

	branchIndex = builder.getOrSetBranchIndex(parentIndex)

	switch {
	case isBranchOff && !isShiftedInCell:
		cellType = cgtCrossLine
		exists = true

		maxParentIndex := 0
		for parentIndex := range builder.parentCommitIndexes {
			if parentIndex > maxParentIndex {
				maxParentIndex = parentIndex
			}
		}

		branchIndex = builder.getOrSetBranchIndex(maxParentIndex)
	case isMergeCrossLine:
		cellType = cgtCrossLine
		exists = true

		if parentIndex == len(builder.parentCommits)-1 {
			branchIndex = builder.getOrSetBranchIndex(len(builder.parentCommits))
		} else if builder.isShiftedInCell(parentIndex) {
			prevParentIndex := builder.previousParentCommitIndex(parentIndex)
			branchIndex = builder.getOrSetBranchIndex(prevParentIndex)
		} else {
			branchIndex = builder.getOrSetBranchIndex(parentIndex + 1)
		}
	case isShiftedInCell:
		cellType = cgtCrossLine
		exists = true

		nextParentIndex := builder.nextParentCommitIndex(parentIndex)
		branchIndex = builder.getOrSetBranchIndex(nextParentIndex)
	case notLastParentCommit:
		cellType = cgtEmpty
		exists = true
	}

	return
}

func (builder *commitGraphRowBuilder) previousParentCommitIndex(parentIndex int) (prevParentIndex int) {
	for prevParentIndex = parentIndex; prevParentIndex > -1; prevParentIndex-- {
		if builder.parentCommits[prevParentIndex] != nil {
			return
		}
	}

	return
}

func (builder *commitGraphRowBuilder) nextParentCommitIndex(parentIndex int) (nextParentIndex int) {
	for nextParentIndex = parentIndex; nextParentIndex > -1; nextParentIndex-- {
		if builder.parentCommits[nextParentIndex] != nil {
			return
		}
	}

	for nextParentIndex = parentIndex + 1; nextParentIndex < len(builder.parentCommits); nextParentIndex++ {
		if builder.parentCommits[nextParentIndex] != nil {
			return
		}
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

func (builder *commitGraphRowBuilder) nextBranchIndex() int {
	maxBranchIndex := 0

	if len(builder.branchIndexes) == 0 {
		return maxBranchIndex
	}

	for _, branchIndex := range builder.branchIndexes {
		if branchIndex > maxBranchIndex {
			maxBranchIndex = branchIndex
		}
	}

	return maxBranchIndex + 1
}

func (builder *commitGraphRowBuilder) getOrSetBranchIndex(parentIndex int) (branchIndex int) {
	if parentIndex < len(builder.branchIndexes) {
		return builder.branchIndexes[parentIndex]
	} else if parentIndex != len(builder.branchIndexes) {
		log.Errorf("Invalid parentIndex: %v, len(builder.branchIndexes): %v", parentIndex, len(builder.branchIndexes))
	}

	branchIndex = builder.nextBranchIndex()
	builder.branchIndexes = append(builder.branchIndexes, branchIndex)

	return
}

func (builder *commitGraphRowBuilder) build() *commitGraphRow {
	row := newCommitGraphRow()

	for parentIndex, parentCommit := range builder.parentCommits {
		var branchIndex int

		if parentCommit == nil {
			nextParentIndex := builder.nextParentCommitIndex(parentIndex)
			branchIndex = builder.getOrSetBranchIndex(nextParentIndex)

			if builder.isShiftedInCell(parentIndex) {
				row.add(cgtCrossLine, branchIndex)
			} else {
				row.add(cgtShiftIn, branchIndex)
			}

		} else {
			branchIndex = builder.getOrSetBranchIndex(parentIndex)

			if builder.isParentCommit(parentIndex) {
				row.add(builder.determineParentCommitCellType(parentIndex), branchIndex)
				builder.parentsSeen++
			} else {
				if builder.isShiftedInCell(parentIndex) {
					row.add(cgtShiftDown, branchIndex)
				} else {
					row.add(cgtParentLine, branchIndex)
				}
			}
		}

		if cellSeparatorType, branchIndex, exists := builder.determineSeparatorCellType(parentIndex); exists {
			row.add(cellSeparatorType, branchIndex)
		}
	}

	rowIsEmpty := row.isEmpty()
	if rowIsEmpty {
		row.add(builder.commitCellType, 0)
	}
	if builder.commitCellType == cgtMergeCommit {
		if !builder.lastCellIsBranchOff() && !builder.lastCellIsShiftedIn() {
			branchIndex := builder.getOrSetBranchIndex(len(builder.parentCommits))
			if rowIsEmpty {
				if len(builder.parentCommits) == 0 {
					branchIndex = builder.getOrSetBranchIndex(1)
				}

				row.add(cgtCrossLine, branchIndex)
			}

			row.add(cgtMergeCommitLine, branchIndex)
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

	builder := newCommitGraphRowBuilder(commitGraph.parentCommits, commitCellType, parentCommitIndexes, commitGraph.branchIndexes)
	row := builder.build()
	commitGraph.branchIndexes = builder.branchIndexes

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

			branchIndex := commitGraph.branchIndexes[i]
			commitGraph.branchIndexes[i] = commitGraph.branchIndexes[i+1]
			commitGraph.branchIndexes[i+1] = branchIndex
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
		commitGraph.branchIndexes = commitGraph.branchIndexes[:nilsIndex+1]
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
	var themeComponentID ThemeComponentID

	for _, cell := range row.cells {
		themeComponentID = branchThemeComponentIDs[cell.branchIndex%len(branchThemeComponentIDs)]

		switch cell.cellType {
		case cgtEmpty:
			lineBuilder.AppendWithStyle(themeComponentID, " ")
		case cgtCommit:
			lineBuilder.AppendACSChar(AcsDiamond, CmpCommitviewGraphCommit)
		case cgtMergeCommit:
			lineBuilder.AppendWithStyle(CmpCommitviewGraphMergeCommit, "M")
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

// Clear removes all rows
func (commitGraph *CommitGraph) Clear() {
	commitGraph.lock.Lock()
	defer commitGraph.lock.Unlock()

	commitGraph.rows = nil
	commitGraph.parentCommits = nil
	commitGraph.branchIndexes = nil
}
