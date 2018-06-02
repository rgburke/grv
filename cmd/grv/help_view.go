package main

import (
	"fmt"
	"sync"

	log "github.com/Sirupsen/logrus"
)

// HelpTable contains help information in a tabular format
type HelpTable struct {
	title          string
	tableFormatter *TableFormatter
}

func (helpTable *HelpTable) rows() uint {
	return helpTable.tableFormatter.RenderedRows()
}

// HelpView displays help information
type HelpView struct {
	*AbstractWindowView
	activeViewPos     ViewPos
	lastViewDimension ViewDimension
	helpTables        []*HelpTable
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
	helpView.helpTables = helpView.config.GenerateHelpTables()

	for _, helpTable := range helpView.helpTables {
		helpView.totalRows += helpTable.rows()
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

	for _, helpTable := range helpView.helpTables {
		rows += helpTable.rows()

		if rowIndex < rows {
			tableRowIndex := rowIndex - prevRows
			winStartRowIndex := (prevRows - viewStartRowIndex) + 1

			return helpTable.tableFormatter.RenderRow(win, winStartRowIndex, tableRowIndex, startColumn, true)
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
