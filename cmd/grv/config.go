package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"

	log "github.com/Sirupsen/logrus"
	slice "github.com/bradfitz/slice"
)

const (
	cfDefaultConfigHomeDir                = "/.config"
	cfGrvConfigDir                        = "/grv"
	cfGrvrcFile                           = "/grvrc"
	cfTabWidthMinValue                    = 1
	cfTabWidthDefaultValue                = 8
	cfClassicThemeName                    = "classic"
	cfSolarizedThemeName                  = "solarized"
	cfMouseDefaultValue                   = false
	cfMouseScrollRowsDefaultValue         = 3
	cfCommitGraphDefaultValue             = false
	cfConfirmCheckoutDefaultValue         = true
	cfPromptHistorySizeDefaultValue       = 1000
	cfGitBinaryFilePathDefaultValue       = ""
	cfCommitLimitDefaultValue             = "100000"
	cfDefaultViewDefaultValue             = ""
	cfDiffDisplayDefaultValue             = "fancy"
	cfInputPromptAfterCommandDefaultValue = true

	cfAllView             = "All"
	cfMainView            = "MainView"
	cfContainerView       = "ContainerView"
	cfWindowContainerView = "WindowContainerView"
	cfHistoryView         = "HistoryView"
	cfStatusView          = "StatusView"
	cfSummaryView         = "SummaryView"
	cfGRVStatusView       = "GRVStatusView"
	cfRefView             = "RefView"
	cfCommitView          = "CommitView"
	cfDiffView            = "DiffView"
	cfStatusBarView       = "StatusBarView"
	cfHelpBarView         = "HelpBarView"
	cfErrorView           = "ErrorView"
	cfGitStatusView       = "GitStatusView"
	cfContextMenuView     = "ContextMenuView"
	cfCommandOutputView   = "CommandOutputView"
	cfMessageBoxView      = "MessageBoxView"
	cfHelpView            = "HelpView"
	cfGRVVariableView     = "GRVVariableView"
	cfRemoteView          = "RemoteView"
	cfGitSummaryView      = "GitSummaryView"
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
	// CfPromptHistorySize stores the maximum number of prompt entries retained
	CfPromptHistorySize ConfigVariable = "prompt-history-size"
	// CfGitBinaryFilePath stores the file path to the git binary
	CfGitBinaryFilePath ConfigVariable = "git-binary-file-path"
	// CfCommitLimit stores the limit on the number of commits to load
	CfCommitLimit ConfigVariable = "commit-limit"
	// CfDefaultView stores the command to generate the default view
	CfDefaultView ConfigVariable = "default-view"
	// CfDiffDisplay stores the way diffs are displayed
	CfDiffDisplay ConfigVariable = "diff-display"
	// CfInputPromptAfterCommand stores whether the user is prompted for input after a command
	CfInputPromptAfterCommand ConfigVariable = "input-prompt-after-command"
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
	cfAllView:             ViewAll,
	cfMainView:            ViewMain,
	cfContainerView:       ViewContainer,
	cfWindowContainerView: ViewWindowContainer,
	cfHistoryView:         ViewHistory,
	cfStatusView:          ViewStatus,
	cfSummaryView:         ViewSummary,
	cfGRVStatusView:       ViewGRVStatus,
	cfRefView:             ViewRef,
	cfCommitView:          ViewCommit,
	cfDiffView:            ViewDiff,
	cfStatusBarView:       ViewStatusBar,
	cfHelpBarView:         ViewHelpBar,
	cfErrorView:           ViewError,
	cfGitStatusView:       ViewGitStatus,
	cfContextMenuView:     ViewContextMenu,
	cfCommandOutputView:   ViewCommandOutput,
	cfMessageBoxView:      ViewMessageBox,
	cfHelpView:            ViewHelp,
	cfGRVVariableView:     ViewGRVVariable,
	cfRemoteView:          ViewRemote,
	cfGitSummaryView:      ViewGitSummary,
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

	cfCommitView + ".Title":                  CmpCommitviewTitle,
	cfCommitView + ".Footer":                 CmpCommitviewFooter,
	cfCommitView + ".ShortOid":               CmpCommitviewShortOid,
	cfCommitView + ".Date":                   CmpCommitviewDate,
	cfCommitView + ".Author":                 CmpCommitviewAuthor,
	cfCommitView + ".Summary":                CmpCommitviewSummary,
	cfCommitView + ".Tag":                    CmpCommitviewTag,
	cfCommitView + ".LocalBranch":            CmpCommitviewLocalBranch,
	cfCommitView + ".RemoteBranch":           CmpCommitviewRemoteBranch,
	cfCommitView + ".CommitGraphCommit":      CmpCommitviewGraphCommit,
	cfCommitView + ".CommitGraphMergeCommit": CmpCommitviewGraphMergeCommit,
	cfCommitView + ".CommitGraphBranch1":     CmpCommitviewGraphBranch1,
	cfCommitView + ".CommitGraphBranch2":     CmpCommitviewGraphBranch2,
	cfCommitView + ".CommitGraphBranch3":     CmpCommitviewGraphBranch3,
	cfCommitView + ".CommitGraphBranch4":     CmpCommitviewGraphBranch4,
	cfCommitView + ".CommitGraphBranch5":     CmpCommitviewGraphBranch5,
	cfCommitView + ".CommitGraphBranch6":     CmpCommitviewGraphBranch6,
	cfCommitView + ".CommitGraphBranch7":     CmpCommitviewGraphBranch7,

	cfDiffView + ".Title":                   CmpDiffviewTitle,
	cfDiffView + ".Footer":                  CmpDiffviewFooter,
	cfDiffView + ".Normal":                  CmpDiffviewDifflineNormal,
	cfDiffView + ".CommitAuthor":            CmpDiffviewDifflineDiffCommitAuthor,
	cfDiffView + ".CommitAuthorDate":        CmpDiffviewDifflineDiffCommitAuthorDate,
	cfDiffView + ".CommitCommitter":         CmpDiffviewDifflineDiffCommitCommitter,
	cfDiffView + ".CommitCommitterDate":     CmpDiffviewDifflineDiffCommitCommitterDate,
	cfDiffView + ".CommitMessage":           CmpDiffviewDifflineDiffCommitMessage,
	cfDiffView + ".StatsFile":               CmpDiffviewDifflineDiffStatsFile,
	cfDiffView + ".GitDiffHeader":           CmpDiffviewDifflineGitDiffHeader,
	cfDiffView + ".GitDiffExtendedHeader":   CmpDiffviewDifflineGitDiffExtendedHeader,
	cfDiffView + ".UnifiedDiffHeader":       CmpDiffviewDifflineUnifiedDiffHeader,
	cfDiffView + ".HunkStart":               CmpDiffviewDifflineHunkStart,
	cfDiffView + ".HunkHeader":              CmpDiffviewDifflineHunkHeader,
	cfDiffView + ".AddedLine":               CmpDiffviewDifflineLineAdded,
	cfDiffView + ".RemovedLine":             CmpDiffviewDifflineLineRemoved,
	cfDiffView + ".FancySeparator":          CmpDiffviewFancyDiffLineSeparator,
	cfDiffView + ".FancyFile":               CmpDiffviewFancyDiffLineFile,
	cfDiffView + ".FancyLineAdded":          CmpDiffviewFancyDifflineLineAdded,
	cfDiffView + ".FancyLineRemoved":        CmpDiffviewFancyDifflineLineRemoved,
	cfDiffView + ".FancyLineAddedChange":    CmpDiffviewFancyDifflineLineAddedChange,
	cfDiffView + ".FancyLineRemovedChange":  CmpDiffviewFancyDifflineLineRemovedChange,
	cfDiffView + ".FancyEmptyLineAdded":     CmpDiffviewFancyDifflineEmptyLineAdded,
	cfDiffView + ".FancyEmptyLineRemoved":   CmpDiffviewFancyDifflineEmptyLineRemoved,
	cfDiffView + ".FancyTrailingWhitespace": CmpDiffviewFancyDifflineTrailingWhitespace,

	cfGitStatusView + ".Message":         CmpGitStatusMessage,
	cfGitStatusView + ".StagedTitle":     CmpGitStatusStagedTitle,
	cfGitStatusView + ".UnstagedTitle":   CmpGitStatusUnstagedTitle,
	cfGitStatusView + ".UntrackedTitle":  CmpGitStatusUntrackedTitle,
	cfGitStatusView + ".ConflictedTitle": CmpGitStatusConflictedTitle,
	cfGitStatusView + ".StagedFile":      CmpGitStatusStagedFile,
	cfGitStatusView + ".UnstagedFile":    CmpGitStatusUnstagedFile,
	cfGitStatusView + ".UntrackedFile":   CmpGitStatusUntrackedFile,
	cfGitStatusView + ".ConflictedFile":  CmpGitStatusConflictedFile,

	cfHelpView + ".Title":                      CmpHelpViewTitle,
	cfHelpView + ".IndexTitle":                 CmpHelpViewIndexTitle,
	cfHelpView + ".IndexSubTitle":              CmpHelpViewIndexSubTitle,
	cfHelpView + ".SectionTitle":               CmpHelpViewSectionTitle,
	cfHelpView + ".SectionSubTitle":            CmpHelpViewSectionSubTitle,
	cfHelpView + ".SectionDescription":         CmpHelpViewSectionDescription,
	cfHelpView + ".SectionCodeBlock":           CmpHelpViewSectionCodeBlock,
	cfHelpView + ".SectionTableHeader":         CmpHelpViewSectionTableHeader,
	cfHelpView + ".SectionTableRow":            CmpHelpViewSectionTableRow,
	cfHelpView + ".SectionTableRowHighlighted": CmpHelpViewSectionTableRowHighlighted,
	cfHelpView + ".SectionTableCellSeparator":  CmpHelpViewSectionTableCellSeparator,
	cfHelpView + ".Footer":                     CmpHelpViewFooter,

	cfStatusBarView + ".Normal": CmpStatusbarviewNormal,

	cfHelpBarView + ".Special": CmpHelpbarviewSpecial,
	cfHelpBarView + ".Normal":  CmpHelpbarviewNormal,

	cfErrorView + ".Title":  CmpErrorViewTitle,
	cfErrorView + ".Footer": CmpErrorViewFooter,
	cfErrorView + ".Errors": CmpErrorViewErrors,

	cfContextMenuView + ".Title":      CmpContextMenuTitle,
	cfContextMenuView + ".Content":    CmpContextMenuContent,
	cfContextMenuView + ".KeyMapping": CmpContextMenuKeyMapping,
	cfContextMenuView + ".Footer":     CmpContextMenuFooter,

	cfCommandOutputView + ".Title":   CmpCommandOutputTitle,
	cfCommandOutputView + ".Command": CmpCommandOutputCommand,
	cfCommandOutputView + ".Normal":  CmpCommandOutputNormal,
	cfCommandOutputView + ".Error":   CmpCommandOutputError,
	cfCommandOutputView + ".Success": CmpCommandOutputSuccess,
	cfCommandOutputView + ".Footer":  CmpCommandOutputFooter,

	cfMessageBoxView + ".Title":          CmpMessageBoxTitle,
	cfMessageBoxView + ".Content":        CmpMessageBoxContent,
	cfMessageBoxView + ".SelectedButton": CmpMessageBoxSelectedButton,

	cfGRVVariableView + ".Title":    CmpGRVVariableViewTitle,
	cfGRVVariableView + ".Variable": CmpGRVVariableViewVariable,
	cfGRVVariableView + ".Value":    CmpGRVVariableViewValue,
	cfGRVVariableView + ".Footer":   CmpGRVVariableViewFooter,

	cfRemoteView + ".Title":  CmpRemoteViewTitle,
	cfRemoteView + ".Remote": CmpRemoteViewRemote,
	cfRemoteView + ".Footer": CmpRemoteViewFooter,

	cfGitSummaryView + ".Header":          CmpSummaryViewHeader,
	cfGitSummaryView + ".Normal":          CmpSummaryViewNormal,
	cfGitSummaryView + ".BranchAhead":     CmpSummaryViewBranchAhead,
	cfGitSummaryView + ".BranchBehind":    CmpSummaryViewBranchBehind,
	cfGitSummaryView + ".StagedFile":      CmpSummaryViewStagedFile,
	cfGitSummaryView + ".UnstagedFile":    CmpSummaryViewUnstagedFile,
	cfGitSummaryView + ".NoModifiedFiles": CmpSummaryViewNoModifiedFiles,
}

