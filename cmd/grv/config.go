package main

import (
	"encoding/hex"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"
)

const (
	cfDefaultConfigHomeDir        = "/.config"
	cfGrvConfigDir                = "/grv"
	cfGrvrcFile                   = "/grvrc"
	cfTabWidthMinValue            = 1
	cfTabWidthDefaultValue        = 8
	cfClassicThemeName            = "classic"
	cfColdThemeName               = "cold"
	cfSolarizedThemeName          = "solarized"
	cfMouseDefaultValue           = false
	cfMouseScrollRowsDefaultValue = 3
	cfCommitGraphDefaultValue     = false
	cfConfirmCheckoutValue        = true

	cfAllView       = "All"
	cfMainView      = "MainView"
	cfHistoryView   = "HistoryView"
	cfStatusView    = "StatusView"
	cfGRVStatusView = "GRVStatusView"
	cfRefView       = "RefView"
	cfCommitView    = "CommitView"
	cfDiffView      = "DiffView"
	cfStatusBarView = "StatusBarView"
	cfHelpBarView   = "HelpBarView"
	cfErrorView     = "ErrorView"
	cfGitStatusView = "GitStatusView"
)

// ConfigVariable stores a config variable name
type ConfigVariable string

const (
	// CfTabWidth stores the tab width variable name
	CfTabWidth ConfigVariable = "tabwidth"
	// CfTheme stores the theme variable name
	CfTheme ConfigVariable = "theme"
	// CfMouse stores whether mouse support is enabled
	CfMouse ConfigVariable = "mouse"
	// CfMouseScrollRows stores the number of rows a view will scroll when a scroll mouse event is received
	CfMouseScrollRows ConfigVariable = "mouse-scroll-rows"
	// CfCommitGraph stores whether the commit graph is visible or not
	CfCommitGraph ConfigVariable = "commit-graph"
	// CfConfirmCheckout stores whether checkouts should be confirmed
	CfConfirmCheckout ConfigVariable = "confirm-checkout"
)

var systemColorValues = map[string]SystemColorValue{
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
	cfMainView:      ViewMain,
	cfHistoryView:   ViewHistory,
	cfStatusView:    ViewStatus,
	cfGRVStatusView: ViewGRVStatus,
	cfRefView:       ViewRef,
	cfCommitView:    ViewCommit,
	cfDiffView:      ViewDiff,
	cfStatusBarView: ViewStatusBar,
	cfHelpBarView:   ViewHelpBar,
	cfErrorView:     ViewError,
	cfGitStatusView: ViewGitStatus,
}

var themeComponents = map[string]ThemeComponentID{
	cfAllView + ".Default":                 CmpAllviewDefault,
	cfAllView + ".SearchMatch":             CmpAllviewSearchMatch,
	cfAllView + ".ActiveViewSelectedRow":   CmpAllviewActiveViewSelectedRow,
	cfAllView + ".InactiveViewSelectedRow": CmpAllviewInactiveViewSelectedRow,

	cfMainView + ".ActiveView": CmpMainviewActiveView,
	cfMainView + ".NormalView": CmpMainviewNormalView,

	cfRefView + ".Title":                CmpRefviewTitle,
	cfRefView + ".Footer":               CmpRefviewFooter,
	cfRefView + ".LocalBranchesHeader":  CmpRefviewLocalBranchesHeader,
	cfRefView + ".RemoteBranchesHeader": CmpRefviewRemoteBranchesHeader,
	cfRefView + ".LocalBranch":          CmpRefviewLocalBranch,
	cfRefView + ".Head":                 CmpRefviewHead,
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

	cfDiffView + ".Title":                 CmpDiffviewTitle,
	cfDiffView + ".Footer":                CmpDiffviewFooter,
	cfDiffView + ".Normal":                CmpDiffviewDifflineNormal,
	cfDiffView + ".CommitAuthor":          CmpDiffviewDifflineDiffCommitAuthor,
	cfDiffView + ".CommitAuthorDate":      CmpDiffviewDifflineDiffCommitAuthorDate,
	cfDiffView + ".CommitCommitter":       CmpDiffviewDifflineDiffCommitCommitter,
	cfDiffView + ".CommitCommitterDate":   CmpDiffviewDifflineDiffCommitCommitterDate,
	cfDiffView + ".CommitMessage":         CmpDiffviewDifflineDiffCommitMessage,
	cfDiffView + ".StatsFile":             CmpDiffviewDifflineDiffStatsFile,
	cfDiffView + ".GitDiffHeader":         CmpDiffviewDifflineGitDiffHeader,
	cfDiffView + ".GitDiffExtendedHeader": CmpDiffviewDifflineGitDiffExtendedHeader,
	cfDiffView + ".UnifiedDiffHeader":     CmpDiffviewDifflineUnifiedDiffHeader,
	cfDiffView + ".HunkStart":             CmpDiffviewDifflineHunkStart,
	cfDiffView + ".HunkHeader":            CmpDiffviewDifflineHunkHeader,
	cfDiffView + ".AddedLine":             CmpDiffviewDifflineLineAdded,
	cfDiffView + ".RemovedLine":           CmpDiffviewDifflineLineRemoved,

	cfGitStatusView + ".StagedTitle":     CmpGitStatusStagedTitle,
	cfGitStatusView + ".UnstagedTitle":   CmpGitStatusUnstagedTitle,
	cfGitStatusView + ".UntrackedTitle":  CmpGitStatusUntrackedTitle,
	cfGitStatusView + ".ConflictedTitle": CmpGitStatusConflictedTitle,
	cfGitStatusView + ".StagedFile":      CmpGitStatusStagedFile,
	cfGitStatusView + ".UnstagedFile":    CmpGitStatusUnstagedFile,
	cfGitStatusView + ".UntrackedFile":   CmpGitStatusUntrackedFile,
	cfGitStatusView + ".ConflictedFile":  CmpGitStatusConflictedFile,

	cfStatusBarView + ".Normal": CmpStatusbarviewNormal,

	cfHelpBarView + ".Special": CmpHelpbarviewSpecial,
	cfHelpBarView + ".Normal":  CmpHelpbarviewNormal,

	cfErrorView + ".Title":  CmpErrorViewTitle,
	cfErrorView + ".Footer": CmpErrorViewFooter,
	cfErrorView + ".Errors": CmpErrorViewErrors,
}

