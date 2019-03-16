package main

import (
	"fmt"
	"sync"

	log "github.com/Sirupsen/logrus"
)

// GRVVariableView represents the variable view
type GRVVariableView struct {
	*AbstractWindowView
	variables         GRVVariableSetter
	activeViewPos     ViewPos
	lastViewDimension ViewDimension
	lock              sync.Mutex
}

// NewGRVVariableView creates a new instance
func NewGRVVariableView(channels Channels, config Config, variables GRVVariableSetter) *GRVVariableView {
	grvVariableView := &GRVVariableView{
		variables:     variables,
		activeViewPos: NewViewPosition(),
	}

	grvVariableView.AbstractWindowView = NewAbstractWindowView(grvVariableView, channels, config, variables, &grvVariableView.lock, "variable")

	return grvVariableView
}

// Render generates and writes variable view to the provided window
func (grvVariableView *GRVVariableView) Render(win RenderWindow) (err error) {
	grvVariableView.lock.Lock()
	defer grvVariableView.lock.Unlock()

	grvVariableView.lastViewDimension = win.ViewDimensions()

	variables := grvVariableView.variables.VariableValues()
	viewRows := grvVariableView.rows()
	rows := win.Rows() - 2

	viewPos := grvVariableView.viewPos()
	viewPos.DetermineViewStartRow(rows, viewRows)
	viewRowIndex := viewPos.ViewStartRowIndex()
	startColumn := viewPos.ViewStartColumn()

	var lineBuilder *LineBuilder
	for rowIndex := uint(0); rowIndex < rows && viewRowIndex < viewRows; rowIndex++ {
		variable := GRVVariable(viewRowIndex)
		variableName := GRVVariableName(variable)
		variableValue := variables[variable]

		if lineBuilder, err = win.LineBuilder(rowIndex+1, startColumn); err != nil {
			return
		}

		lineBuilder.Append(" ").
			AppendWithStyle(CmpGRVVariableViewVariable, "%v: ", variableName).
			AppendWithStyle(CmpGRVVariableViewValue, "%v", variableValue)

		viewRowIndex++
	}

	if err = win.SetSelectedRow(viewPos.SelectedRowIndex()+1, grvVariableView.viewState); err != nil {
		return
	}

	win.DrawBorder()

	if err = win.SetTitle(CmpGRVVariableViewTitle, "Variables"); err != nil {
		return
	}

	if err = win.SetFooter(CmpGRVVariableViewFooter, "Variable %v of %v", viewPos.ActiveRowIndex()+1, grvVariableView.rows()); err != nil {
		return
	}

	if searchActive, searchPattern, lastSearchFoundMatch := grvVariableView.viewSearch.SearchActive(); searchActive && lastSearchFoundMatch {
		if err = win.Highlight(searchPattern, CmpAllviewSearchMatch); err != nil {
			return
		}
	}

	return
}

// ViewID returns the ViewID for the grv variable view
func (grvVariableView *GRVVariableView) ViewID() ViewID {
	return ViewGRVVariable
}

func (grvVariableView *GRVVariableView) line(lineIndex uint) (line string) {
	if lineIndex < grvVariableView.rows() {
		variables := grvVariableView.variables.VariableValues()

		variable := GRVVariable(lineIndex)
		variableName := GRVVariableName(variable)
		variableValue := variables[variable]

		line = fmt.Sprintf("%v: %v", variableName, variableValue)
	}

	return
}

func (grvVariableView *GRVVariableView) viewPos() ViewPos {
	return grvVariableView.activeViewPos
}

func (grvVariableView *GRVVariableView) rows() uint {
	return uint(VarCount)
}

func (grvVariableView *GRVVariableView) viewDimension() ViewDimension {
	return grvVariableView.lastViewDimension
}

func (grvVariableView *GRVVariableView) onRowSelected(rowIndex uint) (err error) {
	grvVariableView.channels.UpdateDisplay()
	return
}

// HandleAction checks if the grv variable view supports this action and if it does executes it
func (grvVariableView *GRVVariableView) HandleAction(action Action) (err error) {
	grvVariableView.lock.Lock()
	defer grvVariableView.lock.Unlock()

	var handled bool
	if handled, err = grvVariableView.AbstractWindowView.HandleAction(action); handled {
		log.Debugf("Action handled by AbstractWindowView")
	} else {
		log.Debugf("Action not handled")
	}

	return
}
