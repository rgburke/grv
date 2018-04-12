package main

import (
	"fmt"
	"os"
	"path/filepath"

	log "github.com/Sirupsen/logrus"
	git "gopkg.in/libgit2/git2go.v26"
)

// RepoSupplier supplies a git2go repository instance
type RepoSupplier interface {
	RepositoryInstance() *git.Repository
}

// RepositoryInitialiser creates a git2go repository instance
type RepositoryInitialiser struct {
	repoPath     string
	workTreePath string
	repo         *git.Repository
}

// NewRepositoryInitialiser creates a new instance
func NewRepositoryInitialiser(repoPath, workTreePath string) *RepositoryInitialiser {
	return &RepositoryInitialiser{
		repoPath:     repoPath,
		workTreePath: workTreePath,
	}
}

// RepositoryInstance returns a git2go repository instance
func (initialiser *RepositoryInitialiser) RepositoryInstance() *git.Repository {
	return initialiser.repo
}

// CreateRepositoryInstance creates a git2go repository instance using the
// repoPath and workTreePath values it was created with
func (initialiser *RepositoryInitialiser) CreateRepositoryInstance() (err error) {
	repoPath, err := initialiser.processRepoPath(initialiser.repoPath)
	if err != nil {
		return
	}
	workTreePath := initialiser.processWorkTreePath(initialiser.workTreePath)

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
