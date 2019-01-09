package main

import (
	"fmt"
	"regexp"

	log "github.com/Sirupsen/logrus"
)

// WindowViewFactory provides a generic interface
// for creating view instances
type WindowViewFactory struct {
	repoData       RepoData
	repoController RepoController
	channels       Channels
	config         Config
	variables      GRVVariableSetter
}

var hexRegexp = regexp.MustCompile(`^[[:xdigit:]]+$`)

// NewWindowViewFactory creates a new instance
func NewWindowViewFactory(repoData RepoData, repoController RepoController, channels Channels, config Config, variables GRVVariableSetter) *WindowViewFactory {
	return &WindowViewFactory{
		repoData:       repoData,
		repoController: repoController,
		channels:       channels,
		config:         config,
		variables:      variables,
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
	case ViewGRVVariable:
		windowView = windowViewFactory.createGRVVariableView()
	case ViewRemote:
		windowView = windowViewFactory.createRemoteView()
	default:
		err = fmt.Errorf("Unsupported view type: %v", viewID)
	}

	return
}

func (windowViewFactory *WindowViewFactory) createRefView() *RefView {
	log.Info("Created RefView instance")
	return NewRefView(windowViewFactory.repoData, windowViewFactory.repoController, windowViewFactory.channels,
		windowViewFactory.config, windowViewFactory.variables)
}

func (windowViewFactory *WindowViewFactory) createCommitView(args []interface{}) (commitView *CommitView, err error) {
	ref, err := windowViewFactory.getRef(args)
	if err != nil {
		return
	}

	commitView = NewCommitView(windowViewFactory.repoData, windowViewFactory.repoController, windowViewFactory.channels,
		windowViewFactory.config, windowViewFactory.variables)

	log.Info("Created CommitView instance")

	if ref == nil {
		ref = windowViewFactory.repoData.Head()
	}

	log.Debugf("Providing Ref to CommitView instance %v:%v", ref.Name(), ref.Oid())
	err = commitView.OnRefSelect(ref)

	return
}

func (windowViewFactory *WindowViewFactory) createDiffView(args []interface{}) (diffView *DiffView, err error) {
	ref, err := windowViewFactory.getRef(args)
	if err != nil {
		return
	}

	diffView = NewDiffView(windowViewFactory.repoData, windowViewFactory.channels, windowViewFactory.config, windowViewFactory.variables)

	log.Info("Created DiffView instance")

	if ref != nil {
		var commit *Commit
		commit, err = windowViewFactory.repoData.Commit(ref.Oid())

		if err == nil {
			log.Debugf("Providing Commit to DiffView instance %v", commit.oid)
			err = diffView.OnCommitSelected(commit)
		}
	}

	return
}

func (windowViewFactory *WindowViewFactory) createGitStatusView() *GitStatusView {
	gitStatusView := NewGitStatusView(windowViewFactory.repoData, windowViewFactory.repoController, windowViewFactory.channels,
		windowViewFactory.config, windowViewFactory.variables)

	if status := windowViewFactory.repoData.Status(); status != nil {
		gitStatusView.OnStatusChanged(status)
	}

	log.Info("Created GitStatusView instance")

	return gitStatusView
}

func (windowViewFactory *WindowViewFactory) createGRVVariableView() *GRVVariableView {
	log.Info("Created GRVVariableView instance")
	return NewGRVVariableView(windowViewFactory.channels, windowViewFactory.config, windowViewFactory.variables)
}

func (windowViewFactory *WindowViewFactory) createRemoteView() *RemoteView {
	log.Info("Created GRVVariableView instance")
	return NewRemoteView(windowViewFactory.repoData, windowViewFactory.repoController,
		windowViewFactory.channels, windowViewFactory.config, windowViewFactory.variables)
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
		log.Debugf("Found ref %v for input %v", ref.Name(), refName)
		return
	}

	if hexRegexp.MatchString(refName) {
		var commit *Commit
		commit, err = windowViewFactory.repoData.CommitByOid(refName)
		if err != nil {
			err = fmt.Errorf("Invalid oid: %v - %v", refName, err)
			return
		}

		log.Debugf("Found commit %v for input %v", commit.oid, refName)
		ref = &HEAD{oid: commit.oid}
	} else {
		log.Debug("Input is not oid")
	}

	return
}

// GenerateWindowViewFactoryHelpSection generates a help documentation table of supported views
func GenerateWindowViewFactoryHelpSection(config Config) *HelpSection {
	headers := []TableHeader{
		{text: "View", themeComponentID: CmpHelpViewSectionTableHeader},
		{text: "Args", themeComponentID: CmpHelpViewSectionTableHeader},
	}

	tableFormatter := NewTableFormatterWithHeaders(headers, config)
	tableFormatter.SetGridLines(true)

	type viewConstructor struct {
		viewID ViewID
		args   string
	}

	viewConstructors := []viewConstructor{
		{
			viewID: ViewCommit,
			args:   "ref or oid",
		},
		{
			viewID: ViewDiff,
			args:   "oid",
		},
		{
			viewID: ViewGitStatus,
			args:   "none",
		},
		{
			viewID: ViewRef,
			args:   "none",
		},
	}

	tableFormatter.Resize(uint(len(viewConstructors)))

	for rowIndex, constructor := range viewConstructors {
		tableFormatter.SetCellWithStyle(uint(rowIndex), 0, CmpHelpViewSectionTableRow, "%v", ViewName(constructor.viewID))
		tableFormatter.SetCellWithStyle(uint(rowIndex), 1, CmpHelpViewSectionTableRow, "%v", constructor.args)
	}

	return &HelpSection{
		tableFormatter: tableFormatter,
	}
}
