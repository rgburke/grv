package main

import (
	"fmt"
	"sync"

	log "github.com/Sirupsen/logrus"
)

const (
	hvTitleRows = 3
)

// HelpSection contains help information about a specific topic
type HelpSection struct {
	title          string
	description    []string
	tableFormatter *TableFormatter
}

func (helpSection *HelpSection) rows() uint {
	return hvTitleRows + helpSection.descriptionRows() + helpSection.tableFormatter.RenderedRows()
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

		lineBuilder.Append("  ").AppendWithStyle(CmpHelpViewSectionTitle, "%v", helpSection.title)
	}

	return
}

func (helpSection *HelpSection) renderDescription(win RenderWindow, winStartRowIndex, helpSectionRowIndex, startColumn uint) (err error) {
	rowIndex := helpSectionRowIndex - hvTitleRows

	if rowIndex < helpSection.descriptionRows()-1 {
		var lineBuilder *LineBuilder
		if lineBuilder, err = win.LineBuilder(winStartRowIndex+helpSectionRowIndex, startColumn); err != nil {
			return
		}

		lineBuilder.Append("  ").AppendWithStyle(CmpHelpViewSectionDescription, "%v", helpSection.description[rowIndex])
	}

	return
}

func (helpSection *HelpSection) renderRow(win RenderWindow, winStartRowIndex, helpSectionRowIndex, startColumn uint) (err error) {
	if helpSectionRowIndex < hvTitleRows {
		return helpSection.renderTitle(win, winStartRowIndex, helpSectionRowIndex, startColumn)
	} else if helpSectionRowIndex < hvTitleRows+helpSection.descriptionRows() {
		return helpSection.renderDescription(win, winStartRowIndex, helpSectionRowIndex, startColumn)
	}

	tableOffset := hvTitleRows + helpSection.descriptionRows()
	winStartRowIndex += tableOffset
	helpSectionRowIndex -= tableOffset

	return helpSection.tableFormatter.RenderRow(win, winStartRowIndex, helpSectionRowIndex, startColumn, true)
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
	helpView.helpSections = helpView.config.GenerateHelpSections()

	for _, helpSection := range helpView.helpSections {
		helpSection.tableFormatter.SetBorderColumnWidth(2)
		if err = helpSection.tableFormatter.PadCells(true); err != nil {
			return
		}

		helpView.totalRows += helpSection.rows()
	}

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

	if err = win.SetTitle(CmpCommandOutputTitle, "Help"); err != nil {
		return
	}

	if err = win.SetFooter(CmpCommandOutputFooter, "Line %v of %v", viewPos.SelectedRowIndex()+1, viewRows); err != nil {
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
