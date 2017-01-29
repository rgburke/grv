package main

type CommitView struct {
	repoData          RepoData
	activeBranch      *Oid
	activeCommitIndex map[*Oid]uint
}

func NewCommitView(repoData RepoData) *CommitView {
	return &CommitView{
		repoData:          repoData,
		activeCommitIndex: make(map[*Oid]uint),
	}
}

func (commitView *CommitView) Initialise() (err error) {
	return
}

func (commitView *CommitView) Render(win RenderWindow) (err error) {
	rowIndex := uint(1)
	commits := commitView.repoData.Commits(commitView.activeBranch)
	activeCommitIndex := commitView.activeCommitIndex[commitView.activeBranch]

	for rowIndex < win.Rows() && activeCommitIndex < uint(len(commits)) {
		commit := commits[rowIndex]
		author := commit.commit.Author()

		if err = win.SetRow(rowIndex, "%v %s %s", author.When, author.Name, commit.commit.Summary()); err != nil {
			break
		}

		rowIndex++
	}

	return err
}

func (commitView *CommitView) OnRefSelect(oid *Oid) (err error) {
	if _, ok := commitView.activeCommitIndex[oid]; ok {
		return
	}

	if err = commitView.repoData.LoadCommits(oid); err != nil {
		return
	}

	commitView.activeBranch = oid
	commitView.activeCommitIndex[oid] = 0
	return
}

func (commitView *CommitView) Handle(keyPressEvent KeyPressEvent) (err error) {
	return
}
