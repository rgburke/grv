package main

import (
	"fmt"
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
)

type promptType int

const (
	ptNone promptType = iota
	ptCommand
	ptSearch
	ptFilter
)

// StatusBarView manages the display of the status bar
type StatusBarView struct {
	repoData      RepoData
	channels      *Channels
	config        ConfigSetter
	active        bool
	promptType    promptType
	pendingStatus string
	lock          sync.Mutex
}

// NewStatusBarView creates a new instance
func NewStatusBarView(repoData RepoData, channels *Channels, config ConfigSetter) *StatusBarView {
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

// HandleEvent does nothing
func (statusBarView *StatusBarView) HandleEvent(event Event) (err error) {
	return
}

// HandleAction checks if the status bar view supports the provided action and executes it if so
func (statusBarView *StatusBarView) HandleAction(action Action) (err error) {
	switch action.ActionType {
	case ActionPrompt:
		statusBarView.showCommandPrompt()
	case ActionSearchPrompt:
		statusBarView.showSearchPrompt(SearchPromptText, ActionSearch)
	case ActionReverseSearchPrompt:
		statusBarView.showSearchPrompt(ReverseSearchPromptText, ActionReverseSearch)
	case ActionFilterPrompt:
		statusBarView.showFilterPrompt()
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

func (statusBarView *StatusBarView) showCommandPrompt() {
	statusBarView.promptType = ptCommand
	input := Prompt(PromptText)
	errors := statusBarView.config.Evaluate(input)
	statusBarView.channels.ReportErrors(errors)
	statusBarView.promptType = ptNone
}

func (statusBarView *StatusBarView) showSearchPrompt(prompt string, actionType ActionType) {
	statusBarView.promptType = ptSearch
	input := Prompt(prompt)

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

func (statusBarView *StatusBarView) showFilterPrompt() {
	statusBarView.promptType = ptFilter
	input := Prompt(FilterPromptText)

	if input != "" {
		statusBarView.channels.DoAction(Action{
			ActionType: ActionAddFilter,
			Args:       []interface{}{input},
		})
	}

	statusBarView.promptType = ptNone
}

// OnActiveChange updates the active state of this view
func (statusBarView *StatusBarView) OnActiveChange(active bool) {
	statusBarView.lock.Lock()
	defer statusBarView.lock.Unlock()

	log.Debugf("StatusBarView active: %v", active)
	statusBarView.active = active
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

	if statusBarView.active {
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
	}

	if message != "" {
		lineBuilder.AppendWithStyle(CmpHelpbarviewSpecial, message)
	}

	return
}
