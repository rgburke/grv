package main

import (
	"bufio"
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"sync"

	log "github.com/Sirupsen/logrus"
)

type diffViewHandler func(*DiffView, Action) error

type diffLineType int

const (
	dltUnset diffLineType = iota
	dltNormal
	dltDiffCommitAuthor
	dltDiffCommitAuthorDate
	dltDiffCommitCommitter
	dltDiffCommitCommitterDate
	dltDiffCommitMessage
	dltDiffStatsFile
	dltDiffStatsSummary
	dltGitDiffHeaderDiff
	dltGitDiffHeaderIndex
	dltGitDiffHeaderNewFile
	dltGitDiffHeaderOldFile
	dltGitDiffHeaderNewMode
	dltGitDiffHeaderOldMode
	dltGitDiffHeaderNewFileMode
	dltGitDiffHeaderDeletedFileMode
	dltGitDiffHeaderSimilarityIndex
	dltGitDiffHeaderRenameFrom
	dltGitDiffHeaderRenameTo
	dltGitDiffHeaderBinaryFile
	dltHunkStart
	dltLineAdded
	dltLineRemoved
)

const (
	dvDateFormat                 = "Mon Jan 2 15:04:05 2006 -0700"
	dvDiffLoadRequestChannelSize = 100
)

var diffStatsSummaryRegex = regexp.MustCompile(`^\d+\sfiles?\schanged,`)

type diffLineSection struct {
	text             string
	char             AcsChar
	themeComponentID ThemeComponentID
}

type diffLineData struct {
	sections []*diffLineSection
	line     string
	lineType diffLineType
}

func newEmptyDiffLineData() *diffLineData {
	return newNormalDiffLineData("")
}

func newNormalDiffLineData(line string) *diffLineData {
	return newDiffLineData(line, dltNormal, CmpDiffviewDifflineNormal)
}

func newDiffLineData(line string, lineType diffLineType, themeComponentID ThemeComponentID) *diffLineData {
	sections := []*diffLineSection{
		&diffLineSection{
			text:             line,
			themeComponentID: themeComponentID,
		},
	}

	return newSectionedDiffLineData(sections, lineType)
}

func newSectionedDiffLineData(sections []*diffLineSection, lineType diffLineType) *diffLineData {
	diffLine := &diffLineData{
		sections: sections,
		lineType: lineType,
	}

	var buf bytes.Buffer
	for _, section := range sections {
		buf.WriteString(section.text)
	}
	diffLine.line = buf.String()

	return diffLine
}

type diffLines struct {
	rawLines []*diffLineData
	lines    []*diffLineData
	diffType diffProcessorType
	viewPos  ViewPos
}

type diffLoadRequest interface {
	diffID() diffID
}

type commitDiffLoadRequest struct {
	commit *Commit
}

func (commitDiffLoadRequest *commitDiffLoadRequest) diffID() diffID {
	return diffID(commitDiffLoadRequest.commit.oid.String())
}

type fileDiffLoadRequest struct {
	statusType StatusType
	filePath   string
}

func (fileDiffLoadRequest *fileDiffLoadRequest) diffID() diffID {
	return diffID(fileDiffLoadRequest.filePath)
}

type stageDiffLoadRequest struct {
	statusType StatusType
}

func (stageDiffLoadRequest *stageDiffLoadRequest) diffID() diffID {
	return diffID(fmt.Sprintf("%v files", strings.ToLower(StatusTypeDisplayName(stageDiffLoadRequest.statusType))))
}

type diffID string

type diffProcessor interface {
	processDiff([]*diffLineData) ([]*diffLineData, error)
}

type gitDiffProcessor struct{}

func (gitDiffProcessor *gitDiffProcessor) processDiff(lines []*diffLineData) ([]*diffLineData, error) {
	return lines, nil
}

type diffProcessorType int

const (
	dptGit diffProcessorType = iota
	dptFancy
)

var diffProcessorNames = map[string]diffProcessorType{
	"git":   dptGit,
	"fancy": dptFancy,
}

var diffProcessors = map[diffProcessorType]diffProcessor{
	dptGit:   &gitDiffProcessor{},
	dptFancy: &fancyDiffProcessor{},
}

