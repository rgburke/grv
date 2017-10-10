package main

import (
	"fmt"
	"sync"

	log "github.com/Sirupsen/logrus"
)

type gitStatusViewHandler func(*GitStatusView, Action) error

var statusTypeTitle = map[StatusType]*renderedStatusEntry{
	StStaged: {
		text:             "Changes to be committed:",
		themeComponentID: CmpGitStatusStagedTitle,
	},
	StUnstaged: {
		text:             "Changes not staged for commit:",
		themeComponentID: CmpGitStatusUnstagedTitle,
	},
	StUntracked: {
		text:             "Untracked files:",
		themeComponentID: CmpGitStatusUntrackedTitle,
	},
	StConflicted: {
		text:             "Unmerged paths:",
		themeComponentID: CmpGitStatusConflictedTitle,
	},
}

var statusTypeFileStyle = map[StatusType]ThemeComponentID{
	StStaged:     CmpGitStatusStagedFile,
	StUnstaged:   CmpGitStatusUnstagedFile,
	StUntracked:  CmpGitStatusUntrackedFile,
	StConflicted: CmpGitStatusConflictedFile,
}

var emptyStatusLine = &renderedStatusEntry{}

type renderedStatusEntry struct {
	text             string
	themeComponentID ThemeComponentID
	statusType       StatusType
	StatusEntry      *StatusEntry
}

// GitStatusView manages displaying git status data
type GitStatusView struct {
	repoData       RepoData
	channels       *Channels
	status         *Status
	renderedStatus []*renderedStatusEntry
	viewPos        ViewPos
	handlers       map[ActionType]gitStatusViewHandler
	active         bool
	lock           sync.Mutex
}

// NewGitStatusView created a new GitStatusView
func NewGitStatusView(repoData RepoData, channels *Channels) *GitStatusView {
	gitStatusView := &GitStatusView{
		repoData: repoData,
		channels: channels,
		viewPos:  NewViewPosition(),
		handlers: map[ActionType]gitStatusViewHandler{
			ActionPrevLine: moveUpStatusEntry,
			ActionNextLine: moveDownStatusEntry,
		},
	}

	repoData.RegisterStatusListener(gitStatusView)

	return gitStatusView
}

// Initialise does nothing
func (gitStatusView *GitStatusView) Initialise() (err error) {
	log.Debug("Initialising GitStatusView")
	return
}

// Render generates and writes the git status view to the provided window
func (gitStatusView *GitStatusView) Render(win RenderWindow) (err error) {
	gitStatusView.lock.Lock()
	defer gitStatusView.lock.Unlock()

	renderedStatus := gitStatusView.renderedStatus
	renderedStatusNum := uint(len(renderedStatus))
	rows := win.Rows() - 2

	viewPos := gitStatusView.viewPos
	viewPos.DetermineViewStartRow(rows, renderedStatusNum)
	renderedStatusIndex := viewPos.ViewStartRowIndex()
	startColumn := viewPos.ViewStartColumn()

	for rowIndex := uint(0); rowIndex < rows && renderedStatusIndex < renderedStatusNum; rowIndex++ {
		renderedStatusEntry := renderedStatus[renderedStatusIndex]

		if err = win.SetRow(rowIndex+1, startColumn, renderedStatusEntry.themeComponentID, " %v", renderedStatusEntry.text); err != nil {
			return
		}

		renderedStatusIndex++
	}

	if err = win.SetSelectedRow(viewPos.SelectedRowIndex()+1, gitStatusView.active); err != nil {
		return
	}

	win.DrawBorder()

	if err = win.SetTitle(CmpCommitviewTitle, "Status"); err != nil {
		return
	}

	return
}

func (gitStatusView *GitStatusView) renderStatusEntries(statusEntries []*StatusEntry) (err error) {
	return
}

// HandleKeyPress does nothing
func (gitStatusView *GitStatusView) HandleKeyPress(keystring string) (err error) {
	return
}

// HandleAction checks if git status view supports this action and if it does executes it
func (gitStatusView *GitStatusView) HandleAction(action Action) (err error) {
	gitStatusView.lock.Lock()
	defer gitStatusView.lock.Unlock()

	if handler, ok := gitStatusView.handlers[action.ActionType]; ok {
		log.Debugf("GitStatusView handling action %v", action)
		err = handler(gitStatusView, action)
	}

	return
}

