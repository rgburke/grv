package main

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"
)

const (
	cfDefaultConfigHomeDir = "/.config"
	cfGrvConfigDir         = "/grv"
	cfGrvrcFile            = "/grvrc"
	cfTabWidthMinValue     = 1
	cfTabWidthDefaultValue = 8
	cfThemeDefaultValue    = "default"
	cfColdThemeName        = "cold"

	cfAllView       = "All"
	cfHistoryView   = "HistoryView"
	cfStatusView    = "StatusView"
	cfRefView       = "RefView"
	cfCommitView    = "CommitView"
	cfDiffView      = "DiffView"
	cfStatusBarView = "StatusBarView"
	cfHelpBarView   = "HelpBarView"
	cfErrorView     = "ErrorView"
)

// ConfigVariable stores a config variable name
type ConfigVariable string

const (
	// CfTabWidth stores the tab width variable name
	CfTabWidth ConfigVariable = "tabWidth"
	// CfTheme stores the theme variable name
	CfTheme ConfigVariable = "theme"
)

var themeColors = map[string]ThemeColor{
	"None":    ColorNone,
	"Black":   ColorBlack,
	"Red":     ColorRed,
	"Green":   ColorGreen,
	"Yellow":  ColorYellow,
	"Blue":    ColorBlue,
	"Magenta": ColorMagenta,
	"Cyan":    ColorCyan,
	"White":   ColorWhite,
}

var viewIDNames = map[string]ViewID{
	cfAllView:       ViewAll,
	cfHistoryView:   ViewHistory,
	cfStatusView:    ViewStatus,
	cfRefView:       ViewRef,
	cfCommitView:    ViewCommit,
	cfDiffView:      ViewDiff,
	cfStatusBarView: ViewStatusBar,
	cfHelpBarView:   ViewHelpBar,
	cfErrorView:     ViewError,
}

var themeComponents = map[string]ThemeComponentID{
	cfAllView + ".SearchMatch":             CmpAllviewSearchMatch,
	cfAllView + ".ActiveViewSelectedRow":   CmpAllviewActiveViewSelectedRow,
	cfAllView + ".InactiveViewSelectedRow": CmpAllviewInactiveViewSelectedRow,

	cfRefView + ".Title":                CmpRefviewTitle,
	cfRefView + ".Footer":               CmpRefviewFooter,
	cfRefView + ".LocalBranchesHeader":  CmpRefviewLocalBranchesHeader,
	cfRefView + ".RemoteBranchesHeader": CmpRefviewRemoteBranchesHeader,
	cfRefView + ".LocalBranch":          CmpRefviewLocalBranch,
	cfRefView + ".RemoteBranch":         CmpRefviewRemoteBranch,
	cfRefView + ".TagsHeader":           CmpRefviewTagsHeader,
	cfRefView + ".Tag":                  CmpRefviewTag,

	cfCommitView + ".Title":        CmpCommitviewTitle,
	cfCommitView + ".Footer":       CmpCommitviewFooter,
	cfCommitView + ".ShortOid":     CmpCommitviewShortOid,
	cfCommitView + ".Date":         CmpCommitviewDate,
	cfCommitView + ".Author":       CmpCommitviewAuthor,
	cfCommitView + ".Summary":      CmpCommitviewSummary,
	cfCommitView + ".Tag":          CmpCommitviewTag,
	cfCommitView + ".LocalBranch":  CmpCommitviewLocalBranch,
	cfCommitView + ".RemoteBranch": CmpCommitviewRemoteBranch,

	cfDiffView + ".Normal":                CmpDiffviewDifflineNormal,
	cfDiffView + ".CommitAuthor":          CmpDiffviewDifflineDiffCommitAuthor,
	cfDiffView + ".CommitAuthorDate":      CmpDiffviewDifflineDiffCommitAuthorDate,
	cfDiffView + ".CommitCommitter":       CmpDiffviewDifflineDiffCommitCommitter,
	cfDiffView + ".CommitCommitterDate":   CmpDiffviewDifflineDiffCommitCommitterDate,
	cfDiffView + ".CommitSummary":         CmpDiffviewDifflineDiffCommitSummary,
	cfDiffView + ".StatsFile":             CmpDiffviewDifflineDiffStatsFile,
	cfDiffView + ".GitDiffHeader":         CmpDiffviewDifflineGitDiffHeader,
	cfDiffView + ".GitDiffExtendedHeader": CmpDiffviewDifflineGitDiffExtendedHeader,
	cfDiffView + ".UnifiedDiffHeader":     CmpDiffviewDifflineUnifiedDiffHeader,
	cfDiffView + ".HunkStart":             CmpDiffviewDifflineHunkStart,
	cfDiffView + ".HunkHeader":            CmpDiffviewDifflineHunkHeader,
	cfDiffView + ".AddedLine":             CmpDiffviewDifflineLineAdded,
	cfDiffView + ".RemovedLine":           CmpDiffviewDifflineLineRemoved,

	cfStatusBarView + ".Normal": CmpStatusbarviewNormal,

	cfHelpBarView + ".Special": CmpHelpbarviewSpecial,
	cfHelpBarView + ".Normal":  CmpHelpbarviewNormal,

	cfErrorView + ".Title":  CmpErrorViewTitle,
	cfErrorView + ".Footer": CmpErrorViewFooter,
	cfErrorView + ".Errors": CmpErrorViewErrors,
}

