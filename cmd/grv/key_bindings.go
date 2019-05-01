package main

import (
	"fmt"
	"io"
	"os/exec"
	"regexp"
	"strings"

	slice "github.com/bradfitz/slice"
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
	ActionSleep
	ActionPrompt
	ActionSearchPrompt
	ActionReverseSearchPrompt
	ActionFilterPrompt
	ActionQuestionPrompt
	ActionBranchNamePrompt
	ActionTagNamePrompt
	ActionCustomPrompt
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
	ActionNextTab
	ActionPrevTab
	ActionSelectTabByName
	ActionRemoveView
	ActionAddFilter
	ActionRemoveFilter
	ActionCenterView
	ActionScrollCursorTop
	ActionScrollCursorBottom
	ActionCursorTopView
	ActionCursorMiddleView
	ActionCursorBottomView
	ActionNewTab
	ActionRemoveTab
	ActionAddView
	ActionSplitView
	ActionMouseSelect
	ActionMouseScrollDown
	ActionMouseScrollUp
	ActionCheckoutRef
	ActionCheckoutPreviousRef
	ActionCheckoutCommit
	ActionCreateBranch
	ActionCreateBranchAndCheckout
	ActionCreateTag
	ActionCreateAnnotatedTag
	ActionCreateContextMenu
	ActionCreateCommandOutputView
	ActionCreateMessageBoxView
	ActionShowAvailableActions
	ActionStageFile
	ActionUnstageFile
	ActionCheckoutFile
	ActionCommit
	ActionAmendCommit
	ActionPullRemote
	ActionPushRef
	ActionDeleteRef
	ActionMergeRef
	ActionRebase
	ActionShowHelpView
	ActionNextButton
	ActionPrevButton
)

// ActionCategory defines the type of an action
type ActionCategory int

// The set of ActionCategory values
const (
	ActionCategoryNone ActionCategory = iota
	ActionCategoryMovement
	ActionCategorySearch
	ActionCategoryViewNavigation
	ActionCategoryGeneral
	ActionCategoryViewSpecific
)

// ActionDescriptor describes an action
type ActionDescriptor struct {
	actionKey      string
	actionCategory ActionCategory
	promptAction   bool
	description    string
	keyBindings    map[ViewID][]string
}

