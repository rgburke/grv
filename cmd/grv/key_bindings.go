package main

import (
	"fmt"
	"io"

	pt "github.com/tchap/go-patricia/patricia"
)

// QuestionResponse represents a response to a question
type QuestionResponse int

// The set of currently supported question responses
const (
	ResponseNone QuestionResponse = iota
	ResponseYes
	ResponseNo
)

// ActionType represents an action to be performed
type ActionType int

// The set of actions possible supported by grv
const (
	ActionNone ActionType = iota
	ActionExit
	ActionSuspend
	ActionRunCommand
	ActionPrompt
	ActionSearchPrompt
	ActionReverseSearchPrompt
	ActionFilterPrompt
	ActionQuestionPrompt
	ActionBranchNamePrompt
	ActionSearch
	ActionReverseSearch
	ActionSearchFindNext
	ActionSearchFindPrev
	ActionClearSearch
	ActionShowStatus
	ActionNextLine
	ActionPrevLine
	ActionNextPage
	ActionPrevPage
	ActionNextHalfPage
	ActionPrevHalfPage
	ActionScrollRight
	ActionScrollLeft
	ActionFirstLine
	ActionLastLine
	ActionSelect
	ActionNextView
	ActionPrevView
	ActionFullScreenView
	ActionToggleViewLayout
	ActionAddFilter
	ActionRemoveFilter
	ActionCenterView
	ActionScrollCursorTop
	ActionScrollCursorBottom
	ActionCursorTopView
	ActionCursorMiddleView
	ActionCursorBottomView
	ActionNextTab
	ActionPrevTab
	ActionNewTab
	ActionRemoveTab
	ActionAddView
	ActionSplitView
	ActionRemoveView
	ActionMouseSelect
	ActionMouseScrollDown
	ActionMouseScrollUp
	ActionCheckoutRef
	ActionCheckoutCommit
	ActionCreateBranch
	ActionCreateContextMenu
	ActionShowAvailableActions
	ActionStageFile
	ActionUnstageFile
	ActionCommit
)

// Action represents a type of actions and its arguments to be executed
type Action struct {
	ActionType ActionType
	Args       []interface{}
}

// CreateViewArgs contains the fields required to create and configure a view
type CreateViewArgs struct {
	viewID               ViewID
	viewArgs             []interface{}
	registerViewListener RegisterViewListener
}

// ActionAddViewArgs contains arguments the ActionAddView action requires
type ActionAddViewArgs struct {
	CreateViewArgs
}

// ActionSplitViewArgs contains arguments the ActionSplitView action requires
type ActionSplitViewArgs struct {
	CreateViewArgs
	orientation ContainerOrientation
}

// ActionPromptArgs contains arguments to an action that displays a prompt
type ActionPromptArgs struct {
	keys       string
	terminated bool
}

// ActionQuestionPromptArgs contains arguments to configure a question prompt
type ActionQuestionPromptArgs struct {
	question      string
	answers       []string
	defaultAnswer string
	onAnswer      func(string)
}

// ActionCreateContextMenuArgs contains arguments to create and configure a context menu
type ActionCreateContextMenuArgs struct {
	config        ContextMenuConfig
	viewDimension ViewDimension
}

// ActionRunCommandArgs contains arguments to run a command and process
// the status and output
type ActionRunCommandArgs struct {
	command    string
	stdin      io.Reader
	stdout     io.Writer
	stderr     io.Writer
	onComplete func(err error, exitStatus int)
}