var colorNumberPattern = regexp.MustCompile(`[0-9]{1,3}`)
var hexColorPattern = regexp.MustCompile(`[a-fA-F0-9]{6}`)
var systemColorPattern = regexp.MustCompile(`[a-zA-Z]+`)

var commandBodyVariablePattern = regexp.MustCompile(`\$(\d+|\{\d+\}|@|\{@\})`)
var argBracketsPattern = regexp.MustCompile(`\{|\}`)

var viewNames = map[ViewID]string{}

func init() {
	for viewName, viewID := range viewIDNames {
		viewNames[viewID] = viewName
	}
}

// ViewName returns the name of the view with the provided ID
func ViewName(viewID ViewID) (viewName string) {
	viewName, _ = viewNames[viewID]
	return
}

// ThemeComponentNames returns the names of all theme components
func ThemeComponentNames() (themeComponentNames []string) {
	for themeComponent := range themeComponents {
		themeComponentNames = append(themeComponentNames, themeComponent)
	}

	slice.Sort(themeComponentNames, func(i, j int) bool {
		return themeComponentNames[i] < themeComponentNames[j]
	})

	return
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
	KeyStrings(ActionType, ViewHierarchy) []BoundKeyString
	GenerateHelpSections() []*HelpSection
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
	defaultValue      interface{}
	value             interface{}
	validator         ConfigVariableValidator
	description       string
	onChangeListeners []ConfigVariableOnChangeListener
}

