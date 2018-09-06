package main

import (
	"sync"

	log "github.com/Sirupsen/logrus"
)

// RemoteView displays remotes
type RemoteView struct {
	*AbstractWindowView
	channels          Channels
	repoData          RepoData
	config            Config
	activeViewPos     ViewPos
	lastViewDimension ViewDimension
	variables         GRVVariableSetter
	remotes           []string
	lock              sync.Mutex
}

// NewRemoteView creates a new remote view instance
func NewRemoteView(repoData RepoData, channels Channels, config Config, variables GRVVariableSetter) *RemoteView {
	remoteView := &RemoteView{
		repoData:      repoData,
		channels:      channels,
		config:        config,
		activeViewPos: NewViewPosition(),
		variables:     variables,
	}

	remoteView.AbstractWindowView = NewAbstractWindowView(remoteView, channels, config, variables, &remoteView.lock, "remote")

	return remoteView
}

// Initialise does an initial remote load
func (remoteView *RemoteView) Initialise() (err error) {
	if loadErr := remoteView.repoData.LoadRemotes(); loadErr != nil {
		log.Debugf("Failed to load remotes %v", loadErr)
	} else {
		remoteView.remotes = remoteView.repoData.Remotes()
	}

	return
}

// Render generates and writes the remote view to the provided window
func (remoteView *RemoteView) Render(win RenderWindow) (err error) {
	remoteView.lock.Lock()
	defer remoteView.lock.Unlock()

	remoteView.lastViewDimension = win.ViewDimensions()
	remoteView.remotes = remoteView.repoData.Remotes()

	remoteNum := remoteView.rows()
	if remoteNum == 0 {
		return remoteView.AbstractWindowView.renderEmptyView(win, "No Remotes")
	}

	rows := win.Rows() - 2
	viewPos := remoteView.activeViewPos
	viewPos.DetermineViewStartRow(rows, remoteNum)

	lineIndex := viewPos.ViewStartRowIndex()
	startColumn := viewPos.ViewStartColumn()

	for rowIndex := uint(0); rowIndex < rows && lineIndex < remoteNum; rowIndex++ {
		remoteName := remoteView.remotes[lineIndex]
		if err = win.SetRow(rowIndex+1, startColumn, CmpRemoteViewRemote, "  %v", remoteName); err != nil {
			return
		}

		lineIndex++
	}

	if err = win.SetSelectedRow(viewPos.SelectedRowIndex()+1, remoteView.active); err != nil {
		return
	}

	win.DrawBorder()

	if err = win.SetTitle(CmpRemoteViewTitle, "Remotes"); err != nil {
		return
	}

	if err = win.SetFooter(CmpRemoteViewFooter, "Remote %v of %v", viewPos.ActiveRowIndex()+1, remoteNum); err != nil {
		return
	}

	if searchActive, searchPattern, lastSearchFoundMatch := remoteView.viewSearch.SearchActive(); searchActive && lastSearchFoundMatch {
		if err = win.Highlight(searchPattern, CmpAllviewSearchMatch); err != nil {
			return
		}
	}

	return
}

// ViewID returns the diff views ID
func (remoteView *RemoteView) ViewID() ViewID {
	return ViewRemote
}

func (remoteView *RemoteView) viewPos() ViewPos {
	return remoteView.activeViewPos
}

func (remoteView *RemoteView) line(lineIndex uint) (line string) {
	if lineIndex >= remoteView.rows() {
		return
	}

	line = remoteView.remotes[lineIndex]

	return
}

func (remoteView *RemoteView) rows() uint {
	return uint(len(remoteView.remotes))
}

func (remoteView *RemoteView) viewDimension() ViewDimension {
	return remoteView.lastViewDimension
}

func (remoteView *RemoteView) onRowSelected(rowIndex uint) (err error) {
	return
}

// HandleAction checks if the remote view supports the provided action and executes it if so
func (remoteView *RemoteView) HandleAction(action Action) (err error) {
	remoteView.lock.Lock()
	defer remoteView.lock.Unlock()

	var handled bool
	if handled, err = remoteView.AbstractWindowView.HandleAction(action); handled {
		log.Debugf("Action handled by AbstractWindowView")
	} else {
		log.Debugf("Action not handled")
	}

	return
}
