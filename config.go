package main

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"os"
	"reflect"
	"strconv"
)

const (
	CF_DEFAULT_CONFIG_DIR  = "/.config"
	CF_GRV_CONFIG_FILE     = "/grv/grvrc"
	CV_TAB_WIDTH_MIN_VALUE = 1
	CV_THEME_DEFALT_VALUE  = "default"
)

type ConfigVariable string

const (
	CV_TAB_WIDTH ConfigVariable = "tabWidth"
	CV_THEME     ConfigVariable = "theme"
)

var themeColors = map[string]ThemeColor{
	"NONE":    COLOR_NONE,
	"BLACK":   COLOR_BLACK,
	"RED":     COLOR_RED,
	"GREEN":   COLOR_GREEN,
	"YELLOW":  COLOR_YELLOW,
	"BLUE":    COLOR_BLUE,
	"MAGENTA": COLOR_MAGENTA,
	"CYAN":    COLOR_CYAN,
	"WHITE":   COLOR_WHITE,
}

var themeComponents = map[string]ThemeComponentId{
	"CommitView.Title":   CMP_COMMITVIEW_TITLE,
	"CommitView.Footer":  CMP_COMMITVIEW_FOOTER,
	"CommitView.Date":    CMP_COMMITVIEW_DATE,
	"CommitView.Author":  CMP_COMMITVIEW_AUTHOR,
	"CommitView.Summary": CMP_COMMITVIEW_SUMMARY,
}

type Config interface {
	GetBool(ConfigVariable) bool
	GetString(ConfigVariable) string
	GetInt(ConfigVariable) int
	GetFloat(ConfigVariable) float64
	GetTheme() Theme
	AddOnChangeListener(ConfigVariable, ConfigVariableOnChangeListener)
}

type ConfigVariableValidator interface {
	validate(value string) (processedValue interface{}, err error)
}

type ConfigVariableOnChangeListener interface {
	onConfigVariableChange(ConfigVariable)
}

type ConfigurationVariable struct {
	value             interface{}
	validator         ConfigVariableValidator
	onChangeListeners []ConfigVariableOnChangeListener
}

type Configuration struct {
	variables map[ConfigVariable]*ConfigurationVariable
	themes    map[string]MutableTheme
}