var colorNumberPattern = regexp.MustCompile(`[0-9]{1,3}`)
var hexColorPattern = regexp.MustCompile(`[a-fA-F0-9]{6}`)
var systemColorPattern = regexp.MustCompile(`[a-zA-Z]+`)

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
			cfClassicThemeName:   NewClassicTheme(),
			cfColdThemeName:      NewColdTheme(),
			cfSolarizedThemeName: NewSolarizedTheme(),
		},
		channels: channels,
	}

	config.variables = map[ConfigVariable]*ConfigurationVariable{
		CfTabWidth: {
			value:     cfTabWidthDefaultValue,
			validator: tabWidithValidator{},
		},
		CfTheme: {
			value: cfSolarizedThemeName,
			validator: themeValidator{
				config: config,
			},
		},
		CfMouse: {
			value: cfMouseDefaultValue,
			validator: booleanValueValidator{
				variableName: string(CfMouse),
			},
		},
		CfMouseScrollRows: {
			value:     cfMouseScrollRowsDefaultValue,
			validator: mouseScrollRowsValidator{},
		},
		CfCommitGraph: {
			value: cfCommitGraphDefaultValue,
			validator: booleanValueValidator{
				variableName: string(CfCommitGraph),
			},
		},
		CfConfirmCheckout: {
			value: cfConfirmCheckoutValue,
			validator: booleanValueValidator{
				variableName: string(CfConfirmCheckout),
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
	case *UnmapCommand:
		err = config.processUnmapCommand(command, inputSource)
	case *QuitCommand:
		err = config.processQuitCommand()
	case *NewTabCommand:
		err = config.processNewTabCommand(command, inputSource)
	case *RemoveTabCommand:
		err = config.processRemoveTabCommand()
	case *AddViewCommand:
		err = config.processAddViewCommand(command, inputSource)
	case *SplitViewCommand:
		err = config.processSplitViewCommand(command, inputSource)
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

	oldValue := variable.value

	log.Infof("Setting %v = %v", configVariable, value)
	variable.value = value
	config.channels.ReportStatus("Set %v = %v", configVariable, value)

	if oldValue != value {
		log.Infof("Value of config variable %v has changed from %v to %v", configVariable, oldValue, value)

		if len(variable.onChangeListeners) > 0 {
			log.Debugf("Firing on change listeners for config variable %v", configVariable)
			for _, listener := range variable.onChangeListeners {
				listener.onConfigVariableChange(configVariable)
			}
		} else {
			log.Debugf("Config variable %v has no change listeners registered", configVariable)
		}
	} else {
		log.Infof("Value of config variable %v has not changed, therefore no listeners will be notified", configVariable)
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
	switch {
	case hexColorPattern.MatchString(color.value):
		return getRGBColor(color.value)
	case colorNumberPattern.MatchString(color.value):
		return getColorNumber(color.value)
	case systemColorPattern.MatchString(color.value):
		return getSystemColor(color.value)
	}

	return nil, generateConfigError(inputSource, color, "Invalid color: %v", color.value)
}

func getColorNumber(colorNumberString string) (ThemeColor, error) {
	colorNumber, err := strconv.Atoi(colorNumberString)
	if err != nil || colorNumber < 0 || colorNumber > 255 {
		return nil, fmt.Errorf("Invalid color number: %v, Must be decimal integer in range 0 - 255", colorNumberString)
	}

	return NewColorNumber(int16(colorNumber)), nil
}

func getRGBColor(hexColorString string) (ThemeColor, error) {
	rgb, err := hex.DecodeString(hexColorString)
	if err != nil || len(rgb) != 3 {
		return nil, fmt.Errorf("Invalid hex color: %v, must be 3 byte hex value", hexColorString)
	}

	return NewRGBColor(rgb[0], rgb[1], rgb[2]), nil
}

func getSystemColor(systemColorString string) (ThemeColor, error) {
	systemColorValue, ok := systemColorValues[systemColorString]
	if !ok {
		return nil, fmt.Errorf("Invalid system color: %v, must be one of: %v",
			systemColorString, reflect.ValueOf(systemColorValues).MapKeys())
	}

	return NewSystemColor(systemColorValue), nil
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

func (config *Configuration) processUnmapCommand(unmapCommand *UnmapCommand, inputSource string) (err error) {
	viewID, ok := viewIDNames[unmapCommand.view.value]
	if !ok {
		return generateConfigError(inputSource, unmapCommand.view, "Invalid view: %v", unmapCommand.view.value)
	}

	if unmapCommand.from.value == "" {
		return generateConfigError(inputSource, unmapCommand.from, "from keystring cannot be empty")
	}

	if config.keyBindings.RemoveBinding(viewID, unmapCommand.from.value) {
		log.Infof("Unmapped \"%v\" for view %v", unmapCommand.from.value, unmapCommand.view.value)
	} else {
		log.Infof("Attempted to unmap \"%v\" for view %v but no binding existed", unmapCommand.from.value, unmapCommand.view.value)
	}

	return
}

func (config *Configuration) processQuitCommand() (err error) {
	log.Info("Processed quit command")
	config.channels.DoAction(Action{ActionType: ActionExit})
	return
}

func (config *Configuration) processNewTabCommand(newTabCommand *NewTabCommand, inputSource string) (err error) {
	if newTabCommand.tabName.value == "" {
		return generateConfigError(inputSource, newTabCommand.tabName, "tab name cannot be empty")
	}

	log.Infof("Processed new tab command with tab name: %v", newTabCommand.tabName.value)

	config.channels.DoAction(Action{
		ActionType: ActionNewTab,
		Args:       []interface{}{newTabCommand.tabName.value},
	})

	return
}

func (config *Configuration) processRemoveTabCommand() (err error) {
	log.Info("Processed remove tab command")
	config.channels.DoAction(Action{ActionType: ActionRemoveTab})
	return
}

func (config *Configuration) generateViewArgs(view *ConfigToken, args []*ConfigToken, inputSource string) (createViewArgs CreateViewArgs, err error) {
	viewID, ok := viewIDNames[view.value]
	if !ok {
		err = generateConfigError(inputSource, view, "Invalid view: %v", view.value)
		return
	}

	var viewArgs []interface{}
	for _, token := range args {
		viewArgs = append(viewArgs, token.value)
	}

	createViewArgs.viewID = viewID
	createViewArgs.viewArgs = viewArgs

	return
}

func (config *Configuration) processAddViewCommand(addViewCommand *AddViewCommand, inputSource string) (err error) {
	createViewArgs, err := config.generateViewArgs(addViewCommand.view, addViewCommand.args, inputSource)
	if err != nil {
		return
	}

	log.Infof("Processing addview command: %v %v", addViewCommand.view.value, createViewArgs.viewArgs)

	config.channels.DoAction(Action{
		ActionType: ActionAddView,
		Args: []interface{}{
			ActionAddViewArgs{
				CreateViewArgs: createViewArgs,
			},
		},
	})

	return
}

func (config *Configuration) processSplitViewCommand(splitViewCommand *SplitViewCommand, inputSource string) (err error) {
	createViewArgs, err := config.generateViewArgs(splitViewCommand.view, splitViewCommand.args, inputSource)
	if err != nil {
		return
	}

	log.Infof("Processing split view command: %v %v %v",
		splitViewCommand.orientation, splitViewCommand.view.value, createViewArgs.viewArgs)

	config.channels.DoAction(Action{
		ActionType: ActionSplitView,
		Args: []interface{}{
			ActionSplitViewArgs{
				CreateViewArgs: createViewArgs,
				orientation:    splitViewCommand.orientation,
			},
		},
	})

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

type booleanValueValidator struct {
	variableName string
}

func (booleanValueValidator booleanValueValidator) validate(value string) (processedValue interface{}, err error) {
	switch value {
	case "true":
		processedValue = true
	case "false":
		processedValue = false
	default:
		err = fmt.Errorf("%v must be set to either true or false but found: %v", booleanValueValidator.variableName, value)
	}

	return
}

type mouseScrollRowsValidator struct{}

func (mouseScrollRowsValidator mouseScrollRowsValidator) validate(value string) (processedValue interface{}, err error) {
	var mouseScrollRows int

	if mouseScrollRows, err = strconv.Atoi(value); err != nil {
		err = fmt.Errorf("%v must be an integer value greater than 0", CfMouseScrollRows)
	} else if mouseScrollRows <= 0 {
		err = fmt.Errorf("%v must be greater than 0", CfMouseScrollRows)
	} else {
		processedValue = mouseScrollRows
	}

	return
}
