package main

import (
	"errors"
	"fmt"
	"sync"

	log "github.com/Sirupsen/logrus"
	git "gopkg.in/libgit2/git2go.v26"
)

// CheckoutRefResultHandler is notified when the checkout is complete
type CheckoutRefResultHandler func(Ref, error)

// CheckoutCommitResultHandler is notified when the checkout is complete
type CheckoutCommitResultHandler func(err error)

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
	StageFiles(filePaths []string) error
	UnstageFiles(filePaths []string) error
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
func (repoController *ReadOnlyRepositoryController) CreateBranch(branchName string, oid *Oid) (err error) {
	return errReadOnly
}

// StageFiles returns a read only error
func (repoController *ReadOnlyRepositoryController) StageFiles(filePaths []string) (err error) {
	return errReadOnly
}

// UnstageFiles returns a read only error
func (repoController *ReadOnlyRepositoryController) UnstageFiles(filePaths []string) (err error) {
	return errReadOnly
}

// RepositoryController implements the RepoController interface
type RepositoryController struct {
	repo     *git.Repository
	repoData RepoData
	channels Channels
	lock     sync.Mutex
}

// NewRepoController creates a new instance
func NewRepoController(repoData RepoData, channels Channels) RepoController {
	return &RepositoryController{
		repoData: repoData,
		channels: channels,
	}
}

// Initialise performs setup
func (repoController *RepositoryController) Initialise(repoSupplier RepoSupplier) {
	repoController.repo = repoSupplier.RepositoryInstance()
}

// CheckoutCommit checks out the provided commit and makes HEAD detached at the commit oid
func (repoController *RepositoryController) CheckoutCommit(commit *Commit, resultHandler CheckoutCommitResultHandler) {
	go func() {
		repoController.lock.Lock()
		defer repoController.lock.Unlock()

		err := repoController.checkoutTree(commit)
		if err != nil {
			resultHandler(err)
			return
		}

		if err = repoController.repo.SetHeadDetached(commit.oid.oid); err != nil {
			err = fmt.Errorf("Unable to update HEAD: %v", err)
		}

		repoController.repoData.LoadRefs(func([]Ref) (err error) {
			resultHandler(err)
			return
		})
	}()
}

// CheckoutRef checks out the provided ref and sets HEAD equal to the ref
func (repoController *RepositoryController) CheckoutRef(ref Ref, resultHandler CheckoutRefResultHandler) {
	go func() {
		repoController.lock.Lock()
		defer repoController.lock.Unlock()

		refName, oid, err := repoController.checkoutRef(ref)
		if err != nil {
			resultHandler(nil, err)
			return
		}

		repoController.repoData.LoadRefs(func(refs []Ref) (err error) {
			head := repoController.repoData.Head()
			if head, isDetached := head.(*HEAD); isDetached && head.Oid().Equal(oid) {
				resultHandler(head, nil)
				return
			}

			for _, ref := range refs {
				if ref.Name() == refName {
					resultHandler(ref, nil)
					return
				}
			}

			resultHandler(nil, fmt.Errorf("Unable to find checked out ref %v", refName))
			return
		})
	}()
}

func (repoController *RepositoryController) checkoutRef(ref Ref) (refName string, oid *Oid, err error) {
	oid = ref.Oid()
	refName = ref.Name()

	switch refInstance := ref.(type) {
	case *Tag:
	case *LocalBranch:
	case *RemoteBranch:
		localBranch := repoController.localBranch(refInstance)
		if localBranch == nil {
			log.Debugf("No local branch exists for %v, creating local tracking branch", refName)

			var newBranch *git.Branch
			newBranch, err = repoController.createBranch(refInstance.ShorthandWithoutRemote(), oid)
			if err != nil {
				err = fmt.Errorf("Checkout failed - %v", err)
				return
			}

			if err = newBranch.SetUpstream(refInstance.Shorthand()); err != nil {
				err = fmt.Errorf("Checkout failed - Unable to set upstream for branch %v: %v", refInstance.ShorthandWithoutRemote(), err)
				return
			}

			log.Debugf("Updated branch %v to track %v", refInstance.ShorthandWithoutRemote(), refName)

			refName = newBranch.Reference.Name()
		} else {
			oid = localBranch.Oid()
			refName = localBranch.Name()
		}

		log.Debugf("Checking out local branch %v", refName)
	default:
		return
	}

	log.Debugf("Checking out ref %v with oid %v", refName, oid)

	commit, err := repoController.repoData.Commit(oid)
	if err != nil {
		err = fmt.Errorf("Checkout failed - Unable to load commit with oid %v: %v", oid, err)
		return
	}

	if err = repoController.checkoutTree(commit); err != nil {
		return
	}

	if err = repoController.repo.SetHead(refName); err != nil {
		err = fmt.Errorf("Checkout failed - Unable to update HEAD: %v", err)
	}

	log.Info("Checked out %v", refName)

	return
}