// Config exposes a read only interface for configuration
type Config interface {
	GetBool(ConfigVariable) bool
	GetString(ConfigVariable) string
	GetInt(ConfigVariable) int
	GetFloat(ConfigVariable) float64
	GetTheme() Theme
	AddOnChangeListener(ConfigVariable, ConfigVariableOnChangeListener)
	ConfigDir() string
}

// ConfigSetter extends the config interface and exposes the ability to set config values
type ConfigSetter interface {
	Config
	Evaluate(config string) []error
}

// ConfigVariableValidator validates a new value for a config variable
type ConfigVariableValidator interface {
	validate(value string) (processedValue interface{}, err error)
}

// ConfigVariableOnChangeListener is notified when a config variable changes value
type ConfigVariableOnChangeListener interface {
	onConfigVariableChange(ConfigVariable)
}

// ConfigurationVariable represents a config variable
type ConfigurationVariable struct {
	value             interface{}
	validator         ConfigVariableValidator
	onChangeListeners []ConfigVariableOnChangeListener
}

// Configuration contains all configuration state
type Configuration struct {
	variables    map[ConfigVariable]*ConfigurationVariable
	themes       map[string]MutableTheme
	keyBindings  KeyBindings
	grvConfigDir string
	channels     *Channels
}

// NewConfiguration creates a Configuration instance with default values
func NewConfiguration(keyBindings KeyBindings, channels *Channels) *Configuration {
	config := &Configuration{
		keyBindings: keyBindings,
		themes: map[string]MutableTheme{
			cfThemeDefaultValue: NewDefaultTheme(),
			cfColdThemeName:     NewColdTheme(),
		},
		channels: channels,
	}

	config.variables = map[ConfigVariable]*ConfigurationVariable{
		CfTabWidth: {
			value:     cfTabWidthDefaultValue,
			validator: tabWidithValidator{},
		},
		CfTheme: {
			value: cfThemeDefaultValue,
			validator: themeValidator{
				config: config,
			},
		},
	}

	return config
}

// Initialise loads the grvrc config file (if it exists)
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
		configHomeDir = home + cfDefaultConfigHomeDir
	} else {
		log.Debugf("XDG_CONFIG_HOME: %v", configHomeDir)
	}

	grvConfigDir := configHomeDir + cfGrvConfigDir

	if err := os.MkdirAll(grvConfigDir, 0755); err != nil {
		log.Errorf("Unable to create config home directory %v: %v", grvConfigDir, err)
		return nil
	}

	config.grvConfigDir = grvConfigDir

	grvConfig := grvConfigDir + cfGrvrcFile

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

// ConfigDir returns the directory grv looks for config in
func (config *Configuration) ConfigDir() string {
	return config.grvConfigDir
}

// LoadFile loads the configuration file at by the provided file path
func (config *Configuration) LoadFile(filePath string) []error {
	file, err := os.Open(filePath)
	if err != nil {
		log.Errorf("Unable to open GRV config file %v for reading: %v", filePath, err.Error())
		return []error{err}
	}

	log.Infof("Loading config file %v", filePath)

	return config.processCommands(NewConfigParser(file, filePath))
}

// Evaluate processes configuration in string format
func (config *Configuration) Evaluate(configString string) (errs []error) {
	if configString == "" {
		return
	}

	reader := strings.NewReader(configString)
	parser := NewConfigParser(reader, "")

	return config.processCommands(parser)
}

