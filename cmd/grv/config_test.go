package main

import (
	"testing"

	"github.com/stretchr/testify/mock"
)

type MockInputConsumer struct {
	mock.Mock
}

func (inputConsumer *MockInputConsumer) ProcessInput(input string) {
	inputConsumer.Called(input)
}

func TestConfigVariablesHaveRequiredFieldsSet(t *testing.T) {
	config := NewConfiguration(&MockKeyBindings{}, &MockChannels{}, &MockGRVVariableSetter{}, &MockInputConsumer{})

	for configVariableName, configVariable := range config.configVariables {
		if configVariable.defaultValue == nil {
			t.Errorf("Config variable \"%v\" has no default value set", configVariableName)
		}
		if configVariable.description == "" {
			t.Errorf("Config variable \"%v\" has no description", configVariableName)
		}
	}
}

func TestThemeComponentMapContainsEntriesForAllThemeComponents(t *testing.T) {
	themeComponentNames := map[ThemeComponentID]string{}

	for themeComponentName, themeComponentID := range themeComponents {
		themeComponentNames[themeComponentID] = themeComponentName
	}

	for themeComponentID := ThemeComponentID(1); themeComponentID < CmpCount; themeComponentID++ {
		if _, exists := themeComponentNames[themeComponentID]; !exists {
			t.Errorf("No entry in themeComponents map for ThemeComponenetID %v", themeComponentID)
		}
	}
}

func TestViewNamesContainsEntriesForAllViews(t *testing.T) {
	for viewID := ViewID(0); viewID < ViewCount; viewID++ {
		if _, exists := viewNames[viewID]; !exists {
			t.Errorf("No entry in viewNames map for ViewID %v", viewID)
		}
	}
}

func TestCommandBodyArgumentsAreExpanded(t *testing.T) {
	config := NewConfiguration(&MockKeyBindings{}, &MockChannels{}, &MockGRVVariableSetter{}, &MockInputConsumer{})

	args := []string{"do", "re", "mi"}

	commandBody := `
	123
	abc
	$1
	$1 "$2"
	${1}2
	"12${3}45"
	$@
	"${@}"
	$$1
	$$$1
	$$$$1
	$$$$$1
	`
	expectedProcessedCommandBody := `
	123
	abc
	do
	do "re"
	do2
	"12mi45"
	do re mi
	"do re mi"
	$1
	$do
	$$1
	$$do
	`

	actualProcessedCommandBody := config.processConfigCommandBody(commandBody, args)

	if expectedProcessedCommandBody != actualProcessedCommandBody {
		t.Errorf("Command body did not match expected value. Expected: %v, Actual: %v", expectedProcessedCommandBody, actualProcessedCommandBody)
	}
}
