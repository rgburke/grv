package main

import (
	"fmt"
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

		lineBuilder.Append("  ").AppendWithStyle(themeComponentID, "%v", descriptionLine.text)
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

// HelpView displays help information
type HelpView struct {
	*AbstractWindowView
	activeViewPos     ViewPos
	lastViewDimension ViewDimension
	helpSections      []*HelpSection
	totalRows         uint
	lock              sync.Mutex
}

// NewHelpView creates a new instance
func NewHelpView(channels Channels, config Config) *HelpView {
	helpView := &HelpView{
		activeViewPos: NewViewPosition(),
	}

	helpView.AbstractWindowView = NewAbstractWindowView(helpView, channels, config, "help line")

	return helpView
}

// Initialise does nothing
func (helpView *HelpView) Initialise() (err error) {
	helpSections := []*HelpSection{helpView.introductionHelpSection()}
	helpSections = append(helpSections, helpView.config.GenerateHelpSections()...)

	for _, helpSection := range helpSections {
		if err = helpSection.initialise(); err != nil {
			return
		}

		helpView.totalRows += helpSection.rows()
	}

	helpView.helpSections = helpSections

	return
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

	if err = win.SetSelectedRow(viewPos.SelectedRowIndex()+1, true); err != nil {
		return
	}

	win.DrawBorder()

	if err = win.SetTitle(CmpHelpViewTitle, "Help"); err != nil {
		return
	}

	if err = win.SetFooter(CmpHelpViewFooter, "Line %v of %v", viewPos.ActiveRowIndex()+1, viewRows); err != nil {
		return
	}

	return
}

func (helpView *HelpView) renderRow(win RenderWindow, viewStartRowIndex, rowIndex, startColumn uint) (err error) {
	rows := uint(0)
	prevRows := uint(0)

	for _, helpSection := range helpView.helpSections {
		rows += helpSection.rows()

		if rowIndex < rows {
			tableRowIndex := rowIndex - prevRows
			winStartRowIndex := (prevRows - viewStartRowIndex) + 1

			return helpSection.renderRow(win, winStartRowIndex, tableRowIndex, startColumn)
		}

		prevRows = rows
	}

	return fmt.Errorf("Unable to render row with index: %v", rowIndex)
}

func (helpView *HelpView) viewPos() ViewPos {
	return helpView.activeViewPos
}

func (helpView *HelpView) rows() uint {
	return helpView.totalRows
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
	if handled, err = helpView.AbstractWindowView.HandleAction(action); handled {
		log.Debugf("Action handled by AbstractWindowView")
	} else {
		log.Debugf("Action not handled")
	}

	return
}

func (helpView *HelpView) introductionHelpSection() *HelpSection {
	return &HelpSection{
		title: HelpSectionText{text: "Introduction"},
		description: []HelpSectionText{
			HelpSectionText{text: "GRV - Git Repository Viewer - is a TUI for viewing and modifying git repositories."},
			HelpSectionText{text: "The sections below provide a brief overview of the ways to configure and interact with GRV."},
			HelpSectionText{text: "For full documentation please visit: https://github.com/rgburke/grv/blob/master/doc/documentation.md"},
		},
	}
}
