package main

import (
	"errors"
	"os"
)

// CheckoutRefResultHandler is notified when the checkout is complete
type CheckoutRefResultHandler func(Ref, error)

// CheckoutCommitResultHandler is notified when the checkout is complete
type CheckoutCommitResultHandler func(err error)

// CommitResultHandler is notified when commit creation is complete
type CommitResultHandler func(oid *Oid, err error)

const (
	checkoutPercentageDiffReportThreshold = 10.0
)

var errReadOnly = errors.New("Invalid operation in read only mode")

// RepoController performs actions on a repository
// and modifies repository state
type RepoController interface {
	Initialise(RepoSupplier)
	CheckoutRef(Ref, CheckoutRefResultHandler)
	CheckoutCommit(*Commit, CheckoutCommitResultHandler)
	CreateBranch(branchName string, oid *Oid) error
	CreateBranchAndCheckout(branchName string, oid *Oid, resultHandler CheckoutRefResultHandler)
	CheckoutPreviousRef(CheckoutRefResultHandler)
	StageFiles(filePaths []string) error
	UnstageFiles(filePaths []string) error
	CheckoutFiles(filePaths []string) error
	CommitMessageFile() (*os.File, error)
	Commit(CommitResultHandler)
}

// ReadOnlyRepositoryController does not permit any
// repository modification and returns a read only error
// when such operations are attempted
type ReadOnlyRepositoryController struct{}

// NewReadOnlyRepositoryController creates a new instance
func NewReadOnlyRepositoryController() RepoController {
	return &ReadOnlyRepositoryController{}
}

// Initialise does nothing
func (repoController *ReadOnlyRepositoryController) Initialise(RepoSupplier) {}

// CheckoutRef returns a read only error
func (repoController *ReadOnlyRepositoryController) CheckoutRef(ref Ref, resultHandler CheckoutRefResultHandler) {
	go resultHandler(nil, errReadOnly)
}

// CheckoutCommit returns a read only error
func (repoController *ReadOnlyRepositoryController) CheckoutCommit(commit *Commit, resultHandler CheckoutCommitResultHandler) {
	go resultHandler(errReadOnly)
}

// CreateBranch returns a read only error
func (repoController *ReadOnlyRepositoryController) CreateBranch(string, *Oid) error {
	return errReadOnly
}

// CreateBranchAndCheckout returns a read only error
func (repoController *ReadOnlyRepositoryController) CreateBranchAndCheckout(branchName string, oid *Oid, resultHandler CheckoutRefResultHandler) {
	go resultHandler(nil, errReadOnly)
}

// CheckoutPreviousRef returns a read only error
func (repoController *ReadOnlyRepositoryController) CheckoutPreviousRef(resultHandler CheckoutRefResultHandler) {
	go resultHandler(nil, errReadOnly)
}

// StageFiles returns a read only error
func (repoController *ReadOnlyRepositoryController) StageFiles([]string) error {
	return errReadOnly
}

// UnstageFiles returns a read only error
func (repoController *ReadOnlyRepositoryController) UnstageFiles([]string) error {
	return errReadOnly
}

// CheckoutFiles returns a read only error
func (repoController *ReadOnlyRepositoryController) CheckoutFiles([]string) error {
	return errReadOnly
}

// CommitMessageFile returns a read only error
func (repoController *ReadOnlyRepositoryController) CommitMessageFile() (file *os.File, err error) {
	return file, errReadOnly
}

// Commit returns a read only error
func (repoController *ReadOnlyRepositoryController) Commit(resultHandler CommitResultHandler) {
	go resultHandler(nil, errReadOnly)
}