var actionDescriptors = map[ActionType]ActionDescriptor{
	ActionNone: {
		actionCategory: ActionCategoryGeneral,
		description:    "Perform no action (NOP)",
	},
	ActionExit: {
		actionKey:      "<grv-exit>",
		actionCategory: ActionCategoryGeneral,
		description:    "Exit GRV",
	},
	ActionSuspend: {
		actionKey:      "<grv-suspend>",
		actionCategory: ActionCategoryGeneral,
		description:    "Suspend GRV",
		keyBindings: map[ViewID][]string{
			ViewAll: {"<C-z>"},
		},
	},
	ActionRunCommand: {
		actionCategory: ActionCategoryGeneral,
		description:    "Run a shell command",
	},
	ActionSleep: {
		actionCategory: ActionCategoryGeneral,
		description:    "Sleep for a specified time",
	},
	ActionPrompt: {
		actionKey:      "<grv-prompt>",
		actionCategory: ActionCategoryGeneral,
		promptAction:   true,
		description:    "GRV Command prompt",
		keyBindings: map[ViewID][]string{
			ViewAll: {PromptText},
		},
	},
	ActionSearchPrompt: {
		actionKey:      "<grv-search-prompt>",
		actionCategory: ActionCategorySearch,
		promptAction:   true,
		description:    "Search forwards",
		keyBindings: map[ViewID][]string{
			ViewAll: {SearchPromptText},
		},
	},
	ActionReverseSearchPrompt: {
		actionKey:      "<grv-reverse-search-prompt>",
		actionCategory: ActionCategorySearch,
		promptAction:   true,
		description:    "Search backwards",
		keyBindings: map[ViewID][]string{
			ViewAll: {ReverseSearchPromptText},
		},
	},
	ActionFilterPrompt: {
		actionKey:      "<grv-filter-prompt>",
		actionCategory: ActionCategoryViewSpecific,
		promptAction:   true,
		description:    "Add filter",
		keyBindings: map[ViewID][]string{
			ViewCommit: {"<C-q>"},
			ViewRef:    {"<C-q>"},
		},
	},
	ActionQuestionPrompt: {
		actionCategory: ActionCategoryGeneral,
		promptAction:   true,
		description:    "Prompt the user with a question",
	},
	ActionBranchNamePrompt: {
		actionCategory: ActionCategoryGeneral,
		promptAction:   true,
		description:    "Specify branch name",
	},
	ActionTagNamePrompt: {
		actionCategory: ActionCategoryGeneral,
		promptAction:   true,
		description:    "Specify tag name",
	},
	ActionCustomPrompt: {
		actionCategory: ActionCategoryGeneral,
		promptAction:   true,
		description:    "Custom prompt for user input",
	},
	ActionSearch: {
		actionCategory: ActionCategorySearch,
		description:    "Perform search forwards",
	},
	ActionReverseSearch: {
		actionCategory: ActionCategorySearch,
		description:    "Perform search backwards",
	},
	ActionSearchFindNext: {
		actionKey:      "<grv-search-find-next>",
		actionCategory: ActionCategorySearch,
		description:    "Move to next search match",
		keyBindings: map[ViewID][]string{
			ViewAll: {"n"},
		},
	},
	ActionSearchFindPrev: {
		actionKey:      "<grv-search-find-prev>",
		actionCategory: ActionCategorySearch,
		description:    "Move to previous search match",
		keyBindings: map[ViewID][]string{
			ViewAll: {"N"},
		},
	},
	ActionClearSearch: {
		actionKey:      "<grv-clear-search>",
		actionCategory: ActionCategorySearch,
		description:    "Clear search",
	},
	ActionShowStatus: {
		actionCategory: ActionCategoryGeneral,
		description:    "Display message in status bar",
	},
	ActionNextLine: {
		actionKey:      "<grv-next-line>",
		actionCategory: ActionCategoryMovement,
		description:    "Move down one line",
		keyBindings: map[ViewID][]string{
			ViewAll: {"<Down>", "j"},
		},
	},
	ActionPrevLine: {
		actionKey:      "<grv-prev-line>",
		actionCategory: ActionCategoryMovement,
		description:    "Move up one line",
		keyBindings: map[ViewID][]string{
			ViewAll: {"<Up>", "k"},
		},
	},
	ActionNextPage: {
		actionKey:      "<grv-next-page>",
		actionCategory: ActionCategoryMovement,
		description:    "Move one page down",
		keyBindings: map[ViewID][]string{
			ViewAll: {"<PageDown>", "<C-f>"},
		},
	},
	ActionPrevPage: {
		actionKey:      "<grv-prev-page>",
		actionCategory: ActionCategoryMovement,
		description:    "Move one page up",
		keyBindings: map[ViewID][]string{
			ViewAll: {"<PageUp>", "<C-b>"},
		},
	},
	ActionNextHalfPage: {
		actionKey:      "<grv-next-half-page>",
		actionCategory: ActionCategoryMovement,
		description:    "Move half page down",
		keyBindings: map[ViewID][]string{
			ViewAll: {"<C-d>"},
		},
	},
	ActionPrevHalfPage: {
		actionKey:      "<grv-prev-half-page>",
		actionCategory: ActionCategoryMovement,
		description:    "Move half page up",
		keyBindings: map[ViewID][]string{
			ViewAll: {"<C-u>"},
		},
	},
	ActionScrollRight: {
		actionKey:      "<grv-scroll-right>",
		actionCategory: ActionCategoryMovement,
		description:    "Scroll right",
		keyBindings: map[ViewID][]string{
			ViewAll: {"<Right>", "l"},
		},
	},
	ActionScrollLeft: {
		actionKey:      "<grv-scroll-left>",
		actionCategory: ActionCategoryMovement,
		description:    "Scroll left",
		keyBindings: map[ViewID][]string{
			ViewAll: {"<Left>", "h"},
		},
	},
	ActionFirstLine: {
		actionKey:      "<grv-first-line>",
		actionCategory: ActionCategoryMovement,
		description:    "Move to first line",
		keyBindings: map[ViewID][]string{
			ViewAll: {"gg"},
		},
	},
	ActionLastLine: {
		actionKey:      "<grv-last-line>",
		actionCategory: ActionCategoryMovement,
		description:    "Move to last line",
		keyBindings: map[ViewID][]string{
			ViewAll: {"G"},
		},
	},
	ActionSelect: {
		actionKey:      "<grv-select>",
		actionCategory: ActionCategoryGeneral,
		description:    "Select item (opens listener view if none exists)",
		keyBindings: map[ViewID][]string{
			ViewAll: {"<Enter>"},
		},
	},
	ActionNextView: {
		actionKey:      "<grv-next-view>",
		actionCategory: ActionCategoryViewNavigation,
		description:    "Move to next view",
		keyBindings: map[ViewID][]string{
			ViewAll: {"<C-w>w", "<C-w><C-w>", "<Tab>"},
		},
	},
	ActionPrevView: {
		actionKey:      "<grv-prev-view>",
		actionCategory: ActionCategoryViewNavigation,
		description:    "Move to previous view",
		keyBindings: map[ViewID][]string{
			ViewAll: {"<C-w>W", "<S-Tab>"},
		},
	},
	ActionFullScreenView: {
		actionKey:      "<grv-full-screen-view>",
		actionCategory: ActionCategoryViewNavigation,
		description:    "Toggle current view full screen",
		keyBindings: map[ViewID][]string{
			ViewAll: {"<C-w>o", "<C-w><C-o>", "f"},
		},
	},
	ActionToggleViewLayout: {
		actionKey:      "<grv-toggle-view-layout>",
		actionCategory: ActionCategoryViewNavigation,
		description:    "Toggle view layout",
		keyBindings: map[ViewID][]string{
			ViewAll: {"<C-w>t"},
		},
	},
	ActionNextTab: {
		actionKey:      "<grv-next-tab>",
		actionCategory: ActionCategoryViewNavigation,
		description:    "Move to next tab",
		keyBindings: map[ViewID][]string{
			ViewAll: {"gt"},
		},
	},
	ActionPrevTab: {
		actionKey:      "<grv-prev-tab>",
		actionCategory: ActionCategoryViewNavigation,
		description:    "Move to previous tab",
		keyBindings: map[ViewID][]string{
			ViewAll: {"gT"},
		},
	},
	ActionSelectTabByName: {
		actionCategory: ActionCategoryViewNavigation,
		description:    "Select active tab by name",
	},
	ActionRemoveView: {
		actionKey:      "<grv-remove-view>",
		actionCategory: ActionCategoryViewNavigation,
		description:    "Close view (or close tab if empty)",
		keyBindings: map[ViewID][]string{
			ViewAll: {"q"},
		},
	},
	ActionAddFilter: {
		actionCategory: ActionCategoryViewSpecific,
		description:    "Add filter",
	},
	ActionRemoveFilter: {
		actionKey:      "<grv-remove-filter>",
		actionCategory: ActionCategoryViewSpecific,
		description:    "Remove filter",
		keyBindings: map[ViewID][]string{
			ViewCommit: {"<C-r>"},
			ViewRef:    {"<C-r>"},
		},
	},
	ActionCenterView: {
		actionKey:      "<grv-center-view>",
		actionCategory: ActionCategoryMovement,
		description:    "Center view",
		keyBindings: map[ViewID][]string{
			ViewAll: {"z.", "zz"},
		},
	},
	ActionScrollCursorTop: {
		actionKey:      "<grv-scroll-cursor-top>",
		actionCategory: ActionCategoryMovement,
		description:    "Scroll the screen so cursor is at the top",
		keyBindings: map[ViewID][]string{
			ViewAll: {"zt"},
		},
	},
	ActionScrollCursorBottom: {
		actionKey:      "<grv-scroll-cursor-bottom>",
		actionCategory: ActionCategoryMovement,
		description:    "Scroll the screen so cursor is at the bottom",
		keyBindings: map[ViewID][]string{
			ViewAll: {"zb"},
		},
	},
	ActionCursorTopView: {
		actionKey:      "<grv-cursor-top-view>",
		actionCategory: ActionCategoryMovement,
		description:    "Move to the first line of the page",
		keyBindings: map[ViewID][]string{
			ViewAll: {"H"},
		},
	},
	ActionCursorMiddleView: {
		actionKey:      "<grv-cursor-middle-view>",
		actionCategory: ActionCategoryMovement,
		description:    "Move to the middle line of the page",
		keyBindings: map[ViewID][]string{
			ViewAll: {"M"},
		},
	},
	ActionCursorBottomView: {
		actionKey:      "<grv-cursor-bottom-view>",
		actionCategory: ActionCategoryMovement,
		description:    "Move to the last line of the page",
		keyBindings: map[ViewID][]string{
			ViewAll: {"L"},
		},
	},
	ActionNewTab: {
		actionCategory: ActionCategoryGeneral,
		description:    "Add a new tab",
	},
	ActionRemoveTab: {
		actionKey:      "<grv-remove-tab>",
		actionCategory: ActionCategoryGeneral,
		description:    "Remove the active tab",
	},
	ActionAddView: {
		actionCategory: ActionCategoryGeneral,
		description:    "Add a new view",
	},
	ActionSplitView: {
		actionCategory: ActionCategoryGeneral,
		description:    "Split the current view with a new view",
	},
	ActionMouseSelect: {
		actionCategory: ActionCategoryGeneral,
		description:    "Mouse select",
	},
	ActionMouseScrollDown: {
		actionCategory: ActionCategoryGeneral,
		description:    "Mouse scroll down",
	},
	ActionMouseScrollUp: {
		actionCategory: ActionCategoryGeneral,
		description:    "Mouse scroll up",
	},
	ActionCheckoutRef: {
		actionKey:      "<grv-checkout-ref>",
		actionCategory: ActionCategoryViewSpecific,
		description:    "Checkout ref",
		keyBindings: map[ViewID][]string{
			ViewRef: {"c"},
		},
	},
	ActionCheckoutPreviousRef: {
		actionKey:      "<grv-checkout-previous-ref>",
		actionCategory: ActionCategoryViewSpecific,
		description:    "Checkout previous ref",
		keyBindings: map[ViewID][]string{
			ViewRef: {"-"},
		},
	},
	ActionCheckoutCommit: {
		actionKey:      "<grv-checkout-commit>",
		actionCategory: ActionCategoryViewSpecific,
		description:    "Checkout commit",
		keyBindings: map[ViewID][]string{
			ViewCommit: {"c"},
		},
	},
	ActionCreateBranch: {
		actionKey:      "<grv-create-branch>",
		actionCategory: ActionCategoryViewSpecific,
		description:    "Create a new branch",
		keyBindings: map[ViewID][]string{
			ViewRef:    {"b"},
			ViewCommit: {"b"},
		},
	},
	ActionCreateBranchAndCheckout: {
		actionKey:      "<grv-create-branch-and-checkout>",
		actionCategory: ActionCategoryViewSpecific,
		description:    "Create a new branch and checkout",
		keyBindings: map[ViewID][]string{
			ViewRef:    {"B"},
			ViewCommit: {"B"},
		},
	},
	ActionCreateTag: {
		actionKey:      "<grv-create-tag>",
		actionCategory: ActionCategoryViewSpecific,
		description:    "Create a new tag",
		keyBindings: map[ViewID][]string{
			ViewRef:    {"t"},
			ViewCommit: {"t"},
		},
	},
	ActionCreateAnnotatedTag: {
		actionKey:      "<grv-create-annotated-tag>",
		actionCategory: ActionCategoryViewSpecific,
		description:    "Create a new annotated tag",
		keyBindings: map[ViewID][]string{
			ViewRef:    {"T"},
			ViewCommit: {"T"},
		},
	},
	ActionCreateContextMenu: {
		actionCategory: ActionCategoryGeneral,
		description:    "Create a context menu",
	},
	ActionCreateCommandOutputView: {
		actionCategory: ActionCategoryGeneral,
		description:    "Create a command output view",
	},
	ActionCreateMessageBoxView: {
		actionCategory: ActionCategoryGeneral,
		description:    "Create a message box view",
	},
	ActionShowAvailableActions: {
		actionKey:      "<grv-show-available-actions>",
		actionCategory: ActionCategoryGeneral,
		description:    "Show available actions for the selected row",
		keyBindings: map[ViewID][]string{
			ViewAll: {"<C-a>"},
		},
	},
	ActionStageFile: {
		actionKey:      "<grv-stage-file>",
		actionCategory: ActionCategoryViewSpecific,
		description:    "Stage",
		keyBindings: map[ViewID][]string{
			ViewGitStatus: {"a"},
		},
	},
	ActionUnstageFile: {
		actionKey:      "<grv-unstage-file>",
		actionCategory: ActionCategoryViewSpecific,
		description:    "Unstage",
		keyBindings: map[ViewID][]string{
			ViewGitStatus: {"u"},
		},
	},
	ActionCheckoutFile: {
		actionKey:      "<grv-checkout-file>",
		actionCategory: ActionCategoryViewSpecific,
		description:    "Checkout",
		keyBindings: map[ViewID][]string{
			ViewGitStatus: {"c"},
		},
	},
	ActionCommit: {
		actionKey:      "<grv-action-commit>",
		actionCategory: ActionCategoryViewSpecific,
		description:    "Commit",
		keyBindings: map[ViewID][]string{
			ViewGitStatus: {"C"},
		},
	},
	ActionAmendCommit: {
		actionKey:      "<grv-action-amend-commit>",
		actionCategory: ActionCategoryViewSpecific,
		description:    "Amend commit",
		keyBindings: map[ViewID][]string{
			ViewGitStatus: {"A"},
		},
	},
	ActionPullRemote: {
		actionKey:      "<grv-pull-remote>",
		actionCategory: ActionCategoryViewSpecific,
		description:    "Pull remote",
		keyBindings: map[ViewID][]string{
			ViewRemote: {"p"},
		},
	},
	ActionPushRef: {
		actionKey:      "<grv-push-ref>",
		actionCategory: ActionCategoryViewSpecific,
		description:    "Push ref to remote",
		keyBindings: map[ViewID][]string{
			ViewRef: {"p"},
		},
	},
	ActionDeleteRef: {
		actionKey:      "<grv-delete-ref>",
		actionCategory: ActionCategoryViewSpecific,
		description:    "Delete ref",
		keyBindings: map[ViewID][]string{
			ViewRef: {"D"},
		},
	},
	ActionMergeRef: {
		actionKey:      "<grv-merge-ref>",
		actionCategory: ActionCategoryViewSpecific,
		description:    "Merge ref into current branch",
		keyBindings: map[ViewID][]string{
			ViewRef: {"m"},
		},
	},
	ActionRebase: {
		actionKey:      "<grv-rebase>",
		actionCategory: ActionCategoryViewSpecific,
		description:    "Rebase current branch onto selected branch",
		keyBindings: map[ViewID][]string{
			ViewRef: {"r"},
		},
	},
	ActionShowHelpView: {
		actionKey:      "<grv-show-help>",
		actionCategory: ActionCategoryGeneral,
		description:    "Show the help view",
	},
	ActionNextButton: {
		actionKey:      "<grv-next-button>",
		actionCategory: ActionCategoryViewSpecific,
		description:    "Select the next button",
		keyBindings: map[ViewID][]string{
			ViewMessageBox: {"<Right>", "l", "<Tab>"},
		},
	},
	ActionPrevButton: {
		actionKey:      "<grv-prev-button>",
		actionCategory: ActionCategoryViewSpecific,
		description:    "Select the previous button",
		keyBindings: map[ViewID][]string{
			ViewMessageBox: {"<Left>", "h", "<S-Tab>"},
		},
	},
}