var actionKeys = map[string]ActionType{
	"<grv-nop>":                    ActionNone,
	"<grv-exit>":                   ActionExit,
	"<grv-suspend>":                ActionSuspend,
	"<grv-run-command>":            ActionRunCommand,
	"<grv-prompt>":                 ActionPrompt,
	"<grv-search-prompt>":          ActionSearchPrompt,
	"<grv-reverse-search-prompt>":  ActionReverseSearchPrompt,
	"<grv-filter-prompt>":          ActionFilterPrompt,
	"<grv-question-prompt>":        ActionQuestionPrompt,
	"<grv-branch-name-prompt>":     ActionBranchNamePrompt,
	"<grv-search>":                 ActionSearch,
	"<grv-reverse-search>":         ActionReverseSearch,
	"<grv-search-find-next>":       ActionSearchFindNext,
	"<grv-search-find-prev>":       ActionSearchFindPrev,
	"<grv-clear-search>":           ActionClearSearch,
	"<grv-show-status>":            ActionShowStatus,
	"<grv-next-line>":              ActionNextLine,
	"<grv-prev-line>":              ActionPrevLine,
	"<grv-next-page>":              ActionNextPage,
	"<grv-prev-page>":              ActionPrevPage,
	"<grv-next-half-page>":         ActionNextHalfPage,
	"<grv-prev-half-page>":         ActionPrevHalfPage,
	"<grv-scroll-right>":           ActionScrollRight,
	"<grv-scroll-left>":            ActionScrollLeft,
	"<grv-first-line>":             ActionFirstLine,
	"<grv-last-line>":              ActionLastLine,
	"<grv-select>":                 ActionSelect,
	"<grv-next-view>":              ActionNextView,
	"<grv-prev-view>":              ActionPrevView,
	"<grv-full-screen-view>":       ActionFullScreenView,
	"<grv-toggle-view-layout>":     ActionToggleViewLayout,
	"<grv-add-filter>":             ActionAddFilter,
	"<grv-remove-filter>":          ActionRemoveFilter,
	"<grv-center-view>":            ActionCenterView,
	"<grv-scroll-cursor-top>":      ActionScrollCursorTop,
	"<grv-scroll-cursor-bottom>":   ActionScrollCursorBottom,
	"<grv-cursor-top-view>":        ActionCursorTopView,
	"<grv-cursor-middle-view>":     ActionCursorMiddleView,
	"<grv-cursor-bottom-view>":     ActionCursorBottomView,
	"<grv-next-tab>":               ActionNextTab,
	"<grv-prev-tab>":               ActionPrevTab,
	"<grv-add-tab>":                ActionNewTab,
	"<grv-remove-tab>":             ActionRemoveTab,
	"<grv-add-view>":               ActionAddView,
	"<grv-split-view>":             ActionSplitView,
	"<grv-remove-view>":            ActionRemoveView,
	"<grv-mouse-select>":           ActionMouseSelect,
	"<grv-checkout-ref>":           ActionCheckoutRef,
	"<grv-checkout-commit>":        ActionCheckoutCommit,
	"<grv-create-branch>":          ActionCreateBranch,
	"<grv-create-context-menu>":    ActionCreateContextMenu,
	"<grv-show-available-actions>": ActionShowAvailableActions,
	"<grv-stage-file>":             ActionStageFile,
	"<grv-unstage-file>":           ActionUnstageFile,
	"<grv-action-commit>":          ActionCommit,
}

var promptActions = map[ActionType]bool{
	ActionPrompt:              true,
	ActionSearchPrompt:        true,
	ActionReverseSearchPrompt: true,
	ActionFilterPrompt:        true,
	ActionQuestionPrompt:      true,
	ActionBranchNamePrompt:    true,
}

