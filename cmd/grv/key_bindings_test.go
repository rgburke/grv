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

func TestKeyStringsReturnsExpectedBoundKeys(t *testing.T) {
	keyBindings := NewKeyBindingManager()

	keyBindings.SetActionBinding(ViewAll, "aaa", ActionFirstLine)
	keyBindings.SetKeystringBinding(ViewAll, "bbb", "<grv-first-line>")

	expectedKeystrings := []BoundKeyString{
		{
			keystring:          "gg",
			userDefinedBinding: false,
		},
		{
			keystring:          "aaa",
			userDefinedBinding: true,
		},
		{
			keystring:          "bbb",
			userDefinedBinding: true,
		},
	}

	actualKeystrings := keyBindings.KeyStrings(ActionFirstLine, ViewAll)

	if !reflect.DeepEqual(expectedKeystrings, actualKeystrings) {
		t.Errorf("Returned keystrings did not match expected value. Expected: %v, Actual: %v", expectedKeystrings, actualKeystrings)
	}
}

func TestRemoveBindingRemovesBinding(t *testing.T) {
	keyBindings := NewKeyBindingManager()

	keyBindings.SetActionBinding(ViewRef, "aaa", ActionFirstLine)
	removed := keyBindings.RemoveBinding(ViewRef, "aaa")
	binding, isPrefix := keyBindings.Binding(ViewHierarchy([]ViewID{ViewMain, ViewHistory, ViewRef}), "aaa")

	if !removed {
		t.Errorf("Expected binding to be removed")
	}

	expectedBinding := newActionBinding(ActionNone)
	checkBinding(binding, isPrefix, expectedBinding, false, t)
}

func TestRemoveBindingRemovesNothingWhenNoBindingExists(t *testing.T) {
	keyBindings := NewKeyBindingManager()

	removed := keyBindings.RemoveBinding(ViewRef, "aaa")

	if removed {
		t.Errorf("Expected no binding to be removed")
	}
}

func TestRemoveBindingDoesNotAffectSubTreeBindings(t *testing.T) {
	keyBindings := NewKeyBindingManager()

	keyBindings.SetActionBinding(ViewRef, "aaa", ActionFirstLine)
	keyBindings.SetActionBinding(ViewRef, "aaaa", ActionLastLine)
	keyBindings.RemoveBinding(ViewRef, "aaa")
	binding, isPrefix := keyBindings.Binding(ViewHierarchy([]ViewID{ViewMain, ViewHistory, ViewRef}), "aaaa")

	expectedBinding := newActionBinding(ActionLastLine)
	checkBinding(binding, isPrefix, expectedBinding, false, t)
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

func TestMouseEventActionReturnsExpectedActions(t *testing.T) {
	mouseEventLeftClick := MouseEvent{
		mouseEventType: MetLeftClick,
		row:            10,
		col:            20,
	}

	tests := map[MouseEvent]Action{
		mouseEventLeftClick: {
			ActionType: ActionMouseSelect,
			Args:       []interface{}{mouseEventLeftClick},
		},
		{mouseEventType: MetScrollDown}: {ActionType: ActionMouseScrollDown},
		{mouseEventType: MetScrollUp}:   {ActionType: ActionMouseScrollUp},
	}

	for mouseEvent, expectedAction := range tests {
		actualAction, err := MouseEventAction(mouseEvent)

		if err != nil {
			t.Errorf("MouseEventAction failed with error: %v", err)
		} else if !reflect.DeepEqual(actualAction, expectedAction) {
			t.Errorf("Returned action did not match expected action. Actual: %v. Expected: %v", actualAction, expectedAction)
		}
	}
}

func TestMouseEventActionReturnsAnErrorForAnInvalidMouseEventType(t *testing.T) {
	_, err := MouseEventAction(MouseEvent{mouseEventType: MouseEventType(-5)})

	if err == nil {
		t.Errorf("Expected MouseEventAction to return error for invalid MouseEventType")
	}
}

func TestGetMouseEventFromActionExtractsMouseEventFromAction(t *testing.T) {
	mouseEventLeftClick := MouseEvent{
		mouseEventType: MetLeftClick,
		row:            10,
		col:            20,
	}
	action := Action{
		ActionType: ActionMouseSelect,
		Args:       []interface{}{mouseEventLeftClick},
	}

	actualMouseEvent, err := GetMouseEventFromAction(action)

	if err != nil {
		t.Errorf("GetMouseEventFromAction failed with error: %v", err)
	} else if !reflect.DeepEqual(actualMouseEvent, mouseEventLeftClick) {
		t.Errorf("Returned MouseEvent did not match expected event. Actual: %v. Expected: %v", actualMouseEvent, mouseEventLeftClick)
	}
}

func TestGetMouseEventReturnsErrorsForInvalidActions(t *testing.T) {
	if _, err := GetMouseEventFromAction(Action{}); err == nil {
		t.Errorf("Expected GetMouseEventFromAction to return error for action with empty Args")
	}

	if _, err := GetMouseEventFromAction(Action{Args: []interface{}{5}}); err == nil {
		t.Errorf("Expected GetMouseEventFromAction to return error for action with invalid Args")
	}
}

func TestActionDescriptorsHaveADescription(t *testing.T) {
	for actionType, actionDescriptor := range actionDescriptors {
		if actionDescriptor.actionCategory == ActionCategoryNone {
			t.Errorf("ActionDescriptor for ActionType %v and ActionKey %v has no category specified", actionType, actionDescriptor.actionKey)
		}
		if actionDescriptor.description == "" {
			t.Errorf("ActionDescriptor for ActionType %v and ActionKey %v has no description", actionType, actionDescriptor.actionKey)
		}
	}
}
