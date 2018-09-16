package main

import (
	"errors"
	"os"
)

// RefOperationResultHandler is notified when the checkout is complete
type RefOperationResultHandler func(Ref, error)

// RepoResultHandler is notified when a repo operation completes
type RepoResultHandler func(err error)

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
	CheckoutRef(Ref, RefOperationResultHandler)
	CheckoutCommit(*Commit, RepoResultHandler)
	CreateBranch(branchName string, oid *Oid) error
	CreateBranchAndCheckout(branchName string, oid *Oid, resultHandler RefOperationResultHandler)
	CreateTag(tagName string, oid *Oid) error
	CreateAnnotatedTag(tagName string, oid *Oid, resultHandler RefOperationResultHandler)
	CheckoutPreviousRef(RefOperationResultHandler)
	StageFiles(filePaths []string) error
	UnstageFiles(filePaths []string) error
	CheckoutFiles(filePaths []string) error
	CommitMessageFile() (*os.File, error)
	Commit(CommitResultHandler)
	AmendCommit(CommitResultHandler)
	Pull(remote string, resultHandler RepoResultHandler)
	Push(remote string, ref Ref, track bool, resultHandler RepoResultHandler)
	DeleteLocalRef(ref Ref) error
	DeleteRemoteRef(remote string, ref Ref, resultHandler RepoResultHandler)
	MergeRef(Ref) error
	Rebase(Ref) error
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
func (repoController *ReadOnlyRepositoryController) CheckoutRef(ref Ref, resultHandler RefOperationResultHandler) {
	go resultHandler(nil, errReadOnly)
}

// CheckoutCommit returns a read only error
func (repoController *ReadOnlyRepositoryController) CheckoutCommit(commit *Commit, resultHandler RepoResultHandler) {
	go resultHandler(errReadOnly)
}

// CreateBranch returns a read only error
func (repoController *ReadOnlyRepositoryController) CreateBranch(string, *Oid) error {
	return errReadOnly
}

// CreateBranchAndCheckout returns a read only error
func (repoController *ReadOnlyRepositoryController) CreateBranchAndCheckout(branchName string, oid *Oid, resultHandler RefOperationResultHandler) {
	go resultHandler(nil, errReadOnly)
}

// CreateTag returns a read only error
func (repoController *ReadOnlyRepositoryController) CreateTag(string, *Oid) error {
	return errReadOnly
}

// CreateAnnotatedTag returns a read only error
func (repoController *ReadOnlyRepositoryController) CreateAnnotatedTag(tagName string, oid *Oid, resultHandler RefOperationResultHandler) {
	go resultHandler(nil, errReadOnly)
}

// CheckoutPreviousRef returns a read only error
func (repoController *ReadOnlyRepositoryController) CheckoutPreviousRef(resultHandler RefOperationResultHandler) {
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

// AmendCommit returns a read only error
func (repoController *ReadOnlyRepositoryController) AmendCommit(resultHandler CommitResultHandler) {
	go resultHandler(nil, errReadOnly)
}

// Pull returns a read only error
func (repoController *ReadOnlyRepositoryController) Pull(remote string, resultHandler RepoResultHandler) {
	go resultHandler(errReadOnly)
}

// Push returns a read only error
func (repoController *ReadOnlyRepositoryController) Push(remote string, ref Ref, track bool, resultHandler RepoResultHandler) {
	go resultHandler(errReadOnly)
}

// DeleteLocalRef returns a read only error
func (repoController *ReadOnlyRepositoryController) DeleteLocalRef(ref Ref) error {
	return errReadOnly
}

// DeleteRemoteRef returns a read only error
func (repoController *ReadOnlyRepositoryController) DeleteRemoteRef(remote string, ref Ref, resultHandler RepoResultHandler) {
	go resultHandler(errReadOnly)
}

// MergeRef returns a read only error
func (repoController *ReadOnlyRepositoryController) MergeRef(Ref) error {
	return errReadOnly
}

// Rebase returns a read only error
func (repoController *ReadOnlyRepositoryController) Rebase(Ref) error {
	return errReadOnly
}
