package main

import (
	"bytes"
	"io/ioutil"

	log "github.com/Sirupsen/logrus"
)

// CommitGraph ...
type CommitGraph struct {
	repoData      RepoData
	rows          []*commitGraphRow
	parentCommits []*Commit
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
)

// NewCommitGraph ...
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

func (commitGraphRow *commitGraphRow) commitIndex() int {
	for commitIndex, cellType := range commitGraphRow.cells {
		if cellType == cgtCommit || cellType == cgtMergeCommit {
			return commitIndex / 2
		}
	}

	return -1
}

type rowProperties int

const (
	rpNone           rowProperties = 0
	rpBranchOff      rowProperties = 1 << 0
	rpMultiBranchOff rowProperties = 1 << 1
	rpMergeIn        rowProperties = 1 << 2
	rpMultiMergeIn   rowProperties = 1 << 3
)

type commitGraphRowBuilder struct {
	parentCommits       []*Commit
	commitCellType      commitGraphCellType
	parentCommitIndexes map[int]bool
	properties          rowProperties
	parentsSeen         int
}

func newCommitGraphRowBuilder(parentCommits []*Commit, commitCellType commitGraphCellType, parentCommitIndexes map[int]bool) *commitGraphRowBuilder {
	return &commitGraphRowBuilder{
		parentCommits:       parentCommits,
		commitCellType:      commitCellType,
		parentCommitIndexes: parentCommitIndexes,
	}
}

func (builder *commitGraphRowBuilder) determineRowProperties() {
	if builder.commitCellType == cgtMergeCommit {
		builder.properties &= rpMergeIn
	}

	if len(builder.parentCommits) > 2 {
		builder.properties &= rpMultiMergeIn
	}

	if len(builder.parentCommitIndexes) > 1 {
		builder.properties &= rpBranchOff
	}

	if len(builder.parentCommitIndexes) > 2 {
		builder.properties &= rpMultiBranchOff
	}

	builder.parentsSeen = 0
}

func (builder *commitGraphRowBuilder) build() *commitGraphRow {
	builder.determineRowProperties()
	row := newCommitGraphRow()

	for parentIndex := range builder.parentCommits {
		if _, isParentCommit := builder.parentCommitIndexes[parentIndex]; isParentCommit {
			if builder.parentsSeen == 0 {
				row.add(builder.commitCellType)
			} else {
				if builder.parentsSeen < len(builder.parentCommitIndexes)-1 {
					row.add(cgtMultiBranchOffLine)
				} else {
					row.add(cgtBranchOffLine)
				}
			}

			builder.parentsSeen++
		} else {
			row.add(cgtParentLine)
		}

		//if parentIndex != len(builder.parentCommits)-1 {
		if (builder.parentsSeen > 0 && builder.commitCellType == cgtMergeCommit) ||
			(builder.parentsSeen > 0 && builder.parentsSeen < len(builder.parentCommitIndexes) && builder.commitCellType == cgtCommit) {
			row.add(cgtCrossLine)
		} else {
			row.add(cgtEmpty)
		}
		//}
	}

	if row.isEmpty() {
		row.add(builder.commitCellType)
	} else if builder.commitCellType == cgtMergeCommit {
		row.add(cgtMergeCommitLine)
	}

	return row
}

// AddCommit ...
func (commitGraph *CommitGraph) AddCommit(commit *Commit) (err error) {
	parentCommits, err := commitGraph.repoData.CommitParents(commit.oid)
	if err != nil {
		return
	}

	commitCellType := commitGraph.determineCommitCellType(parentCommits)
	parentCommitIndexes := commitGraph.determineParentIndexes(commit)

	row := newCommitGraphRowBuilder(commitGraph.parentCommits, commitCellType, parentCommitIndexes).build()
	commitGraph.addRow(row)

	commitGraph.updateParentCommits(row, parentCommits)

	return
}

func (commitGraph *CommitGraph) determineCommitCellType(parentCommits []*Commit) commitGraphCellType {
	if len(parentCommits) > 1 {
		return cgtMergeCommit
	}

	return cgtCommit
}

func (commitGraph *CommitGraph) determineParentIndexes(commit *Commit) (parentIndexes map[int]bool) {
	parentIndexes = make(map[int]bool)

	for parentIndex, parentCommit := range commitGraph.parentCommits {
		if parentCommit != nil && parentCommit.oid.Equal(commit.oid) {
			parentIndexes[parentIndex] = true
		}
	}

	return
}

func (commitGraph *CommitGraph) updateParentCommits(row *commitGraphRow, parentCommits []*Commit) {
	if len(parentCommits) == 0 {
		return
	} else if len(commitGraph.parentCommits) == 0 {
		commitGraph.parentCommits = append(commitGraph.parentCommits, parentCommits...)
		return
	}

	commitIndex := row.commitIndex()
	log.Debugf("GRAPH: rowIndex: %v, commitIndex: %v", len(commitGraph.rows)-1, commitIndex)
	commitGraph.parentCommits[commitIndex] = parentCommits[0]

	for i := 1; i < len(parentCommits); i++ {
		commitGraph.parentCommits = append(commitGraph.parentCommits, parentCommits[i])
	}
}

func (commitGraph *CommitGraph) addRow(row *commitGraphRow) {
	commitGraph.rows = append(commitGraph.rows, row)
}

// WriteToFile ...
func (commitGraph *CommitGraph) WriteToFile(filePath string) error {
	var buf bytes.Buffer

	for _, row := range commitGraph.rows {
		for _, cellType := range row.cells {
			var cellString string

			switch cellType {
			case cgtEmpty:
				cellString = " "
			case cgtCommit:
				cellString = "o"
			case cgtMergeCommit:
				cellString = "M"
			case cgtParentLine:
				cellString = "│"
			case cgtMergeCommitLine:
				cellString = "┐"
			case cgtCrossLine:
				cellString = "─"
			case cgtBranchOffLine:
				cellString = "┘"
			case cgtMultiBranchOffLine:
				cellString = "┴"
			default:
				cellString = "?"
			}

			buf.WriteString(cellString)
		}

		buf.WriteString("\n")
	}

	return ioutil.WriteFile(filePath, buf.Bytes(), 0644)
}
