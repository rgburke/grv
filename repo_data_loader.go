package main

import (
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

func NewRepoDataLoader() *RepoDataLoader {
	return &RepoDataLoader{}
}

func (repoDataLoader *RepoDataLoader) Free() {
	repoDataLoader.repo.Free()
}

func (repoDataLoader *RepoDataLoader) Initialise(repoPath string) (err error) {
	repo, err := git.OpenRepository(repoPath)
	if err == nil {
		repoDataLoader.repo = repo
	}

	return
}

func (repoDataLoader *RepoDataLoader) Head() (oid *Oid, err error) {
	ref, err := repoDataLoader.repo.Head()
	if err != nil {
		return
	}

	oid = &Oid{ref.Target()}

	return
}

func (repoDataLoader *RepoDataLoader) LocalRefs() (branches []*Branch, tags []*Tag, err error) {
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

				branches = append(branches, &Branch{oid, branchName})
			} else if ref.IsTag() {
				tag, err := repoDataLoader.repo.LookupTag(ref.Target())
				if err != nil {
					break
				}

				oid := &Oid{tag.TargetId()}

				tags = append(tags, &Tag{oid, tag})
			}
		}
	}

	return
}

func (repoDataLoader *RepoDataLoader) Commits(oid *Oid) (commits []*Commit, err error) {
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

	return
}
