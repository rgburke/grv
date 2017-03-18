package main

import (
	"strings"
	"testing"
)

const (
	CONFIG_FILE = "~/.config/grv/grvrc"
)

type SetCommandValues struct {
	variable string
	value    string
}

func (setCommandValues *SetCommandValues) Equal(command Command) bool {
	if command == nil {
		return false
	}

	other, ok := command.(*SetCommand)
	if !ok {
		return false
	}

	if other.variable == nil || other.value == nil {
		return false
	}

	return setCommandValues.variable == other.variable.value &&
		setCommandValues.value == other.value.value
}

type ThemeCommandValues struct {
	name      string
	component string
	bgcolor   string
	fgcolour  string
}

func (themeCommandValues *ThemeCommandValues) Equal(command Command) bool {
	if command == nil {
		return false
	}

	other, ok := command.(*ThemeCommand)
	if !ok {
		return false
	}

	if other.name == nil || other.component == nil || other.bgcolor == nil || other.fgcolor == nil {
		return false
	}

	return themeCommandValues.name == other.name.value &&
		themeCommandValues.component == other.component.value &&
		themeCommandValues.bgcolor == other.bgcolor.value &&
		themeCommandValues.fgcolour == other.fgcolor.value
}

func TestParseSingleCommand(t *testing.T) {
	var singleCommandTests = []struct {
		input           string
		expectedCommand Command
	}{
		{
			input: "set theme mytheme",
			expectedCommand: &SetCommandValues{
				variable: "theme",
				value:    "mytheme",
			},
		},
		{
			input: "theme --name mytheme --component CommitView.CommitDate --bgcolor NONE --fgcolor YELLOW\n",
			expectedCommand: &ThemeCommandValues{
				name:      "mytheme",
				component: "CommitView.CommitDate",
				bgcolor:   "NONE",
				fgcolour:  "YELLOW",
			},
		},
	}

	for _, singleCommandTest := range singleCommandTests {
		expectedCommand := singleCommandTest.expectedCommand
		parser := NewParser(strings.NewReader(singleCommandTest.input), CONFIG_FILE)
		command, _, err := parser.Parse()

		if err != nil {
			t.Errorf("Parse failed with error %v", err)
		} else if !expectedCommand.Equal(command) {
			t.Errorf("Command does not match expected value. Expected %v, Actual %v", singleCommandTest.expectedCommand, command)
		}
	}
}

func TestEOFIsSet(t *testing.T) {
	var eofTests = []struct {
		input string
		eof   bool
	}{
		{
			input: "set theme mytheme",
			eof:   true,
		},
		{
			input: "set theme mytheme\nset theme mytheme2",
			eof:   false,
		},
	}

	for _, eofTest := range eofTests {
		parser := NewParser(strings.NewReader(eofTest.input), CONFIG_FILE)
		_, _, err := parser.Parse()

		if err != nil {
			t.Errorf("Parse failed with error %v", err)
		}

		_, eof, err := parser.Parse()

		if err != nil {
			t.Errorf("Parse failed with error %v", err)
		} else if eof != eofTest.eof {
			t.Errorf("EOF value does not match expected value. Expected %v, Actual %v", eofTest.eof, eof)
		}
	}

}

func TestParseMultipleCommands(t *testing.T) {
	var multipleCommandsTests = []struct {
		input            string
		expectedCommands []Command
	}{
		{
			input: " set mouse\ttrue # Enable mouse\n# Set theme\n\tset theme \"my theme 2\" #Custom theme",
			expectedCommands: []Command{
				&SetCommandValues{
					variable: "mouse",
					value:    "true",
				},
				&SetCommandValues{
					variable: "theme",
					value:    "my theme 2",
				},
			},
		},
		{
			input: "theme\t--name mytheme\t--component RefView.LocalBranch \\\n\t--bgcolor BLUE\t--fgcolor YELLOW\nset mouse false\n",
			expectedCommands: []Command{
				&ThemeCommandValues{
					name:      "mytheme",
					component: "RefView.LocalBranch",
					bgcolor:   "BLUE",
					fgcolour:  "YELLOW",
				},
				&SetCommandValues{
					variable: "mouse",
					value:    "false",
				},
			},
		},
	}

	for _, multipleCommandsTest := range multipleCommandsTests {
		parser := NewParser(strings.NewReader(multipleCommandsTest.input), CONFIG_FILE)

		for _, expectedCommand := range multipleCommandsTest.expectedCommands {
			command, _, err := parser.Parse()

			if err != nil {
				t.Errorf("Parse failed with error %v", err)
			} else if !expectedCommand.Equal(command) {
				t.Errorf("Command does not match expected value. Expected %v, Actual %v", expectedCommand, command)
			}
		}
	}
}