// IsValidDiffProcessorName returns true if there exists a diff processor with the
// specified name
func IsValidDiffProcessorName(diffProcessorName string) (isValid bool) {
	_, isValid = diffProcessorNames[diffProcessorName]
	return
}

// DiffView contains all state for the diff view
type DiffView struct {
	*AbstractWindowView
	channels          Channels
	repoData          RepoData
	config            Config
	lastRequestedDiff diffID
	activeDiff        diffID
	diffs             map[diffID]*diffLines
	activeViewPos     ViewPos
	lastViewDimension ViewDimension
	handlers          map[ActionType]diffViewHandler
	diffLoadRequestCh chan diffLoadRequest
	variables         GRVVariableSetter
	waitGroup         sync.WaitGroup
	lock              sync.Mutex
}

// NewDiffView creates a new diff view instance
func NewDiffView(repoData RepoData, channels Channels, config Config, variables GRVVariableSetter) *DiffView {
	diffView := &DiffView{
		repoData:          repoData,
		channels:          channels,
		config:            config,
		activeViewPos:     NewViewPosition(),
		diffs:             make(map[diffID]*diffLines),
		diffLoadRequestCh: make(chan diffLoadRequest, dvDiffLoadRequestChannelSize),
		variables:         variables,
		handlers: map[ActionType]diffViewHandler{
			ActionSelect: selectDiffLine,
		},
	}

	diffView.AbstractWindowView = NewAbstractWindowView(diffView, channels, config, variables, &diffView.lock, "diff line")

	return diffView
}

// Initialise does nothing
func (diffView *DiffView) Initialise() (err error) {
	log.Info("Initialising DiffView")

	diffView.lock.Lock()
	defer diffView.lock.Unlock()

	diffView.waitGroup.Add(1)
	go diffView.processDiffLoadRequests()

	diffView.config.AddOnChangeListener(CfDiffDisplay, diffView)

	return
}

// Dispose of any resources held by the view
func (diffView *DiffView) Dispose() {
	diffView.lock.Lock()

	close(diffView.diffLoadRequestCh)
	diffView.diffLoadRequestCh = nil

	diffView.lock.Unlock()

	diffView.waitGroup.Wait()
}

// Render generates and writes the diff view to the provided window
func (diffView *DiffView) Render(win RenderWindow) (err error) {
	diffView.lock.Lock()
	defer diffView.lock.Unlock()

	diffView.lastViewDimension = win.ViewDimensions()

	if diffView.activeDiff == "" {
		return diffView.AbstractWindowView.renderEmptyView(win, "No diff to display")
	} else if diffView.activeDiff != diffView.lastRequestedDiff {
		return diffView.AbstractWindowView.renderEmptyView(win, "Loading diff...")
	}

	rows := win.Rows() - 2
	viewPos := diffView.activeViewPos
	diffLines, ok := diffView.diffs[diffView.activeDiff]
	if !ok {
		log.Errorf("No diff data found for %v", diffView.activeDiff)
		return
	}

	lineNum := uint(len(diffLines.lines))
	viewPos.DetermineViewStartRow(rows, lineNum)

	lineIndex := viewPos.ViewStartRowIndex()
	startColumn := viewPos.ViewStartColumn()

	var lineBuilder *LineBuilder
	for rowIndex := uint(0); rowIndex < rows && lineIndex < lineNum; rowIndex++ {
		diffLine := diffLines.lines[lineIndex]

		if lineBuilder, err = win.LineBuilder(rowIndex+1, startColumn); err != nil {
			return
		}

		lineBuilder.Append(" ")
		for _, section := range diffLine.sections {
			if section.char != 0 {
				lineBuilder.AppendACSChar(section.char, section.themeComponentID)
			} else {
				lineBuilder.AppendWithStyle(section.themeComponentID, section.text)
			}
		}

		lineIndex++
	}

	if err = win.SetSelectedRow(viewPos.SelectedRowIndex()+1, diffView.viewState); err != nil {
		return
	}

	win.DrawBorder()

	if err = win.SetTitle(CmpDiffviewTitle, "Diff for %v", diffView.activeDiff); err != nil {
		return
	}

	if err = win.SetFooter(CmpDiffviewFooter, "Line %v of %v", viewPos.ActiveRowIndex()+1, lineNum); err != nil {
		return
	}

	if searchActive, searchPattern, lastSearchFoundMatch := diffView.viewSearch.SearchActive(); searchActive && lastSearchFoundMatch {
		if err = win.Highlight(searchPattern, CmpAllviewSearchMatch); err != nil {
			return
		}
	}

	return
}

