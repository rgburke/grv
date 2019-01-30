package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	log "github.com/Sirupsen/logrus"
)

// GitCommandRepoController uses git shell commands
// to update the repository
type GitCommandRepoController struct {
	repoData RepoData
	channels Channels
	config   Config
}

// NewGitCommandRepoController creates a new instance
func NewGitCommandRepoController(repoData RepoData, channels Channels, config Config) *GitCommandRepoController {
	return &GitCommandRepoController{
		repoData: repoData,
		channels: channels,
		config:   config,
	}
}

// Initialise does nothing
func (controller *GitCommandRepoController) Initialise(RepoSupplier) {

}

// CheckoutRef does a git checkout on the provided ref
func (controller *GitCommandRepoController) CheckoutRef(ref Ref, resultHandler RefOperationResultHandler) {
	go func() {
		err := controller.runGitCommand("checkout", ref.Shorthand())
		if err == nil {
			controller.repoData.LoadRefs(nil)
		}

		resultHandler(ref, err)
	}()
}

// CheckoutCommit does a git checkout on the provided commit
func (controller *GitCommandRepoController) CheckoutCommit(commit *Commit, resultHandler RepoResultHandler) {
	go func() {
		err := controller.runGitCommand("checkout", commit.oid.String())
		if err == nil {
			controller.repoData.LoadRefs(nil)
		}

		resultHandler(err)
	}()
}

// CreateBranch uses git branch to create a new branch with the provided name and oid
func (controller *GitCommandRepoController) CreateBranch(branchName string, oid *Oid) (err error) {
	if err = controller.runGitCommand("branch", branchName, oid.String()); err == nil {
		controller.repoData.LoadRefs(nil)
	}

	return
}

// CreateBranchAndCheckout uses git checkout -b to create and checkout a branch with the provided name and oid
func (controller *GitCommandRepoController) CreateBranchAndCheckout(branchName string, oid *Oid, resultHandler RefOperationResultHandler) {
	err := controller.runGitCommand("checkout", "-b", branchName, oid.String())
	if err == nil {
		controller.findRef(resultHandler, branchName, func(ref Ref) bool {
			branch, isBranch := ref.(*LocalBranch)
			return isBranch && branchName == branch.Shorthand()
		})
	}

	return
}

// CreateTag uses git tag to create a new tag with the provided name
func (controller *GitCommandRepoController) CreateTag(tagName string, oid *Oid) (err error) {
	if err = controller.runGitCommand("tag", tagName, oid.String()); err == nil {
		controller.repoData.LoadRefs(nil)
	}

	return
}

// CreateAnnotatedTag uses git tag -a to create a new annotated tag with the provided name
func (controller *GitCommandRepoController) CreateAnnotatedTag(tagName string, oid *Oid, resultHandler RefOperationResultHandler) {
	controller.runInteractiveGitCommand(func(commandErr error, exitStatus int) (err error) {
		if commandErr != nil || exitStatus != 0 {
			resultHandler(nil, fmt.Errorf("Failed to create tag. Command Status: %v, Error: %v", exitStatus, commandErr))
			return
		}

		controller.findRef(resultHandler, tagName, func(ref Ref) bool {
			tag, isTag := ref.(*Tag)
			return isTag && tag.Shorthand() == tagName
		})

		return
	}, "tag", "-a", tagName, oid.String())

	return
}

// CheckoutPreviousRef uses git checkout - to checkout the previous ref
func (controller *GitCommandRepoController) CheckoutPreviousRef(resultHandler RefOperationResultHandler) {
	go func() {
		var err error
		if err = controller.runGitCommand("checkout", "-"); err == nil {
			if err = controller.repoData.LoadHead(); err == nil {
				head := controller.repoData.Head()

				controller.findRef(resultHandler, head.Name(), func(ref Ref) bool {
					return ref.Name() == head.Name() && ref.Oid().Equal(head.Oid())
				})
			}
		}

		if err != nil {
			resultHandler(nil, fmt.Errorf("Failed to checkout previous ref: %v", err))
		}
	}()
}

// StageFiles uses git add with the provided file paths
func (controller *GitCommandRepoController) StageFiles(filePaths []string) (err error) {
	args := append([]string{"add", "--"}, filePaths...)
	if err = controller.runGitCommand(args...); err == nil {
		err = controller.repoData.LoadStatus()
	}

	return
}

// UnstageFiles does a git reset HEAD with the provided file paths
func (controller *GitCommandRepoController) UnstageFiles(filePaths []string) (err error) {
	args := append([]string{"reset", "HEAD", "--"}, filePaths...)
	if err = controller.runGitCommand(args...); err == nil {
		err = controller.repoData.LoadStatus()
	}

	return
}

// CheckoutFiles does a git checkout with the provided file paths
func (controller *GitCommandRepoController) CheckoutFiles(filePaths []string) (err error) {
	args := append([]string{"checkout", "--"}, filePaths...)
	if err = controller.runGitCommand(args...); err == nil {
		err = controller.repoData.LoadStatus()
	}

	return
}

// CommitMessageFile creates and truncates the COMMIT_EDITMSG file so that a new
// commit message file is ready to be written
func (controller *GitCommandRepoController) CommitMessageFile() (file *os.File, err error) {
	repoPath := controller.repoData.Path()
	commitMessageFilePath := fmt.Sprintf("%v/%v", repoPath, "COMMIT_EDITMSG")

	file, err = os.OpenFile(commitMessageFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		err = fmt.Errorf("Unable to open file %v for writing: %v", commitMessageFilePath, err)
	}

	return
}

