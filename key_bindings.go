package main

import (
	pt "github.com/tchap/go-patricia/patricia"
)

type Action int

const (
	ACTION_NONE Action = iota
	ACTION_EXIT
	ACTION_PROMPT
	ACTION_NEXT_LINE
	ACTION_PREV_LINE
	ACTION_SCROLL_RIGHT
	ACTION_SCROLL_LEFT
	ACTION_FIRST_LINE
	ACTION_LAST_LINE
	ACTION_SELECT
	ACTION_NEXT_VIEW
	ACTION_PREV_VIEW
	ACTION_FULL_SCREEN_VIEW
)

var actionKeys = map[string]Action{
	"<grv-nop>":              ACTION_NONE,
	"<grv-exit>":             ACTION_EXIT,
	"<grv-prompt>":           ACTION_PROMPT,
	"<grv-next-line>":        ACTION_NEXT_LINE,
	"<grv-prev-line>":        ACTION_PREV_LINE,
	"<grv-scroll-right>":     ACTION_SCROLL_RIGHT,
	"<grv-scroll-left>":      ACTION_SCROLL_LEFT,
	"<grv-first-line>":       ACTION_FIRST_LINE,
	"<grv-last-line>":        ACTION_LAST_LINE,
	"<grv-select>":           ACTION_SELECT,
	"<grv-next-view>":        ACTION_NEXT_VIEW,
	"<grv-prev-view>":        ACTION_PREV_VIEW,
	"<grv-full-screen-view>": ACTION_FULL_SCREEN_VIEW,
}

type ViewHierarchy []ViewId

type BindingType int

const (
	BT_ACTION BindingType = iota
	BT_KEYSTRING
)

type Binding struct {
	bindingType BindingType
	action      Action
	keystring   string
}

func NewActionBinding(action Action) Binding {
	return Binding{
		bindingType: BT_ACTION,
		action:      action,
	}
}

func NewKeystringBinding(keystring string) Binding {
	return Binding{
		bindingType: BT_KEYSTRING,
		keystring:   keystring,
		action:      ACTION_NONE,
	}
}

type KeyBindings interface {
	Binding(viewHierarchy ViewHierarchy, keystring string) (binding Binding, isPrefix bool)
	SetActionBinding(viewId ViewId, keystring string, action Action)
	SetKeystringBinding(viewId ViewId, keystring, mappedKeystring string)
}

type KeyBindingManager struct {
	bindings map[ViewId]*pt.Trie
}

func NewKeyBindingManager() KeyBindings {
	keyBindingManager := &KeyBindingManager{
		bindings: make(map[ViewId]*pt.Trie),
	}

	keyBindingManager.setDefaultKeyBindings()

	return keyBindingManager
}

func (keyBindingManager *KeyBindingManager) Binding(viewHierarchy ViewHierarchy, keystring string) (Binding, bool) {
	viewHierarchy = append(viewHierarchy, VIEW_ALL)
	isPrefix := false

	for _, viewId := range viewHierarchy {
		if viewBindings, ok := keyBindingManager.bindings[viewId]; ok {
			if binding := viewBindings.Get(pt.Prefix(keystring)); binding != nil {
				return binding.(Binding), false
			} else if viewBindings.MatchSubtree(pt.Prefix(keystring)) {
				isPrefix = true
			}
		}
	}

	return NewActionBinding(ACTION_NONE), isPrefix
}

func (keyBindingManager *KeyBindingManager) SetActionBinding(viewId ViewId, keystring string, action Action) {
	viewBindings := keyBindingManager.getOrCreateViewBindings(viewId)
	viewBindings.Set(pt.Prefix(keystring), NewActionBinding(action))
}

func (keyBindingManager *KeyBindingManager) SetKeystringBinding(viewId ViewId, keystring, mappedKeystring string) {
	viewBindings := keyBindingManager.getOrCreateViewBindings(viewId)
	viewBindings.Set(pt.Prefix(keystring), NewKeystringBinding(mappedKeystring))
}

func (keyBindingManager *KeyBindingManager) getOrCreateViewBindings(viewId ViewId) *pt.Trie {
	if viewBindings, ok := keyBindingManager.bindings[viewId]; ok {
		return viewBindings
	} else {
		viewBindings = pt.NewTrie()
		keyBindingManager.bindings[viewId] = viewBindings
		return viewBindings
	}
}

func (keyBindingManager *KeyBindingManager) setDefaultKeyBindings() {
	for actionKey, action := range actionKeys {
		keyBindingManager.SetActionBinding(VIEW_ALL, actionKey, action)
	}

	// Main View
	keyBindingManager.SetActionBinding(VIEW_MAIN, PROMPT_TEXT, ACTION_PROMPT)

	// All Views
	keyBindingManager.SetActionBinding(VIEW_ALL, "<Up>", ACTION_PREV_LINE)
	keyBindingManager.SetActionBinding(VIEW_ALL, "k", ACTION_PREV_LINE)

	keyBindingManager.SetActionBinding(VIEW_ALL, "<Down>", ACTION_NEXT_LINE)
	keyBindingManager.SetActionBinding(VIEW_ALL, "j", ACTION_NEXT_LINE)

	keyBindingManager.SetActionBinding(VIEW_ALL, "<Right>", ACTION_SCROLL_RIGHT)
	keyBindingManager.SetActionBinding(VIEW_ALL, "l", ACTION_SCROLL_RIGHT)

	keyBindingManager.SetActionBinding(VIEW_ALL, "<Left>", ACTION_SCROLL_LEFT)
	keyBindingManager.SetActionBinding(VIEW_ALL, "h", ACTION_SCROLL_LEFT)

	keyBindingManager.SetActionBinding(VIEW_ALL, "gg", ACTION_FIRST_LINE)
	keyBindingManager.SetActionBinding(VIEW_ALL, "G", ACTION_LAST_LINE)

	keyBindingManager.SetActionBinding(VIEW_ALL, "<Tab>", ACTION_NEXT_VIEW)
	keyBindingManager.SetActionBinding(VIEW_ALL, "<C-w><C-w>", ACTION_NEXT_VIEW)

	keyBindingManager.SetActionBinding(VIEW_ALL, "f", ACTION_FULL_SCREEN_VIEW)
	keyBindingManager.SetActionBinding(VIEW_ALL, "<C-w>o", ACTION_FULL_SCREEN_VIEW)

	keyBindingManager.SetActionBinding(VIEW_ALL, "<Return>", ACTION_SELECT)
}

func isValidAction(action string) bool {
	_, valid := actionKeys[action]
	return valid
}