// RenderHelpBar renders help information for the diff view
func (diffView *DiffView) RenderHelpBar(lineBuilder *LineBuilder) (err error) {
	diffView.lock.Lock()
	defer diffView.lock.Unlock()

	if diffView.activeDiff == "" {
		return
	}

	diffLines, ok := diffView.diffs[diffView.activeDiff]
	if !ok {
		return
	}

	lineIndex := diffView.activeViewPos.ActiveRowIndex()

	if lineIndex < uint(len(diffLines.lines)) {
		line := diffLines.lines[lineIndex]

		if line.lineType == dltDiffStatsFile {
			RenderKeyBindingHelp(diffView.ViewID(), lineBuilder, diffView.config, []ActionMessage{
				{action: ActionSelect, message: "Jump to file diff"},
			})
		}
	}

	return
}

// OnStateChange updates whether this view is currently active
func (diffView *DiffView) OnStateChange(viewState ViewState) {
	diffView.AbstractWindowView.OnStateChange(viewState)

	diffView.lock.Lock()
	defer diffView.lock.Unlock()

	if viewState != ViewStateInvisible {
		diffView.setVariables()
	}
}

// ViewID returns the diff views ID
func (diffView *DiffView) ViewID() ViewID {
	return ViewDiff
}

// OnCommitSelected loads/fetches the diff for the selected commit and refreshes the display
func (diffView *DiffView) OnCommitSelected(commit *Commit) (err error) {
	log.Debugf("DiffView loading diff for selected commit %v", commit.commit.Id())

	diffID := diffID(commit.oid.String())

	diffView.lock.Lock()
	diffView.lastRequestedDiff = diffID

	if diffView.switchToDiffIfExists(diffID) {
		diffView.lock.Unlock()
		return
	}

	diffView.lock.Unlock()

	diffView.addDiffLoadRequest(&commitDiffLoadRequest{
		commit: commit,
	})

	return
}

func (diffView *DiffView) switchToDiffIfExists(diffID diffID) (exists bool) {
	diffLines, exists := diffView.diffs[diffID]
	if exists {
		if diffLines.diffType != diffView.currentDiffProcessorType() {
			diffProcessor, diffType := diffView.currentDiffProcessor()
			lines, err := diffProcessor.processDiff(diffLines.rawLines)

			if err != nil {
				log.Errorf("Failed to convert diff to format %v: %v", diffType, err)
				diffLines.diffType = dptGit
				diffLines.lines = diffLines.rawLines
			} else {
				diffLines.diffType = diffType
				diffLines.lines = lines
			}
		}

		diffView.activeDiff = diffID
		diffView.activeViewPos = diffLines.viewPos
		diffView.setVariables()
		diffView.channels.UpdateDisplay()
	}

	return
}

// OnFileSelected loads/fetches the diff for the selected file and refreshes the display
func (diffView *DiffView) OnFileSelected(statusType StatusType, filePath string) {
	log.Debugf("DiffView loading diff for file %v", filePath)

	request := &fileDiffLoadRequest{
		statusType: statusType,
		filePath:   filePath,
	}

	diffView.lock.Lock()
	diffView.lastRequestedDiff = request.diffID()
	diffView.lock.Unlock()

	diffView.addDiffLoadRequest(request)
}

// OnStageGroupSelected does nothing
func (diffView *DiffView) OnStageGroupSelected(statusType StatusType) {
	log.Debugf("DiffView loading diff for stage %v", statusType)

	request := &stageDiffLoadRequest{
		statusType: statusType,
	}

	diffView.lock.Lock()
	diffView.lastRequestedDiff = request.diffID()
	diffView.lock.Unlock()

	diffView.addDiffLoadRequest(request)
}

