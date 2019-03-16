package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
	"unicode"

	log "github.com/Sirupsen/logrus"
)

type gitStatusViewHandler func(*GitStatusView, Action) error

const (
	gsvLastModifyThresholdMillis = 500
)

var commitMessageFileCommentLineRegex = regexp.MustCompile(`^\s*#`)

var statusTypeTitle = map[StatusType]*renderedStatusEntry{
	StStaged: {
		entryType:        rsetHeader,
		text:             "Changes to be committed:",
		themeComponentID: CmpGitStatusStagedTitle,
		statusType:       StStaged,
	},
	StUnstaged: {
		entryType:        rsetHeader,
		text:             "Changes not staged for commit:",
		themeComponentID: CmpGitStatusUnstagedTitle,
		statusType:       StUnstaged,
	},
	StUntracked: {
		entryType:        rsetHeader,
		text:             "Untracked files:",
		themeComponentID: CmpGitStatusUntrackedTitle,
		statusType:       StUntracked,
	},
	StConflicted: {
		entryType:        rsetHeader,
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

type renderedStatusEntryType int

const (
	rsetEmpty renderedStatusEntryType = iota
	rsetStatusMessage
	rsetHeader
	rsetFile
)

type renderedStatusEntry struct {
	entryType        renderedStatusEntryType
	text             string
	filePath         string
	themeComponentID ThemeComponentID
	statusType       StatusType
	StatusEntry      *StatusEntry
}

func (renderedStatusEntry *renderedStatusEntry) isSelectable() bool {
	return renderedStatusEntry.entryType != rsetEmpty &&
		renderedStatusEntry.entryType != rsetStatusMessage
}

func (renderedStatusEntry *renderedStatusEntry) isFileEntry() bool {
	return renderedStatusEntry.entryType == rsetFile
}

// GitStatusViewListener is notified when either a file
// or a non-file entry is selected in the GitStatusView
type GitStatusViewListener interface {
	OnFileSelected(statusType StatusType, path string)
	OnStageGroupSelected(statusType StatusType)
	OnNoEntrySelected()
}

// GitStatusView manages displaying git status data
type GitStatusView struct {
	*SelectableRowView
	repoData               RepoData
	repoController         RepoController
	channels               Channels
	config                 Config
	status                 *Status
	renderedStatus         []*renderedStatusEntry
	activeViewPos          ViewPos
	handlers               map[ActionType]gitStatusViewHandler
	gitStatusViewListeners []GitStatusViewListener
	lastViewDimension      ViewDimension
	lastModify             time.Time
	variables              GRVVariableSetter
	lock                   sync.Mutex
}

// NewGitStatusView created a new GitStatusView
func NewGitStatusView(repoData RepoData, repoController RepoController, channels Channels, config Config, variables GRVVariableSetter) *GitStatusView {
	gitStatusView := &GitStatusView{
		repoData:       repoData,
		repoController: repoController,
		channels:       channels,
		config:         config,
		activeViewPos:  NewViewPosition(),
		lastModify:     time.Now(),
		variables:      variables,
		handlers: map[ActionType]gitStatusViewHandler{
			ActionSelect:       selectGitStatusEntry,
			ActionStageFile:    stageFile,
			ActionUnstageFile:  unstageFile,
			ActionCheckoutFile: checkoutFile,
			ActionCommit:       commit,
			ActionAmendCommit:  amendCommit,
		},
	}

	gitStatusView.SelectableRowView = NewSelectableRowView(gitStatusView, channels, config, variables, &gitStatusView.lock, "status row")
	repoData.RegisterStatusListener(gitStatusView)
	repoData.RegisterRefStateListener(gitStatusView)

	return gitStatusView
}

// Initialise does nothing
func (gitStatusView *GitStatusView) Initialise() (err error) {
	log.Debug("Initialising GitStatusView")

	go gitStatusView.repoData.LoadStatus()

	return
}

// Dispose of any resources held by the view
func (gitStatusView *GitStatusView) Dispose() {

}

// Render generates and writes the git status view to the provided window
func (gitStatusView *GitStatusView) Render(win RenderWindow) (err error) {
	gitStatusView.lock.Lock()
	defer gitStatusView.lock.Unlock()

	gitStatusView.lastViewDimension = win.ViewDimensions()

	if gitStatusView.status == nil {
		return gitStatusView.AbstractWindowView.renderEmptyView(win, "Loading status...")
	}

	renderedStatus := gitStatusView.renderedStatus
	renderedStatusNum := uint(len(renderedStatus))
	rows := win.Rows() - 2

	viewPos := gitStatusView.viewPos()
	viewPos.DetermineViewStartRow(rows, renderedStatusNum)
	renderedStatusIndex := viewPos.ViewStartRowIndex()
	startColumn := viewPos.ViewStartColumn()

	var rowIndex uint
	for rowIndex = 0; rowIndex < rows && renderedStatusIndex < renderedStatusNum; rowIndex++ {
		renderedStatusEntry := renderedStatus[renderedStatusIndex]

		if err = win.SetRow(rowIndex+1, startColumn, renderedStatusEntry.themeComponentID, " %v", renderedStatusEntry.text); err != nil {
			return
		}

		renderedStatusIndex++
	}

	if !gitStatusView.status.IsEmpty() {
		if err = win.SetSelectedRow(viewPos.SelectedRowIndex()+1, gitStatusView.viewState); err != nil {
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

// HandleEvent reacts to an event
func (gitStatusView *GitStatusView) HandleEvent(event Event) (err error) {
	gitStatusView.lock.Lock()
	defer gitStatusView.lock.Unlock()

	switch event.EventType {
	case ViewRemovedEvent:
		gitStatusView.removeGitStatusViewListeners(event.Args)
	}

	return
}

func (gitStatusView *GitStatusView) removeGitStatusViewListeners(views []interface{}) {
	for _, view := range views {
		if gitStatusViewListener, ok := view.(GitStatusViewListener); ok {
			gitStatusView.removeGitStatusViewListener(gitStatusViewListener)
		}
	}
}

func (gitStatusView *GitStatusView) removeGitStatusViewListener(gitStatusViewListener GitStatusViewListener) {
	for index, listener := range gitStatusView.gitStatusViewListeners {
		if gitStatusViewListener == listener {
			log.Debugf("Removing GitStatusViewListener %T", gitStatusViewListener)
			gitStatusView.gitStatusViewListeners = append(gitStatusView.gitStatusViewListeners[:index], gitStatusView.gitStatusViewListeners[index+1:]...)
			break
		}
	}
}

// OnStateChange updates whether this view is currently active
func (gitStatusView *GitStatusView) OnStateChange(viewState ViewState) {
	gitStatusView.AbstractWindowView.OnStateChange(viewState)

	gitStatusView.lock.Lock()
	defer gitStatusView.lock.Unlock()

	if viewState == ViewStateActive {
		if err := gitStatusView.selectEntry(gitStatusView.activeViewPos.ActiveRowIndex()); err != nil {
			gitStatusView.channels.ReportError(err)
		}
	}
}

// ViewID returns the ViewID for the git status view
func (gitStatusView *GitStatusView) ViewID() ViewID {
	return ViewGitStatus
}

// RenderHelpBar does nothing
func (gitStatusView *GitStatusView) RenderHelpBar(lineBuilder *LineBuilder) (err error) {
	RenderKeyBindingHelp(gitStatusView.ViewID(), lineBuilder, gitStatusView.config, []ActionMessage{
		{action: ActionStageFile, message: "Stage"},
		{action: ActionUnstageFile, message: "Unstage"},
		{action: ActionCheckoutFile, message: "Checkout"},
		{action: ActionCommit, message: "Commit"},
		{action: ActionAmendCommit, message: "Amend Commit"},
	})

	return
}

// RegisterGitStatusFileSelectedListener registers a listener to be notified when the selected entry changes
func (gitStatusView *GitStatusView) RegisterGitStatusFileSelectedListener(gitStatusViewListener GitStatusViewListener) {
	if gitStatusViewListener == nil {
		return
	}

	log.Debugf("Registering GitStatusViewListener %T", gitStatusViewListener)

	gitStatusView.lock.Lock()
	defer gitStatusView.lock.Unlock()

	gitStatusView.gitStatusViewListeners = append(gitStatusView.gitStatusViewListeners, gitStatusViewListener)
}

func (gitStatusView *GitStatusView) notifyFileEntrySelected(renderedStatus *renderedStatusEntry) {
	log.Debugf("Notifying git status file selected listeners that file is selected")

	go func() {
		for _, gitStatusViewListener := range gitStatusView.gitStatusViewListeners {
			gitStatusViewListener.OnFileSelected(renderedStatus.statusType, renderedStatus.StatusEntry.NewFilePath())
		}
	}()

	return
}

func (gitStatusView *GitStatusView) notifyStageGroupSelected(statusType StatusType) {
	log.Debugf("Notifying git status file selected listeners that a stage group is selected")

	go func() {
		for _, gitStatusViewListener := range gitStatusView.gitStatusViewListeners {
			gitStatusViewListener.OnStageGroupSelected(statusType)
		}
	}()

	return
}

func (gitStatusView *GitStatusView) notifyNoEntrySelected() {
	log.Debugf("Notifying git status file selected listeners that no entry is selected")

	go func() {
		for _, gitStatusViewListener := range gitStatusView.gitStatusViewListeners {
			gitStatusViewListener.OnNoEntrySelected()
		}
	}()

	return
}

func (gitStatusView *GitStatusView) selectEntry(index uint) (err error) {
	renderedStatusNum := uint(len(gitStatusView.renderedStatus))

	if index > 0 && index >= renderedStatusNum {
		return fmt.Errorf("Invalid rendered status index: %v out of %v entries", index, renderedStatusNum)
	}

	gitStatusView.viewPos().SetActiveRowIndex(index)

	if gitStatusView.status == nil || gitStatusView.status.IsEmpty() {
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

	gitStatusView.setVariables()

	return
}

func (gitStatusView *GitStatusView) line(lineIndex uint) (line string) {
	rows := gitStatusView.rows()
	if lineIndex >= rows {
		log.Errorf("Invalid lineIndex: %v >= %v", lineIndex, rows)
		return
	}

	renderedStatusEntry := gitStatusView.renderedStatus[lineIndex]

	return renderedStatusEntry.text
}

func (gitStatusView *GitStatusView) viewPos() ViewPos {
	return gitStatusView.activeViewPos
}

// OnStatusChanged updates the git status view with the latest git status
func (gitStatusView *GitStatusView) OnStatusChanged(status *Status) {
	gitStatusView.lock.Lock()
	defer gitStatusView.lock.Unlock()

	gitStatusView.status = status
	gitStatusView.generateRenderedStatus()

	renderedStatus := gitStatusView.renderedStatus
	renderedStatusNum := uint(len(renderedStatus))
	viewPos := gitStatusView.viewPos()
	index := viewPos.ActiveRowIndex()

	if status.IsEmpty() {
		index = 0
	} else if viewPos.ActiveRowIndex() >= renderedStatusNum {
		index = renderedStatusNum - 1
	}

	if err := gitStatusView.selectEntry(index); err != nil {
		log.Errorf("Error when attempting to selected status entry at index %v out of %v entries", index, renderedStatusNum)
	}

	if status.IsEmpty() {
		gitStatusView.notifyNoEntrySelected()
	} else if (!gitStatusView.isSelectableRow(index)) ||
		time.Now().Before(gitStatusView.lastModify.Add(time.Millisecond*gsvLastModifyThresholdMillis)) {

		gitStatusView.channels.ReportError(gitStatusView.selectNextFileEntry())
	}
}

// OnRefsChanged updates the status message when refs have changed
func (gitStatusView *GitStatusView) OnRefsChanged(addedRefs, removedRefs []Ref, updatedRefs []*UpdatedRef) {
	gitStatusView.lock.Lock()
	defer gitStatusView.lock.Unlock()

	gitStatusView.generateRenderedStatus()
	gitStatusView.channels.UpdateDisplay()
}

// OnHeadChanged updates the status message when head has changed
func (gitStatusView *GitStatusView) OnHeadChanged(oldHead, newHead Ref) {
	gitStatusView.lock.Lock()
	defer gitStatusView.lock.Unlock()

	gitStatusView.generateRenderedStatus()
	gitStatusView.channels.UpdateDisplay()
}

// OnTrackingBranchesUpdated updates the status message when tracking branch data has been refreshed
func (gitStatusView *GitStatusView) OnTrackingBranchesUpdated(trackingBranches []*LocalBranch) {
	gitStatusView.lock.Lock()
	defer gitStatusView.lock.Unlock()

	gitStatusView.generateRenderedStatus()
	gitStatusView.channels.UpdateDisplay()
}

func (gitStatusView *GitStatusView) generateRenderedStatus() {
	renderedStatus := gitStatusView.generateRenderedBranchStatus()
	status := gitStatusView.status

	switch {
	case status == nil:
	case status.IsEmpty():
		renderedStatus = append(renderedStatus, &renderedStatusEntry{
			entryType:        rsetStatusMessage,
			text:             "nothing to commit, working tree clean",
			themeComponentID: CmpGitStatusMessage,
		})
	default:
		renderedStatus = append(renderedStatus, emptyStatusLine)
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

					text = fmt.Sprintf("%v%v", prefix, statusEntry.NewFilePath())
				case SetModified:
					text = fmt.Sprintf("modified:   %v", statusEntry.NewFilePath())
				case SetDeleted:
					text = fmt.Sprintf("deleted:   %v", statusEntry.NewFilePath())
				case SetRenamed:
					text = fmt.Sprintf("renamed:   %v -> %v", statusEntry.OldFilePath(), statusEntry.NewFilePath())
				case SetTypeChange:
					text = fmt.Sprintf("typechange: %v", statusEntry.NewFilePath())
				case SetConflicted:
					text = fmt.Sprintf("both modified:   %v", statusEntry.NewFilePath())
				}

				renderedStatus = append(renderedStatus, &renderedStatusEntry{
					entryType:        rsetFile,
					text:             "\t" + text,
					filePath:         statusEntry.NewFilePath(),
					themeComponentID: themeComponentID,
					statusType:       statusType,
					StatusEntry:      statusEntry,
				})
			}

			if statusTypeIndex != len(statusTypes)-1 {
				renderedStatus = append(renderedStatus, emptyStatusLine)
			}
		}
	}

	gitStatusView.renderedStatus = renderedStatus

	gitStatusView.selectNearestSelectableRow()
	gitStatusView.setVariables()
}

func (gitStatusView *GitStatusView) generateRenderedBranchStatus() (renderedStatus []*renderedStatusEntry) {
	branchStatus := gitStatusView.generateBranchStatus()

	for _, branchStatusLine := range branchStatus {
		renderedStatus = append(renderedStatus, &renderedStatusEntry{
			entryType:        rsetStatusMessage,
			text:             branchStatusLine,
			themeComponentID: CmpGitStatusMessage,
		})
	}

	return
}

func (gitStatusView *GitStatusView) setVariables() {
	gitStatusView.SelectableRowView.setVariables()

	rowIndex := gitStatusView.viewPos().ActiveRowIndex()

	if !gitStatusView.isFileEntry(rowIndex) {
		return
	}

	filePath := gitStatusView.renderedStatus[rowIndex].filePath
	gitStatusView.variables.SetViewVariable(VarFile, filePath, gitStatusView.viewState)
}

func (gitStatusView *GitStatusView) rows() uint {
	return uint(len(gitStatusView.renderedStatus))
}

func (gitStatusView *GitStatusView) viewDimension() ViewDimension {
	return gitStatusView.lastViewDimension
}

func (gitStatusView *GitStatusView) onRowSelected(rowIndex uint) error {
	return gitStatusView.selectEntry(rowIndex)
}

func (gitStatusView *GitStatusView) isSelectableRow(rowIndex uint) (isSelectable bool) {
	renderedStatus := gitStatusView.renderedStatus

	if rowIndex >= uint(len(renderedStatus)) {
		return
	}

	return renderedStatus[rowIndex].isSelectable()
}

func (gitStatusView *GitStatusView) isFileEntry(rowIndex uint) (isFileEntry bool) {
	renderedStatus := gitStatusView.renderedStatus

	if rowIndex >= uint(len(renderedStatus)) {
		return
	}

	return renderedStatus[rowIndex].isFileEntry()
}

func (gitStatusView *GitStatusView) createGitStatusViewListener() {
	createViewArgs := CreateViewArgs{
		viewID: ViewDiff,
		registerViewListener: func(observer interface{}) (err error) {
			if observer == nil {
				return fmt.Errorf("Invalid GitStatusViewListener: %v", observer)
			}

			if listener, ok := observer.(GitStatusViewListener); ok {
				gitStatusView.RegisterGitStatusFileSelectedListener(listener)
				gitStatusView.HandleAction(Action{
					ActionType: ActionSelect,
				})
			} else {
				err = fmt.Errorf("Observer is not a GitStatusViewListener but has type %T", observer)
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

func (gitStatusView *GitStatusView) userEditor() (editor string, err error) {
	if editor, err = gitStatusView.repoData.UserEditor(); err != nil || editor != "" {
		return
	} else if editor = os.Getenv("EDITOR"); editor != "" {
		return
	} else if editor = os.Getenv("VISUAL"); editor != "" {
		return
	} else {
		editor = "vi"
	}

	return
}

func (gitStatusView *GitStatusView) generateBranchStatus() (lines []string) {
	head := gitStatusView.repoData.Head()

	if _, isDetached := head.(*HEAD); isDetached {
		lines = append(lines, GetDetachedHeadDisplayValue(head.Oid()))
		return
	}

	ref, err := gitStatusView.repoData.Ref(head.Name())
	if err != nil {
		return
	}

	branch, isLocalBranch := ref.(*LocalBranch)
	if !isLocalBranch {
		return
	}

	lines = append(lines, fmt.Sprintf("On branch %v", branch.Shorthand()))

	if !branch.IsTrackingBranch() {
		return
	}

	remoteBranch, err := gitStatusView.repoData.Ref(branch.remoteBranch)
	if err != nil {
		return
	}

	var trackingStatus string

	switch {
	case branch.ahead == 0 && branch.behind == 0:
		trackingStatus = fmt.Sprintf("Your branch is up-to-date with '%v'.", remoteBranch.Shorthand())
	case branch.ahead > 0 && branch.behind > 0:
		trackingStatus = fmt.Sprintf("Your branch and '%v' have diverged, "+
			"and have %v and %v different commits each, respectively",
			remoteBranch.Shorthand(), branch.ahead, branch.behind)
	case branch.ahead > 0:
		multiple := ""
		if branch.ahead > 0 {
			multiple = "s"
		}
		trackingStatus = fmt.Sprintf("Your branch is ahead of '%v' by %v commit%v.",
			remoteBranch.Shorthand(), branch.ahead, multiple)
	case branch.behind > 0:
		multiple := ""
		if branch.behind > 0 {
			multiple = "s"
		}
		trackingStatus = fmt.Sprintf("Your branch is behind '%v' by %v commit%v.",
			remoteBranch.Shorthand(), branch.behind, multiple)
	}

	lines = append(lines, trackingStatus)

	return
}

func (gitStatusView *GitStatusView) generateCommitMessageFile() (filePath string, err error) {
	commitMessageFile, err := gitStatusView.repoController.CommitMessageFile()
	if err != nil {
		return
	}
	defer commitMessageFile.Close()
	filePath = commitMessageFile.Name()

	lines := []string{
		"Please enter the commit message for your changes. Lines starting",
		"with '#' will be ignored, and an empty message aborts the commit.",
		"",
	}

	lines = append(lines, gitStatusView.generateBranchStatus()...)
	lines = append(lines, "")

	for _, renderedStatusEntry := range gitStatusView.renderedStatus {
		lines = append(lines, renderedStatusEntry.text)
	}

	lines = append(lines, "")

	writer := bufio.NewWriter(commitMessageFile)

	if _, err = writer.WriteString("\n"); err != nil {
		err = fmt.Errorf("Failed to write commit message file: %v", err)
		return
	}

	for _, line := range lines {
		if _, err = writer.WriteString(fmt.Sprintf("# %v\n", line)); err != nil {
			err = fmt.Errorf("Failed to write commit message file: %v", err)
			return
		}
	}

	if err = writer.Flush(); err != nil {
		err = fmt.Errorf("Failed to write commit message file: %v", err)
	}

	return
}

func (gitStatusView *GitStatusView) processCommitMessageFile(filePath string) (commitMessage string, err error) {
	file, err := os.Open(filePath)
	if err != nil {
		err = fmt.Errorf("Unable to open commit message file for reading: %v", err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	commitMessageLines := []string{}

	for scanner.Scan() {
		line := scanner.Text()
		if !commitMessageFileCommentLineRegex.MatchString(line) {
			commitMessageLines = append(commitMessageLines, line)
		}
	}

	if err = scanner.Err(); err != nil {
		err = fmt.Errorf("Error when reading commit message file: %v", err)
		return
	}

	commitMessage = strings.Join(commitMessageLines, "\n")

	trimmedCommitMessage := strings.Map(func(char rune) rune {
		if unicode.IsSpace(char) {
			return -1
		}

		return char
	}, commitMessage)

	if trimmedCommitMessage == "" {
		err = fmt.Errorf("Aborting commit due to empty commit message")
		return
	}

	return
}

// HandleAction checks if git status view supports this action and if it does executes it
func (gitStatusView *GitStatusView) HandleAction(action Action) (err error) {
	gitStatusView.lock.Lock()
	defer gitStatusView.lock.Unlock()

	if gitStatusView.status == nil {
		log.Debugf("Status not set. Cannot perform any actions yet")
		return
	}

	var handled bool
	if handler, ok := gitStatusView.handlers[action.ActionType]; ok {
		log.Debugf("Action handled by GitStatusView")
		err = handler(gitStatusView, action)
		gitStatusView.lastModify = time.Now()
	} else if handled, err = gitStatusView.SelectableRowView.HandleAction(action); handled {
		log.Debugf("Action handled by SelectableRowView")
	} else {
		log.Debugf("Action not handled")
	}

	return
}

func (gitStatusView *GitStatusView) selectNextFileEntry() (err error) {
	rows := gitStatusView.rows()
	if rows == 0 {
		return
	}

	selectedRowIndex := gitStatusView.activeViewPos.ActiveRowIndex()

	if gitStatusView.isFileEntry(selectedRowIndex) {
		return
	}

	defer gitStatusView.channels.UpdateDisplay()

	for rowIndex := selectedRowIndex + 1; rowIndex < rows; rowIndex++ {
		if gitStatusView.isFileEntry(rowIndex) {
			return gitStatusView.selectEntry(rowIndex)
		}
	}

	if selectedRowIndex > 0 {
		for rowIndex := selectedRowIndex - 1; rowIndex > 0; rowIndex-- {
			if gitStatusView.isFileEntry(rowIndex) {
				return gitStatusView.selectEntry(rowIndex)
			}
		}
	}

	return gitStatusView.selectEntry(0)
}

func selectGitStatusEntry(gitStatusView *GitStatusView, action Action) (err error) {
	if len(gitStatusView.gitStatusViewListeners) == 0 {
		gitStatusView.createGitStatusViewListener()
	} else {
		viewPos := gitStatusView.viewPos()
		gitStatusView.selectEntry(viewPos.ActiveRowIndex())
	}

	return
}

func stageFile(gitStatusView *GitStatusView, action Action) (err error) {
	if gitStatusView.rows() == 0 {
		return
	}

	renderedStatus := gitStatusView.renderedStatus
	statusEntry := renderedStatus[gitStatusView.activeViewPos.ActiveRowIndex()]

	if !statusEntry.isSelectable() || statusEntry.statusType == StStaged {
		return
	}

	var filePaths []string

	if statusEntry.entryType == rsetFile {
		filePaths = append(filePaths, statusEntry.filePath)
	} else if statusEntry.entryType == rsetHeader {
		filePaths = gitStatusView.status.FilePaths(statusEntry.statusType)
	}

	if len(filePaths) == 0 {
		return
	}

	if err = gitStatusView.repoController.StageFiles(filePaths); err != nil {
		return
	}

	_, err = gitStatusView.SelectableRowView.HandleAction(Action{ActionType: ActionNextLine})
	gitStatusView.channels.UpdateDisplay()

	return
}

func unstageFile(gitStatusView *GitStatusView, action Action) (err error) {
	if gitStatusView.rows() == 0 {
		return
	}

	renderedStatus := gitStatusView.renderedStatus
	statusEntry := renderedStatus[gitStatusView.activeViewPos.ActiveRowIndex()]

	if !statusEntry.isSelectable() || statusEntry.statusType != StStaged {
		return
	}

	var filePaths []string

	if statusEntry.entryType == rsetFile {
		filePaths = append(filePaths, statusEntry.filePath)
	} else if statusEntry.entryType == rsetHeader {
		filePaths = gitStatusView.status.FilePaths(StStaged)
	}

	if len(filePaths) == 0 {
		return
	}

	if err = gitStatusView.repoController.UnstageFiles(filePaths); err != nil {
		return
	}

	_, err = gitStatusView.SelectableRowView.HandleAction(Action{ActionType: ActionPrevLine})
	gitStatusView.channels.UpdateDisplay()

	return
}

func checkoutFile(gitStatusView *GitStatusView, action Action) (err error) {
	if gitStatusView.rows() == 0 {
		return
	}

	renderedStatus := gitStatusView.renderedStatus
	statusEntry := renderedStatus[gitStatusView.activeViewPos.ActiveRowIndex()]

	if !statusEntry.isSelectable() || statusEntry.statusType != StUnstaged {
		return
	}

	var filePaths []string

	if statusEntry.entryType == rsetFile {
		filePaths = append(filePaths, statusEntry.filePath)
	} else if statusEntry.entryType == rsetHeader {
		filePaths = gitStatusView.status.FilePaths(StUnstaged)
	}

	if len(filePaths) == 0 {
		return
	}

	performCheckout := func(paths []string) (err error) {
		if err = gitStatusView.repoController.CheckoutFiles(paths); err != nil {
			return
		}

		gitStatusView.channels.DoAction(Action{ActionType: ActionPrevLine})
		return
	}

	if gitStatusView.config.GetBool(CfConfirmCheckout) {
		var question string
		if len(filePaths) == 1 {
			question = fmt.Sprintf("Are you sure you want to checkout %v?", filepath.Base(filePaths[0]))
		} else {
			question = fmt.Sprintf("Are you sure you want to checkout %v files?", len(filePaths))
		}

		gitStatusView.channels.DoAction(YesNoQuestion(question, func(response QuestionResponse) {
			if response == ResponseYes {
				gitStatusView.channels.ReportError(performCheckout(filePaths))
			}
		}))
	} else {
		err = performCheckout(filePaths)
	}

	return
}

func commit(gitStatusView *GitStatusView, action Action) (err error) {
	if gitStatusView.rows() == 0 || len(gitStatusView.status.FilePaths(StStaged)) == 0 {
		return fmt.Errorf("No files to commit")
	} else if len(gitStatusView.status.FilePaths(StConflicted)) > 0 {
		return fmt.Errorf("Committing is not possible due to unmerged files - Resolve conflicts before commiting")
	}

	gitStatusView.repoController.Commit(func(oid *Oid, err error) {
		if err == nil {
			gitStatusView.channels.ReportStatus("Created commit %v", oid.ShortID())
		} else {
			gitStatusView.channels.ReportError(fmt.Errorf("Commit failed: %v", err))
		}
	})

	return
}

func amendCommit(gitStatusView *GitStatusView, action Action) (err error) {
	if len(gitStatusView.status.FilePaths(StConflicted)) > 0 {
		return fmt.Errorf("Committing is not possible due to unmerged files - Resolve conflicts before commiting")
	}

	gitStatusView.repoController.AmendCommit(func(oid *Oid, err error) {
		if err == nil {
			gitStatusView.channels.ReportStatus("Amended commit. New oid: %v", oid.ShortID())
		} else {
			gitStatusView.channels.ReportError(fmt.Errorf("Amending commit failed: %v", err))
		}
	})

	return
}