var whitespaceBindingRegex = regexp.MustCompile(`^(.*\s+.*)+$`)

var actionKeys = map[string]ActionType{}

func init() {
	for actionType, actionDescriptor := range actionDescriptors {
		if actionDescriptor.actionKey != "" {
			actionKeys[actionDescriptor.actionKey] = actionType
		}
	}
}

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

// ActionCreateCommandOutputViewArgs contains arguments to create and configure a command output view
type ActionCreateCommandOutputViewArgs struct {
	command       string
	viewDimension ViewDimension
	onCreation    func(commandOutputProcessor CommandOutputProcessor)
}

// ActionCreateMessageBoxViewArgs contains arguments to create and configure a message box view
type ActionCreateMessageBoxViewArgs struct {
	config MessageBoxConfig
}

// ActionRunCommandArgs contains arguments to run a command and process
// the status and output
type ActionRunCommandArgs struct {
	command        string
	args           []string
	interactive    bool
	promptForInput bool
	noShell        bool
	stdin          io.Reader
	stdout         io.Writer
	stderr         io.Writer
	beforeStart    func(cmd *exec.Cmd)
	onStart        func(cmd *exec.Cmd)
	onComplete     func(err error, exitStatus int) error
}

// ActionCustomPromptArgs contains arguments to display a custom prompt
// and handle the user input
type ActionCustomPromptArgs struct {
	prompt       string
	inputHandler func(input string)
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
	KeyStrings(actionType ActionType, viewID ViewID) (keystrings []BoundKeyString)
	GenerateHelpSections(Config) []*HelpSection
}

