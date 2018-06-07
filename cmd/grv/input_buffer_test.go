package main

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/mock"
)

type MockKeyBindings struct {
	mock.Mock
}

func (keyBindings *MockKeyBindings) Binding(viewHierarchy ViewHierarchy, keystring string) (binding Binding, isPrefix bool) {
	args := keyBindings.Called(viewHierarchy, keystring)
	return args.Get(0).(Binding), args.Bool(1)
}

func (keyBindings *MockKeyBindings) SetActionBinding(viewID ViewID, keystring string, actionType ActionType) {
	keyBindings.Called(viewID, keystring, actionType)
}

func (keyBindings *MockKeyBindings) SetKeystringBinding(viewID ViewID, keystring, mappedKeystring string) {
	keyBindings.Called(viewID, keystring, mappedKeystring)
}

func (keyBindings *MockKeyBindings) RemoveBinding(viewID ViewID, keystring string) (removed bool) {
	args := keyBindings.Called(viewID, keystring)
	return args.Bool(0)
}

func (keyBindings *MockKeyBindings) KeyStrings(actionType ActionType, viewID ViewID) []BoundKeyString {
	args := keyBindings.Called(actionType, viewID)
	return args.Get(0).([]BoundKeyString)
}

func (keyBindings *MockKeyBindings) GenerateHelpSections(config Config) []*HelpSection {
	args := keyBindings.Called(config)
	return args.Get(0).([]*HelpSection)
}

func checkProcessResult(expectedAction Action, expectedKeystring string, actualAction Action, actualKeystring string, t *testing.T) {
	if !reflect.DeepEqual(expectedAction, actualAction) {
		t.Errorf("Returned action does not match expected value. Expected: %v, Actual: %v", expectedAction, actualAction)
	}

	if expectedKeystring != actualKeystring {
		t.Errorf("Returned keystring does not match expected value. Expected: %v, Actual: %v", expectedKeystring, actualKeystring)
	}
}

func TestEmptyInputBufferReturnsNoAction(t *testing.T) {
	keyBindings := &MockKeyBindings{}
	inputBuffer := NewInputBuffer(keyBindings)

	viewHierarchy := ViewHierarchy([]ViewID{ViewMain, ViewHistory, ViewCommit})
	action, keyString := inputBuffer.Process(viewHierarchy)

	checkProcessResult(Action{ActionType: ActionNone}, "", action, keyString, t)
}

func TestSingleKeyPressIsMappedToBinding(t *testing.T) {
	keyBindings := &MockKeyBindings{}
	inputBuffer := NewInputBuffer(keyBindings)

	viewHierarchy := ViewHierarchy([]ViewID{ViewMain, ViewHistory, ViewCommit})

	keyBindings.On("Binding", viewHierarchy, "a").Return(newActionBinding(ActionFirstLine), false)

	inputBuffer.Append("a")
	action, keyString := inputBuffer.Process(viewHierarchy)

	checkProcessResult(Action{ActionType: ActionFirstLine}, "a", action, keyString, t)
}

func TestMultiKeyPressIsMappedToBinding(t *testing.T) {
	keyBindings := &MockKeyBindings{}
	inputBuffer := NewInputBuffer(keyBindings)

	viewHierarchy := ViewHierarchy([]ViewID{ViewMain, ViewHistory, ViewCommit})

	keyBindings.On("Binding", viewHierarchy, "a").Return(newActionBinding(ActionNone), true)
	keyBindings.On("Binding", viewHierarchy, "ab").Return(newActionBinding(ActionNone), true)
	keyBindings.On("Binding", viewHierarchy, "abc").Return(newActionBinding(ActionFirstLine), false)

	inputBuffer.Append("a")
	action, keyString := inputBuffer.Process(viewHierarchy)
	checkProcessResult(Action{ActionType: ActionNone}, "", action, keyString, t)

	inputBuffer.Append("b")
	action, keyString = inputBuffer.Process(viewHierarchy)
	checkProcessResult(Action{ActionType: ActionNone}, "", action, keyString, t)

	inputBuffer.Append("c")
	action, keyString = inputBuffer.Process(viewHierarchy)
	checkProcessResult(Action{ActionType: ActionFirstLine}, "abc", action, keyString, t)
}