// OnNoEntrySelected clears the diff view
func (diffView *DiffView) OnNoEntrySelected() {
	log.Debugf("No entry selected to display diff for")

	diffView.lock.Lock()
	defer diffView.lock.Unlock()

	diffView.activeDiff = diffID("")
	diffView.channels.UpdateDisplay()
}

func (diffView *DiffView) addDiffLoadRequest(request diffLoadRequest) {
	if diffView.diffLoadRequestCh != nil {
		diffView.diffLoadRequestCh <- request
	}
}

func (diffView *DiffView) processDiffLoadRequests() {
	defer diffView.waitGroup.Done()

	for request := range diffView.diffLoadRequestCh {
		request = diffView.retrieveLatestDiffLoadRequest(request)
		var err error

		switch req := request.(type) {
		case *commitDiffLoadRequest:
			err = diffView.loadCommitDiffAndMakeActive(req)
		case *fileDiffLoadRequest:
			err = diffView.loadFileDiffAndMakeActive(req)
		case *stageDiffLoadRequest:
			err = diffView.loadStageDiffAndMakeActive(req)
		default:
			log.Errorf("Unknown diff load request type: %T", request)
		}

		if err != nil {
			diffView.channels.ReportError(err)
		}

		diffView.channels.UpdateDisplay()
	}
}

func (diffView *DiffView) retrieveLatestDiffLoadRequest(request diffLoadRequest) diffLoadRequest {
	requestFound := true

	for requestFound {
		select {
		case request = <-diffView.diffLoadRequestCh:
		default:
			requestFound = false
		}
	}

	return request
}

func (diffView *DiffView) loadCommitDiffAndMakeActive(request *commitDiffLoadRequest) (err error) {
	commit := request.commit

	lines, err := diffView.generateDiffLinesForCommit(commit)
	if err != nil {
		log.Errorf("Unable to store commit diff: %v", err)
		return
	}

	diffView.storeDiff(request.diffID(), lines)

	return
}

func (diffView *DiffView) loadFileDiffAndMakeActive(request *fileDiffLoadRequest) (err error) {
	statusType := request.statusType
	filePath := request.filePath

	diff, err := diffView.repoData.DiffFile(statusType, filePath)
	if err != nil {
		log.Errorf("Unable to load file diff: %v", err)
		return
	}

	lines, err := diffView.generateDiffLinesForDiff(diff)
	if err != nil {
		log.Errorf("Unable to store file diff: %v", err)
		return
	}

	diffView.storeDiff(request.diffID(), lines)

	return
}

func (diffView *DiffView) loadStageDiffAndMakeActive(request *stageDiffLoadRequest) (err error) {
	statusType := request.statusType

	diff, err := diffView.repoData.DiffStage(statusType)
	if err != nil {
		log.Errorf("Unable to load diff for stage %v: %v", statusType, err)
		return
	}

	lines, err := diffView.generateDiffLinesForDiff(diff)
	if err != nil {
		log.Errorf("Unable to store stage diff: %v", err)
		return
	}

	diffView.storeDiff(request.diffID(), lines)

	return
}

func (diffView *DiffView) storeDiff(diffID diffID, lines []*diffLineData) {
	diffView.lock.Lock()
	defer diffView.lock.Unlock()

	rawLines := lines
	diffProcessor, diffType := diffView.currentDiffProcessor()
	lines, err := diffProcessor.processDiff(lines)
	if err != nil {
		log.Errorf("Failed to convert diff to format %v: %v", diffType, err)
		lines = rawLines
	}

	diffLines := &diffLines{
		rawLines: rawLines,
		lines:    lines,
		diffType: diffType,
		viewPos:  NewViewPosition(),
	}

	diffView.diffs[diffID] = diffLines

	if diffID != diffView.lastRequestedDiff {
		return
	}

	diffView.activeDiff = diffID
	diffView.activeViewPos = diffLines.viewPos
	diffView.setVariables()

	return
}