// Commit uses git commit to create a new commit
func (controller *GitCommandRepoController) Commit(resultHandler CommitResultHandler) {
	controller.runInteractiveGitCommand(controller.onGitCommit(resultHandler), "commit")
}

// AmendCommit uses git commit --amend to ammend the last commit
func (controller *GitCommandRepoController) AmendCommit(resultHandler CommitResultHandler) {
	controller.runInteractiveGitCommand(controller.onGitCommit(resultHandler), "commit", "--amend")
}

func (controller *GitCommandRepoController) onGitCommit(resultHandler CommitResultHandler) func(error, int) error {
	return func(commandErr error, exitStatus int) (err error) {
		if commandErr != nil || exitStatus != 0 {
			resultHandler(nil, fmt.Errorf("Command Status: %v, Error: %v", exitStatus, commandErr))
			return
		}

		var resultError error
		var oid *Oid

		if resultError = controller.repoData.LoadHead(); resultError == nil {
			oid = controller.repoData.Head().Oid()
		}

		resultHandler(oid, resultError)
		return
	}
}

// Pull performs a git pull for the provided remote
func (controller *GitCommandRepoController) Pull(remote string, resultHandler RepoResultHandler) {
	go func() {
		resultHandler(controller.runGitCommand("pull", remote))
	}()
}

// Push performs a git push on the provided remote and ref
func (controller *GitCommandRepoController) Push(remote string, ref Ref, track bool, resultHandler RepoResultHandler) {
	go func() {
		args := []string{"push"}

		if track {
			args = append(args, "-u")
		}

		args = append(args, remote, ref.Shorthand())

		resultHandler(controller.runGitCommand(args...))
	}()
}

// DeleteLocalRef uses git branch -D and git tag -d to delete a local branch or tag respectively
func (controller *GitCommandRepoController) DeleteLocalRef(ref Ref) (err error) {
	switch ref.(type) {
	case *LocalBranch:
		return controller.runGitCommand("branch", "-D", ref.Shorthand())
	case *Tag:
		return controller.runGitCommand("tag", "-d", ref.Shorthand())
	}

	return fmt.Errorf("Invalid ref type %T", ref)
}

// DeleteRemoteRef uses git push --delete to delete a remote branch or tag
func (controller *GitCommandRepoController) DeleteRemoteRef(remote string, ref Ref, resultHandler RepoResultHandler) {
	go func() {
		var refName string

		switch rawRef := ref.(type) {
		case *RemoteBranch:
			refName = rawRef.ShorthandWithoutRemote()
		case *LocalBranch:
			refName = ref.Shorthand()
		case *Tag:
			refName = ref.Shorthand()
		default:
			resultHandler(fmt.Errorf("Invalid ref type %T", ref))
			return
		}

		resultHandler(controller.runGitCommand("push", "--delete", remote, refName))
	}()
}

// MergeRef uses git merge to merge a branch with the provided ref
func (controller *GitCommandRepoController) MergeRef(ref Ref) (err error) {
	return controller.runGitCommand("merge", "--no-edit", ref.Shorthand())
}

// Rebase uses git rebase to rebase a branch onto the provided ref
func (controller *GitCommandRepoController) Rebase(ref Ref) (err error) {
	return controller.runGitCommand("rebase", ref.Shorthand())
}

func (controller *GitCommandRepoController) findRef(resultHandler RefOperationResultHandler, refName string, refPredicate func(Ref) bool) {
	controller.repoData.LoadRefs(func(refs []Ref) error {
		for _, ref := range refs {
			if refPredicate(ref) {
				resultHandler(ref, nil)
				return nil
			}
		}

		resultHandler(nil, fmt.Errorf("Unable to find ref %v", refName))
		return nil
	})
}

func (controller *GitCommandRepoController) gitBinary() string {
	if gitBinary := controller.config.GetString(CfGitBinaryFilePath); gitBinary != "" {
		return gitBinary
	}

	return "git"
}

func (controller *GitCommandRepoController) runGitCommand(args ...string) (err error) {
	gitBinary := controller.gitBinary()
	log.Debugf("Running command: %v %v", gitBinary, strings.Join(args, " "))

	cmd := exec.Command(gitBinary, args...)
	cmd.Env, cmd.Dir = controller.repoData.GenerateGitCommandEnvironment()

	if err = cmd.Run(); err != nil {
		err = fmt.Errorf("Git command failed: %v", err)
	}

	return
}

func (controller *GitCommandRepoController) runInteractiveGitCommand(onComplete func(error, int) error, args ...string) {
	gitBinary := controller.gitBinary()
	log.Debugf("Running interactive command: %v %v", gitBinary, strings.Join(args, " "))

	controller.channels.DoAction(Action{ActionType: ActionRunCommand, Args: []interface{}{
		ActionRunCommandArgs{
			command:     gitBinary,
			args:        args,
			noShell:     true,
			interactive: true,
			stdin:       os.Stdin,
			stdout:      os.Stdout,
			stderr:      os.Stderr,
			onComplete:  onComplete,
		},
	}})
}
