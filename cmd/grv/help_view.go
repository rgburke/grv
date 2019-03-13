package main

import (
	"fmt"
	"strings"
	"sync"

	log "github.com/Sirupsen/logrus"
)

// HelpSectionText specifies help section text and its style
type HelpSectionText struct {
	text             string
	themeComponentID ThemeComponentID
}

// HelpSection contains help information about a specific topic
type HelpSection struct {
	title          HelpSectionText
	description    []HelpSectionText
	tableFormatter *TableFormatter
}

func (helpSection *HelpSection) initialise() (err error) {
	if helpSection.tableFormatter != nil {
		helpSection.tableFormatter.SetBorderColumnWidth(2)
		err = helpSection.tableFormatter.PadCells(true)
	}

	return
}

func (helpSection *HelpSection) rows() uint {
	rows := helpSection.titleRows() + helpSection.descriptionRows()

	if helpSection.tableFormatter != nil {
		rows += helpSection.tableFormatter.RenderedRows() + 1
	}

	return rows
}

func (helpSection *HelpSection) titleRows() uint {
	if helpSection.title.text != "" {
		return 3
	}

	return 0
}

func (helpSection *HelpSection) descriptionRows() uint {
	rows := uint(len(helpSection.description))

	if rows > 0 {
		rows++
	}

	return rows
}

func (helpSection *HelpSection) renderTitle(win RenderWindow, winStartRowIndex, helpSectionRowIndex, startColumn uint) (err error) {
	if helpSectionRowIndex == 1 {
		var lineBuilder *LineBuilder
		if lineBuilder, err = win.LineBuilder(winStartRowIndex+helpSectionRowIndex, startColumn); err != nil {
			return
		}

		themeComponentID := CmpHelpViewSectionTitle
		if helpSection.title.themeComponentID != CmpNone {
			themeComponentID = helpSection.title.themeComponentID
		}

		lineBuilder.Append("  ").AppendWithStyle(themeComponentID, "%v", helpSection.title.text)
	}

	return
}

func (helpSection *HelpSection) renderDescription(win RenderWindow, winStartRowIndex, helpSectionRowIndex, startColumn uint) (err error) {
	rowIndex := helpSectionRowIndex - helpSection.titleRows()

	if rowIndex < helpSection.descriptionRows()-1 {
		var lineBuilder *LineBuilder
		if lineBuilder, err = win.LineBuilder(winStartRowIndex+helpSectionRowIndex, startColumn); err != nil {
			return
		}

		descriptionLine := helpSection.description[rowIndex]

		themeComponentID := CmpHelpViewSectionDescription
		if descriptionLine.themeComponentID != CmpNone {
			themeComponentID = descriptionLine.themeComponentID
		}

		spacing := "  "
		if themeComponentID == CmpHelpViewSectionCodeBlock {
			spacing += "  "
		}

		lineBuilder.Append(spacing).AppendWithStyle(themeComponentID, "%v", descriptionLine.text)
	}

	return
}

func (helpSection *HelpSection) renderRow(win RenderWindow, winStartRowIndex, helpSectionRowIndex, startColumn uint) (err error) {
	if helpSectionRowIndex < helpSection.titleRows() {
		return helpSection.renderTitle(win, winStartRowIndex, helpSectionRowIndex, startColumn)
	} else if helpSectionRowIndex < helpSection.titleRows()+helpSection.descriptionRows() {
		return helpSection.renderDescription(win, winStartRowIndex, helpSectionRowIndex, startColumn)
	} else if helpSection.tableFormatter != nil {
		tableOffset := helpSection.titleRows() + helpSection.descriptionRows()
		winStartRowIndex += tableOffset
		helpSectionRowIndex -= tableOffset

		if helpSectionRowIndex < helpSection.tableFormatter.RenderedRows() {
			return helpSection.tableFormatter.RenderRow(win, winStartRowIndex, helpSectionRowIndex, startColumn, true)
		}
	}

	return
}

