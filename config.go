package main

import (
	"fmt"
	"strconv"
)

type ConfigVariable int

const (
	CV_TAB_WIDTH ConfigVariable = iota
)

type Config interface {
	GetBool(ConfigVariable) bool
	GetString(ConfigVariable) string
	GetInt(ConfigVariable) int
	GetFloat(ConfigVariable) float64
}

type ConfigVariableValidator func(value string) (processedValue interface{}, err error)

type ConfigurationVariable struct {
	name      string
	value     interface{}
	validator ConfigVariableValidator
}

type Configuration struct {
	variables map[ConfigVariable]*ConfigurationVariable
}

func NewConfiguration() *Configuration {
	return &Configuration{
		variables: map[ConfigVariable]*ConfigurationVariable{
			CV_TAB_WIDTH: &ConfigurationVariable{
				name:      "tabWidth",
				value:     8,
				validator: tabwidthValidator,
			},
		},
	}
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
	return strconv.Atoi(value)
}