func (config *Configuration) processCommands(parser *ConfigParser) []error {
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

func (config *Configuration) processCommand(command ConfigCommand, inputSource string) (err error) {
	switch command := command.(type) {
	case *SetCommand:
		err = config.processSetCommand(command, inputSource)
	case *ThemeCommand:
		err = config.processThemeCommand(command, inputSource)
	case *MapCommand:
		err = config.processMapCommand(command, inputSource)
	case *QuitCommand:
		err = config.processQuitCommand()
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
	themeComponentID, componentIDExists := themeComponents[themeCommand.component.value]

	if !componentIDExists {
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

	themeComponent := theme.CreateOrGetComponent(themeComponentID)
	themeComponent.bgcolor = bgThemeColor
	themeComponent.fgcolor = fgThemeColor

	return
}

func getThemeColor(color *ConfigToken, inputSource string) (ThemeColor, error) {
	themeColor, ok := themeColors[color.value]

	if !ok {
		return ColorNone, generateConfigError(inputSource, color, "Invalid color: %v", color.value)
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
	viewID, ok := viewIDNames[mapCommand.view.value]
	if !ok {
		return generateConfigError(inputSource, mapCommand.view, "Invalid view: %v", mapCommand.view.value)
	}

	if mapCommand.from.value == "" {
		return generateConfigError(inputSource, mapCommand.from, "from keystring cannot be empty")
	} else if mapCommand.to.value == "" {
		return generateConfigError(inputSource, mapCommand.to, "to keystring cannot be empty")
	}

	config.keyBindings.SetKeystringBinding(viewID, mapCommand.from.value, mapCommand.to.value)

	log.Infof("Mapped \"%v\" to \"%v\" for view %v", mapCommand.from.value, mapCommand.to.value, mapCommand.view.value)

	return
}

func (config *Configuration) processQuitCommand() (err error) {
	log.Info("Processed quit command")
	config.channels.DoAction(Action{ActionType: ActionExit})
	return
}

// AddOnChangeListener adds a listener to be notified when a configuration variable changes value
func (config *Configuration) AddOnChangeListener(configVariable ConfigVariable, listener ConfigVariableOnChangeListener) {
	variable := config.getVariable(configVariable)
	variable.onChangeListeners = append(variable.onChangeListeners, listener)
}

// GetBool returns the boolean value of the specified configuration variable
func (config *Configuration) GetBool(configVariable ConfigVariable) bool {
	switch value := config.getVariable(configVariable).value.(type) {
	case bool:
		return value
	}

	panic(fmt.Sprintf("ConfigVariable with ID %v does not have a boolean value", configVariable))
}

// GetString returns the string value of the specified configuration variable
func (config *Configuration) GetString(configVariable ConfigVariable) string {
	switch value := config.getVariable(configVariable).value.(type) {
	case string:
		return value
	}

	panic(fmt.Sprintf("ConfigVariable with ID %v does not have a string value", configVariable))
}

// GetInt returns the integer value of the specified configuration variable
func (config *Configuration) GetInt(configVariable ConfigVariable) int {
	switch value := config.getVariable(configVariable).value.(type) {
	case int:
		return value
	}

	panic(fmt.Sprintf("ConfigVariable with ID %v does not have an integer value", configVariable))
}

// GetFloat returns the floating point value of the specified configuration variable
func (config *Configuration) GetFloat(configVariable ConfigVariable) float64 {
	switch value := config.getVariable(configVariable).value.(type) {
	case float64:
		return value
	}

	panic(fmt.Sprintf("ConfigVariable with ID %v does not have a floating point value", configVariable))
}

// GetTheme returns the currently active theme
func (config *Configuration) GetTheme() Theme {
	themeName := config.GetString(CfTheme)
	theme, ok := config.themes[themeName]

	if !ok {
		panic(fmt.Sprintf("No theme exists with name %v", themeName))
	}

	return theme
}

type tabWidithValidator struct{}

func (tabwidthValidator tabWidithValidator) validate(value string) (processedValue interface{}, err error) {
	var tabWidth int

	if tabWidth, err = strconv.Atoi(value); err != nil {
		err = fmt.Errorf("%v must be an integer value greater than %v", CfTabWidth, cfTabWidthMinValue-1)
	} else if tabWidth < cfTabWidthMinValue {
		err = fmt.Errorf("%v must be greater than %v", CfTabWidth, cfTabWidthMinValue-1)
	} else {
		processedValue = tabWidth
	}

	return
}

type themeValidator struct {
	config *Configuration
}

func (themeValidator themeValidator) validate(value string) (processedValue interface{}, err error) {
	if _, ok := themeValidator.config.themes[value]; !ok {
		err = fmt.Errorf("No theme exists with name %v", value)
	} else {
		processedValue = value
	}

	return
}
