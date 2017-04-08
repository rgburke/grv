package main

import (
	"fmt"
	gc "github.com/rthornton128/goncurses"
)

type Action int

const (
	ACTION_NONE Action = iota

	ACTION_HISTORY_VIEW_NEXT_VIEW

	ACTION_REF_VIEW_SELECT_REF
	ACTION_REF_VIEW_PREV_REF
	ACTION_REF_VIEW_NEXT_REF
	ACTION_REF_VIEW_SCROLL_RIGHT
	ACTION_REF_VIEW_SCROLL_LEFT

	ACTION_COMMIT_VIEW_PREV_COMMIT
	ACTION_COMMIT_VIEW_NEXT_COMMIT
	ACTION_COMMIT_VIEW_SCROLL_RIGHT
	ACTION_COMMIT_VIEW_SCROLL_LEFT

	ACTION_DIFF_VIEW_PREV_LINE
	ACTION_DIFF_VIEW_NEXT_LINE
	ACTION_DIFF_VIEW_SCROLL_RIGHT
	ACTION_DIFF_VIEW_SCROLL_LEFT
)

type ViewHierarchy []ViewId

type KeyBindings interface {
	Action(ViewHierarchy, gc.Key) Action
	SetKeyBinding(gc.Key, ViewId, Action)
}

type KeyBindingManager struct {
	bindings map[ViewId]map[gc.Key]Action
}

func NewKeyBindingManager() KeyBindings {
	return &KeyBindingManager{
		bindings: map[ViewId]map[gc.Key]Action{
			VIEW_MAIN: map[gc.Key]Action{},
			VIEW_HISTORY: map[gc.Key]Action{
				gc.KEY_TAB: ACTION_HISTORY_VIEW_NEXT_VIEW,
			},
			VIEW_REF: map[gc.Key]Action{
				gc.KEY_UP:     ACTION_REF_VIEW_PREV_REF,
				gc.KEY_DOWN:   ACTION_REF_VIEW_NEXT_REF,
				gc.KEY_RIGHT:  ACTION_REF_VIEW_SCROLL_RIGHT,
				gc.KEY_LEFT:   ACTION_REF_VIEW_SCROLL_LEFT,
				gc.KEY_RETURN: ACTION_REF_VIEW_SELECT_REF,
			},
			VIEW_COMMIT: map[gc.Key]Action{
				gc.KEY_UP:    ACTION_COMMIT_VIEW_PREV_COMMIT,
				gc.KEY_DOWN:  ACTION_COMMIT_VIEW_NEXT_COMMIT,
				gc.KEY_RIGHT: ACTION_COMMIT_VIEW_SCROLL_RIGHT,
				gc.KEY_LEFT:  ACTION_COMMIT_VIEW_SCROLL_LEFT,
			},
			VIEW_DIFF: map[gc.Key]Action{
				gc.KEY_UP:    ACTION_DIFF_VIEW_PREV_LINE,
				gc.KEY_DOWN:  ACTION_DIFF_VIEW_NEXT_LINE,
				gc.KEY_RIGHT: ACTION_DIFF_VIEW_SCROLL_RIGHT,
				gc.KEY_LEFT:  ACTION_DIFF_VIEW_SCROLL_LEFT,
			},
		},
	}
}

func (keyBindingManager *KeyBindingManager) Action(viewHierarchy ViewHierarchy, key gc.Key) Action {
	for _, viewId := range viewHierarchy {
		if keyBindings, ok := keyBindingManager.bindings[viewId]; ok {
			if action, ok := keyBindings[key]; ok {
				return action
			}
		} else {
			panic(fmt.Sprintf("No key bindings map defined for view with Id %v", viewId))
		}
	}

	return ACTION_NONE
}

func (keyBindingManager *KeyBindingManager) SetKeyBinding(key gc.Key, viewId ViewId, action Action) {
	if keyBindings, ok := keyBindingManager.bindings[viewId]; ok {
		keyBindings[key] = action
	} else {
		panic(fmt.Sprintf("No key bindings map defined for view with Id %v", viewId))
	}
}