// BoundKeyString is a keystring bound to an action
type BoundKeyString struct {
	keystring          string
	userDefinedBinding bool
}

// KeyBindingManager manages key bindings in grv
type KeyBindingManager struct {
	bindings           map[ViewID]*pt.Trie
	helpFormat         map[ActionType]map[ViewID][]BoundKeyString
	userDefinedBinding bool
}

// NewKeyBindingManager creates a new instance
func NewKeyBindingManager() KeyBindings {
	keyBindingManager := &KeyBindingManager{
		bindings:   make(map[ViewID]*pt.Trie),
		helpFormat: make(map[ActionType]map[ViewID][]BoundKeyString),
	}

	keyBindingManager.setDefaultKeyBindings()
	keyBindingManager.userDefinedBinding = true

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
	keyBindingManager.updateHelpFormat(actionType, viewID, keystring)
}

// SetKeystringBinding allows a key sequence to be bound to the provided key sequence and view
func (keyBindingManager *KeyBindingManager) SetKeystringBinding(viewID ViewID, keystring, mappedKeystring string) {
	keyBindingManager.RemoveBinding(viewID, keystring)

	viewBindings := keyBindingManager.getOrCreateViewBindings(viewID)
	viewBindings.Set(pt.Prefix(keystring), newKeystringBinding(mappedKeystring))

	if actionType, ok := actionKeys[mappedKeystring]; ok {
		keyBindingManager.updateHelpFormat(actionType, viewID, keystring)
	}
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

func (keyBindingManager *KeyBindingManager) updateHelpFormat(actionType ActionType, viewID ViewID, keystring string) {
	if strings.HasPrefix(keystring, "<grv-") {
		return
	}

	viewBindings, ok := keyBindingManager.helpFormat[actionType]
	if !ok {
		viewBindings = map[ViewID][]BoundKeyString{}
		keyBindingManager.helpFormat[actionType] = viewBindings
	}

	keystrings, ok := viewBindings[viewID]
	if !ok {
		keystrings = []BoundKeyString{}
	}

	viewBindings[viewID] = append(keystrings, BoundKeyString{
		keystring:          keystring,
		userDefinedBinding: keyBindingManager.userDefinedBinding,
	})
}

// RemoveBinding removes the binding for the provided keystring if it exists
func (keyBindingManager *KeyBindingManager) RemoveBinding(viewID ViewID, keystring string) (removed bool) {
	binding, _ := keyBindingManager.Binding([]ViewID{viewID}, keystring)

	if viewBindings, ok := keyBindingManager.bindings[viewID]; ok {
		removed = viewBindings.Delete(pt.Prefix(keystring))
	}

	if binding.actionType != ActionNone || binding.keystring != "" {
		keyBindingManager.removeHelpFormatEntry(binding, viewID, keystring)
	}

	return
}

func (keyBindingManager *KeyBindingManager) removeHelpFormatEntry(binding Binding, viewID ViewID, keystring string) {
	var actionType ActionType

	if binding.bindingType == BtAction {
		actionType = binding.actionType
	} else if binding.bindingType == BtKeystring {
		if mappedActionType, ok := actionKeys[binding.keystring]; ok {
			actionType = mappedActionType
		}
	}

	viewBindings, ok := keyBindingManager.helpFormat[actionType]
	if !ok {
		return
	}

	keystrings, ok := viewBindings[viewID]
	if !ok {
		return
	}

	updatedKeystrings := []BoundKeyString{}

	for _, key := range keystrings {
		if key.keystring != keystring {
			updatedKeystrings = append(updatedKeystrings, key)
		}
	}

	viewBindings[viewID] = updatedKeystrings
}

// KeyStrings returns the keystrings bound to the provided action and view
func (keyBindingManager *KeyBindingManager) KeyStrings(actionType ActionType, viewID ViewID) (keystrings []BoundKeyString) {
	viewBindings, ok := keyBindingManager.helpFormat[actionType]
	if !ok {
		return
	}

	keystrings, _ = viewBindings[viewID]
	return
}

func (keyBindingManager *KeyBindingManager) setDefaultKeyBindings() {
	for actionKey, actionType := range actionKeys {
		keyBindingManager.SetActionBinding(ViewAll, actionKey, actionType)
	}

	for actionType, actionDescriptor := range actionDescriptors {
		for viewID, keys := range actionDescriptor.keyBindings {
			for _, key := range keys {
				keyBindingManager.SetActionBinding(viewID, key, actionType)
			}
		}
	}
}

// GenerateHelpSections generates key binding help sections
func (keyBindingManager *KeyBindingManager) GenerateHelpSections(config Config) []*HelpSection {
	helpSections := []*HelpSection{
		{
			title: HelpSectionText{text: "Key Bindings"},
			description: []HelpSectionText{
				{text: "The following tables contain default and user configured key bindings"},
			},
		},
	}

	type KeyBindingSection struct {
		title        string
		actionFilter actionFilter
	}

	keyBindingSections := []KeyBindingSection{
		{
			title: "Movement",
			actionFilter: func(actionDescriptor ActionDescriptor) bool {
				return actionDescriptor.actionCategory == ActionCategoryMovement
			},
		},
		{
			title: "Search",
			actionFilter: func(actionDescriptor ActionDescriptor) bool {
				return actionDescriptor.actionCategory == ActionCategorySearch
			},
		},
		{
			title: "View Navigation",
			actionFilter: func(actionDescriptor ActionDescriptor) bool {
				return actionDescriptor.actionCategory == ActionCategoryViewNavigation
			},
		},
		{
			title: "General",
			actionFilter: func(actionDescriptor ActionDescriptor) bool {
				return actionDescriptor.actionCategory == ActionCategoryGeneral
			},
		},
	}

	for _, KeyBindingSection := range keyBindingSections {
		helpSections = append(helpSections, &HelpSection{
			description: []HelpSectionText{
				{text: KeyBindingSection.title, themeComponentID: CmpHelpViewSectionSubTitle},
			},
			tableFormatter: keyBindingManager.generateKeyBindingsTable(config, KeyBindingSection.actionFilter, ViewAll),
		})
	}

	helpSections = append(helpSections, keyBindingManager.generateViewSpecificKeyBindingHelpSections(config)...)

	return helpSections
}

func (keyBindingManager *KeyBindingManager) generateViewSpecificKeyBindingHelpSections(config Config) (helpSections []*HelpSection) {
	viewSpecificIDMap := map[ViewID]bool{}

	for _, actionDescriptor := range actionDescriptors {
		if actionDescriptor.actionCategory == ActionCategoryViewSpecific {
			for viewID := range actionDescriptor.keyBindings {
				viewSpecificIDMap[viewID] = true
			}
		}
	}

	viewSpecificIDs := []ViewID{}
	for viewID := range viewSpecificIDMap {
		viewSpecificIDs = append(viewSpecificIDs, viewID)
	}

	slice.Sort(viewSpecificIDs, func(i, j int) bool {
		return viewSpecificIDs[i] < viewSpecificIDs[j]
	})

	for _, viewID := range viewSpecificIDs {
		helpSections = append(helpSections, &HelpSection{
			description: []HelpSectionText{
				{text: fmt.Sprintf("%v Specific", ViewName(viewID)), themeComponentID: CmpHelpViewSectionSubTitle},
			},
			tableFormatter: keyBindingManager.generateKeyBindingsTable(config, func(actionDescriptor ActionDescriptor) bool {
				if actionDescriptor.actionCategory == ActionCategoryViewSpecific {
					if actionDescriptor.keyBindings != nil {
						if _, ok := actionDescriptor.keyBindings[viewID]; ok {
							return true
						}
					}
				}

				return false
			}, viewID),
		})
	}

	return
}

type actionFilter func(ActionDescriptor) bool

func (keyBindingManager *KeyBindingManager) generateKeyBindingsTable(config Config, filter actionFilter, viewID ViewID) *TableFormatter {
	headers := []TableHeader{
		{text: "Key Bindings", themeComponentID: CmpHelpViewSectionTableHeader},
		{text: "Action", themeComponentID: CmpHelpViewSectionTableHeader},
		{text: "Description", themeComponentID: CmpHelpViewSectionTableHeader},
	}

	tableFormatter := NewTableFormatterWithHeaders(headers, config)
	tableFormatter.SetGridLines(true)

	type matchingActionDescriptor struct {
		actionType       ActionType
		actionDescriptor ActionDescriptor
	}

	matchingActionDescriptors := []matchingActionDescriptor{}

	for actionType, actionDescriptor := range actionDescriptors {
		if actionDescriptor.actionKey != "" && filter(actionDescriptor) {
			matchingActionDescriptors = append(matchingActionDescriptors, matchingActionDescriptor{
				actionType:       actionType,
				actionDescriptor: actionDescriptor,
			})
		}
	}

	slice.Sort(matchingActionDescriptors, func(i, j int) bool {
		return matchingActionDescriptors[i].actionDescriptor.actionKey < matchingActionDescriptors[j].actionDescriptor.actionKey
	})

	tableFormatter.Resize(uint(len(matchingActionDescriptors)))

	for rowIndex, matchingActionDescriptor := range matchingActionDescriptors {
		seenKeyBindings := map[string]bool{}
		keyBindings := []BoundKeyString{}

		viewIDs := []ViewID{}
		if len(matchingActionDescriptor.actionDescriptor.keyBindings) == 0 {
			viewIDs = append(viewIDs, ViewAll)
		} else {
			viewIDs = append(viewIDs, viewID)

			if viewID != ViewAll {
				viewIDs = append(viewIDs, ViewAll)
			}
		}

		for _, viewID := range viewIDs {
			for _, keyBinding := range keyBindingManager.KeyStrings(matchingActionDescriptor.actionType, viewID) {
				if _, exists := seenKeyBindings[keyBinding.keystring]; !exists {
					keyBindings = append(keyBindings, keyBinding)
					seenKeyBindings[keyBinding.keystring] = true
				}
			}
		}

		if len(keyBindings) == 0 {
			tableFormatter.SetCellWithStyle(uint(rowIndex), 0, CmpHelpViewSectionTableCellSeparator, "%v", "None")
		} else {
			for bindingIndex, keyBinding := range keyBindings {
				themeComponentID := CmpHelpViewSectionTableRow
				if keyBinding.userDefinedBinding {
					themeComponentID = CmpHelpViewSectionTableRowHighlighted
				}

				keystringContainsWhitespace := whitespaceBindingRegex.MatchString(keyBinding.keystring)

				if keystringContainsWhitespace {
					tableFormatter.AppendToCellWithStyle(uint(rowIndex), 0, themeComponentID, `"`)
				}

				tableFormatter.AppendToCellWithStyle(uint(rowIndex), 0, themeComponentID, "%v", keyBinding.keystring)

				if keystringContainsWhitespace {
					tableFormatter.AppendToCellWithStyle(uint(rowIndex), 0, themeComponentID, `"`)
				}

				if bindingIndex != len(keyBindings)-1 {
					tableFormatter.AppendToCellWithStyle(uint(rowIndex), 0, CmpHelpViewSectionTableCellSeparator, "%v", ", ")
				}
			}
		}

		tableFormatter.SetCellWithStyle(uint(rowIndex), 1, CmpHelpViewSectionTableRow, "%v", matchingActionDescriptor.actionDescriptor.actionKey)
		tableFormatter.SetCellWithStyle(uint(rowIndex), 2, CmpHelpViewSectionTableRow, "%v", matchingActionDescriptor.actionDescriptor.description)
	}

	return tableFormatter
}

func isValidAction(action string) bool {
	_, valid := actionKeys[action]
	return valid
}

// IsPromptAction returns true if the action presents a prompt
func IsPromptAction(actionType ActionType) bool {
	if actionDescriptor, exists := actionDescriptors[actionType]; exists {
		return actionDescriptor.promptAction
	}

	return false
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
		ActionType: ActionCreateMessageBoxView,
		Args: []interface{}{ActionCreateMessageBoxViewArgs{
			config: MessageBoxConfig{
				Title:   "Confirm",
				Message: question,
				Buttons: []MessageBoxButton{ButtonYes, ButtonNo},
				OnSelect: func(button MessageBoxButton) {
					var response QuestionResponse

					if button == ButtonYes {
						response = ResponseYes
					} else {
						response = ResponseNo
					}

					onResponse(response)
				},
			},
		}},
	}
}