func (repoController *RepositoryController) checkoutTree(commit *Commit) (err error) {
	tree, err := commit.commit.Tree()
	if err != nil {
		err = fmt.Errorf("Checkout failed - Unable to load tree for commit with oid %v: %v", commit.oid, err)
		return
	}
	defer tree.Free()

	lastReportedCheckoutPercentage := 0.0

	checkoutOpts := &git.CheckoutOpts{
		Strategy: git.CheckoutSafe | git.CheckoutRecreateMissing,
		ProgressCallback: func(path string, completed, total uint) git.ErrorCode {
			percentageComplete := (float64(completed) * 100.0) / float64(total)

			if percentageComplete-lastReportedCheckoutPercentage > checkoutPercentageDiffReportThreshold {
				repoController.channels.ReportStatus("Checkout %v%% complete...", uint(percentageComplete))
				lastReportedCheckoutPercentage = percentageComplete
			}

			return git.ErrOk
		},
	}

	if err = repoController.repo.CheckoutTree(tree, checkoutOpts); err != nil {
		err = fmt.Errorf("Checkout failed - Unable to checkout to commit %v: %v", commit.oid, err)
		return
	}

	repoController.channels.ReportStatus("Checkout complete")

	return
}

func (repoController *RepositoryController) localBranch(remoteBranch *RemoteBranch) (localBranch *LocalBranch) {
	localBranches := repoController.repoData.LocalBranches(remoteBranch)
	if len(localBranches) == 0 {
		return
	}

	localBranchName := remoteBranch.ShorthandWithoutRemote()

	for _, localBranch := range localBranches {
		if localBranch.Shorthand() == localBranchName {
			return localBranch
		}
	}

	return localBranches[0]
}

// CreateBranch creates a new local branch with the specified name pointing to the provided oid
func (repoController *RepositoryController) CreateBranch(branchName string, oid *Oid) (err error) {
	repoController.lock.Lock()
	defer repoController.lock.Unlock()

	_, err = repoController.createBranch(branchName, oid)

	return
}

func (repoController *RepositoryController) createBranch(branchName string, oid *Oid) (branch *git.Branch, err error) {
	commit, err := repoController.repoData.Commit(oid)
	if err != nil {
		err = fmt.Errorf("Create branch failed - Unable to load commit with oid %v: %v", oid, err)
		return
	}

	branch, err = repoController.repo.CreateBranch(branchName, commit.commit, false)
	if err != nil {
		err = fmt.Errorf("Create branch failed - Unable to create branch %v: %v", branchName, err)
		return
	}

	log.Info("Created local branch %v", branchName)

	return
}

// StageFiles stages the specified filea
func (repoController *RepositoryController) StageFiles(filePaths []string) (err error) {
	index, err := repoController.repo.Index()
	if err != nil {
		return fmt.Errorf("Unable to stage file: %v", err)
	}

	if err = index.AddAll(filePaths, git.IndexAddDefault, nil); err != nil {
		return fmt.Errorf("Unable to stage files: %v", err)
	}

	if err = index.Write(); err != nil {
		return fmt.Errorf("Unable to stage file: %v", err)
	}

	go repoController.repoData.LoadStatus()

	return
}

// UnstageFiles removes files from the staged area
func (repoController *RepositoryController) UnstageFiles(filePaths []string) (err error) {
	head := repoController.repoData.Head()
	commit, err := repoController.repoData.Commit(head.Oid())
	if err != nil {
		return fmt.Errorf("Unable to unstage file: %v", err)
	}

	index, err := repoController.repo.Index()
	if err != nil {
		return fmt.Errorf("Unable to unstage file: %v", err)
	}

	if err = repoController.repo.ResetDefaultToCommit(commit.commit, filePaths); err != nil {
		return fmt.Errorf("Unable to unstage file: %v", err)
	}

	if err = index.Write(); err != nil {
		return fmt.Errorf("Unable to unstage file: %v", err)
	}

	go repoController.repoData.LoadStatus()

	return
}
