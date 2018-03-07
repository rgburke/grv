// +build ignore

package main

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

func NewCommitGraph() *CommitGraph {
	return &CommitGraph{}
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
			return commitIndex
		}
	}

	return -1
}

func (commitGraph *CommitGraph) AddCommit(commit *Commit) (err error) {
	parentCommits, err := commitGraph.repoData.CommitParents(commit.oid)
	if err != nil {
		return
	}

	cellType := commitGraph.determineCommitCellType(parentCommits)
	parentIndexes := commitGraph.determineParentIndex(commit)
	row := newCommitGraphRow()

	parentsSeen := 0
	for parentIndex, parentCommit := range commitGraph.parentCommits {
		if _, isParentCommit := parentIndexes[parentIndex]; isParentCommit {
			if parentsSeen == 0 {
				row.add(cellType)
				parentsSeen++
			} else {
			}
		} else {

		}

		if parentsSeen > 0 && parentsSeen < len(parentIndexes) && cellType == cgtMergeCommit {
			row.add(cgtBranchOffLine)
		} else {
			row.add(cgtEmpty)
		}
	}

	if commitGraphRow.isEmpty() {
		row.add(cellType)
	}

	commitGraph.updateParentCommits(parentCommits)
}

func (commitGraph *CommitGraph) determineCommitCellType(parentCommits *[]Commit) commitCellGraphType {
	if len(parentCommits) > 1 {
		return cgtMergeCommit
	}

	return cgtCommit
}

func (commitGraph *CommitGraph) determineParentIndexes(commit *Commit) (parentIndexes map[int]bool) {
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
	}

	commitIndex := row.commitIndex()
	commitGraph.parentCommits[0] = parentCommits[0]

	for i := 1; i < len(parentCommits); i++ {
		commitGraph.parentCommits = append(commitGraph.parentCommits, parentCommits[i])
	}
}
