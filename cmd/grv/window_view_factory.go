package main

import (
	"fmt"
)

// WindowViewFactory provides a generic interface
// for creating view instances
type WindowViewFactory struct {
	repoData RepoData
	channels *Channels
	config   Config
}

// NewWindowViewFactory creates a new instance
func NewWindowViewFactory(repoData RepoData, channels *Channels, config Config) *WindowViewFactory {
	return &WindowViewFactory{
		repoData: repoData,
		channels: channels,
		config:   config,
	}
}

// CreateWindowView creates a window view instance identified by the provided view id
func (windowViewFactory *WindowViewFactory) CreateWindowView(viewID ViewID) (WindowView, error) {
	return windowViewFactory.CreateWindowViewWithArgs(viewID, nil)
}

// CreateWindowViewWithArgs creates a window view instance identified by the provided view id and instantiated with the provided args
func (windowViewFactory *WindowViewFactory) CreateWindowViewWithArgs(viewID ViewID, args []interface{}) (windowView WindowView, err error) {
	switch viewID {
	case ViewRef:
		windowView = windowViewFactory.createRefView()
	case ViewCommit:
		windowView, err = windowViewFactory.createCommitView(args)
	case ViewDiff:
		windowView, err = windowViewFactory.createDiffView(args)
	case ViewGitStatus:
		windowView = windowViewFactory.createGitStatusView()
	default:
		err = fmt.Errorf("Unsupported view type: %v", viewID)
	}

	return
}

func (windowViewFactory *WindowViewFactory) createRefView() *RefView {
	return NewRefView(windowViewFactory.repoData, windowViewFactory.channels)
}

func (windowViewFactory *WindowViewFactory) createCommitView(args []interface{}) (commitView *CommitView, err error) {
	ref, err := windowViewFactory.getRef(args)
	if err != nil {
		return
	}

	commitView = NewCommitView(windowViewFactory.repoData, windowViewFactory.channels)

	if ref != nil {
		err = commitView.OnRefSelect(ref)
	}

	return
}

func (windowViewFactory *WindowViewFactory) createDiffView(args []interface{}) (diffView *DiffView, err error) {
	ref, err := windowViewFactory.getRef(args)
	if err != nil {
		return
	}

	diffView = NewDiffView(windowViewFactory.repoData, windowViewFactory.channels)

	if ref != nil {
		var commit *Commit
		commit, err = windowViewFactory.repoData.Commit(ref.Oid())

		if err == nil {
			err = diffView.OnCommitSelected(commit)
		}
	}

	return
}

func (windowViewFactory *WindowViewFactory) createGitStatusView() *GitStatusView {
	gitStatusView := NewGitStatusView(windowViewFactory.repoData, windowViewFactory.channels)

	status := windowViewFactory.repoData.Status()
	gitStatusView.OnStatusChanged(status)

	return gitStatusView
}

func (windowViewFactory *WindowViewFactory) getRef(args []interface{}) (ref Ref, err error) {
	if len(args) == 0 {
		return
	}

	refName, ok := args[0].(string)
	if !ok {
		err = fmt.Errorf("Expected refName argument of type string but got type %T", args[0])
		return
	}

	if ref, err = windowViewFactory.repoData.Ref(refName); err == nil {
		return
	}

	commit, err := windowViewFactory.repoData.CommitByOid(refName)
	if err != nil {
		err = fmt.Errorf("Invalid oid: %v - %v", refName, err)
	}

	ref = &HEAD{oid: commit.oid}

	return
}