var defaultKeyBindings = map[ActionType]map[ViewID][]string{
	ActionPrompt: {
		ViewMain: {PromptText},
	},
	ActionSearchPrompt: {
		ViewMain: {SearchPromptText},
	},
	ActionReverseSearchPrompt: {
		ViewMain: {ReverseSearchPromptText},
	},
	ActionSuspend: {
		ViewAll: {"<C-z>"},
	},
	ActionSearchFindNext: {
		ViewAll: {"n"},
	},
	ActionSearchFindPrev: {
		ViewAll: {"N"},
	},
	ActionNextLine: {
		ViewAll: {"<Down>", "j"},
	},
	ActionPrevLine: {
		ViewAll: {"<Up>", "k"},
	},
	ActionNextPage: {
		ViewAll: {"<PageDown>", "<C-f>"},
	},
	ActionPrevPage: {
		ViewAll: {"<PageUp>", "<C-b>"},
	},
	ActionNextHalfPage: {
		ViewAll: {"<C-d>"},
	},
	ActionPrevHalfPage: {
		ViewAll: {"<C-u>"},
	},
	ActionScrollRight: {
		ViewAll: {"<Right>", "l"},
	},
	ActionScrollLeft: {
		ViewAll: {"<Left>", "h"},
	},
	ActionFirstLine: {
		ViewAll: {"gg"},
	},
	ActionLastLine: {
		ViewAll: {"G"},
	},
	ActionNextView: {
		ViewAll: {"<Tab>", "<C-w>w", "<C-w><C-w>"},
	},
	ActionPrevView: {
		ViewAll: {"<S-Tab>", "<C-w>W"},
	},
	ActionFullScreenView: {
		ViewAll: {"f", "<C-w>o", "<C-w><C-o>"},
	},
	ActionToggleViewLayout: {
		ViewAll: {"<C-w>t"},
	},
	ActionSelect: {
		ViewAll: {"<Enter>"},
	},
	ActionFilterPrompt: {
		ViewCommit: {"<C-q>"},
		ViewRef:    {"<C-q>"},
	},
	ActionRemoveFilter: {
		ViewCommit: {"<C-r>"},
		ViewRef:    {"<C-r>"},
	},
	ActionCenterView: {
		ViewAll: {"zz", "z."},
	},
	ActionScrollCursorTop: {
		ViewAll: {"zt"},
	},
	ActionScrollCursorBottom: {
		ViewAll: {"zb"},
	},
	ActionCursorTopView: {
		ViewAll: {"H"},
	},
	ActionCursorMiddleView: {
		ViewAll: {"M"},
	},
	ActionCursorBottomView: {
		ViewAll: {"L"},
	},
	ActionNextTab: {
		ViewAll: {"gt"},
	},
	ActionPrevTab: {
		ViewAll: {"gT"},
	},
	ActionRemoveView: {
		ViewAll: {"q"},
	},
	ActionCheckoutRef: {
		ViewRef: {"c"},
	},
	ActionCheckoutCommit: {
		ViewCommit: {"c"},
	},
	ActionBranchNamePrompt: {
		ViewRef:    {"b"},
		ViewCommit: {"b"},
	},
	ActionShowAvailableActions: {
		ViewAll: {"<C-a>"},
	},
	ActionStageFile: {
		ViewGitStatus: {"a"},
	},
	ActionUnstageFile: {
		ViewGitStatus: {"u"},
	},
	ActionCommit: {
		ViewGitStatus: {"c"},
	},
}

// ViewHierarchy is a list of views parent to child
type ViewHierarchy []ViewID

// BindingType specifies the type a key sequence is bound to
type BindingType int

// The types a key sequence can by bound to
const (
	BtAction BindingType = iota
	BtKeystring
)

// Binding is the entity a key sequence is bound to
// This is either an action or a key sequence
type Binding struct {
	bindingType BindingType
	actionType  ActionType
	keystring   string
}

func newActionBinding(actionType ActionType) Binding {
	return Binding{
		bindingType: BtAction,
		actionType:  actionType,
	}
}

func newKeystringBinding(keystring string) Binding {
	return Binding{
		bindingType: BtKeystring,
		keystring:   keystring,
		actionType:  ActionNone,
	}
}

// KeyBindings exposes key bindings that have been configured and allows new bindings to be set
type KeyBindings interface {
	Binding(viewHierarchy ViewHierarchy, keystring string) (binding Binding, isPrefix bool)
	SetActionBinding(viewID ViewID, keystring string, actionType ActionType)
	SetKeystringBinding(viewID ViewID, keystring, mappedKeystring string)
	RemoveBinding(viewID ViewID, keystring string) (removed bool)
}

// KeyBindingManager manages key bindings in grv
type KeyBindingManager struct {
	bindings map[ViewID]*pt.Trie
}

// NewKeyBindingManager creates a new instance
func NewKeyBindingManager() KeyBindings {
	keyBindingManager := &KeyBindingManager{
		bindings: make(map[ViewID]*pt.Trie),
	}

	keyBindingManager.setDefaultKeyBindings()

	return keyBindingManager
}

// Binding returns the Binding bound to the provided key sequence for the view hierarchy provided
// If no binding exists or the provided key sequence is a prefix to a binding then an action binding with action ActionNone is returned and a boolean indicating whether there is a prefix match
func (keyBindingManager *KeyBindingManager) Binding(viewHierarchy ViewHierarchy, keystring string) (Binding, bool) {
	viewHierarchy = append(viewHierarchy, ViewAll)
	isPrefix := false

	for _, viewID := range viewHierarchy {
		if viewBindings, ok := keyBindingManager.bindings[viewID]; ok {
			if binding := viewBindings.Get(pt.Prefix(keystring)); binding != nil {
				return binding.(Binding), false
			} else if viewBindings.MatchSubtree(pt.Prefix(keystring)) {
				isPrefix = true
			}
		}
	}

	return newActionBinding(ActionNone), isPrefix
}