// OnActiveChange updates whether this view is currently active
func (gitStatusView *GitStatusView) OnActiveChange(active bool) {
	log.Debugf("GitStatusView active: %v", active)
	gitStatusView.lock.Lock()
	defer gitStatusView.lock.Unlock()

	gitStatusView.active = active
}

// ViewID returns the ViewID for the git status view
func (gitStatusView *GitStatusView) ViewID() ViewID {
	return ViewGitStatus
}

// RenderHelpBar does nothing
func (gitStatusView *GitStatusView) RenderHelpBar(*LineBuilder) (err error) {
	return
}

// OnStatusChanged updates the git status view with the latest git status
func (gitStatusView *GitStatusView) OnStatusChanged(status *Status) {
	gitStatusView.lock.Lock()
	defer gitStatusView.lock.Unlock()

	gitStatusView.status = status
	gitStatusView.generateRenderedStatus()

	renderedStatus := gitStatusView.renderedStatus
	renderedStatusNum := uint(len(renderedStatus))
	viewPos := gitStatusView.viewPos

	if renderedStatusNum == 0 {
		viewPos.SetActiveRowIndex(0)
	} else if viewPos.ActiveRowIndex() >= renderedStatusNum {
		viewPos.SetActiveRowIndex(renderedStatusNum - 1)
	}
}

func (gitStatusView *GitStatusView) generateRenderedStatus() {
	var renderedStatus []*renderedStatusEntry
	status := gitStatusView.status
	statusTypes := status.StatusTypes()

	for statusTypeIndex, statusType := range statusTypes {
		renderedStatus = append(renderedStatus, statusTypeTitle[statusType], emptyStatusLine)

		themeComponentID := statusTypeFileStyle[statusType]
		statusEntries := status.Entries(statusType)

		for _, statusEntry := range statusEntries {
			var text string

			switch statusEntry.statusEntryType {
			case SetNew:
				prefix := ""

				if statusType == StStaged {
					prefix = "new file:   "
				}

				text = fmt.Sprintf("%v%v", prefix, statusEntry.diffDelta.NewFile.Path)
			case SetModified:
				text = fmt.Sprintf("modified:   %v", statusEntry.diffDelta.NewFile.Path)
			case SetDeleted:
				text = fmt.Sprintf("deleted:   %v", statusEntry.diffDelta.NewFile.Path)
			case SetRenamed:
				text = fmt.Sprintf("renamed:   %v -> %v", statusEntry.diffDelta.OldFile.Path, statusEntry.diffDelta.NewFile.Path)
			case SetTypeChange:
				text = fmt.Sprintf("typechange: %v", statusEntry.diffDelta.NewFile.Path)
			case SetConflicted:
				text = fmt.Sprintf("both modified:   %v", statusEntry.diffDelta.NewFile.Path)
			}

			renderedStatus = append(renderedStatus, &renderedStatusEntry{
				text:             "\t" + text,
				themeComponentID: themeComponentID,
				statusType:       statusType,
				StatusEntry:      statusEntry,
			})
		}

		if statusTypeIndex != len(statusTypes)-1 {
			renderedStatus = append(renderedStatus, emptyStatusLine)
		}
	}

	gitStatusView.renderedStatus = renderedStatus
}

func moveUpStatusEntry(gitStatusView *GitStatusView, action Action) (err error) {
	viewPos := gitStatusView.viewPos
	renderedStatus := gitStatusView.renderedStatus

	for viewPos.ActiveRowIndex() > 0 {
		if !viewPos.MoveLineUp() {
			return
		}

		if renderedStatus[viewPos.ActiveRowIndex()].text != "" {
			break
		}
	}

	log.Debug("Moved up one status entry")
	gitStatusView.channels.UpdateDisplay()

	return
}

func moveDownStatusEntry(gitStatusView *GitStatusView, action Action) (err error) {
	log.Debug("Moving up one status entry")

	viewPos := gitStatusView.viewPos
	renderedStatus := gitStatusView.renderedStatus
	renderedStatusNum := uint(len(renderedStatus))

	if renderedStatusNum == 0 {
		return
	}

	for viewPos.ActiveRowIndex() < renderedStatusNum-1 {
		if !viewPos.MoveLineDown(renderedStatusNum) {
			return
		}

		if renderedStatus[viewPos.ActiveRowIndex()].text != "" {
			break
		}
	}

	log.Debug("Moved down one status entry")
	gitStatusView.channels.UpdateDisplay()

	return
}
