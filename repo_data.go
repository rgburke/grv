package main

import (
	"github.com/bradfitz/slice"
)

type RepoData interface {
	LoadHead() error
	LoadLocalRefs() error
	LoadCommits(*Oid) error
	Head() *Oid
	LocalBranches() []*Branch
	LocalTags() []*Tag
	Commits(*Oid) []*Commit
}

type RepositoryData struct {
	repoDataLoader    *RepoDataLoader
	head              *Oid
	localBranches     map[*Oid]*Branch
	localBranchesList []*Branch
	localTags         map[*Oid]*Tag
	localTagsList     []*Tag
	commits           map[*Oid][]*Commit
}

func NewRepositoryData(repoDataLoader *RepoDataLoader) *RepositoryData {
	return &RepositoryData{
		repoDataLoader: repoDataLoader,
		localBranches:  make(map[*Oid]*Branch),
		localTags:      make(map[*Oid]*Tag),
		commits:        make(map[*Oid][]*Commit),
	}
}

func (repoData *RepositoryData) Free() {
	repoData.repoDataLoader.Free()
}

func (repoData *RepositoryData) Initialise(repoPath string) (err error) {
	if err = repoData.repoDataLoader.Initialise(repoPath); err != nil {
		return
	}

	return
}

func (repoData *RepositoryData) LoadHead() (err error) {
	head, err := repoData.repoDataLoader.Head()
	if err != nil {
		return
	}

	repoData.head = head

	return
}

func (repoData *RepositoryData) LoadLocalRefs() (err error) {
	branches, tags, err := repoData.repoDataLoader.LocalRefs()
	if err != nil {
		return
	}

	for _, branch := range branches {
		repoData.localBranches[branch.oid] = branch
		repoData.localBranchesList = append(repoData.localBranchesList, branch)
	}

	slice.Sort(repoData.localBranchesList, func(i, j int) bool {
		return repoData.localBranchesList[i].name < repoData.localBranchesList[j].name
	})

	for _, tag := range tags {
		repoData.localTags[tag.oid] = tag
		repoData.localTagsList = append(repoData.localTagsList, tag)
	}

	slice.Sort(repoData.localTagsList, func(i, j int) bool {
		return repoData.localTagsList[i].tag.Name() < repoData.localTagsList[j].tag.Name()
	})

	return
}

func (repoData *RepositoryData) LoadCommits(oid *Oid) error {
	if _, ok := repoData.commits[oid]; ok {
		return nil
	}

	commits, err := repoData.repoDataLoader.Commits(oid)
	if err == nil {
		repoData.commits[oid] = commits
	}

	return err
}

func (repoData *RepositoryData) Head() *Oid {
	return repoData.head
}

func (repoData *RepositoryData) LocalBranches() []*Branch {
	return repoData.localBranchesList
}

func (repoData *RepositoryData) LocalTags() []*Tag {
	return repoData.localTagsList
}

func (repoData *RepositoryData) Commits(oid *Oid) []*Commit {
	return repoData.commits[oid]
}