// Configuration contains all configuration state
type Configuration struct {
	configVariables map[ConfigVariable]*ConfigurationVariable
	themes          map[string]MutableTheme
	keyBindings     KeyBindings
	grvConfigDir    string
	channels        Channels
	variables       GRVVariableGetter
	customCommands  map[string]string
	inputConsumer   InputConsumer
}

// NewConfiguration creates a Configuration instance with default values
func NewConfiguration(keyBindings KeyBindings, channels Channels, variables GRVVariableGetter, inputConsumer InputConsumer) *Configuration {
	config := &Configuration{
		keyBindings:    keyBindings,
		channels:       channels,
		variables:      variables,
		inputConsumer:  inputConsumer,
		customCommands: map[string]string{},
		themes: map[string]MutableTheme{
			cfClassicThemeName:   NewClassicTheme(),
			cfSolarizedThemeName: NewSolarizedTheme(),
		},
	}

	config.configVariables = map[ConfigVariable]*ConfigurationVariable{
		CfTabWidth: {
			defaultValue: cfTabWidthDefaultValue,
			validator:    tabWidithValidator{},
			description:  fmt.Sprintf("Tab character screen width (minimum value: %v)", cfTabWidthMinValue),
		},
		CfTheme: {
			defaultValue: cfSolarizedThemeName,
			validator: themeValidator{
				config: config,
			},
			description: "The currently active theme",
		},
		CfMouse: {
			defaultValue: cfMouseDefaultValue,
			validator: booleanValueValidator{
				variableName: string(CfMouse),
			},
			description: "Mouse support enabled",
		},
		CfMouseScrollRows: {
			defaultValue: cfMouseScrollRowsDefaultValue,
			validator:    mouseScrollRowsValidator{},
			description:  "Number of rows scrolled for each mouse event",
		},
		CfCommitGraph: {
			defaultValue: cfCommitGraphDefaultValue,
			validator: booleanValueValidator{
				variableName: string(CfCommitGraph),
			},
			description: "Commit graph visible",
		},
		CfConfirmCheckout: {
			defaultValue: cfConfirmCheckoutDefaultValue,
			validator: booleanValueValidator{
				variableName: string(CfConfirmCheckout),
			},
			description: "Confirm before performing git checkout",
		},
		CfPromptHistorySize: {
			defaultValue: cfPromptHistorySizeDefaultValue,
			validator:    promptHistorySizeValidator{},
			description:  "Maximum number of prompt entries retained",
		},
		CfGitBinaryFilePath: {
			defaultValue: cfGitBinaryFilePathDefaultValue,
			description:  "File path to git binary. Required only when git binary is not in $PATH",
		},
		CfCommitLimit: {
			defaultValue: cfCommitLimitDefaultValue,
			description:  "Limit the number of commits loaded. Allowed values: number, date, oid or tag",
		},
		CfDefaultView: {
			defaultValue: cfDefaultViewDefaultValue,
			validator:    &defaultViewValidator{config: config},
			description:  "Command to generate a custom default view on start up",
		},
		CfDiffDisplay: {
			defaultValue: cfDiffDisplayDefaultValue,
			validator:    &diffDisplayValidator{},
			description:  "Diff display format",
		},
		CfInputPromptAfterCommand: {
			defaultValue: cfInputPromptAfterCommandDefaultValue,
			validator: booleanValueValidator{
				variableName: string(CfInputPromptAfterCommand),
			},
			description: `Display "Press any key to continue" after executing external command`,
		},
	}

	for _, configVariable := range config.configVariables {
		configVariable.value = configVariable.defaultValue
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
	case *GitCommand:
		config.processGitCommand(command)
	case *HelpCommand:
		config.processHelpCommand(command)
	case *ShellCommand:
		err = config.processShellCommand(command, inputSource)
	case *DefCommand:
		err = config.processDefCommand(command)
	case *UndefCommand:
		err = config.processUndefCommand(command, inputSource)
	case *CustomCommand:
		err = config.processCustomCommand(command)
	case *EvalKeysCommand:
		err = config.processEvalKeysCommand(command)
	case *SleepCommand:
		err = config.processSleepCommand(command)
	default:
		log.Errorf("Unknown command type %T", command)
	}

	return
}

