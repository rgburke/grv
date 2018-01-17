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
		statusType:       StStaged,
	},
	StUnstaged: {
		text:             "Changes not staged for commit:",
		themeComponentID: CmpGitStatusUnstagedTitle,
		statusType:       StUnstaged,
	},
	StUntracked: {
		text:             "Untracked files:",
		themeComponentID: CmpGitStatusUntrackedTitle,
		statusType:       StUntracked,
	},
	StConflicted: {
		text:             "Unmerged paths:",
		themeComponentID: CmpGitStatusConflictedTitle,
		statusType:       StConflicted,
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

// GitStatusEntrySelectedListener is notified when either a file
// or a non-file entry is selected in the GitStatusView
type GitStatusEntrySelectedListener interface {
	OnFileSelected(statusType StatusType, path string)
	OnStageGroupSelected(statusType StatusType)
	OnNoEntrySelected()
}

// GitStatusView manages displaying git status data
type GitStatusView struct {
	repoData               RepoData
	channels               *Channels
	status                 *Status
	renderedStatus         []*renderedStatusEntry
	viewPos                ViewPos
	handlers               map[ActionType]gitStatusViewHandler
	active                 bool
	entrySelectedListeners []GitStatusEntrySelectedListener
	viewDimension          ViewDimension
	viewSearch             *ViewSearch
	lock                   sync.Mutex
}

// NewGitStatusView created a new GitStatusView
func NewGitStatusView(repoData RepoData, channels *Channels) *GitStatusView {
	gitStatusView := &GitStatusView{
		repoData: repoData,
		channels: channels,
		viewPos:  NewViewPosition(),
		handlers: map[ActionType]gitStatusViewHandler{
			ActionPrevLine:    moveUpGitStatusEntry,
			ActionNextLine:    moveDownGitStatusEntry,
			ActionPrevPage:    moveUpGitStatusPage,
			ActionNextPage:    moveDownGitStatusPage,
			ActionScrollRight: scrollGitStatusViewRight,
			ActionScrollLeft:  scrollGitStatusViewLeft,
			ActionFirstLine:   moveToFirstGitStatusEntry,
			ActionLastLine:    moveToLastGitStatusEntry,
			ActionCenterView:  centerGitStatusView,
			ActionSelect:      selectDiffEntry,
		},
	}

	gitStatusView.viewSearch = NewViewSearch(gitStatusView, channels)
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

	log.Debug("Rendering GitStatusView")

	gitStatusView.viewDimension = win.ViewDimensions()

	renderedStatus := gitStatusView.renderedStatus
	renderedStatusNum := uint(len(renderedStatus))
	rows := win.Rows() - 2

	viewPos := gitStatusView.ViewPos()
	viewPos.DetermineViewStartRow(rows, renderedStatusNum)
	renderedStatusIndex := viewPos.ViewStartRowIndex()
	startColumn := viewPos.ViewStartColumn()

	if renderedStatusNum == 0 {
		if err = win.SetRow(2, startColumn, CmpNone, "   %v", "nothing to commit, working tree clean"); err != nil {
			return
		}
	} else {
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
	}

	win.DrawBorder()

	if err = win.SetTitle(CmpCommitviewTitle, "Status"); err != nil {
		return
	}

	if searchActive, searchPattern, lastSearchFoundMatch := gitStatusView.viewSearch.SearchActive(); searchActive && lastSearchFoundMatch {
		if err = win.Highlight(searchPattern, CmpAllviewSearchMatch); err != nil {
			return
		}
	}

	return
}

// HandleEvent does nothing
func (gitStatusView *GitStatusView) HandleEvent(event Event) (err error) {
	return
}

// OnActiveChange updates whether this view is currently active
func (gitStatusView *GitStatusView) OnActiveChange(active bool) {
	log.Debugf("GitStatusView active: %v", active)
	gitStatusView.lock.Lock()
	defer gitStatusView.lock.Unlock()

	gitStatusView.active = active

	if active {
		if err := gitStatusView.selectEntry(gitStatusView.viewPos.ActiveRowIndex()); err != nil {
			gitStatusView.channels.ReportError(err)
		}
	}
}

// ViewID returns the ViewID for the git status view
func (gitStatusView *GitStatusView) ViewID() ViewID {
	return ViewGitStatus
}

// RenderHelpBar does nothing
func (gitStatusView *GitStatusView) RenderHelpBar(*LineBuilder) (err error) {
	return
}

// RegisterGitStatusFileSelectedListener registers a listener to be notified when the selected entry changes
func (gitStatusView *GitStatusView) RegisterGitStatusFileSelectedListener(entrySelectedListener GitStatusEntrySelectedListener) {
	gitStatusView.lock.Lock()
	defer gitStatusView.lock.Unlock()

	log.Debugf("Registering %T as a GitStatusFileSelectedListener", entrySelectedListener)
	gitStatusView.entrySelectedListeners = append(gitStatusView.entrySelectedListeners, entrySelectedListener)
}

func (gitStatusView *GitStatusView) notifyFileEntrySelected(renderedStatus *renderedStatusEntry) {
	log.Debugf("Notifying git status file selected listeners that file is selected")

	go func() {
		for _, entrySelectedListener := range gitStatusView.entrySelectedListeners {
			entrySelectedListener.OnFileSelected(renderedStatus.statusType, renderedStatus.StatusEntry.diffDelta.NewFile.Path)
		}
	}()

	return
}

func (gitStatusView *GitStatusView) notifyStageGroupSelected(statusType StatusType) {
	log.Debugf("Notifying git status file selected listeners that a stage group is selected")

	go func() {
		for _, entrySelectedListener := range gitStatusView.entrySelectedListeners {
			entrySelectedListener.OnStageGroupSelected(statusType)
		}
	}()

	return
}

func (gitStatusView *GitStatusView) notifyNoEntrySelected() {
	log.Debugf("Notifying git status file selected listeners that no entry is selected")

	go func() {
		for _, entrySelectedListener := range gitStatusView.entrySelectedListeners {
			entrySelectedListener.OnNoEntrySelected()
		}
	}()

	return
}

func (gitStatusView *GitStatusView) selectEntry(index uint) (err error) {
	renderedStatusNum := uint(len(gitStatusView.renderedStatus))

	if index > 0 && index >= renderedStatusNum {
		return fmt.Errorf("Invalid rendered status index: %v out of %v entries", index, renderedStatusNum)
	}

	gitStatusView.ViewPos().SetActiveRowIndex(index)

	if renderedStatusNum == 0 {
		return
	}

	renderedStatusEntry := gitStatusView.renderedStatus[index]
	log.Debugf("Selecting git status entry with index %v: %v", index, renderedStatusEntry.text)

	if renderedStatusEntry.statusType != StUntracked {
		if renderedStatusEntry.StatusEntry != nil {
			gitStatusView.notifyFileEntrySelected(renderedStatusEntry)
		} else {
			gitStatusView.notifyStageGroupSelected(renderedStatusEntry.statusType)
		}
	}

	return
}

// Line returns the rendered line at the specified index
func (gitStatusView *GitStatusView) Line(lineIndex uint) (line string) {
	gitStatusView.lock.Lock()
	defer gitStatusView.lock.Unlock()

	lineNumber := gitStatusView.lineNumber()
	if lineIndex >= lineNumber {
		log.Errorf("Invalid lineIndex: %v >= %v", lineIndex, lineNumber)
		return
	}

	renderedStatusEntry := gitStatusView.renderedStatus[lineIndex]

	return renderedStatusEntry.text
}

// LineNumber returns the number of lines in the view
func (gitStatusView *GitStatusView) LineNumber() (lineNumber uint) {
	gitStatusView.lock.Lock()
	defer gitStatusView.lock.Unlock()

	return gitStatusView.lineNumber()
}

// ViewPos returns the view position for this view
func (gitStatusView *GitStatusView) ViewPos() ViewPos {
	return gitStatusView.viewPos
}

// OnSearchMatch selects the line which matched the search pattern
func (gitStatusView *GitStatusView) OnSearchMatch(startPos ViewPos, matchLineIndex uint) {
	gitStatusView.lock.Lock()
	defer gitStatusView.lock.Unlock()

	viewPos := gitStatusView.ViewPos()

	if viewPos != startPos {
		log.Debugf("Selected git status entry has changed since search started")
		return
	}

	if gitStatusView.renderedStatus[matchLineIndex] == emptyStatusLine {
		log.Debugf("Unable to select empty line")
	} else {
		gitStatusView.selectEntry(matchLineIndex)
		gitStatusView.channels.UpdateDisplay()
	}
}

// OnStatusChanged updates the git status view with the latest git status
func (gitStatusView *GitStatusView) OnStatusChanged(status *Status) {
	gitStatusView.lock.Lock()
	defer gitStatusView.lock.Unlock()

	gitStatusView.status = status
	gitStatusView.generateRenderedStatus()

	renderedStatus := gitStatusView.renderedStatus
	renderedStatusNum := uint(len(renderedStatus))
	viewPos := gitStatusView.ViewPos()
	index := viewPos.ActiveRowIndex()

	if renderedStatusNum == 0 {
		index = 0
	} else if viewPos.ActiveRowIndex() >= renderedStatusNum {
		index = renderedStatusNum - 1
	}

	if err := gitStatusView.selectEntry(index); err != nil {
		log.Errorf("Error when attempting to selected status entry at index %v out of %v entries", index, renderedStatusNum)
	}

	if renderedStatusNum == 0 {
		gitStatusView.notifyNoEntrySelected()
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

func (gitStatusView *GitStatusView) lineNumber() uint {
	return uint(len(gitStatusView.renderedStatus))
}

func (gitStatusView *GitStatusView) createGitStatusViewListener() {
	createViewArgs := CreateViewArgs{
		viewID: ViewDiff,
		registerViewListener: func(observer interface{}) (err error) {
			if observer == nil {
				return fmt.Errorf("Invalid GitStatusEntrySelectedListener: %v", observer)
			}

			if listener, ok := observer.(GitStatusEntrySelectedListener); ok {
				gitStatusView.RegisterGitStatusFileSelectedListener(listener)
				gitStatusView.HandleAction(Action{
					ActionType: ActionSelect,
				})
			} else {
				err = fmt.Errorf("Observer is not a GitStatusEntrySelectedListener but has type %T", observer)
			}

			return
		},
	}

	gitStatusView.channels.DoAction(Action{
		ActionType: ActionSplitView,
		Args: []interface{}{
			ActionSplitViewArgs{
				CreateViewArgs: createViewArgs,
				orientation:    CoDynamic,
			},
		},
	})
}

// HandleAction checks if git status view supports this action and if it does executes it
func (gitStatusView *GitStatusView) HandleAction(action Action) (err error) {
	gitStatusView.lock.Lock()
	defer gitStatusView.lock.Unlock()

	if handler, ok := gitStatusView.handlers[action.ActionType]; ok {
		log.Debugf("GitStatusView handling action %v", action)
		err = handler(gitStatusView, action)
	} else {
		_, err = gitStatusView.viewSearch.HandleAction(action)
	}

	return
}

func moveUpGitStatusEntry(gitStatusView *GitStatusView, action Action) (err error) {
	viewPos := gitStatusView.ViewPos()
	renderedStatus := gitStatusView.renderedStatus

	for viewPos.ActiveRowIndex() > 0 {
		if !viewPos.MoveLineUp() {
			return
		}

		if renderedStatus[viewPos.ActiveRowIndex()].text != "" {
			break
		}
	}

	if action.ActionType == ActionPrevLine {
		gitStatusView.selectEntry(viewPos.ActiveRowIndex())
		log.Debug("Moved up one status entry")
		gitStatusView.channels.UpdateDisplay()
	}

	return
}

func moveDownGitStatusEntry(gitStatusView *GitStatusView, action Action) (err error) {
	viewPos := gitStatusView.ViewPos()
	renderedStatus := gitStatusView.renderedStatus
	renderedStatusNum := gitStatusView.lineNumber()

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

	if action.ActionType == ActionNextLine {
		gitStatusView.selectEntry(viewPos.ActiveRowIndex())
		log.Debug("Moved down one status entry")
		gitStatusView.channels.UpdateDisplay()
	}

	return
}

func moveUpGitStatusPage(gitStatusView *GitStatusView, action Action) (err error) {
	pageSize := gitStatusView.viewDimension.rows - 2
	viewPos := gitStatusView.ViewPos()

	for viewPos.ActiveRowIndex() > 0 && pageSize > 0 {
		if err = moveUpGitStatusEntry(gitStatusView, action); err != nil {
			return
		}

		pageSize--
	}

	if err = gitStatusView.selectEntry(viewPos.ActiveRowIndex()); err != nil {
		return
	}

	log.Debug("Moved up one page")
	gitStatusView.channels.UpdateDisplay()

	return
}

func moveDownGitStatusPage(gitStatusView *GitStatusView, action Action) (err error) {
	pageSize := gitStatusView.viewDimension.rows - 2
	viewPos := gitStatusView.ViewPos()
	renderedStatusNum := gitStatusView.lineNumber()

	for viewPos.ActiveRowIndex()+1 < renderedStatusNum && pageSize > 0 {
		if err = moveDownGitStatusEntry(gitStatusView, action); err != nil {
			return
		}

		pageSize--
	}

	if err = gitStatusView.selectEntry(viewPos.ActiveRowIndex()); err != nil {
		return
	}

	log.Debug("Moved down one page")
	gitStatusView.channels.UpdateDisplay()

	return
}

func scrollGitStatusViewRight(gitStatusView *GitStatusView, action Action) (err error) {
	viewPos := gitStatusView.ViewPos()
	viewPos.MovePageRight(gitStatusView.viewDimension.cols)
	log.Debugf("Scrolling right. View starts at column %v", viewPos.ViewStartColumn())
	gitStatusView.channels.UpdateDisplay()

	return
}

func scrollGitStatusViewLeft(gitStatusView *GitStatusView, action Action) (err error) {
	viewPos := gitStatusView.ViewPos()

	if viewPos.MovePageLeft(gitStatusView.viewDimension.cols) {
		log.Debugf("Scrolling left. View starts at column %v", viewPos.ViewStartColumn())
		gitStatusView.channels.UpdateDisplay()
	}

	return
}

func moveToFirstGitStatusEntry(gitStatusView *GitStatusView, action Action) (err error) {
	viewPos := gitStatusView.ViewPos()

	if viewPos.MoveToFirstLine() {
		if err = gitStatusView.selectEntry(viewPos.ActiveRowIndex()); err != nil {
			return
		}

		log.Debug("Selected first entry")
		gitStatusView.channels.UpdateDisplay()
	}

	return
}

func moveToLastGitStatusEntry(gitStatusView *GitStatusView, action Action) (err error) {
	lineNumber := gitStatusView.lineNumber()
	viewPos := gitStatusView.ViewPos()

	if viewPos.MoveToLastLine(lineNumber) {
		if err = gitStatusView.selectEntry(viewPos.ActiveRowIndex()); err != nil {
			return
		}

		log.Debug("Moved to last entry")
		gitStatusView.channels.UpdateDisplay()
	}

	return
}

func centerGitStatusView(gitStatusView *GitStatusView, action Action) (err error) {
	viewPos := gitStatusView.ViewPos()

	if viewPos.CenterActiveRow(gitStatusView.viewDimension.rows - 2) {
		log.Debug("Centering GitStatusView")
		gitStatusView.channels.UpdateDisplay()
	}

	return
}

func selectDiffEntry(gitStatusView *GitStatusView, action Action) (err error) {
	if len(gitStatusView.entrySelectedListeners) == 0 {
		gitStatusView.createGitStatusViewListener()
	} else {
		viewPos := gitStatusView.ViewPos()
		gitStatusView.selectEntry(viewPos.ActiveRowIndex())
	}

	return
}
