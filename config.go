package main

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"os"
	"reflect"
	"strconv"
	"strings"
)

const (
	CF_DEFAULT_CONFIG_HOME_DIR = "/.config"
	CF_GRV_CONFIG_DIR          = "/grv"
	CF_GRVRC_FILE              = "/grvrc"
	CV_TAB_WIDTH_MIN_VALUE     = 1
	CV_THEME_DEFALT_VALUE      = "default"

	CV_ALL_VIEW        = "All"
	CV_MAIN_VIEW       = "MainView"
	CV_HISTORY_VIEW    = "HistoryView"
	CV_STATUS_VIEW     = "StatusView"
	CV_REF_VIEW        = "RefView"
	CV_COMMIT_VIEW     = "CommitView"
	CV_DIFF_VIEW       = "DiffView"
	CV_STATUS_BAR_VIEW = "StatusBarView"
	CV_HELP_BAR_VIEW   = "HelpBarView"
	CV_ERROR_VIEW      = "ErrorView"
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

var viewIdNames = map[string]ViewId{
	CV_ALL_VIEW:        VIEW_ALL,
	CV_MAIN_VIEW:       VIEW_MAIN,
	CV_HISTORY_VIEW:    VIEW_HISTORY,
	CV_STATUS_VIEW:     VIEW_STATUS,
	CV_REF_VIEW:        VIEW_REF,
	CV_COMMIT_VIEW:     VIEW_COMMIT,
	CV_DIFF_VIEW:       VIEW_DIFF,
	CV_STATUS_BAR_VIEW: VIEW_STATUS_BAR,
	CV_HELP_BAR_VIEW:   VIEW_HELP_BAR,
	CV_ERROR_VIEW:      VIEW_ERROR,
}

var themeComponents = map[string]ThemeComponentId{
	CV_ALL_VIEW + ".SearchMatch": CMP_ALLVIEW_SEARCH_MATCH,

	CV_REF_VIEW + ".Title":          CMP_REFVIEW_TITLE,
	CV_REF_VIEW + ".Footer":         CMP_REFVIEW_FOOTER,
	CV_REF_VIEW + ".BranchesHeader": CMP_REFVIEW_BRANCHES_HEADER,
	CV_REF_VIEW + ".Branch":         CMP_REFVIEW_BRANCH,
	CV_REF_VIEW + ".TagsHeader":     CMP_REFVIEW_TAGS_HEADER,
	CV_REF_VIEW + ".Tag":            CMP_REFVIEW_TAG,

	CV_COMMIT_VIEW + ".Title":   CMP_COMMITVIEW_TITLE,
	CV_COMMIT_VIEW + ".Footer":  CMP_COMMITVIEW_FOOTER,
	CV_COMMIT_VIEW + ".Date":    CMP_COMMITVIEW_DATE,
	CV_COMMIT_VIEW + ".Author":  CMP_COMMITVIEW_AUTHOR,
	CV_COMMIT_VIEW + ".Summary": CMP_COMMITVIEW_SUMMARY,

	CV_DIFF_VIEW + ".Normal":                CMP_DIFFVIEW_DIFFLINE_NORMAL,
	CV_DIFF_VIEW + ".CommitAuthor":          CMP_DIFFVIEW_DIFFLINE_DIFF_COMMIT_AUTHOR,
	CV_DIFF_VIEW + ".CommitAuthorDate":      CMP_DIFFVIEW_DIFFLINE_DIFF_COMMIT_AUTHOR_DATE,
	CV_DIFF_VIEW + ".CommitCommitter":       CMP_DIFFVIEW_DIFFLINE_DIFF_COMMIT_COMMITTER,
	CV_DIFF_VIEW + ".CommitCommitterDate":   CMP_DIFFVIEW_DIFFLINE_DIFF_COMMIT_COMMITTER_DATE,
	CV_DIFF_VIEW + ".CommitSummary":         CMP_DIFFVIEW_DIFFLINE_DIFF_COMMIT_SUMMARY,
	CV_DIFF_VIEW + ".StatsFile":             CMP_DIFFVIEW_DIFFLINE_DIFF_STATS_FILE,
	CV_DIFF_VIEW + ".GitDiffHeader":         CMP_DIFFVIEW_DIFFLINE_GIT_DIFF_HEADER,
	CV_DIFF_VIEW + ".GitDiffExtendedHeader": CMP_DIFFVIEW_DIFFLINE_GIT_DIFF_EXTENDED_HEADER,
	CV_DIFF_VIEW + ".UnifiedDiffHeader":     CMP_DIFFVIEW_DIFFLINE_UNIFIED_DIFF_HEADER,
	CV_DIFF_VIEW + ".HunkStart":             CMP_DIFFVIEW_DIFFLINE_HUNK_START,
	CV_DIFF_VIEW + ".HunkHeader":            CMP_DIFFVIEW_DIFFLINE_HUNK_HEADER,
	CV_DIFF_VIEW + ".AddedLine":             CMP_DIFFVIEW_DIFFLINE_LINE_ADDED,
	CV_DIFF_VIEW + ".RemovedLine":           CMP_DIFFVIEW_DIFFLINE_LINE_REMOVED,

	CV_STATUS_BAR_VIEW + ".Normal": CMP_STATUSBARVIEW_NORMAL,

	CV_HELP_BAR_VIEW + ".Special": CMP_HELPBARVIEW_SPECIAL,
	CV_HELP_BAR_VIEW + ".Normal":  CMP_HELPBARVIEW_NORMAL,

	CV_ERROR_VIEW + ".Title":  CMP_ERROR_VIEW_TITLE,
	CV_ERROR_VIEW + ".Footer": CMP_ERROR_VIEW_FOOTER,
	CV_ERROR_VIEW + ".Errors": CMP_ERROR_VIEW_ERRORS,
}

type Config interface {
	GetBool(ConfigVariable) bool
	GetString(ConfigVariable) string
	GetInt(ConfigVariable) int
	GetFloat(ConfigVariable) float64
	GetTheme() Theme
	AddOnChangeListener(ConfigVariable, ConfigVariableOnChangeListener)
	ConfigDir() string
}

type ConfigSetter interface {
	Config
	Evaluate(config string) []error
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
	variables    map[ConfigVariable]*ConfigurationVariable
	themes       map[string]MutableTheme
	keyBindings  KeyBindings
	grvConfigDir string
	channels     *Channels
}

func NewConfiguration(keyBindings KeyBindings, channels *Channels) *Configuration {
	config := &Configuration{
		keyBindings: keyBindings,
		themes: map[string]MutableTheme{
			CV_THEME_DEFALT_VALUE: NewDefaultTheme(),
		},
		channels: channels,
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
	configHomeDir, configHomeDirSet := os.LookupEnv("XDG_CONFIG_HOME")

	if !configHomeDirSet {
		log.Debug("XDG_CONFIG_HOME not set")
		home, homeSet := os.LookupEnv("HOME")

		if !homeSet {
			log.Info("Unable to determine config directory")
			return nil
		}

		log.Debugf("HOME directory: %v", home)
		configHomeDir = home + CF_DEFAULT_CONFIG_HOME_DIR
	} else {
		log.Debugf("XDG_CONFIG_HOME: %v", configHomeDir)
	}

	grvConfigDir := configHomeDir + CF_GRV_CONFIG_DIR

	if err := os.MkdirAll(grvConfigDir, 0755); err != nil {
		log.Errorf("Unable to create config home directory %v: %v", grvConfigDir, err)
		return nil
	}

	config.grvConfigDir = grvConfigDir

	grvConfig := grvConfigDir + CF_GRVRC_FILE

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

func (config *Configuration) ConfigDir() string {
	return config.grvConfigDir
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

func (config *Configuration) Evaluate(configString string) (errs []error) {
	if configString == "" {
		return
	}

	reader := strings.NewReader(configString)
	parser := NewParser(reader, "")

	return config.processCommands(parser)
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
	case *MapCommand:
		err = config.processMapCommand(command, inputSource)
	case *QuitCommand:
		err = config.processQuitCommand(command)
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

func (config *Configuration) processMapCommand(mapCommand *MapCommand, inputSource string) (err error) {
	viewId, ok := viewIdNames[mapCommand.view.value]
	if !ok {
		return generateConfigError(inputSource, mapCommand.view, "Invalid view: %v", mapCommand.view.value)
	}

	if mapCommand.from.value == "" {
		return generateConfigError(inputSource, mapCommand.from, "from keystring cannot be empty")
	} else if mapCommand.to.value == "" {
		return generateConfigError(inputSource, mapCommand.to, "to keystring cannot be empty")
	}

	config.keyBindings.SetKeystringBinding(viewId, mapCommand.from.value, mapCommand.to.value)

	log.Infof("Mapped \"%v\" to \"%v\" for view %v", mapCommand.from.value, mapCommand.to.value, mapCommand.view.value)

	return
}

func (config *Configuration) processQuitCommand(quitCommand *QuitCommand) (err error) {
	log.Info("Processed quit command")
	config.channels.DoAction(Action{ActionType: ACTION_EXIT})
	return
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
