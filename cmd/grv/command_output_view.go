package main

import (
	"fmt"
	"sync"

	log "github.com/Sirupsen/logrus"
)

type outputLineType int

const (
	oltNormal outputLineType = iota
	oltCommand
	oltError
	oltSuccess
)

var outputLineThemeComponentIDs = map[outputLineType]ThemeComponentID{
	oltNormal:  CmpCommandOutputNormal,
	oltCommand: CmpCommandOutputCommand,
	oltError:   CmpCommandOutputError,
	oltSuccess: CmpCommandOutputSuccess,
}

type outputLine struct {
	line     string
	lineType outputLineType
}

// CommandOutputProcessor receives the output and status of a command
type CommandOutputProcessor interface {
	AddOutputLine(line string)
	OnCommandExecutionError(err error)
	OnCommandComplete(exitCode int)
}

// CommandOutputView is a view for displaying command output
type CommandOutputView struct {
	*AbstractWindowView
	activeViewPos     ViewPos
	lastViewDimension ViewDimension
	outputLines       []outputLine
	lock              sync.Mutex
}

// NewCommandOutputView creates a new instance
func NewCommandOutputView(command string, channels Channels, config Config, variables GRVVariableSetter) *CommandOutputView {
	commandOutputView := &CommandOutputView{
		activeViewPos: NewViewPosition(),
	}

	commandOutputView.AbstractWindowView = NewAbstractWindowView(commandOutputView, channels, config, variables, &commandOutputView.lock, "output line")

	commandOutputView.addOutputLine(outputLine{
		line:     fmt.Sprintf("$ %v", command),
		lineType: oltCommand,
	})

	return commandOutputView
}

// ViewID returns the ViewID of the command output view
func (commandOutputView *CommandOutputView) ViewID() ViewID {
	return ViewCommandOutput
}

// Render generates the comamnd output view and writes it to the provided window
func (commandOutputView *CommandOutputView) Render(win RenderWindow) (err error) {
	commandOutputView.lock.Lock()
	defer commandOutputView.lock.Unlock()

	commandOutputView.lastViewDimension = win.ViewDimensions()

	winRows := win.Rows() - 2
	viewPos := commandOutputView.viewPos()

	viewRows := commandOutputView.rows()
	viewPos.DetermineViewStartRow(winRows, viewRows)

	viewRowIndex := viewPos.ViewStartRowIndex()
	startColumn := viewPos.ViewStartColumn()

	win.ApplyStyle(CmpCommandOutputNormal)

	for rowIndex := uint(0); rowIndex < winRows && viewRowIndex < viewRows; rowIndex++ {
		outputLine := commandOutputView.outputLines[viewRowIndex]
		themeComponentID := outputLineThemeComponentIDs[outputLine.lineType]

		if err = win.SetRow(rowIndex+1, startColumn, themeComponentID, " %v", outputLine.line); err != nil {
			return
		}

		viewRowIndex++
	}

	if err = win.SetSelectedRow(viewPos.SelectedRowIndex()+1, ViewStateActive); err != nil {
		return
	}

	win.DrawBorderWithStyle(CmpCommandOutputNormal)

	if err = win.SetTitle(CmpCommandOutputTitle, "Command Output"); err != nil {
		return
	}

	if err = win.SetFooter(CmpCommandOutputFooter, "Line %v of %v", viewPos.SelectedRowIndex()+1, viewRows); err != nil {
		return
	}

	return
}

// RenderHelpBar renders a help message for the command output view
func (commandOutputView *CommandOutputView) RenderHelpBar(lineBuilder *LineBuilder) (err error) {
	quitKeys := commandOutputView.config.KeyStrings(ActionRemoveView, ViewHierarchy{ViewCommandOutput, ViewAll})

	if len(quitKeys) > 0 {
		quitKeyText := fmt.Sprintf("Press %v to close command output", quitKeys[len(quitKeys)-1].keystring)
		lineBuilder.AppendWithStyle(CmpHelpbarviewSpecial, " %v", quitKeyText)
	}

	return
}

// AddOutputLine receives a line of command output
func (commandOutputView *CommandOutputView) AddOutputLine(line string) {
	commandOutputView.lock.Lock()
	defer commandOutputView.lock.Unlock()

	commandOutputView.addOutputLine(outputLine{
		line:     line,
		lineType: oltNormal,
	})
}

// OnCommandExecutionError is called when command execution has failed
func (commandOutputView *CommandOutputView) OnCommandExecutionError(err error) {
	commandOutputView.lock.Lock()
	defer commandOutputView.lock.Unlock()

	commandOutputView.addOutputLine(
		outputLine{
			line:     "",
			lineType: oltNormal,
		},
		outputLine{
			line:     fmt.Sprintf("Command execution failed: %v", err),
			lineType: oltError,
		},
	)
}

// OnCommandComplete is called when a command has completed and it's exit status is available
func (commandOutputView *CommandOutputView) OnCommandComplete(exitCode int) {
	commandOutputView.lock.Lock()
	defer commandOutputView.lock.Unlock()

	var lineType outputLineType

	if exitCode == 0 {
		lineType = oltSuccess
	} else {
		lineType = oltError
	}

	commandOutputView.addOutputLine(
		outputLine{
			line:     "",
			lineType: oltNormal,
		},
		outputLine{
			line:     fmt.Sprintf("Command exited with status %v", exitCode),
			lineType: lineType,
		},
	)
}

func (commandOutputView *CommandOutputView) addOutputLine(outputLines ...outputLine) {
	commandOutputView.outputLines = append(commandOutputView.outputLines, outputLines...)
	commandOutputView.activeViewPos.SetActiveRowIndex(commandOutputView.rows() - 1)
	commandOutputView.channels.UpdateDisplay()
}

func (commandOutputView *CommandOutputView) viewPos() ViewPos {
	return commandOutputView.activeViewPos
}

func (commandOutputView *CommandOutputView) rows() uint {
	return uint(len(commandOutputView.outputLines))
}

func (commandOutputView *CommandOutputView) viewDimension() ViewDimension {
	return commandOutputView.lastViewDimension
}

func (commandOutputView *CommandOutputView) onRowSelected(rowIndex uint) (err error) {
	return
}

func (commandOutputView *CommandOutputView) line(lineIndex uint) (line string) {
	if lineIndex < commandOutputView.rows() {
		line = commandOutputView.outputLines[lineIndex].line
	}

	return
}

// HandleAction handles the action if supported
func (commandOutputView *CommandOutputView) HandleAction(action Action) (err error) {
	commandOutputView.lock.Lock()
	defer commandOutputView.lock.Unlock()

	var handled bool
	if handled, err = commandOutputView.AbstractWindowView.HandleAction(action); handled {
		log.Debugf("Action handled by AbstractWindowView")
	} else {
		log.Debugf("Action not handled")
	}

	return
}