func (diffView *DiffView) currentDiffProcessorType() diffProcessorType {
	diffDisplay := diffView.config.GetString(CfDiffDisplay)
	if diffType, exists := diffProcessorNames[diffDisplay]; exists {
		return diffType
	}

	return dptGit
}

func (diffView *DiffView) currentDiffProcessor() (diffProcessor, diffProcessorType) {
	diffType := diffView.currentDiffProcessorType()
	return diffProcessors[diffType], diffType
}

func (diffView *DiffView) generateDiffLinesForCommit(commit *Commit) (lines []*diffLineData, err error) {
	author := commit.commit.Author()
	committer := commit.commit.Committer()

	lines = append(lines,
		newDiffLineData(
			fmt.Sprintf("Author:\t%v <%v>", author.Name, author.Email),
			dltDiffCommitAuthor,
			CmpDiffviewDifflineDiffCommitAuthor,
		),
		newDiffLineData(
			fmt.Sprintf("AuthorDate:\t%v", author.When.Format(dvDateFormat)),
			dltDiffCommitAuthorDate,
			CmpDiffviewDifflineDiffCommitAuthorDate,
		),
		newDiffLineData(
			fmt.Sprintf("Committer:\t%v <%v>", committer.Name, committer.Email),
			dltDiffCommitCommitter,
			CmpDiffviewDifflineDiffCommitCommitter,
		),
		newDiffLineData(
			fmt.Sprintf("CommitterDate:\t%v", committer.When.Format(dvDateFormat)),
			dltDiffCommitCommitterDate,
			CmpDiffviewDifflineDiffCommitCommitterDate,
		),
		newEmptyDiffLineData(),
	)

	commitMessageScanner := bufio.NewScanner(strings.NewReader(commit.commit.Message()))

	for commitMessageScanner.Scan() {
		lines = append(lines, newDiffLineData(
			commitMessageScanner.Text(),
			dltDiffCommitMessage,
			CmpDiffviewDifflineDiffCommitMessage),
		)
	}

	lines = append(lines, newEmptyDiffLineData())

	diff, err := diffView.repoData.DiffCommit(commit)
	if err != nil {
		return
	}

	diffContent, err := diffView.generateDiffLinesForDiff(diff)
	if err != nil {
		return
	}

	lines = append(lines, diffContent...)

	return
}

