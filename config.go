package main

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"os"
	"reflect"
	"strconv"
)

const (
	DEFAULT_CONFIG_DIR  = "/.config"
	GRV_CONFIG_FILE     = "/grv/grvrc"
	MIN_TAB_WIDTH_VALUE = 1
)

type ConfigVariable string

const (
	CV_TAB_WIDTH ConfigVariable = "tabWidth"
)

type Config interface {
	GetBool(ConfigVariable) bool
	GetString(ConfigVariable) string
	GetInt(ConfigVariable) int
	GetFloat(ConfigVariable) float64
}

type ConfigVariableValidator func(value string) (processedValue interface{}, err error)

type ConfigurationVariable struct {
	value     interface{}
	validator ConfigVariableValidator
}

type Configuration struct {
	variables map[ConfigVariable]*ConfigurationVariable
}

func NewConfiguration() *Configuration {
	config := &Configuration{
		variables: map[ConfigVariable]*ConfigurationVariable{
			CV_TAB_WIDTH: &ConfigurationVariable{
				value:     8,
				validator: tabwidthValidator,
			},
		},
	}

	return config
}

func (config *Configuration) Initialise() []error {
	configHome, configHomeSet := os.LookupEnv("XDG_CONFIG_HOME")

	if !configHomeSet {
		log.Debug("XDG_CONFIG_HOME not set")
		home, homeSet := os.LookupEnv("HOME")

		if !homeSet {
			log.Info("Unable to determine config directory")
			return nil
		}

		log.Debugf("HOME directory: %v", home)
		configHome = home + DEFAULT_CONFIG_DIR
	} else {
		log.Debugf("XDG_CONFIG_HOME: %v", configHome)
	}

	grvConfig := configHome + GRV_CONFIG_FILE

	if _, err := os.Stat(grvConfig); os.IsNotExist(err) {
		log.Infof("No config file found at: %v", grvConfig)
		return nil
	}

	errors := config.LoadFile(grvConfig)

	if len(errors) > 0 {
		log.Infof("Encountered %v error(s) when loading config file", len(errors))
	}

	return errors
}

func (config *Configuration) LoadFile(filePath string) []error {
	file, err := os.Open(filePath)
	if err != nil {
		log.Errorf("Unable to open GRV config file %v for reading: %v", filePath, err.Error())
		return []error{err}
	}

	log.Infof("Loading config file %v", filePath)

	return config.processCommands(NewParser(file, filePath))
}

func (config *Configuration) processCommands(parser *Parser) []error {
	var configErrors []error

OuterLoop:
	for {
		command, eof, err := parser.Parse()

		switch {
		case err != nil:
			configErrors = append(configErrors, err)
		case eof:
			break OuterLoop
		case command != nil:
			if err = config.processCommand(command, parser.InputSource()); err != nil {
				configErrors = append(configErrors, err)
			}
		default:
			log.Error("Invalid parse state - no error, eof or command returned")
		}
	}

	return configErrors
}

func (config *Configuration) processCommand(command Command, inputSource string) (err error) {
	switch command := command.(type) {
	case *SetCommand:
		err = config.processSetCommand(command, inputSource)
	case *ThemeCommand:
		break
	default:
		log.Errorf("Unknown command type %T", command)
	}

	return
}

func (config *Configuration) processSetCommand(setCommand *SetCommand, inputSource string) error {
	configVariable, ok := config.variables[ConfigVariable(setCommand.variable.value)]
	if !ok {
		return generateConfigError(inputSource, setCommand.variable, "Invalid variable %v", setCommand.variable.value)
	}

	var value interface{}

	if configVariable.validator != nil {
		var err error
		if value, err = configVariable.validator(setCommand.value.value); err != nil {
			return generateConfigError(inputSource, setCommand.value, "%v", err.Error())
		}
	} else {
		value = setCommand.value.value
	}

	expectedType := reflect.TypeOf(configVariable.value)
	actualType := reflect.TypeOf(value)

	if actualType != expectedType {
		return generateConfigError(inputSource, setCommand.value, "Expected type %v but found type %v",
			expectedType, actualType)
	}

	log.Infof("Setting %v = %v", setCommand.variable.value, value)
	configVariable.value = value

	return nil
}

func (config *Configuration) getValue(configVariable ConfigVariable) interface{} {
	if variable, ok := config.variables[configVariable]; ok {
		return variable.value
	} else {
		panic(fmt.Sprintf("No ConfigVariable exists exists for ID %v", configVariable))
	}
}

func (config *Configuration) GetBool(configVariable ConfigVariable) bool {
	switch value := config.getValue(configVariable).(type) {
	case bool:
		return value
	}

	panic(fmt.Sprintf("ConfigVariable with ID %v does not have a boolean value", configVariable))
}

func (config *Configuration) GetString(configVariable ConfigVariable) string {
	switch value := config.getValue(configVariable).(type) {
	case string:
		return value
	}

	panic(fmt.Sprintf("ConfigVariable with ID %v does not have a string value", configVariable))
}

func (config *Configuration) GetInt(configVariable ConfigVariable) int {
	switch value := config.getValue(configVariable).(type) {
	case int:
		return value
	}

	panic(fmt.Sprintf("ConfigVariable with ID %v does not have an integer value", configVariable))
}

func (config *Configuration) GetFloat(configVariable ConfigVariable) float64 {
	switch value := config.getValue(configVariable).(type) {
	case float64:
		return value
	}

	panic(fmt.Sprintf("ConfigVariable with ID %v does not have a floating point value", configVariable))
}

func tabwidthValidator(value string) (processedValue interface{}, err error) {
	var tabWidth int

	if tabWidth, err = strconv.Atoi(value); err != nil {
		err = fmt.Errorf("%v must be an integer value greater than %v", CV_TAB_WIDTH, MIN_TAB_WIDTH_VALUE-1)
	} else if tabWidth < MIN_TAB_WIDTH_VALUE {
		err = fmt.Errorf("%v must be greater than %v", CV_TAB_WIDTH, MIN_TAB_WIDTH_VALUE-1)
	} else {
		processedValue = tabWidth
	}

	return
}
