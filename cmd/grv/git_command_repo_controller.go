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
func (controller *GitCommandRepoController) CheckoutRef(ref Ref, resultHandler CheckoutRefResultHandler) {
	go func() {
		err := controller.runGitCommand("checkout", ref.Shorthand())
		if err == nil {
			controller.repoData.LoadRefs(nil)
		}

		resultHandler(ref, err)
	}()
}

// CheckoutCommit does a git checkout on the provided commit
func (controller *GitCommandRepoController) CheckoutCommit(commit *Commit, resultHandler CheckoutCommitResultHandler) {
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
	err = controller.runGitCommand("branch", branchName, oid.String())
	if err == nil {
		controller.repoData.LoadRefs(nil)
	}

	return
}

// CreateBranchAndCheckout uses git checkout -b to create and checkout a branch with the provided name and oid
func (controller *GitCommandRepoController) CreateBranchAndCheckout(branchName string, oid *Oid, resultHandler CheckoutRefResultHandler) {
	err := controller.runGitCommand("checkout", "-b", branchName, oid.String())
	if err == nil {
		controller.repoData.LoadRefs(func(refs []Ref) error {
			for _, ref := range refs {
				if branchName == ref.Shorthand() {
					resultHandler(ref, nil)
					return nil
				}
			}

			resultHandler(nil, fmt.Errorf("Unable to find ref %v", branchName))
			return nil
		})
	}

	return
}

// CheckoutPreviousRef uses git checkout - to checkout the previous ref
func (controller *GitCommandRepoController) CheckoutPreviousRef(resultHandler CheckoutRefResultHandler) {
	go func() {
		var err error
		if err = controller.runGitCommand("checkout", "-"); err == nil {
			if err = controller.repoData.LoadHead(); err == nil {
				head := controller.repoData.Head()

				controller.repoData.LoadRefs(func(refs []Ref) error {
					for _, ref := range refs {
						if ref.Name() == head.Name() && ref.Oid().Equal(head.Oid()) {
							resultHandler(ref, nil)
							return nil
						}
					}

					resultHandler(nil, fmt.Errorf("Unable to find ref %v", head.Name()))
					return nil
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
func (controller *GitCommandRepoController) Commit(ref Ref, message string) (oid *Oid, err error) {
	if err = controller.runGitCommand("commit", "-m", message); err == nil {
		if err = controller.repoData.LoadHead(); err == nil {
			oid = controller.repoData.Head().Oid()
		}
	}

	return
}

func (controller *GitCommandRepoController) runGitCommand(args ...string) (err error) {
	gitBinary := controller.config.GetString(CfGitBinaryFilePath)
	if gitBinary == "" {
		gitBinary = "git"
	}

	log.Debugf("Running command: %v %v", gitBinary, strings.Join(args, " "))

	cmd := exec.Command(gitBinary, args...)
	cmd.Env = controller.repoData.GenerateGitCommandEnvironment()

	if err = cmd.Run(); err != nil {
		err = fmt.Errorf("Git command failed: %v", err)
	}

	return
}
