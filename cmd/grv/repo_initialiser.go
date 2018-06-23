package main

import (
	"fmt"
	"os"
	"path/filepath"

	log "github.com/Sirupsen/logrus"
	git "gopkg.in/libgit2/git2go.v27"
)

// RepoSupplier supplies a git2go repository instance
type RepoSupplier interface {
	RepositoryInstance() *git.Repository
}

// RepositoryInitialiser creates a git2go repository instance
type RepositoryInitialiser struct {
	repo *git.Repository
}

// NewRepositoryInitialiser creates a new instance
func NewRepositoryInitialiser() *RepositoryInitialiser {
	return &RepositoryInitialiser{}
}

// RepositoryInstance returns a git2go repository instance
func (initialiser *RepositoryInitialiser) RepositoryInstance() *git.Repository {
	return initialiser.repo
}

// Free releases the git2go repository instance
func (initialiser *RepositoryInitialiser) Free() {
	log.Info("Freeing git2go repository instance")

	if initialiser.repo != nil {
		initialiser.repo.Free()
	}
}

// CreateRepositoryInstance creates a git2go repository instance using the
// repoPath and workTreePath values provided
func (initialiser *RepositoryInitialiser) CreateRepositoryInstance(repoPath, workTreePath string) (err error) {
	repoPath, err = initialiser.processRepoPath(repoPath)
	if err != nil {
		return
	}
	workTreePath = initialiser.processWorkTreePath(workTreePath)

	log.Infof("Opening repository at %v", repoPath)

	repo, err := git.OpenRepository(repoPath)
	if err != nil {
		log.Debugf("Failed to open repository: %v", err)
		return err
	}

	if workTreePath != "" {
		if err = repo.SetWorkdir(workTreePath, false); err != nil {
			log.Debugf("Failed to set work dir: %v", err)
			return err
		}
	}

	initialiser.repo = repo

	return
}

func (initialiser *RepositoryInitialiser) processRepoPath(repoPath string) (processedRepoPath string, err error) {
	if gitDir, gitDirSet := os.LookupEnv("GIT_DIR"); gitDirSet {
		return gitDir, nil
	}

	path, err := CanonicalPath(repoPath)
	if err != nil {
		return
	}

	for {
		gitDirPath := filepath.Join(path, GitRepositoryDirectoryName)
		log.Debugf("gitDirPath: %v", gitDirPath)

		if _, err = os.Stat(gitDirPath); err != nil {
			if !os.IsNotExist(err) {
				break
			}
		} else {
			processedRepoPath = gitDirPath
			break
		}

		if path == "/" {
			err = fmt.Errorf("Unable to find a git repository in %v or any of its parent directories", repoPath)
			break
		}

		path = filepath.Dir(path)
	}

	return
}

func (initialiser *RepositoryInitialiser) processWorkTreePath(workTreePath string) string {
	if gitWorkTree, gitWorkTreeSet := os.LookupEnv("GIT_WORK_TREE"); gitWorkTreeSet {
		return gitWorkTree
	}

	return workTreePath
}
