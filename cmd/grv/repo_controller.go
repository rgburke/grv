package main

import (
	git "gopkg.in/libgit2/git2go.v26"
)

// RepoController performs actions on a repository
// and modifies repository state
type RepoController interface{}

// RepositoryController implements the RepoController interface
type RepositoryController struct {
	repo *git.Repository
}

// NewRepoController creates a new instance
func NewRepoController() *RepositoryController {
	return &RepositoryController{}
}

// Initialise performs setup
func (repoController *RepositoryController) Initialise(repoSupplier RepoSupplier) {
	repoController.repo = repoSupplier.RepositoryInstance()
}