func TestErrorsAreReceivedForInvalidTokenSequences(t *testing.T) {
	var errorTests = []struct {
		input                string
		expectedErrorMessage string
	}{
		{
			input:                "--name",
			expectedErrorMessage: CONFIG_FILE + ":1:1 Unexpected Option \"--name\"",
		},
		{
			input:                "\"theme",
			expectedErrorMessage: CONFIG_FILE + ":1:1 Syntax Error: Unterminated string",
		},
		{
			input:                "\n sety theme mytheme",
			expectedErrorMessage: CONFIG_FILE + ":2:2 Invalid command \"sety\"",
		},
		{
			input:                "set theme",
			expectedErrorMessage: CONFIG_FILE + ":1:9 Unexpected EOF",
		},
		{
			input:                "set theme --name mytheme",
			expectedErrorMessage: CONFIG_FILE + ":1:11 Expected Word but got Option: \"--name\"",
		},
		{
			input:                "set theme\nmytheme",
			expectedErrorMessage: CONFIG_FILE + ":1:10 Expected Word but got Terminator: \"\n\"",
		},
		{
			input:                "theme --name mytheme --component CommitView.CommitDate --bgcolour NONE --fgcolour YELLOW\n",
			expectedErrorMessage: CONFIG_FILE + ":1:56 Invalid option for theme command: \"--bgcolour\"",
		},
	}

	for _, errorTest := range errorTests {
		parser := NewParser(strings.NewReader(errorTest.input), CONFIG_FILE)
		_, _, err := parser.Parse()

		if err == nil {
			t.Errorf("Expected Parse to return error: %v", errorTest.expectedErrorMessage)
		} else if err.Error() != errorTest.expectedErrorMessage {
			t.Errorf("Error message does not match expected value. Expected %v, Actual %v", errorTest.expectedErrorMessage, err.Error())
		}
	}
}

func TestInvalidCommandsAreDiscardedAndParsingContinuesOnNextLine(t *testing.T) {
	var invalidCommandTests = []struct {
		input            string
		expectedCommands []Command
	}{
		{
			input: "set theme mytheme\nset theme --name theme\nset mouse false",
			expectedCommands: []Command{
				&SetCommandValues{
					variable: "theme",
					value:    "mytheme",
				},
				&SetCommandValues{
					variable: "mouse",
					value:    "false",
				},
			},
		},
		{
			input: "set theme\nmytheme\nset theme --name theme\nset mouse false",
			expectedCommands: []Command{
				&SetCommandValues{
					variable: "mouse",
					value:    "false",
				},
			},
		},
	}

	for _, invalidCommandTest := range invalidCommandTests {
		parser := NewParser(strings.NewReader(invalidCommandTest.input), CONFIG_FILE)

		var commands []Command

		for {
			command, eof, _ := parser.Parse()

			if command != nil {
				commands = append(commands, command)
			}
			if eof {
				break
			}
		}

		if len(commands) != len(invalidCommandTest.expectedCommands) {
			t.Errorf("Number of commands parsed differs from the number expected. Expected: %v, Actual: %v",
				invalidCommandTest.expectedCommands, commands)
		}

		for i := 0; i < len(invalidCommandTest.expectedCommands); i++ {
			expectedCommand := invalidCommandTest.expectedCommands[i]
			command := commands[i]

			if !expectedCommand.Equal(command) {
				t.Errorf("Command does not match expected value. Expected %v, Actual %v", expectedCommand, command)
			}
		}
	}
}
