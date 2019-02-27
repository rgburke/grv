package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"

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

var shellCommandPrefixOutputType = map[rune]ShellCommandOutputType{
	'@': NoOutput,
	'!': TerminalOutput,
}

// OutputType returns the ShellCommandOutputType the provided shell command prefix maps to
// or WindowOuput if no mapping exists
func OutputType(shellCommandPrefix rune) ShellCommandOutputType {
	if outputType, ok := shellCommandPrefixOutputType[shellCommandPrefix]; ok {
		return outputType
	}

	return WindowOutput
}

type referenceType int

const (
	rtNone referenceType = iota
	rtVariable
	rtPrompt
)

// ShellCommandProcessor processes a shell command before executing it
type ShellCommandProcessor struct {
	channels   Channels
	variables  GRVVariableGetter
	command    []rune
	outputType ShellCommandOutputType
	refs       []referenceOccurence
	lock       sync.Mutex
}

type referencePosition struct {
	startIndex int
	endIndex   int
}

type variableReference struct {
	pos      *referencePosition
	variable GRVVariable
}

func (ref *variableReference) position() *referencePosition {
	return ref.pos
}

type promptReference struct {
	pos    *referencePosition
	prompt string
	value  string
}

func (ref *promptReference) position() *referencePosition {
	return ref.pos
}

type referenceOccurence interface {
	position() *referencePosition
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
	processor.lock.Lock()

	go func() {
		defer processor.lock.Unlock()

		log.Debugf("Processing command: %v", string(processor.command))
		processor.findReferences()
		processor.determinePromptValues()
		command := processor.replaceReferences()
		log.Debugf("Executing command: %v", command)
		processor.executeCommand(command)
	}()
}

func (processor *ShellCommandProcessor) findReferences() {
	for position := 0; position < len(processor.command); position++ {
		refType, candidate := processor.nextReferenceCandidate(position)

		if refType == rtNone || candidate == nil {
			break
		} else if refType == rtVariable {
			if variable := processor.getVariable(candidate); variable != nil {
				log.Debugf("Found variable %v at %v", GRVVariableName(variable.variable), *candidate)
				processor.refs = append(processor.refs, variable)
				position = variable.pos.endIndex
			}
		} else if refType == rtPrompt {
			promptRef := processor.getPrompt(candidate)
			log.Debugf("Found prompt %v at %v", promptRef.prompt, *candidate)
			processor.refs = append(processor.refs, promptRef)
			position = promptRef.pos.endIndex
		}
	}
}