func (diffView *DiffView) generateDiffLinesForDiff(diff *Diff) (lines []*diffLineData, err error) {
	scanner := bufio.NewScanner(&diff.stats)

	for scanner.Scan() {
		line := scanner.Text()

		if diffStatsSummaryRegex.MatchString(line) {
			lines = append(lines, newDiffLineData(line, dltDiffStatsSummary, CmpDiffviewDifflineNormal))
			continue
		}

		filePart, changePart, err := diffView.splitDiffStatsFileLine(line)
		if err != nil {
			log.Warnf("Diff stats line in unexpected format: %v - %v", line, err)
			continue
		}

		sections := []*diffLineSection{
			&diffLineSection{
				text:             filePart + " |",
				themeComponentID: CmpDiffviewDifflineDiffStatsFile,
			},
		}

		if strings.Contains(changePart, " -> ") {
			sections = append(sections, &diffLineSection{
				text:             changePart,
				themeComponentID: CmpDiffviewDifflineNormal,
			})
		} else {
			for _, char := range changePart {
				switch char {
				case '+':
					sections = append(sections, &diffLineSection{
						text:             "+",
						themeComponentID: CmpDiffviewDifflineLineAdded,
					})
				case '-':
					sections = append(sections, &diffLineSection{
						text:             "-",
						themeComponentID: CmpDiffviewDifflineLineRemoved,
					})
				default:
					sections = append(sections, &diffLineSection{
						text:             fmt.Sprintf("%c", char),
						themeComponentID: CmpDiffviewDifflineNormal,
					})
				}
			}
		}

		lines = append(lines, newSectionedDiffLineData(sections, dltDiffStatsFile))
	}

	if len(lines) > 0 {
		lines = append(lines, newEmptyDiffLineData())
	}

	scanner = bufio.NewScanner(&diff.diffText)

	for scanner.Scan() {
		line := scanner.Text()
		var diffLine *diffLineData

		switch {
		case strings.HasPrefix(line, "diff --git"):
			diffLine = newDiffLineData(line, dltGitDiffHeaderDiff, CmpDiffviewDifflineGitDiffHeader)
		case strings.HasPrefix(line, "new mode"):
			diffLine = newDiffLineData(line, dltGitDiffHeaderNewMode, CmpDiffviewDifflineNormal)
		case strings.HasPrefix(line, "old mode"):
			diffLine = newDiffLineData(line, dltGitDiffHeaderOldMode, CmpDiffviewDifflineNormal)
		case strings.HasPrefix(line, "new file mode"):
			diffLine = newDiffLineData(line, dltGitDiffHeaderNewFileMode, CmpDiffviewDifflineNormal)
		case strings.HasPrefix(line, "deleted file mode"):
			diffLine = newDiffLineData(line, dltGitDiffHeaderDeletedFileMode, CmpDiffviewDifflineNormal)
		case strings.HasPrefix(line, "similarity index"):
			diffLine = newDiffLineData(line, dltGitDiffHeaderSimilarityIndex, CmpDiffviewDifflineNormal)
		case strings.HasPrefix(line, "rename from"):
			diffLine = newDiffLineData(line, dltGitDiffHeaderRenameFrom, CmpDiffviewDifflineNormal)
		case strings.HasPrefix(line, "rename to"):
			diffLine = newDiffLineData(line, dltGitDiffHeaderRenameTo, CmpDiffviewDifflineNormal)
		case strings.HasPrefix(line, "Binary files"):
			diffLine = newDiffLineData(line, dltGitDiffHeaderBinaryFile, CmpDiffviewDifflineNormal)
		case strings.HasPrefix(line, "index"):
			diffLine = newDiffLineData(line, dltGitDiffHeaderIndex, CmpDiffviewDifflineGitDiffExtendedHeader)
		case strings.HasPrefix(line, "+++"):
			diffLine = newDiffLineData(line, dltGitDiffHeaderNewFile, CmpDiffviewDifflineUnifiedDiffHeader)
		case strings.HasPrefix(line, "---"):
			diffLine = newDiffLineData(line, dltGitDiffHeaderOldFile, CmpDiffviewDifflineUnifiedDiffHeader)
		case strings.HasPrefix(line, "@@"):
			if lineParts := strings.SplitAfter(line, "@@"); len(lineParts) != 3 {
				log.Warnf("Unable to handle hunk header line: %v", line)
				diffLine = newDiffLineData(line, dltHunkStart, CmpDiffviewDifflineHunkStart)
			} else {
				sections := []*diffLineSection{
					&diffLineSection{
						text:             strings.Join(lineParts[:2], ""),
						themeComponentID: CmpDiffviewDifflineHunkStart,
					},
					&diffLineSection{
						text:             lineParts[2],
						themeComponentID: CmpDiffviewDifflineHunkHeader,
					},
				}

				diffLine = newSectionedDiffLineData(sections, dltHunkStart)
			}
		case strings.HasPrefix(line, "+"):
			diffLine = newDiffLineData(line, dltLineAdded, CmpDiffviewDifflineLineAdded)
		case strings.HasPrefix(line, "-"):
			diffLine = newDiffLineData(line, dltLineRemoved, CmpDiffviewDifflineLineRemoved)
		default:
			diffLine = newNormalDiffLineData(line)
		}

		lines = append(lines, diffLine)
	}

	return
}

// HandleEvent does nothing
func (diffView *DiffView) HandleEvent(event Event) (err error) {
	return
}

func (diffView *DiffView) viewPos() ViewPos {
	return diffView.activeViewPos
}

func (diffView *DiffView) onConfigVariableChange(configVariable ConfigVariable) {
	diffView.lock.Lock()
	defer diffView.lock.Unlock()

	switch configVariable {
	case CfDiffDisplay:
		diffView.switchToDiffIfExists(diffView.activeDiff)
	}
}

