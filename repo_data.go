package main

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/bradfitz/slice"
	"sync"
)

type RepoData interface {
	LoadHead() error
	LoadLocalRefs() error
	LoadCommits(*Oid) error
	Head() *Oid
	LocalBranches() []*Branch
	LocalTags() []*Tag
	CommitSetState(*Oid) CommitSetState
	Commits(oid *Oid, startIndex, count uint) (<-chan *Commit, error)
}

type CommitSet struct {
	commits []*Commit
	loading bool
	lock    sync.Mutex
}

type CommitSetState struct {
	loading   bool
	commitNum uint
}

type RepositoryData struct {
	repoDataLoader    *RepoDataLoader
	head              *Oid
	localBranches     map[*Oid]*Branch
	localBranchesList []*Branch
	localTags         map[*Oid]*Tag
	localTagsList     []*Tag
	commits           map[*Oid]*CommitSet
}

func NewRepositoryData(repoDataLoader *RepoDataLoader) *RepositoryData {
	return &RepositoryData{
		repoDataLoader: repoDataLoader,
		localBranches:  make(map[*Oid]*Branch),
		localTags:      make(map[*Oid]*Tag),
		commits:        make(map[*Oid]*CommitSet),
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

func (repoData *RepositoryData) LoadCommits(oid *Oid) (err error) {
	if _, ok := repoData.commits[oid]; ok {
		log.Debugf("Commits already loaded for oid %v", oid)
		return
	}

	commitCh, err := repoData.repoDataLoader.Commits(oid)
	if err != nil {
		return
	}

	commitSet := &CommitSet{
		loading: true,
		commits: make([]*Commit, 0),
	}
	repoData.commits[oid] = commitSet

	go func() {
		log.Debugf("Receiving commits from RepoDataLoader for oid %v", oid)

		for commit := range commitCh {
			commitSet.lock.Lock()
			commitSet.commits = append(commitSet.commits, commit)
			commitSet.lock.Unlock()
		}

		commitSet.loading = false
		log.Debugf("Finished loading commits for oid %v", oid)
	}()

	return
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

func (repoData *RepositoryData) CommitSetState(oid *Oid) CommitSetState {
	var commitSetState CommitSetState

	if commitSet, ok := repoData.commits[oid]; !ok {
		commitSetState = CommitSetState{
			loading:   false,
			commitNum: 0,
		}
	} else {
		commitSet.lock.Lock()
		commitSetState = CommitSetState{
			loading:   commitSet.loading,
			commitNum: uint(len(commitSet.commits)),
		}
		commitSet.lock.Unlock()
	}

	return commitSetState
}

func (repoData *RepositoryData) Commits(oid *Oid, startIndex, count uint) (<-chan *Commit, error) {
	var commitSet *CommitSet
	var ok bool

	if commitSet, ok = repoData.commits[oid]; !ok {
		return nil, fmt.Errorf("No commits loaded for oid %v", oid)
	}

	commitCh := make(chan *Commit)

	go func() {
		defer close(commitCh)
		var commit *Commit
		index := startIndex

		for {
			commit = nil
			commitSet.lock.Lock()

			if index < uint(len(commitSet.commits)) && index-startIndex < count {
				commit = commitSet.commits[index]
				index++
			}

			commitSet.lock.Unlock()

			if commit != nil {
				commitCh <- commit
			} else {
				return
			}
		}
	}()

	return commitCh, nil
}
