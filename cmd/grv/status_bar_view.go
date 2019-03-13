package main

import (
	"fmt"
	"strings"
	"sync"
	"unicode/utf8"

	log "github.com/Sirupsen/logrus"
)

// The different prompt types grv uses
const (
	PromptText              = ":"
	SearchPromptText        = "/"
	ReverseSearchPromptText = "?"
	FilterPromptText        = "query: "
	BranchNamePromptText    = "branch name: "
	TagNamePromptText       = "tag name: "
)

type promptType int

const (
	ptNone promptType = iota
	ptCommand
	ptSearch
	ptFilter
	ptQuestion
	ptBranchName
	ptTagName
)

// StatusBarView manages the display of the status bar
type StatusBarView struct {
	repoData      RepoData
	channels      Channels
	config        ConfigSetter
	viewState     ViewState
	promptType    promptType
	pendingStatus string
	lock          sync.Mutex
}

// NewStatusBarView creates a new instance
func NewStatusBarView(repoData RepoData, channels Channels, config ConfigSetter) *StatusBarView {
	return &StatusBarView{
		repoData: repoData,
		channels: channels,
		config:   config,
	}
}

// Initialise does nothing
func (statusBarView *StatusBarView) Initialise() (err error) {
	return
}

// Dispose of any resources held by the view
func (statusBarView *StatusBarView) Dispose() {

}

// HandleEvent does nothing
func (statusBarView *StatusBarView) HandleEvent(event Event) (err error) {
	return
}

// HandleAction checks if the status bar view supports the provided action and executes it if so
func (statusBarView *StatusBarView) HandleAction(action Action) (err error) {
	switch action.ActionType {
	case ActionPrompt:
		statusBarView.showCommandPrompt(action)
	case ActionSearchPrompt:
		statusBarView.showSearchPrompt(action, SearchPromptText, ActionSearch)
	case ActionReverseSearchPrompt:
		statusBarView.showSearchPrompt(action, ReverseSearchPromptText, ActionReverseSearch)
	case ActionFilterPrompt:
		statusBarView.showFilterPrompt(action)
	case ActionQuestionPrompt:
		statusBarView.showQuestionPrompt(action)
	case ActionBranchNamePrompt:
		statusBarView.showRefNamePrompt(action, ptBranchName, BranchNamePromptText)
	case ActionTagNamePrompt:
		statusBarView.showRefNamePrompt(action, ptTagName, TagNamePromptText)
	case ActionCustomPrompt:
		statusBarView.showCustomPrompt(action)
	case ActionShowStatus:
		statusBarView.lock.Lock()
		defer statusBarView.lock.Unlock()

		if len(action.Args) > 0 {
			status, ok := action.Args[0].(string)
			if ok {
				statusBarView.pendingStatus = status
				log.Infof("Received status: %v", status)
				statusBarView.channels.UpdateDisplay()
				return
			}
		}

		err = fmt.Errorf("Expected status argument but received: %v", action.Args)
	}

	return
}

func (statusBarView *StatusBarView) showCommandPrompt(action Action) {
	statusBarView.promptType = ptCommand
	input := statusBarView.showPrompt(&PromptArgs{Prompt: PromptText}, action)
	errors := statusBarView.config.Evaluate(input)
	statusBarView.channels.ReportErrors(errors)
	statusBarView.promptType = ptNone
}

func (statusBarView *StatusBarView) showSearchPrompt(action Action, prompt string, actionType ActionType) {
	statusBarView.promptType = ptSearch
	input := statusBarView.showPrompt(&PromptArgs{Prompt: prompt}, action)

	if input == "" {
		statusBarView.channels.DoAction(Action{
			ActionType: ActionClearSearch,
		})
	} else {
		statusBarView.channels.DoAction(Action{
			ActionType: actionType,
			Args:       []interface{}{input},
		})
	}

	statusBarView.promptType = ptNone
}

func (statusBarView *StatusBarView) showFilterPrompt(action Action) {
	statusBarView.promptType = ptFilter
	input := statusBarView.showPrompt(&PromptArgs{Prompt: FilterPromptText}, action)

	if input != "" {
		statusBarView.channels.DoAction(Action{
			ActionType: ActionAddFilter,
			Args:       []interface{}{input},
		})
	}

	statusBarView.promptType = ptNone
}

