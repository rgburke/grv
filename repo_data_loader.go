package main

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/libgit2/git2go"
)

type RepoDataLoader struct {
	repo *git.Repository
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

func NewRepoDataLoader() *RepoDataLoader {
	return &RepoDataLoader{}
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

	oid = &Oid{ref.Target()}
	log.Debugf("Loaded HEAD %v", oid)

	return
}

func (repoDataLoader *RepoDataLoader) LocalRefs() (branches []*Branch, tags []*Tag, err error) {
	log.Debug("Loading local refs")

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

		if !ref.IsRemote() {
			if ref.IsBranch() {
				branch := ref.Branch()
				oid := &Oid{branch.Target()}
				branchName, err := branch.Name()
				branch.Free()

				if err != nil {
					break
				}

				newBranch := &Branch{oid, branchName}
				branches = append(branches, newBranch)

				log.Debugf("Loaded branch %v", newBranch)
			} else if ref.IsTag() {
				tag, err := repoDataLoader.repo.LookupTag(ref.Target())
				if err != nil {
					break
				}

				oid := &Oid{tag.TargetId()}

				newTag := &Tag{oid, tag}
				tags = append(tags, newTag)

				log.Debugf("Loaded tag %v", newTag)
			}
		}
	}

	return
}

func (repoDataLoader *RepoDataLoader) Commits(oid *Oid) (commits []*Commit, err error) {
	log.Debugf("Loading commits for oid %v", oid)

	revWalk, err := repoDataLoader.repo.Walk()
	if err != nil {
		return
	}
	defer revWalk.Free()

	revWalk.Sorting(git.SortTime)
	revWalk.Push(oid.oid)

	revWalk.Iterate(func(commit *git.Commit) bool {
		commits = append(commits, &Commit{commit})
		return true
	})

	log.Debugf("Loaded %v commits", len(commits))

	return
}