func (helpSection *HelpSection) rowText(helpSectionRowIndex uint) (line string) {
	if helpSectionRowIndex < helpSection.titleRows() {
		if helpSectionRowIndex == 1 {
			return helpSection.title.text
		}
	} else if helpSectionRowIndex < helpSection.titleRows()+helpSection.descriptionRows() {
		descriptionLineIndex := helpSectionRowIndex - helpSection.titleRows()

		if descriptionLineIndex < uint(len(helpSection.description)) {
			return helpSection.description[descriptionLineIndex].text
		}
	} else if helpSection.tableFormatter != nil {
		tableRowIndex := helpSectionRowIndex - (helpSection.titleRows() + helpSection.descriptionRows())

		if tableRowIndex < helpSection.tableFormatter.RenderedRows() {
			return helpSection.tableFormatter.RenderRowText(tableRowIndex)
		}
	}

	return
}

// HelpViewSection represents a help view section and
// its child sections
type HelpViewSection struct {
	title    string
	children []*HelpViewSection
}

// HelpViewIndex is an index of help view sections
type HelpViewIndex struct {
	sections    []*HelpViewSection
	rowMap      map[uint]uint
	titleRowMap map[string]uint
	totalRows   uint
}

// NewHelpViewIndex creates a new instance
func NewHelpViewIndex(helpSections []*HelpSection) *HelpViewIndex {
	var current *HelpViewSection
	sections := []*HelpViewSection{}
	totalRows := uint(0)
	indexRowIndex := uint(2)
	rowMap := map[uint]uint{}
	titleRowMap := map[string]uint{}

	for _, helpSection := range helpSections {
		if helpSection.title.text != "" {
			current = &HelpViewSection{
				title: helpSection.title.text,
			}

			sections = append(sections, current)

			indexRowIndex++
			rowMap[indexRowIndex] = totalRows + 1
			indexRowIndex++
		}

		for descriptionLineIndex, descriptionLine := range helpSection.description {
			if descriptionLine.themeComponentID == CmpHelpViewSectionSubTitle {
				current.children = append(current.children, &HelpViewSection{
					title: descriptionLine.text,
				})

				rowIndex := totalRows + helpSection.titleRows() + uint(descriptionLineIndex)
				titleRowMap[strings.ToLower(descriptionLine.text)] = rowIndex
				rowMap[indexRowIndex] = rowIndex
				indexRowIndex++
			}
		}

		totalRows += helpSection.rows()
	}

	return &HelpViewIndex{
		sections:    sections,
		rowMap:      rowMap,
		titleRowMap: titleRowMap,
		totalRows:   totalRows,
	}
}

func (helpViewIndex *HelpViewIndex) applyOffset(preOffset, postOffset uint) {
	helpViewIndex.totalRows += postOffset

	rowMap := map[uint]uint{}
	for index, rowIndex := range helpViewIndex.rowMap {
		rowMap[index+preOffset] = rowIndex + postOffset
	}

	titleRowMap := map[string]uint{}
	for title, rowIndex := range helpViewIndex.titleRowMap {
		titleRowMap[title] = rowIndex + postOffset
	}

	helpViewIndex.rowMap = rowMap
	helpViewIndex.titleRowMap = titleRowMap
}

func (helpViewIndex *HelpViewIndex) generateHelpSection() *HelpSection {
	description := []HelpSectionText{}

	for _, section := range helpViewIndex.sections {
		description = append(description, HelpSectionText{text: section.title, themeComponentID: CmpHelpViewIndexTitle})

		for _, childSection := range section.children {
			description = append(description, HelpSectionText{text: "  - " + childSection.title, themeComponentID: CmpHelpViewIndexSubTitle})
		}

		description = append(description, HelpSectionText{})
	}

	if len(description) > 0 {
		description = description[:len(description)-1]
	}

	return &HelpSection{
		title:       HelpSectionText{text: "Table Of Contents"},
		description: description,
	}
}

func (helpViewIndex *HelpViewIndex) mappedRow(rowIndex uint) (mappedIndex uint, exists bool) {
	mappedIndex, exists = helpViewIndex.rowMap[rowIndex]
	return
}