// HandleAction checks if the diff view supports the provided action and executes it if so
func (diffView *DiffView) HandleAction(action Action) (err error) {
	log.Debugf("DiffView handling action %v", action)
	diffView.lock.Lock()
	defer diffView.lock.Unlock()

	var handled bool
	if handler, ok := diffView.handlers[action.ActionType]; ok {
		log.Debugf("Action handled by DiffView")
		err = handler(diffView, action)
	} else if handled, err = diffView.AbstractWindowView.HandleAction(action); handled {
		log.Debugf("Action handled by AbstractWindowView")
	} else {
		log.Debugf("Action not handled")
	}

	return
}

func (diffView *DiffView) line(lineIndex uint) (line string) {
	diffLines, ok := diffView.diffs[diffView.activeDiff]
	if !ok {
		return
	}

	lineNum := uint(len(diffLines.lines))

	if lineIndex >= lineNum {
		log.Errorf("Invalid lineIndex: %v", lineIndex)
		return
	}

	diffLine := diffLines.lines[lineIndex]
	line = diffLine.line

	return
}

func (diffView *DiffView) rows() uint {
	diffLines, ok := diffView.diffs[diffView.activeDiff]
	if !ok {
		return 0
	}

	return uint(len(diffLines.lines))
}

func (diffView *DiffView) viewDimension() ViewDimension {
	return diffView.lastViewDimension
}

func (diffView *DiffView) onRowSelected(rowIndex uint) (err error) {
	diffView.setVariables()
	return
}

func (diffView *DiffView) setVariables() {
	diffView.AbstractWindowView.setVariables()

	rowIndex := diffView.viewPos().ActiveRowIndex()
	if rowIndex >= diffView.rows() {
		return
	}

	diffLines, ok := diffView.diffs[diffView.activeDiff]
	if !ok {
		return
	}

	diffLine := diffLines.lines[rowIndex]

	if diffLine.lineType == dltDiffStatsFile {
		filePart, _, err := diffView.splitDiffStatsFileLine(diffLine.line)
		if err != nil {
			log.Errorf("Unable to set variables for diff view: %v", err)
			return
		}

		filePart = strings.TrimRight(filePart, " ")
		diffView.variables.SetViewVariable(VarFile, filePart, diffView.viewState)

		if diffView.viewState != ViewStateInvisible {
			diffView.variables.SetViewVariable(VarDiffViewFile, filePart, diffView.viewState)
		}
	} else if diffView.viewState != ViewStateInvisible {
		diffView.variables.ClearViewVariable(VarDiffViewFile, diffView.viewState)
	}

	return
}

func (diffView *DiffView) splitDiffStatsFileLine(line string) (filePart, changePart string, err error) {
	sepIndex := strings.LastIndex(line, "|")

	if sepIndex == -1 || sepIndex >= len(line)-1 {
		err = fmt.Errorf("Unable to determine file path from line: %v", line)
		return
	}

	filePart = line[0:sepIndex]
	changePart = line[sepIndex+1:]

	return
}

func selectDiffLine(diffView *DiffView, action Action) (err error) {
	diffLines, ok := diffView.diffs[diffView.activeDiff]
	if !ok {
		return
	}

	lineIndex := diffView.activeViewPos.ActiveRowIndex()
	if lineIndex >= diffView.rows() {
		return
	}

	diffLine := diffLines.lines[lineIndex]

	if diffLine.lineType != dltDiffStatsFile {
		return
	}

	filePart, _, err := diffView.splitDiffStatsFileLine(diffLine.line)
	if err != nil {
		return
	}

	filePart = strings.TrimRight(filePart, " ")

	for lineIndex++; lineIndex < uint(len(diffLines.lines)); lineIndex++ {
		diffLine = diffLines.lines[lineIndex]

		if diffLine.lineType == dltGitDiffHeaderDiff && strings.Contains(diffLine.line, filePart) {
			break
		}
	}

	if lineIndex >= uint(len(diffLines.lines)) {
		return fmt.Errorf("Unable to find diff for file: %v", filePart)
	}

	diffView.activeViewPos.SetActiveRowIndex(lineIndex)
	diffView.setVariables()
	defer diffView.channels.UpdateDisplay()

	_, err = diffView.AbstractWindowView.HandleAction(Action{ActionType: ActionCenterView})
	return
}
