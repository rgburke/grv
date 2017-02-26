package main

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/libgit2/git2go"
	"sync"
)

const (
	RDL_COMMIT_BUFFER_SIZE = 100
)

type InstanceCache struct {
	oids       map[string]*Oid
	commits    map[string]*Commit
	oidLock    sync.Mutex
	commitLock sync.Mutex
}

type RepoDataLoader struct {
	repo  *git.Repository
	cache *InstanceCache
}

type Oid struct {
	oid *git.Oid
}

type Branch struct {
	oid  *Oid
	name string
}

type Tag struct {
	oid *Oid
	tag *git.Tag
}

type Commit struct {
	commit *git.Commit
}

func (oid Oid) String() string {
	return oid.oid.String()
}

func (branch Branch) String() string {
	return fmt.Sprintf("%v:%v", branch.name, branch.oid)
}

func (tag Tag) String() string {
	return fmt.Sprintf("%v:%v", tag.tag.Name(), tag.oid)
}

func NewInstanceCache() *InstanceCache {
	return &InstanceCache{
		oids:    make(map[string]*Oid),
		commits: make(map[string]*Commit),
	}
}

func (cache *InstanceCache) getOid(rawOid *git.Oid) *Oid {
	cache.oidLock.Lock()
	defer cache.oidLock.Unlock()

	oidStr := rawOid.String()

	if oid, ok := cache.oids[oidStr]; ok {
		return oid
	}

	oid := &Oid{oid: rawOid}
	cache.oids[oidStr] = oid

	return oid
}

func (cache *InstanceCache) getCommit(rawCommit *git.Commit) *Commit {
	cache.commitLock.Lock()
	defer cache.commitLock.Unlock()

	oidStr := rawCommit.Id().String()

	if commit, ok := cache.commits[oidStr]; ok {
		return commit
	}

	commit := &Commit{commit: rawCommit}
	cache.commits[oidStr] = commit

	return commit
}

func NewRepoDataLoader() *RepoDataLoader {
	return &RepoDataLoader{
		cache: NewInstanceCache(),
	}
}

func (repoDataLoader *RepoDataLoader) Free() {
	log.Info("Freeing RepoDataLoader")
	repoDataLoader.repo.Free()
}

func (repoDataLoader *RepoDataLoader) Initialise(repoPath string) (err error) {
	log.Infof("Opening repository at %v", repoPath)

	repo, err := git.OpenRepository(repoPath)
	if err == nil {
		repoDataLoader.repo = repo
	}

	return
}

func (repoDataLoader *RepoDataLoader) Head() (oid *Oid, err error) {
	log.Debug("Loading HEAD")
	ref, err := repoDataLoader.repo.Head()
	if err != nil {
		return
	}

	oid = repoDataLoader.cache.getOid(ref.Target())
	log.Debugf("Loaded HEAD %v", oid)

	return
}

func (repoDataLoader *RepoDataLoader) LoadLocalBranches() (branches []*Branch, err error) {
	log.Debug("Loading local branches")
	return repoDataLoader.LoadBranches(git.BranchLocal)
}

func (repoDataLoader *RepoDataLoader) LoadBranches(branchType git.BranchType) (branches []*Branch, err error) {
	branchIter, err := repoDataLoader.repo.NewBranchIterator(branchType)
	if err != nil {
		return
	}
	defer branchIter.Free()

	err = branchIter.ForEach(func(branch *git.Branch, branchType git.BranchType) error {
		branchName, err := branch.Name()
		if err != nil {
			return err
		}
		oid := repoDataLoader.cache.getOid(branch.Target())

		newBranch := &Branch{oid, branchName}
		branches = append(branches, newBranch)
		log.Debugf("Loaded branch %v", newBranch)

		return nil
	})

	return
}

func (repoDataLoader *RepoDataLoader) LocalTags() (tags []*Tag, err error) {
	log.Debug("Loading local tags")

	refIter, err := repoDataLoader.repo.NewReferenceIterator()
	if err != nil {
		return
	}
	defer refIter.Free()

	for {
		ref, err := refIter.Next()
		if err != nil {
			break
		}

		if !ref.IsRemote() && ref.IsTag() {
			tag, err := repoDataLoader.repo.LookupTag(ref.Target())
			if err != nil {
				break
			}

			oid := repoDataLoader.cache.getOid(tag.TargetId())

			newTag := &Tag{oid, tag}
			tags = append(tags, newTag)

			log.Debugf("Loaded tag %v", newTag)
		}
	}

	return
}

func (repoDataLoader *RepoDataLoader) Commits(oid *Oid) (<-chan *Commit, error) {
	log.Debugf("Loading commits for oid %v", oid)

	revWalk, err := repoDataLoader.repo.Walk()
	if err != nil {
		return nil, err
	}

	revWalk.Sorting(git.SortTime)
	revWalk.Push(oid.oid)

	commitCh := make(chan *Commit, RDL_COMMIT_BUFFER_SIZE)

	go func() {
		commitNum := 0
		revWalk.Iterate(func(commit *git.Commit) bool {
			commitNum++
			commitCh <- repoDataLoader.cache.getCommit(commit)
			return true
		})

		close(commitCh)
		revWalk.Free()
		log.Debugf("Loaded %v commits for oid %v", commitNum, oid)
	}()

	return commitCh, nil
}
