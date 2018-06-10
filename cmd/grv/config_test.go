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
