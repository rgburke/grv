package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"

	log "github.com/Sirupsen/logrus"
)

// ShellCommandOutputType represents a way command output can be handled
type ShellCommandOutputType int

// The set of supported command output types
const (
	NoOutput ShellCommandOutputType = iota
	WindowOutput
	TerminalOutput
	StatusBarOutput
)

// ShellCommandProcessor processes a shell command before executing it
type ShellCommandProcessor struct {
	channels     Channels
	variables    GRVVariableGetter
	command      []rune
	outputType   ShellCommandOutputType
	variableRefs []*variableReference
}

type commandPosition struct {
	startIndex int
	endIndex   int
}

type variableReference struct {
	variable GRVVariable
	position *commandPosition
}

// NewShellCommandProcessor creates a new instance
func NewShellCommandProcessor(channels Channels, variables GRVVariableGetter, command string, outputType ShellCommandOutputType) *ShellCommandProcessor {
	return &ShellCommandProcessor{
		channels:   channels,
		variables:  variables,
		command:    []rune(command),
		outputType: outputType,
	}
}

// Execute performs variable expansion on the command
// and then executes the command using the configured
// output type
func (processor *ShellCommandProcessor) Execute() {
	log.Debugf("Processing command: %v", string(processor.command))
	processor.findVariables()
	command := processor.replaceVariables()
	log.Debugf("Executing command: %v", command)
	processor.executeCommand(command)
}

func (processor *ShellCommandProcessor) findVariables() {
	for position := 0; position < len(processor.command); position++ {
		variableCandidate := processor.nextVariableCandidate(position)
		if variableCandidate == nil {
			break
		}

		if variable := processor.getVariable(variableCandidate); variable != nil {
			log.Debugf("Found variable %v", GRVVariableName(variable.variable))
			processor.variableRefs = append(processor.variableRefs, variable)
			position = variable.position.endIndex
		}
	}
}

func (processor *ShellCommandProcessor) nextVariableCandidate(processorIndex int) *commandPosition {
	var startIndex int

	for i := processorIndex; i+1 < len(processor.command); i++ {
		if processor.command[i] == '$' && processor.command[i+1] == '{' {
			startIndex = i
			break
		}
	}

	var endIndex int

	for i := startIndex + 1; i < len(processor.command); i++ {
		if processor.command[i] == '}' {
			endIndex = i
			break
		}
	}

	if endIndex > startIndex {
		return &commandPosition{
			startIndex: startIndex,
			endIndex:   endIndex,
		}
	}

	return nil
}

func (processor *ShellCommandProcessor) getVariable(variableCandidate *commandPosition) *variableReference {
	variableName := string(processor.command[variableCandidate.startIndex+2 : variableCandidate.endIndex])

	if variable, exists := LookupGRVVariable(variableName); exists {
		return &variableReference{
			variable: variable,
			position: variableCandidate,
		}
	}

	return nil
}

func (processor *ShellCommandProcessor) replaceVariables() (processedCommand string) {
	if len(processor.variableRefs) == 0 {
		return string(processor.command)
	}

	var buf bytes.Buffer
	index := 0

	for _, variableRef := range processor.variableRefs {
		for ; index < variableRef.position.startIndex; index++ {
			buf.WriteRune(processor.command[index])
		}

		value, _ := processor.variables.VariableValue(variableRef.variable)
		buf.WriteString(value)

		index = variableRef.position.endIndex + 1
	}

	for ; index < len(processor.command); index++ {
		buf.WriteRune(processor.command[index])
	}

	return buf.String()
}

func (processor *ShellCommandProcessor) executeCommand(command string) {
	switch processor.outputType {
	case NoOutput:
		panic("NoOutput shell command not implemented")
	case WindowOutput:
		processor.runWindowOutputCommand(command)
	case TerminalOutput:
		processor.runTerminalOutputCommand(command)
	case StatusBarOutput:
		panic("StatusBarOutput shell command not implemented")
	default:
		panic("Invalid shell command output type")
	}
}

func (processor *ShellCommandProcessor) runTerminalOutputCommand(command string) {
	processor.channels.DoAction(Action{ActionType: ActionRunCommand, Args: []interface{}{
		ActionRunCommandArgs{
			command:        command,
			interactive:    true,
			promptForInput: true,
			stdin:          os.Stdin,
			stdout:         os.Stdout,
			stderr:         os.Stderr,
			onComplete: func(commandErr error, exitStatus int) (err error) {
				if commandErr != nil {
					return fmt.Errorf(`Command "%v" failed: %v"`, command, commandErr)
				} else if exitStatus != 0 {
					return fmt.Errorf(`Command "%v" exited with status %v`, command, exitStatus)
				}

				processor.channels.ReportStatus(`Command "%v" exited with status %v`, command, exitStatus)

				return
			},
		},
	}})
}

func (processor *ShellCommandProcessor) runWindowOutputCommand(command string) {
	processor.channels.DoAction(Action{ActionType: ActionCreateCommandOutputView, Args: []interface{}{
		ActionCreateCommandOutputViewArgs{
			command: command,
			viewDimension: ViewDimension{
				cols: 80,
				rows: 24,
			},
			onCreation: func(commandOutputProcessor CommandOutputProcessor) {
				processor.runNonInteractiveCommand(command, commandOutputProcessor)
			},
		},
	}})
}

func (processor *ShellCommandProcessor) runNonInteractiveCommand(command string, commandOutputProcessor CommandOutputProcessor) {
	var scanner *bufio.Scanner

	processor.channels.DoAction(Action{ActionType: ActionRunCommand, Args: []interface{}{
		ActionRunCommandArgs{
			command:     command,
			interactive: false,
			beforeStart: func(cmd *exec.Cmd) {
				stdout, err := cmd.StdoutPipe()
				if err != nil {
					commandOutputProcessor.OnCommandExecutionError(err)
					return
				}

				stderr, err := cmd.StderrPipe()
				if err != nil {
					commandOutputProcessor.OnCommandExecutionError(err)
					return
				}

				scanner = bufio.NewScanner(io.MultiReader(stdout, stderr))
			},
			onStart: func(cmd *exec.Cmd) {
				for scanner.Scan() {
					commandOutputProcessor.AddOutputLine(scanner.Text())
				}
			},
			onComplete: func(commandErr error, exitStatus int) (err error) {
				if commandErr != nil {
					commandOutputProcessor.OnCommandExecutionError(commandErr)
				} else {
					commandOutputProcessor.OnCommandComplete(exitStatus)
				}

				return
			},
		},
	}})
}
