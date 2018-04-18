package main

import (
	"fmt"
	"sync"

	log "github.com/Sirupsen/logrus"
	git "gopkg.in/libgit2/git2go.v26"
)

// CheckoutRefResultHandler is notified when checkout is complete
type CheckoutRefResultHandler func(Ref, error)

const (
	checkoutPercentageDiffReportThreshold = 10.0
)

// RepoController performs actions on a repository
// and modifies repository state
type RepoController interface {
	CheckoutRef(Ref, CheckoutRefResultHandler)
}

// RepositoryController implements the RepoController interface
type RepositoryController struct {
	repo     *git.Repository
	repoData RepoData
	channels *Channels
	lock     sync.Mutex
}

// NewRepoController creates a new instance
func NewRepoController(repoData RepoData, channels *Channels) *RepositoryController {
	return &RepositoryController{
		repoData: repoData,
		channels: channels,
	}
}

// Initialise performs setup
func (repoController *RepositoryController) Initialise(repoSupplier RepoSupplier) {
	repoController.repo = repoSupplier.RepositoryInstance()
}

// CheckoutRef checks out ref and sets HEAD equal to the ref
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

			var commit *Commit
			commit, err = repoController.repoData.Commit(oid)
			if err != nil {
				err = fmt.Errorf("Checkout failed - Unable to load commit with oid %v: %v", oid, err)
				return
			}

			var newBranch *git.Branch
			newBranch, err = repoController.repo.CreateBranch(refInstance.ShorthandWithoutRemote(), commit.commit, false)
			if err != nil {
				err = fmt.Errorf("Checkout failed - Unable to create branch %v: %v", refInstance.ShorthandWithoutRemote(), err)
				return
			}

			log.Debugf("Created local branch %v to track remote branch %v", newBranch.Reference.Name(), refName)

			if err = newBranch.SetUpstream(refInstance.Shorthand()); err != nil {
				err = fmt.Errorf("Checkout failed - Unable to set upstream for branch %v: %v", refInstance.ShorthandWithoutRemote(), err)
				return
			}

			oid = commit.oid
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

	tree, err := commit.commit.Tree()
	if err != nil {
		err = fmt.Errorf("Checkout failed - Unable to load tree for commit with oid %v: %v", oid, err)
		return
	}
	defer tree.Free()

	lastReportedCheckoutPercentage := 0.0

	checkoutOpts := &git.CheckoutOpts{
		Strategy: git.CheckoutSafe | git.CheckoutRecreateMissing | git.CheckoutAllowConflicts | git.CheckoutUseTheirs,
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
		err = fmt.Errorf("Checkout failed - Unable to checkout ref %v: %v", refName, err)
		return
	}

	repoController.channels.ReportStatus("Checkout complete")

	if err = repoController.repo.SetHead(refName); err != nil {
		err = fmt.Errorf("Checkout failed - Unable to update HEAD: %v", err)
	}

	log.Info("Checked out %v", refName)

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