func (statusBarView *StatusBarView) showQuestionPrompt(action Action) {
	if len(action.Args) == 0 {
		log.Errorf("Expected to find ActionQuestionPromptArgs arg but found none")
		return
	}

	args, ok := action.Args[0].(ActionQuestionPromptArgs)
	if !ok {
		log.Errorf("Expected to find type ActionQuestionPromptArgs but found %T", action.Args[0])
		return
	}

	validAnswers := make(map[string]string)

	promptText := fmt.Sprintf("%v (%v)", args.question, strings.Join(args.answers, "|"))

	if args.defaultAnswer != "" {
		promptText = fmt.Sprintf("%v (default=%v)", promptText, args.defaultAnswer)
		validAnswers[""] = args.defaultAnswer
	}

	promptText = fmt.Sprintf(" %v? ", promptText)

	maxAnswerLength := 0

	for _, answer := range args.answers {
		validAnswers[answer] = answer

		answerLength := len([]rune(answer))
		if answerLength > maxAnswerLength {
			maxAnswerLength = answerLength
		}
	}

	promptArgs := PromptArgs{
		Prompt:         promptText,
		NumCharsToRead: maxAnswerLength,
	}

	statusBarView.promptType = ptQuestion

	for {
		answer := statusBarView.showPrompt(&promptArgs, action)

		if validAnswer, isValidAnswer := validAnswers[answer]; isValidAnswer {
			if args.onAnswer != nil {
				args.onAnswer(validAnswer)
			}

			break
		} else if answer == "" {
			break
		}
	}

	statusBarView.promptType = ptNone
}

func (statusBarView *StatusBarView) showRefNamePrompt(action Action, promptType promptType, promptText string) {
	if len(action.Args) == 0 {
		log.Errorf("Expected ActionType argument")
		return
	}

	nextAction, ok := action.Args[0].(ActionType)
	if !ok {
		log.Errorf("Expected ActionType argument but found %T", action.Args[0])
		return
	}

	statusBarView.promptType = promptType
	input := statusBarView.showPrompt(&PromptArgs{Prompt: promptText}, action)

	if input != "" {
		statusBarView.channels.DoAction(Action{
			ActionType: nextAction,
			Args:       []interface{}{input},
		})
	}

	statusBarView.promptType = ptNone
}

func (statusBarView *StatusBarView) showCustomPrompt(action Action) {
	if len(action.Args) == 0 {
		log.Errorf("Expected ActionCustomPromptArgs argument")
		return
	}

	args, ok := action.Args[0].(ActionCustomPromptArgs)
	if !ok {
		log.Errorf("Expected ActionCustomPromptArgs argument but found %T", action.Args[0])
		return
	}

	input := statusBarView.showPrompt(&PromptArgs{Prompt: args.prompt}, action)

	args.inputHandler(input)
}

func (statusBarView *StatusBarView) showPrompt(promptArgs *PromptArgs, action Action) string {
	for _, arg := range action.Args {
		if actionPromptArgs, ok := arg.(ActionPromptArgs); ok {
			if actionPromptArgs.terminated {
				return actionPromptArgs.keys
			}

			promptArgs.InitialBufferText = actionPromptArgs.keys
			break
		}
	}

	return Prompt(promptArgs)
}

// OnStateChange updates the active state of this view
func (statusBarView *StatusBarView) OnStateChange(viewState ViewState) {
	statusBarView.lock.Lock()
	defer statusBarView.lock.Unlock()

	statusBarView.viewState = viewState
}

// ViewID returns the view ID of the status bar view
func (statusBarView *StatusBarView) ViewID() ViewID {
	return ViewStatusBar
}

// Render generates and draws the status view to the provided window
// If the readline prompt is active then this is drawn
func (statusBarView *StatusBarView) Render(win RenderWindow) (err error) {
	statusBarView.lock.Lock()
	defer statusBarView.lock.Unlock()

	lineBuilder, err := win.LineBuilder(0, 1)
	if err != nil {
		return
	}

	if statusBarView.viewState == ViewStateActive {
		promptText, promptInput, promptPoint := PromptState()
		lineBuilder.Append("%v%v", promptText, promptInput)
		bytes := 0
		characters := len(promptText)

		for _, char := range promptInput {
			bytes += utf8.RuneLen(char)

			if bytes > promptPoint {
				break
			}

			characters += RuneWidth(char)
		}

		err = win.SetCursor(0, uint(characters))
	} else {
		lineBuilder.Append(" %v", statusBarView.pendingStatus)
		win.ApplyStyle(CmpStatusbarviewNormal)
	}

	return
}

// RenderHelpBar renders help information for the status bar view
func (statusBarView *StatusBarView) RenderHelpBar(lineBuilder *LineBuilder) (err error) {
	message := ""

	switch statusBarView.promptType {
	case ptCommand:
		message = "Enter a command"
	case ptSearch:
		message = "Enter a regex pattern"
	case ptFilter:
		message = "Enter a filter query"
	case ptQuestion:
		message = "Enter an answer"
	case ptBranchName:
		message = "Enter the new branch name"
	case ptTagName:
		message = "Enter the new tag name"
	}

	if message != "" {
		lineBuilder.AppendWithStyle(CmpHelpbarviewSpecial, message)
	}

	return
}