func (config *Configuration) processSetCommand(setCommand *SetCommand, inputSource string) error {
	configVariable := ConfigVariable(setCommand.variable.value)
	variable, ok := config.configVariables[configVariable]
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
	if variable, ok := config.configVariables[configVariable]; ok {
		return variable
	}

	panic(fmt.Sprintf("No ConfigVariable exists exists for ID %v", configVariable))
}

func (config *Configuration) processMapCommand(mapCommand *MapCommand, inputSource string) (err error) {
	view := mapCommand.view.value

	viewID, ok := viewIDNames[view]
	if !ok {
		return generateConfigError(inputSource, mapCommand.view, "Invalid view: %v", view)
	}

	from := mapCommand.from.value
	to := mapCommand.to.value

	if from == "" {
		return generateConfigError(inputSource, mapCommand.from, "from keystring cannot be empty")
	} else if to == "" {
		return generateConfigError(inputSource, mapCommand.to, "to keystring cannot be empty")
	}

	if (mapCommand.to.tokenType & CtkShellCommand) != 0 {
		to = "<grv-prompt>" + strings.TrimSuffix(to, "<Enter>") + "<Enter>"
	}

	config.keyBindings.SetKeystringBinding(viewID, from, to)

	log.Infof("Mapped \"%v\" to \"%v\" for view %v", from, to, view)

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

func (config *Configuration) processGitCommand(gitCommand *GitCommand) {
	var buffer bytes.Buffer

	buffer.WriteString("git ")

	for _, token := range gitCommand.args {
		buffer.WriteString(token.value)
		buffer.WriteRune(' ')
	}

	buffer.Truncate(buffer.Len() - 1)
	command := buffer.String()

	var outputType ShellCommandOutputType

	if gitCommand.interactive {
		outputType = TerminalOutput
	} else {
		outputType = WindowOutput
	}

	config.runCommand(command, outputType)
}

func (config *Configuration) processShellCommand(shellCommand *ShellCommand, inputSource string) (err error) {
	command := strings.TrimSpace(shellCommand.command.value)
	commandLength := len([]rune(command))

	if commandLength < 1 {
		return generateConfigError(inputSource, shellCommand.command, "Expected command token with preceeding !")
	} else if commandLength == 1 {
		log.Debugf("Empty command, not running")
		return
	}

	prefix, _ := utf8.DecodeRuneInString(command)
	outputType := OutputType(prefix)
	command = command[1:]

	config.runCommand(command, outputType)

	return
}

func (config *Configuration) processDefCommand(defCommand *DefCommand) (err error) {
	if err = DefineCustomCommand(defCommand.commandName); err != nil {
		return
	}

	if _, exists := config.customCommands[defCommand.commandName]; exists {
		log.Debugf("Overriding previous command definition for command %v", defCommand.commandName)
	}

	config.customCommands[defCommand.commandName] = defCommand.functionBody
	config.channels.ReportStatus("Defined user comamnd %v", defCommand.commandName)

	return
}

func (config *Configuration) processUndefCommand(undefCommand *UndefCommand, inputSource string) (err error) {
	if err = UndefineCustomCommand(undefCommand.commandName.value); err != nil {
		return generateConfigError(inputSource, undefCommand.commandName, "%v", err)
	}

	if _, exists := config.customCommands[undefCommand.commandName.value]; exists {
		delete(config.customCommands, undefCommand.commandName.value)
		config.channels.ReportStatus("Undefined user comamnd %v", undefCommand.commandName.value)
	} else {
		log.Warnf("No user defined command %v exists", undefCommand.commandName.value)
	}

	return
}

func (config *Configuration) processCustomCommand(customCommand *CustomCommand) (err error) {
	commandBody, ok := config.customCommands[customCommand.commandName]
	if !ok {
		return fmt.Errorf("No command with name %v exists", customCommand.commandName)
	}

	log.Infof("Executing user defined command %v with args: %v", customCommand.commandName, customCommand.args)
	commandBody = config.processConfigCommandBody(commandBody, customCommand.args)

	if errs := config.Evaluate(commandBody); len(errs) > 0 {
		config.channels.ReportErrors(errs)
		err = fmt.Errorf("Command %v generated errors", customCommand.commandName)
	}

	return
}

func (config *Configuration) processConfigCommandBody(commandBody string, args []string) string {
	allMatchIndexes := commandBodyVariablePattern.FindAllStringSubmatchIndex(commandBody, -1)
	if len(allMatchIndexes) == 0 {
		return commandBody
	}

	var processedCommandBody bytes.Buffer
	lastMatchIndex := 0

	for _, matchIndexes := range allMatchIndexes {
		matchStartIndex := matchIndexes[0]
		matchEndIndex := matchIndexes[1]

		escapeCount := 0
		for bodyIndex := matchStartIndex - 1; bodyIndex > -1 && commandBody[bodyIndex] == '$'; bodyIndex-- {
			escapeCount++
		}

		processedCommandBody.WriteString(commandBody[lastMatchIndex : matchStartIndex-escapeCount])

		if escapeCount > 0 && escapeCount%2 != 0 {
			processedCommandBody.WriteString(strings.Repeat("$", (escapeCount-1)/2))
			processedCommandBody.WriteString(commandBody[matchStartIndex:matchEndIndex])
		} else {
			if escapeCount > 0 {
				processedCommandBody.WriteString(strings.Repeat("$", escapeCount/2))
			}

			argString := argBracketsPattern.ReplaceAllString(commandBody[matchIndexes[2]:matchIndexes[3]], "")

			if argString == "@" {
				processedCommandBody.WriteString(strings.Join(args, " "))
			} else if argNumber, err := strconv.Atoi(argString); err == nil {
				if argNumber > 0 && argNumber-1 < len(args) {
					processedCommandBody.WriteString(args[argNumber-1])
				}
			} else {
				log.Errorf("Failed to parse argument placeholder %v: %v", argString, err)
			}
		}

		lastMatchIndex = matchEndIndex
	}

	processedCommandBody.WriteString(commandBody[lastMatchIndex:])

	return processedCommandBody.String()
}

func (config *Configuration) processEvalKeysCommand(evalKeysCommand *EvalKeysCommand) (err error) {
	log.Debugf("Processing keys: %v", evalKeysCommand.keys)
	config.inputConsumer.ProcessInput(evalKeysCommand.keys)
	return
}

func (config *Configuration) processSleepCommand(sleepCommand *SleepCommand) (err error) {
	config.channels.DoAction(Action{
		ActionType: ActionSleep,
		Args:       []interface{}{sleepCommand.sleepSeconds},
	})

	return
}

func (config *Configuration) runCommand(command string, outputType ShellCommandOutputType) {
	NewShellCommandProcessor(config.channels, config.variables, command, outputType).Execute()
}

func (config *Configuration) processHelpCommand(helpCommand *HelpCommand) {
	args := []interface{}{}
	if helpCommand.searchTerm != "" {
		args = append(args, helpCommand.searchTerm)
	}

	config.channels.DoAction(Action{
		ActionType: ActionShowHelpView,
		Args:       args,
	})
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

// KeyStrings returns the set of keystrings bound to the provided action and view
func (config *Configuration) KeyStrings(actionType ActionType, viewHierarchy ViewHierarchy) (keystrings []BoundKeyString) {
	for i := len(viewHierarchy) - 1; i > -1; i-- {
		keystrings = append(keystrings, config.keyBindings.KeyStrings(actionType, viewHierarchy[i])...)
	}

	sort.Stable(slice.SortInterface(keystrings, func(i, j int) bool {
		return !keystrings[i].userDefinedBinding && keystrings[j].userDefinedBinding
	}))

	return
}

// HandleEvent reacts to an event
func (config *Configuration) HandleEvent(event Event) (err error) {
	switch event.EventType {
	case ViewRemovedEvent:
		for _, view := range event.Args {
			if listener, ok := view.(ConfigVariableOnChangeListener); ok {
				config.removeOnChangeListener(listener)
			}
		}
	}

	return
}

func (config *Configuration) removeOnChangeListener(onChangeListener ConfigVariableOnChangeListener) {
	for _, variable := range config.configVariables {
		for index, listener := range variable.onChangeListeners {
			if onChangeListener == listener {
				variable.onChangeListeners = append(variable.onChangeListeners[:index], variable.onChangeListeners[index+1:]...)
				break
			}
		}
	}
}

// GenerateHelpSections generates all help tables related to configuration
func (config *Configuration) GenerateHelpSections() (helpSections []*HelpSection) {
	helpSections = append(helpSections, config.keyBindings.GenerateHelpSections(config)...)

	helpSections = append(helpSections, config.generateConfigVariableHelpSection())

	helpSections = append(helpSections, GenerateConfigCommandHelpSections(config)...)

	return helpSections
}

func (config *Configuration) generateConfigVariableHelpSection() (helpSection *HelpSection) {
	isDocFile := os.Getenv(MnGenerateDocumentationEnv) != ""

	headers := []TableHeader{
		{text: "Variable", themeComponentID: CmpHelpViewSectionTableHeader},
		{text: "Type", themeComponentID: CmpHelpViewSectionTableHeader},
		{text: "Default Value", themeComponentID: CmpHelpViewSectionTableHeader},
	}

	if !isDocFile {
		headers = append(headers, TableHeader{text: "Current Value", themeComponentID: CmpHelpViewSectionTableHeader})
	}

	headers = append(headers, TableHeader{text: "Description", themeComponentID: CmpHelpViewSectionTableHeader})

	tableFormatter := NewTableFormatterWithHeaders(headers, config)
	tableFormatter.SetGridLines(true)

	configVariableNames := []ConfigVariable{}
	for configVariableName := range config.configVariables {
		configVariableNames = append(configVariableNames, configVariableName)
	}

	slice.Sort(configVariableNames, func(i, j int) bool {
		return configVariableNames[i] < configVariableNames[j]
	})

	tableFormatter.Resize(uint(len(configVariableNames)))

	for rowIndex, configVariableName := range configVariableNames {
		configVariable := config.configVariables[configVariableName]

		tableFormatter.SetCellWithStyle(uint(rowIndex), 0, CmpHelpViewSectionTableRow, "%v", configVariableName)
		tableFormatter.SetCellWithStyle(uint(rowIndex), 1, CmpHelpViewSectionTableRow, "%v", reflect.TypeOf(configVariable.defaultValue))
		tableFormatter.SetCellWithStyle(uint(rowIndex), 2, CmpHelpViewSectionTableRow, "%v", configVariable.defaultValue)

		if isDocFile {
			tableFormatter.SetCellWithStyle(uint(rowIndex), 3, CmpHelpViewSectionTableRow, "%v", configVariable.description)
		} else {
			currentValueThemeComponentID := CmpHelpViewSectionTableRow
			if configVariable.defaultValue != configVariable.value {
				currentValueThemeComponentID = CmpHelpViewSectionTableRowHighlighted
			}

			tableFormatter.SetCellWithStyle(uint(rowIndex), 3, currentValueThemeComponentID, "%v", configVariable.value)
			tableFormatter.SetCellWithStyle(uint(rowIndex), 4, CmpHelpViewSectionTableRow, "%v", configVariable.description)
		}
	}

	return &HelpSection{
		title: HelpSectionText{text: "Configuration Variables"},
		description: []HelpSectionText{
			{text: "Configuration variables allow features to be enabled, disabled and configured."},
			{text: "They are specified using the set command in the grvrc file or at the command prompt"},
		},
		tableFormatter: tableFormatter,
	}
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

type promptHistorySizeValidator struct{}

func (promptHistorySizeValidator promptHistorySizeValidator) validate(value string) (processedValue interface{}, err error) {
	var promptHistorySize int

	if promptHistorySize, err = strconv.Atoi(value); err != nil {
		err = fmt.Errorf("%v must be an integer value greater than or equal to 0", CfPromptHistorySize)
	} else if promptHistorySize < 0 {
		err = fmt.Errorf("%v must be greater than or equal to 0", CfPromptHistorySize)
	} else {
		processedValue = promptHistorySize
	}

	return
}

type defaultViewValidator struct {
	config *Configuration
}

func (defaultViewValidator *defaultViewValidator) validate(value string) (processedValue interface{}, err error) {
	if _, exists := defaultViewValidator.config.customCommands[value]; !exists {
		err = fmt.Errorf("No user defined command with name \"%v\" exists", value)
	} else {
		processedValue = value
	}

	return
}

type diffDisplayValidator struct{}

func (diffDisplayValidator *diffDisplayValidator) validate(value string) (processedValue interface{}, err error) {
	if IsValidDiffProcessorName(value) {
		processedValue = value
	} else {
		err = fmt.Errorf("Invalid %v value %v", CfDiffDisplay, value)
	}

	return
}