func TestKeystringBindingIsExpandedInPlace(t *testing.T) {
	keyBindings := &MockKeyBindings{}
	inputBuffer := NewInputBuffer(keyBindings)

	viewHierarchy := ViewHierarchy([]ViewID{ViewMain, ViewHistory, ViewCommit})

	keyBindings.On("Binding", viewHierarchy, "a").Return(newKeystringBinding("bb"), false)
	keyBindings.On("Binding", viewHierarchy, "b").Return(newActionBinding(ActionNone), true)
	keyBindings.On("Binding", viewHierarchy, "bb").Return(newActionBinding(ActionNone), true)
	keyBindings.On("Binding", viewHierarchy, "bbb").Return(newActionBinding(ActionFirstLine), false)

	inputBuffer.Append("ab")
	action, keyString := inputBuffer.Process(viewHierarchy)
	checkProcessResult(Action{ActionType: ActionFirstLine}, "bbb", action, keyString, t)
}

func TestPotentialPrefixMatchIsReturnedAsSeparateKeysWhenFullInputDoesNotMatchBinding(t *testing.T) {
	keyBindings := &MockKeyBindings{}
	inputBuffer := NewInputBuffer(keyBindings)

	viewHierarchy := ViewHierarchy([]ViewID{ViewMain, ViewHistory, ViewCommit})

	keyBindings.On("Binding", viewHierarchy, "a").Return(newActionBinding(ActionNone), true)
	keyBindings.On("Binding", viewHierarchy, "b").Return(newActionBinding(ActionNone), false)
	keyBindings.On("Binding", viewHierarchy, "ab").Return(newActionBinding(ActionNone), false)

	inputBuffer.Append("ab")

	action, keyString := inputBuffer.Process(viewHierarchy)
	checkProcessResult(Action{ActionType: ActionNone}, "a", action, keyString, t)

	action, keyString = inputBuffer.Process(viewHierarchy)
	checkProcessResult(Action{ActionType: ActionNone}, "b", action, keyString, t)
}

func TestDiscardToOnlyDiscardsInputUntilProvidedKey(t *testing.T) {
	keyBindings := &MockKeyBindings{}
	inputBuffer := NewInputBuffer(keyBindings)

	inputBuffer.Append(`<grv-filter-prompt>committername="John Smith"<Enter><Tab>`)

	discarded, enterFound := inputBuffer.DiscardTo("<Enter>")
	hasInput := inputBuffer.hasInput()

	expectedDiscarded := `<grv-filter-prompt>committername="John Smith"<Enter>`

	if discarded != expectedDiscarded {
		t.Errorf(`Discarded text "%v" does not match expected "%v"`, discarded, expectedDiscarded)
	}

	if !enterFound {
		t.Errorf("Expected enterFound to be true")
	}

	if !hasInput {
		t.Errorf("InputBuffer is empty")
	}
}

func TestDiscardToDiscardsAllInputWhenTargetKeyIsNotPresent(t *testing.T) {
	keyBindings := &MockKeyBindings{}
	inputBuffer := NewInputBuffer(keyBindings)

	inputBuffer.Append(`<grv-filter-prompt>committername="John Smith"`)

	discarded, enterFound := inputBuffer.DiscardTo("<Enter>")
	hasInput := inputBuffer.hasInput()

	expectedDiscarded := `<grv-filter-prompt>committername="John Smith"`

	if discarded != expectedDiscarded {
		t.Errorf(`Discarded text "%v" does not match expected "%v"`, discarded, expectedDiscarded)
	}

	if enterFound {
		t.Errorf("Expected enterFound to be false")
	}

	if hasInput {
		t.Errorf("InputBuffer is not empty")
	}
}

func TestDiscardToReturnsEmptyStringWhenInputBufferIsEmpty(t *testing.T) {
	keyBindings := &MockKeyBindings{}
	inputBuffer := NewInputBuffer(keyBindings)

	discarded, enterFound := inputBuffer.DiscardTo("<Enter>")

	if discarded != "" {
		t.Errorf(`Discarded text "%v" does not match expected "%v"`, discarded, "")
	}

	if enterFound {
		t.Errorf("Expected enterFound to be false")
	}
}
