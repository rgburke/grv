package main

import (
	log "github.com/Sirupsen/logrus"
)

type CommitView struct {
	repoData          RepoData
	activeBranch      *Oid
	activeCommitIndex map[*Oid]uint
	active            bool
}

func NewCommitView(repoData RepoData) *CommitView {
	return &CommitView{
		repoData:          repoData,
		activeCommitIndex: make(map[*Oid]uint),
	}
}

func (commitView *CommitView) Initialise() (err error) {
	log.Info("Initialising CommitView")
	return
}

func (commitView *CommitView) Render(win RenderWindow) (err error) {
	log.Debug("Rendering CommitView")
	rowIndex := uint(1)
	commitIndex := 0
	commits := commitView.repoData.Commits(commitView.activeBranch)

	for rowIndex < win.Rows() && commitIndex < len(commits) {
		commit := commits[commitIndex]
		author := commit.commit.Author()

		if err = win.SetRow(rowIndex, "%v %s %s", author.When, author.Name, commit.commit.Summary()); err != nil {
			break
		}

		commitIndex++
		rowIndex++
	}

	return err
}

func (commitView *CommitView) OnRefSelect(oid *Oid) (err error) {
	log.Debugf("CommitView loading commits for selected oid %v", oid)

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

func (commitView *CommitView) Handle(keyPressEvent KeyPressEvent, channels HandlerChannels) (err error) {
	log.Debugf("CommitView handling key %v", keyPressEvent)
	return
}

func (commitView *CommitView) OnActiveChange(active bool) {
	log.Debugf("CommitView active %v", active)
	commitView.active = active
}