func (helpViewIndex *HelpViewIndex) findSection(section string) (rowIndex uint) {
	section = strings.ToLower(section)
	rowIndex, exists := helpViewIndex.titleRowMap[section]
	if exists {
		return
	}

	matches := []uint{}

	for title, mappedRowIndex := range helpViewIndex.titleRowMap {
		if strings.HasPrefix(title, section) {
			return mappedRowIndex
		} else if strings.Contains(title, section) {
			matches = append(matches, mappedRowIndex)
		}
	}

	if len(matches) > 0 {
		return matches[0]
	}

	return
}

type helpViewHandler func(*HelpView, Action) error

// HelpView displays help information
type HelpView struct {
	*AbstractWindowView
	activeViewPos     ViewPos
	lastViewDimension ViewDimension
	helpSections      []*HelpSection
	helpViewIndex     *HelpViewIndex
	handlers          map[ActionType]helpViewHandler
	lock              sync.Mutex
}

// NewHelpView creates a new instance
func NewHelpView(channels Channels, config Config, variables GRVVariableSetter) *HelpView {
	helpView := &HelpView{
		activeViewPos: NewViewPosition(),
		handlers: map[ActionType]helpViewHandler{
			ActionSelect: selectHelpRow,
		},
	}

	helpView.AbstractWindowView = NewAbstractWindowView(helpView, channels, config, variables, &helpView.lock, "help line")

	return helpView
}

// Initialise does nothing
func (helpView *HelpView) Initialise() (err error) {
	helpView.helpSections, helpView.helpViewIndex, err = GenerateHelpView(helpView.config)
	return
}

// SearchHelp will attempt to find a section or part of the help
// that matches the provided search term
func (helpView *HelpView) SearchHelp(searchTerm string) {
	if searchTerm == "" {
		return
	}

	helpView.lock.Lock()
	defer helpView.lock.Unlock()

	if rowIndex := helpView.helpViewIndex.findSection(searchTerm); rowIndex > 0 {
		helpView.viewPos().SetActiveRowIndex(rowIndex)
		helpView.viewPos().ScrollActiveRowTop()
		return
	}

	helpView.AbstractWindowView.HandleAction(Action{
		ActionType: ActionSearch,
		Args:       []interface{}{searchTerm},
	})
}

// ViewID returns the ViewID of the help view
func (helpView *HelpView) ViewID() ViewID {
	return ViewHelp
}

// Render generates help information and writes it to the provided window
func (helpView *HelpView) Render(win RenderWindow) (err error) {
	helpView.lock.Lock()
	defer helpView.lock.Unlock()

	helpView.lastViewDimension = win.ViewDimensions()

	winRows := win.Rows() - 2
	viewPos := helpView.viewPos()

	viewRows := helpView.rows()
	viewPos.DetermineViewStartRow(winRows, viewRows)

	viewStartRowIndex := viewPos.ViewStartRowIndex()
	viewRowIndex := viewStartRowIndex
	startColumn := viewPos.ViewStartColumn()

	for rowIndex := uint(0); rowIndex < winRows && viewRowIndex < viewRows; rowIndex++ {
		if err = helpView.renderRow(win, viewStartRowIndex, viewRowIndex, startColumn); err != nil {
			return
		}

		viewRowIndex++
	}

	if err = win.SetSelectedRow(viewPos.SelectedRowIndex()+1, helpView.viewState); err != nil {
		return
	}

	win.DrawBorder()

	if err = win.SetTitle(CmpHelpViewTitle, "Help"); err != nil {
		return
	}

	if err = win.SetFooter(CmpHelpViewFooter, "Line %v of %v", viewPos.ActiveRowIndex()+1, viewRows); err != nil {
		return
	}

	if searchActive, searchPattern, lastSearchFoundMatch := helpView.viewSearch.SearchActive(); searchActive && lastSearchFoundMatch {
		if err = win.Highlight(searchPattern, CmpAllviewSearchMatch); err != nil {
			return
		}
	}

	return
}

func (helpView *HelpView) renderRow(win RenderWindow, viewStartRowIndex, rowIndex, startColumn uint) (err error) {
	helpSection, helpSectionStartRowIndex, err := helpView.helpSection(rowIndex)
	if err != nil {
		return
	}

	helpSectionRowIndex := rowIndex - helpSectionStartRowIndex
	winStartRowIndex := (helpSectionStartRowIndex - viewStartRowIndex) + 1

	return helpSection.renderRow(win, winStartRowIndex, helpSectionRowIndex, startColumn)
}