func NewConfiguration() *Configuration {
	config := &Configuration{
		themes: map[string]MutableTheme{
			CV_THEME_DEFALT_VALUE: NewDefaultTheme(),
		},
	}

	config.variables = map[ConfigVariable]*ConfigurationVariable{
		CV_TAB_WIDTH: &ConfigurationVariable{
			value:     8,
			validator: TabWidithValidator{},
		},
		CV_THEME: &ConfigurationVariable{
			value: CV_THEME_DEFALT_VALUE,
			validator: ThemeValidator{
				config: config,
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
		configHome = home + CF_DEFAULT_CONFIG_DIR
	} else {
		log.Debugf("XDG_CONFIG_HOME: %v", configHome)
	}

	grvConfig := configHome + CF_GRV_CONFIG_FILE

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
		err = config.processThemeCommand(command, inputSource)
	default:
		log.Errorf("Unknown command type %T", command)
	}

	return
}

func (config *Configuration) processSetCommand(setCommand *SetCommand, inputSource string) error {
	configVariable := ConfigVariable(setCommand.variable.value)
	variable, ok := config.variables[configVariable]
	if !ok {
		return generateConfigError(inputSource, setCommand.variable, "Invalid variable %v", setCommand.variable.value)
	}

	var value interface{}

	if variable.validator != nil {
		var err error
		if value, err = variable.validator.validate(setCommand.value.value); err != nil {
			return generateConfigError(inputSource, setCommand.value, "%v", err.Error())
		}
	} else {
		value = setCommand.value.value
	}

	expectedType := reflect.TypeOf(variable.value)
	actualType := reflect.TypeOf(value)

	if actualType != expectedType {
		return generateConfigError(inputSource, setCommand.value, "Expected type %v but found type %v",
			expectedType, actualType)
	}

	log.Infof("Setting %v = %v", configVariable, value)
	variable.value = value

	if len(variable.onChangeListeners) > 0 {
		log.Debugf("Firing on change listeners for config variable %v", configVariable)
		for _, listener := range variable.onChangeListeners {
			listener.onConfigVariableChange(configVariable)
		}
	}

	return nil
}

func (config *Configuration) processThemeCommand(themeCommand *ThemeCommand, inputSource string) (err error) {
	themeComponentId, componentIdExists := themeComponents[themeCommand.component.value]

	if !componentIdExists {
		err = generateConfigError(inputSource, themeCommand.component, "Invalid theme component: %v", themeCommand.component.value)
		return
	}

	var bgThemeColor, fgThemeColor ThemeColor

	if bgThemeColor, err = getThemeColor(themeCommand.bgcolor, inputSource); err != nil {
		return
	} else if fgThemeColor, err = getThemeColor(themeCommand.fgcolor, inputSource); err != nil {
		return
	}

	theme, themeExists := config.themes[themeCommand.name.value]

	if !themeExists {
		theme = NewTheme()
		config.themes[themeCommand.name.value] = theme
	}

	log.Infof("Setting bgcolor = %v and fgcolor = %v for component %v in theme %v",
		themeCommand.bgcolor.value, themeCommand.fgcolor.value,
		themeCommand.component.value, themeCommand.name.value)

	themeComponent := theme.CreateOrGetComponent(themeComponentId)
	themeComponent.bgcolor = bgThemeColor
	themeComponent.fgcolor = fgThemeColor

	return
}

func getThemeColor(color *Token, inputSource string) (ThemeColor, error) {
	themeColor, ok := themeColors[color.value]

	if !ok {
		return COLOR_NONE, generateConfigError(inputSource, color, "Invalid color: %v", color.value)
	}

	return themeColor, nil
}

func (config *Configuration) getVariable(configVariable ConfigVariable) *ConfigurationVariable {
	if variable, ok := config.variables[configVariable]; ok {
		return variable
	}

	panic(fmt.Sprintf("No ConfigVariable exists exists for ID %v", configVariable))
}

func (config *Configuration) AddOnChangeListener(configVariable ConfigVariable, listener ConfigVariableOnChangeListener) {
	variable := config.getVariable(configVariable)
	variable.onChangeListeners = append(variable.onChangeListeners, listener)
}

func (config *Configuration) GetBool(configVariable ConfigVariable) bool {
	switch value := config.getVariable(configVariable).value.(type) {
	case bool:
		return value
	}

	panic(fmt.Sprintf("ConfigVariable with ID %v does not have a boolean value", configVariable))
}

func (config *Configuration) GetString(configVariable ConfigVariable) string {
	switch value := config.getVariable(configVariable).value.(type) {
	case string:
		return value
	}

	panic(fmt.Sprintf("ConfigVariable with ID %v does not have a string value", configVariable))
}

func (config *Configuration) GetInt(configVariable ConfigVariable) int {
	switch value := config.getVariable(configVariable).value.(type) {
	case int:
		return value
	}

	panic(fmt.Sprintf("ConfigVariable with ID %v does not have an integer value", configVariable))
}

func (config *Configuration) GetFloat(configVariable ConfigVariable) float64 {
	switch value := config.getVariable(configVariable).value.(type) {
	case float64:
		return value
	}

	panic(fmt.Sprintf("ConfigVariable with ID %v does not have a floating point value", configVariable))
}

func (config *Configuration) GetTheme() Theme {
	themeName := config.GetString(CV_THEME)
	theme, ok := config.themes[themeName]

	if !ok {
		panic(fmt.Sprintf("No theme exists with name %v", themeName))
	}

	return theme
}

type TabWidithValidator struct{}

func (tabwidthValidator TabWidithValidator) validate(value string) (processedValue interface{}, err error) {
	var tabWidth int

	if tabWidth, err = strconv.Atoi(value); err != nil {
		err = fmt.Errorf("%v must be an integer value greater than %v", CV_TAB_WIDTH, CV_TAB_WIDTH_MIN_VALUE-1)
	} else if tabWidth < CV_TAB_WIDTH_MIN_VALUE {
		err = fmt.Errorf("%v must be greater than %v", CV_TAB_WIDTH, CV_TAB_WIDTH_MIN_VALUE-1)
	} else {
		processedValue = tabWidth
	}

	return
}

type ThemeValidator struct {
	config *Configuration
}

func (themeValidator ThemeValidator) validate(value string) (processedValue interface{}, err error) {
	if _, ok := themeValidator.config.themes[value]; !ok {
		err = fmt.Errorf("No theme exists with name %v", value)
	} else {
		processedValue = value
	}

	return
}
