package main

import (
	"testing"
)

func TestConfigVariablesHaveRequiredFieldsSet(t *testing.T) {
	config := NewConfiguration(&MockKeyBindings{}, &MockChannels{})

	for configVariableName, configVariable := range config.variables {
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