func (helpView *HelpView) helpSection(rowIndex uint) (helpSection *HelpSection, helpSectionStartRowIndex uint, err error) {
	rows := uint(0)

	for _, helpSection = range helpView.helpSections {
		rows += helpSection.rows()

		if rowIndex < rows {
			return
		}

		helpSectionStartRowIndex = rows
	}

	err = fmt.Errorf("Unable to find HelpSection with row index: %v", rowIndex)
	return
}

// RenderHelpBar shows key bindings custom to the help view
func (helpView *HelpView) RenderHelpBar(lineBuilder *LineBuilder) (err error) {
	rowIndex := helpView.activeViewPos.ActiveRowIndex()

	if _, exists := helpView.helpViewIndex.mappedRow(rowIndex); exists {
		RenderKeyBindingHelp(helpView.ViewID(), lineBuilder, helpView.config, []ActionMessage{
			{action: ActionSelect, message: "Jump to section"},
		})
	}

	return
}

func (helpView *HelpView) line(lineIndex uint) (line string) {
	helpSection, helpSectionStartRowIndex, err := helpView.helpSection(lineIndex)
	if err != nil {
		return
	}

	helpSectionRowIndex := lineIndex - helpSectionStartRowIndex
	return helpSection.rowText(helpSectionRowIndex)
}

func (helpView *HelpView) viewPos() ViewPos {
	return helpView.activeViewPos
}

func (helpView *HelpView) rows() uint {
	return helpView.helpViewIndex.totalRows
}

func (helpView *HelpView) viewDimension() ViewDimension {
	return helpView.lastViewDimension
}

func (helpView *HelpView) onRowSelected(rowIndex uint) (err error) {
	return
}

// HandleAction handles the action if supported
func (helpView *HelpView) HandleAction(action Action) (err error) {
	helpView.lock.Lock()
	defer helpView.lock.Unlock()

	var handled bool
	if handler, ok := helpView.handlers[action.ActionType]; ok {
		log.Debugf("Action handled by HelpView")
		err = handler(helpView, action)
	} else if handled, err = helpView.AbstractWindowView.HandleAction(action); handled {
		log.Debugf("Action handled by AbstractWindowView")
	} else {
		log.Debugf("Action not handled")
	}

	return
}

func selectHelpRow(helpView *HelpView, action Action) (err error) {
	rowIndex := helpView.activeViewPos.ActiveRowIndex()
	mappedRowIndex, exists := helpView.helpViewIndex.mappedRow(rowIndex)
	if !exists {
		return
	}

	helpView.activeViewPos.SetActiveRowIndex(mappedRowIndex)
	helpView.AbstractWindowView.HandleAction(Action{ActionType: ActionScrollCursorTop})
	helpView.channels.UpdateDisplay()

	return
}

// GenerateHelpView generates the help view
func GenerateHelpView(config Config) (helpSections []*HelpSection, helpViewIndex *HelpViewIndex, err error) {
	introHelpSection := &HelpSection{
		title: HelpSectionText{text: "Introduction"},
		description: []HelpSectionText{
			{text: "GRV - Git Repository Viewer - is a TUI for viewing and modifying Git repositories."},
			{text: "The sections below provide an overview of the ways to configure and interact with GRV."},
		},
	}

	helpSections = append(helpSections, GenerateCommandLineArgumentsHelpSections())
	helpSections = append(helpSections, config.GenerateHelpSections()...)
	helpSections = append(helpSections, GenerateShellCommandHelpSections(config)...)
	helpSections = append(helpSections, GenerateFilterQueryLanguageHelpSections(config)...)

	helpViewIndex = NewHelpViewIndex(helpSections)
	indexHelpSection := helpViewIndex.generateHelpSection()
	helpViewIndex.applyOffset(introHelpSection.rows(), introHelpSection.rows()+indexHelpSection.rows())

	helpSections = append([]*HelpSection{introHelpSection, indexHelpSection}, helpSections...)

	for _, helpSection := range helpSections {
		if err = helpSection.initialise(); err != nil {
			return
		}
	}

	return
}
