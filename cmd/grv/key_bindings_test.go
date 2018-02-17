package main

import (
	"reflect"
	"testing"
)

func checkBinding(expectedBinding Binding, expectedIsPrefix bool, actualBinding Binding, actualIsPrefix bool, t *testing.T) {
	if expectedBinding != actualBinding {
		t.Errorf("Binding does not match expected value. Expected: %v, Actual: %v", expectedBinding, actualBinding)
	}

	if expectedIsPrefix != actualIsPrefix {
		t.Errorf("isPrefix does not match expected value. Expected: %v, Actual: %v", expectedIsPrefix, actualIsPrefix)
	}
}

func TestDefaultKeyBindingsAreSetOnCreation(t *testing.T) {
	keyBindings := NewKeyBindingManager()

	expectedBinding := newActionBinding(ActionNextLine)
	binding, isPrefix := keyBindings.Binding(ViewHierarchy([]ViewID{ViewMain, ViewHistory, ViewCommit}), "j")
	checkBinding(binding, isPrefix, expectedBinding, false, t)

	expectedBinding = newActionBinding(ActionNextView)
	binding, isPrefix = keyBindings.Binding(ViewHierarchy([]ViewID{ViewMain, ViewHistory, ViewDiff}), "<Tab>")
	checkBinding(binding, isPrefix, expectedBinding, false, t)
}

func TestActionBindingCanBeSet(t *testing.T) {
	keyBindings := NewKeyBindingManager()

	keyBindings.SetActionBinding(ViewRef, "aaa", ActionFirstLine)
	binding, isPrefix := keyBindings.Binding(ViewHierarchy([]ViewID{ViewMain, ViewHistory, ViewRef}), "aaa")

	expectedBinding := newActionBinding(ActionFirstLine)
	checkBinding(binding, isPrefix, expectedBinding, false, t)
}

func TestKeystringBindingCanBeSet(t *testing.T) {
	keyBindings := NewKeyBindingManager()

	keyBindings.SetKeystringBinding(ViewRef, "aaa", "bbb")
	binding, isPrefix := keyBindings.Binding(ViewHierarchy([]ViewID{ViewMain, ViewHistory, ViewRef}), "aaa")

	expectedBinding := newKeystringBinding("bbb")
	checkBinding(binding, isPrefix, expectedBinding, false, t)
}

func TestBindingPrefixIsRecognised(t *testing.T) {
	keyBindings := NewKeyBindingManager()

	keyBindings.SetActionBinding(ViewRef, "aaa", ActionFirstLine)
	binding, isPrefix := keyBindings.Binding(ViewHierarchy([]ViewID{ViewMain, ViewHistory, ViewRef}), "aa")

	expectedBinding := newActionBinding(ActionNone)
	checkBinding(binding, isPrefix, expectedBinding, true, t)
}

func TestNonExistentBindingReturnsNoAction(t *testing.T) {
	keyBindings := NewKeyBindingManager()

	binding, isPrefix := keyBindings.Binding(ViewHierarchy([]ViewID{ViewMain, ViewHistory, ViewRef}), "aaaaaaaa")

	expectedBinding := newActionBinding(ActionNone)
	checkBinding(binding, isPrefix, expectedBinding, false, t)
}

func TestParentViewBindingHasPriorityOverChildBinding(t *testing.T) {
	keyBindings := NewKeyBindingManager()

	keyBindings.SetActionBinding(ViewRef, "aaa", ActionFirstLine)
	keyBindings.SetActionBinding(ViewHistory, "aaa", ActionLastLine)

	binding, isPrefix := keyBindings.Binding(ViewHierarchy([]ViewID{ViewMain, ViewHistory, ViewRef}), "aaa")

	expectedBinding := newActionBinding(ActionLastLine)
	checkBinding(binding, isPrefix, expectedBinding, false, t)
}

func TestViewAllBindingIsAvailableInAllViews(t *testing.T) {
	keyBindings := NewKeyBindingManager()

	keyBindings.SetActionBinding(ViewAll, "aaa", ActionFirstLine)
	expectedBinding := newActionBinding(ActionFirstLine)

	binding, isPrefix := keyBindings.Binding(ViewHierarchy([]ViewID{ViewMain, ViewHistory, ViewRef}), "aaa")
	checkBinding(binding, isPrefix, expectedBinding, false, t)

	binding, isPrefix = keyBindings.Binding(ViewHierarchy([]ViewID{ViewMain, ViewHistory, ViewCommit}), "aaa")
	checkBinding(binding, isPrefix, expectedBinding, false, t)

	binding, isPrefix = keyBindings.Binding(ViewHierarchy([]ViewID{ViewMain, ViewHistory}), "aaa")
	checkBinding(binding, isPrefix, expectedBinding, false, t)
}

func TestDefaultKeyBindingsReturnsBindings(t *testing.T) {
	tests := map[ActionType][]string{
		ActionFirstLine:    {"gg"},
		ActionScrollLeft:   {"<Left>", "h"},
		ActionFilterPrompt: {"<C-q>"},
	}

	for actionType, expectedKeys := range tests {
		actualKeys := DefaultKeyBindings(actionType, ViewCommit)

		if !reflect.DeepEqual(expectedKeys, actualKeys) {
			t.Errorf("DefaultKeyBindings result did not match expected result. Expected: %v, Actual: %v", expectedKeys, actualKeys)
		}
	}
}

func TestIsPromptActionCorrectlyIdentifiesPromptActions(t *testing.T) {
	tests := map[ActionType]bool{
		ActionPrompt:              true,
		ActionSearchPrompt:        true,
		ActionReverseSearchPrompt: true,
		ActionFilterPrompt:        true,
		ActionSearch:              false,
		ActionLastLine:            false,
		ActionRemoveTab:           false,
		ActionPrevPage:            false,
	}

	for actionType, expectedValue := range tests {
		if IsPromptAction(actionType) != expectedValue {
			t.Errorf("IsPromptAction did not returns expected value %v for action %v", expectedValue, actionType)
		}
	}
}