func (processor *ShellCommandProcessor) nextReferenceCandidate(processorIndex int) (referenceType, *referencePosition) {
	var startIndex int

	for i := processorIndex; i+1 < len(processor.command); i++ {
		if (processor.command[i] == '$' || processor.command[i] == '?') && processor.command[i+1] == '{' {
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
		refType := rtVariable
		if processor.command[startIndex] == '?' {
			refType = rtPrompt
		}

		return refType, &referencePosition{
			startIndex: startIndex,
			endIndex:   endIndex,
		}
	}

	return rtNone, nil
}

func (processor *ShellCommandProcessor) getVariable(variableCandidate *referencePosition) *variableReference {
	variableName := string(processor.command[variableCandidate.startIndex+2 : variableCandidate.endIndex])

	if variable, exists := LookupGRVVariable(variableName); exists {
		return &variableReference{
			variable: variable,
			pos:      variableCandidate,
		}
	}

	return nil
}

func (processor *ShellCommandProcessor) getPrompt(position *referencePosition) *promptReference {
	prompt := string(processor.command[position.startIndex+2 : position.endIndex])

	return &promptReference{
		prompt: prompt,
		pos:    position,
	}
}

func (processor *ShellCommandProcessor) determinePromptValues() {
	if len(processor.refs) == 0 {
		return
	}

	var waitGroup sync.WaitGroup

	for _, ref := range processor.refs {
		promptRef, isPromptRef := ref.(*promptReference)
		if !isPromptRef {
			continue
		}

		waitGroup.Add(1)

		processor.channels.DoAction(Action{ActionType: ActionCustomPrompt, Args: []interface{}{
			ActionCustomPromptArgs{
				prompt: promptRef.prompt,
				inputHandler: func(input string) {
					promptRef.value = input
					waitGroup.Done()
				},
			},
		}})
	}

	waitGroup.Wait()
}

func (processor *ShellCommandProcessor) replaceReferences() (processedCommand string) {
	if len(processor.refs) == 0 {
		return string(processor.command)
	}

	var buf bytes.Buffer
	index := 0

	for _, ref := range processor.refs {
		for ; index < ref.position().startIndex; index++ {
			buf.WriteRune(processor.command[index])
		}

		var value string

		if variableRef, isVariableRef := ref.(*variableReference); isVariableRef {
			value, _ = processor.variables.VariableValue(variableRef.variable)
		} else if promptRef, isPromptRef := ref.(*promptReference); isPromptRef {
			value = promptRef.value
		}

		buf.WriteString(value)

		index = ref.position().endIndex + 1
	}

	for ; index < len(processor.command); index++ {
		buf.WriteRune(processor.command[index])
	}

	return buf.String()
}

func (processor *ShellCommandProcessor) executeCommand(command string) {
	switch processor.outputType {
	case NoOutput:
		processor.runNoOutputCommand(command)
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

func (processor *ShellCommandProcessor) runNoOutputCommand(command string) {
	processor.channels.DoAction(Action{ActionType: ActionRunCommand, Args: []interface{}{
		ActionRunCommandArgs{
			command: command,
			onComplete: func(commandErr error, exitStatus int) (err error) {
				if commandErr != nil {
					return fmt.Errorf(`Command "%v" failed: %v"`, command, commandErr)
				} else if exitStatus != 0 {
					return fmt.Errorf(`Command "%v" exited with status %v`, command, exitStatus)
				}

				return
			},
		},
	}})
}

// GenerateShellCommandHelpSections generates help information for shell commands
func GenerateShellCommandHelpSections(config Config) (helpSections []*HelpSection) {
	description := []HelpSectionText{
		{text: "Shell commands can be specified by using the following prefixes: ! or @."},
		{text: "The ! prefix runs the command and displays the output in a pop-up window."},
		{text: "The @ prefix runs the command silently and does not display any output."},
		{text: "For example, to run the command 'git pull' and see the output in a window:"},
		{},
		{text: ":!git pull", themeComponentID: CmpHelpViewSectionCodeBlock},
		{},
		{text: "Alternatively to run the command 'git pull' without seeing the output:"},
		{},
		{text: ":@git pull", themeComponentID: CmpHelpViewSectionCodeBlock},
		{},
		{text: "Key sequences can be mapped to shell commands."},
		{text: "For example, to map 'gp' to run the command 'git pull' in the background:"},
		{},
		{text: ":map All gp @git pull", themeComponentID: CmpHelpViewSectionCodeBlock},
		{},
		{text: "GRV maintains a set of variables that can be embedded in shell commands."},
		{text: "These variables represent the current state of the visible views."},
		{text: "The set of variables available is:"},
	}

	helpSections = append(helpSections, &HelpSection{
		title:       HelpSectionText{text: "Shell Commands"},
		description: description,
	})

	helpSections = append(helpSections, GenerateGRVVariablesHelpSection(config))

	helpSections = append(helpSections, &HelpSection{
		description: []HelpSectionText{
			{text: "Variables can be specified in shell commands using the syntax:"},
			{},
			{text: "${variable}", themeComponentID: CmpHelpViewSectionCodeBlock},
			{},
			{text: "For example, to cherry-pick the currently selected commit:"},
			{},
			{text: ":!git cherry-pick ${commit}", themeComponentID: CmpHelpViewSectionCodeBlock},
			{},
			{text: "User input can also be specified in shell commands by specifying a custom prompt."},
			{text: "The syntax for a custom prompt is:"},
			{},
			{text: "?{prompt text}", themeComponentID: CmpHelpViewSectionCodeBlock},
			{},
			{text: "For example, to create a new branch from the currently selected commit:"},
			{},
			{text: ":!git branch ?{New Branch Name: } ${commit}", themeComponentID: CmpHelpViewSectionCodeBlock},
			{},
			{text: "When the above command is run the user will be shown the prompt: 'New Branch Name: '"},
			{text: "The value entered into the prompt will be substituted into the command when executed."},
		},
	})

	return
}