// SetActionBinding allows an action to be bound to the provided key sequence and view
func (keyBindingManager *KeyBindingManager) SetActionBinding(viewID ViewID, keystring string, actionType ActionType) {
	viewBindings := keyBindingManager.getOrCreateViewBindings(viewID)
	viewBindings.Set(pt.Prefix(keystring), newActionBinding(actionType))
}

// SetKeystringBinding allows a key sequence to be bound to the provided key sequence and view
func (keyBindingManager *KeyBindingManager) SetKeystringBinding(viewID ViewID, keystring, mappedKeystring string) {
	viewBindings := keyBindingManager.getOrCreateViewBindings(viewID)
	viewBindings.Set(pt.Prefix(keystring), newKeystringBinding(mappedKeystring))
}

func (keyBindingManager *KeyBindingManager) getOrCreateViewBindings(viewID ViewID) *pt.Trie {
	viewBindings, ok := keyBindingManager.bindings[viewID]
	if ok {
		return viewBindings
	}

	viewBindings = pt.NewTrie()
	keyBindingManager.bindings[viewID] = viewBindings
	return viewBindings
}

// RemoveBinding removes the binding for the provided keystring if it exists
func (keyBindingManager *KeyBindingManager) RemoveBinding(viewID ViewID, keystring string) (removed bool) {
	if viewBindings, ok := keyBindingManager.bindings[viewID]; ok {
		return viewBindings.Delete(pt.Prefix(keystring))
	}

	return
}

func (keyBindingManager *KeyBindingManager) setDefaultKeyBindings() {
	for actionKey, actionType := range actionKeys {
		keyBindingManager.SetActionBinding(ViewAll, actionKey, actionType)
	}

	for actionType, viewKeys := range defaultKeyBindings {
		for viewID, keys := range viewKeys {
			for _, key := range keys {
				keyBindingManager.SetActionBinding(viewID, key, actionType)
			}
		}
	}
}

func isValidAction(action string) bool {
	_, valid := actionKeys[action]
	return valid
}

// IsPromptAction returns true if the action presents a prompt
func IsPromptAction(actionType ActionType) bool {
	_, isPrompt := promptActions[actionType]
	return isPrompt
}

// DefaultKeyBindings returns the default key sequences that are bound to an action for the provided view
func DefaultKeyBindings(actionType ActionType, viewID ViewID) (keyBindings []string) {
	viewKeys, ok := defaultKeyBindings[actionType]
	if !ok {
		return
	}

	keys, ok := viewKeys[viewID]
	if !ok {
		keys, ok = viewKeys[ViewAll]

		if !ok {
			return
		}
	}

	return keys
}

// MouseEventAction maps a mouse event to an action
func MouseEventAction(mouseEvent MouseEvent) (action Action, err error) {
	switch mouseEvent.mouseEventType {
	case MetLeftClick:
		action = Action{
			ActionType: ActionMouseSelect,
			Args:       []interface{}{mouseEvent},
		}
	case MetScrollDown:
		action = Action{ActionType: ActionMouseScrollDown}
	case MetScrollUp:
		action = Action{ActionType: ActionMouseScrollUp}
	default:
		err = fmt.Errorf("Unknown MouseEventType %v", mouseEvent.mouseEventType)
	}

	return
}

// GetMouseEventFromAction converts a MouseEvent into an Action that can be processed by a view
func GetMouseEventFromAction(action Action) (mouseEvent MouseEvent, err error) {
	if len(action.Args) == 0 {
		err = fmt.Errorf("Expected MouseEvent arg")
		return
	}

	mouseEvent, ok := action.Args[0].(MouseEvent)
	if !ok {
		err = fmt.Errorf("Expected first argument to have type MouseEvent but has type: %T", action.Args[0])
	}

	return
}

// YesNoQuestion generates an action that will prompt the user for a yes/no response
// The onResponse handler is called when an answer is received
func YesNoQuestion(question string, onResponse func(QuestionResponse)) Action {
	return Action{
		ActionType: ActionQuestionPrompt,
		Args: []interface{}{ActionQuestionPromptArgs{
			question: question,
			answers:  []string{"y", "n"},
			onAnswer: func(answer string) {
				var response QuestionResponse
				switch answer {
				case "y":
					response = ResponseYes
				case "n":
					response = ResponseNo
				default:
					response = ResponseNone
				}

				onResponse(response)
			},
		}},
	}
}
