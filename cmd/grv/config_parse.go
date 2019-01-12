package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	slice "github.com/bradfitz/slice"
)

const (
	setCommand            = "set"
	themeCommand          = "theme"
	mapCommand            = "map"
	unmapCommand          = "unmap"
	quitCommand           = "q"
	addtabCommand         = "addtab"
	removetabCommand      = "rmtab"
	addviewCommand        = "addview"
	vsplitCommand         = "vsplit"
	hsplitCommand         = "hsplit"
	splitCommand          = "split"
	gitCommand            = "git"
	gitInteractiveCommand = "giti"
	helpCommand           = "help"
	defCommand            = "def"
	undefCommand          = "undef"
	evalkeysCommand       = "evalkeys"
	sleepCommand          = "sleep"
)

const (
	openingBrace = "{"
	closingBrace = "}"
)

var isIdentifier = regexp.MustCompile(`[[:alnum:]]+`).MatchString
var commentTokens = map[ConfigTokenType]bool{
	CtkComment: true,
}
var whiteSpaceTokens = map[ConfigTokenType]bool{
	CtkWhiteSpace: true,
	CtkComment:    true,
}
var whiteSpaceAndTerminatorTokens = map[ConfigTokenType]bool{
	CtkWhiteSpace: true,
	CtkComment:    true,
	CtkTerminator: true,
}

type commandConstructor func(parser *ConfigParser, commandToken *ConfigToken, tokens []*ConfigToken) (ConfigCommand, error)

// ConfigCommand represents a config command
type ConfigCommand interface {
	configCommand()
}

// SetCommand contains state for setting a config variable to a value
type SetCommand struct {
	variable *ConfigToken
	value    *ConfigToken
}

func (setCommand *SetCommand) configCommand() {}

// ThemeCommand contains state for setting a components values for on a theme
type ThemeCommand struct {
	name      *ConfigToken
	component *ConfigToken
	bgcolor   *ConfigToken
	fgcolor   *ConfigToken
}

func (themeCommand *ThemeCommand) configCommand() {}

// MapCommand contains state for mapping a key sequence to another
type MapCommand struct {
	view *ConfigToken
	from *ConfigToken
	to   *ConfigToken
}

func (mapCommand *MapCommand) configCommand() {}

// UnmapCommand contains state for unmapping a key sequence
type UnmapCommand struct {
	view *ConfigToken
	from *ConfigToken
}

func (unmapCommand *UnmapCommand) configCommand() {}

// QuitCommand represents the command to quit grv
type QuitCommand struct{}

func (quitCommand *QuitCommand) configCommand() {}

// NewTabCommand represents the command to create a new tab
type NewTabCommand struct {
	tabName *ConfigToken
}

func (newTabCommand *NewTabCommand) configCommand() {}

// RemoveTabCommand represents the command to remove the currently active tab
type RemoveTabCommand struct{}

func (removeTabCommand *RemoveTabCommand) configCommand() {}

// AddViewCommand represents the command to add a new view
// to the currently active view
type AddViewCommand struct {
	view *ConfigToken
	args []*ConfigToken
}

func (addViewCommand *AddViewCommand) configCommand() {}

// SplitViewCommand represents the command split the currently
// active view with a new view
type SplitViewCommand struct {
	orientation ContainerOrientation
	view        *ConfigToken
	args        []*ConfigToken
}

func (splitViewCommand *SplitViewCommand) configCommand() {}

// GitCommand represents a git command
type GitCommand struct {
	interactive bool
	args        []*ConfigToken
}

func (gitCommand *GitCommand) configCommand() {}

// HelpCommand represents the command to show the help view
type HelpCommand struct {
	searchTerm string
}

func (helpCommand *HelpCommand) configCommand() {}

// ShellCommand represents a shell command
type ShellCommand struct {
	command *ConfigToken
}

func (shellCommand *ShellCommand) configCommand() {}

// DefCommand represents a function definition command
type DefCommand struct {
	commandName  string
	functionBody string
}

func (defCommand *DefCommand) configCommand() {}

// UndefCommand represents the command to undefine a command
type UndefCommand struct {
	commandName *ConfigToken
}

func (undefCommand *UndefCommand) configCommand() {}

// CustomCommand represents an invocation of a user defined command
type CustomCommand struct {
	commandName string
	args        []string
}

func (customCommand *CustomCommand) configCommand() {}

// EvalKeysCommand represents a key evaluation command
type EvalKeysCommand struct {
	keys string
}

func (evalKeysCommand *EvalKeysCommand) configCommand() {}

// SleepCommand represents a command to sleep
type SleepCommand struct {
	sleepSeconds float64
}

func (sleepCommand *SleepCommand) configCommand() {}

type commandHelpGenerator func(config Config) []*HelpSection
type commandCustomParser func(parser *ConfigParser) (tokens []*ConfigToken, err error)

type commandDescriptor struct {
	tokenTypes           []ConfigTokenType
	constructor          commandConstructor
	commandHelpGenerator commandHelpGenerator
	customParser         commandCustomParser
	userDefined          bool
}

// DefineCustomCommand allows a custom command to be parsed
func DefineCustomCommand(commandName string) (err error) {
	if existingDescriptor, ok := commandDescriptors[commandName]; ok && !existingDescriptor.userDefined {
		return fmt.Errorf("Cannot override built in command %v", commandName)
	}

	commandDescriptors[commandName] = &commandDescriptor{
		customParser: parseVarArgsCommand(),
		constructor:  customCommandConstructor,
		userDefined:  true,
	}

	return
}

// UndefineCustomCommand invalidates a custom command
func UndefineCustomCommand(commandName string) (err error) {
	if existingDescriptor, ok := commandDescriptors[commandName]; ok && !existingDescriptor.userDefined {
		return fmt.Errorf("Cannot undefine built in command %v", commandName)
	}

	delete(commandDescriptors, commandName)

	return
}

var commandDescriptors = map[string]*commandDescriptor{
	setCommand: {
		tokenTypes:           []ConfigTokenType{CtkWord, CtkWord},
		constructor:          setCommandConstructor,
		commandHelpGenerator: GenerateSetCommandHelpSections,
	},
	themeCommand: {
		tokenTypes:           []ConfigTokenType{CtkOption, CtkWord, CtkOption, CtkWord, CtkOption, CtkWord, CtkOption, CtkWord},
		constructor:          themeCommandConstructor,
		commandHelpGenerator: GenerateThemeCommandHelpSections,
	},
	mapCommand: {
		tokenTypes:           []ConfigTokenType{CtkWord, CtkWord, CtkWord | CtkShellCommand},
		constructor:          mapCommandConstructor,
		commandHelpGenerator: GenerateMapCommandHelpSections,
	},
	unmapCommand: {
		tokenTypes:           []ConfigTokenType{CtkWord, CtkWord},
		constructor:          unmapCommandConstructor,
		commandHelpGenerator: GenerateUnmapCommandHelpSections,
	},
	quitCommand: {
		constructor:          quitCommandConstructor,
		commandHelpGenerator: GenerateQuitCommandHelpSections,
	},
	addtabCommand: {
		tokenTypes:           []ConfigTokenType{CtkWord},
		constructor:          newTabCommandConstructor,
		commandHelpGenerator: GenerateAddTabCommandHelpSections,
	},
	removetabCommand: {
		constructor:          newRemoveTabCommandConstructor,
		commandHelpGenerator: GenerateRmTabCommandHelpSections,
	},
	addviewCommand: {
		customParser:         parseVarArgsCommand(),
		constructor:          addViewCommandConstructor,
		commandHelpGenerator: GenerateAddViewCommandHelpSections,
	},
	vsplitCommand: {
		customParser:         parseVarArgsCommand(),
		constructor:          splitViewCommandConstructor,
		commandHelpGenerator: GenerateVSplitCommandHelpSections,
	},
	hsplitCommand: {
		customParser:         parseVarArgsCommand(),
		constructor:          splitViewCommandConstructor,
		commandHelpGenerator: GenerateHSplitCommandHelpSections,
	},
	splitCommand: {
		customParser:         parseVarArgsCommand(),
		constructor:          splitViewCommandConstructor,
		commandHelpGenerator: GenerateSplitCommandHelpSections,
	},
	gitCommand: {
		customParser:         parseVarArgsCommand(),
		constructor:          gitCommandConstructor,
		commandHelpGenerator: GenerateGitCommandHelpSections,
	},
	gitInteractiveCommand: {
		customParser:         parseVarArgsCommand(),
		constructor:          gitCommandConstructor,
		commandHelpGenerator: GenerateGitiCommandHelpSections,
	},
	helpCommand: {
		customParser:         parseVarArgsCommand(),
		constructor:          helpCommandConstructor,
		commandHelpGenerator: GenerateHelpCommandHelpSections,
	},
	defCommand: {
		customParser:         parseDefCommand,
		constructor:          defCommandConstructor,
		commandHelpGenerator: GenerateDefCommandHelpSections,
	},
	undefCommand: {
		tokenTypes:           []ConfigTokenType{CtkWord},
		constructor:          undefCommandConstructor,
		commandHelpGenerator: GenerateUndefCommandHelpSections,
	},
	evalkeysCommand: {
		customParser:         parseVarArgsParserGenerator(false),
		constructor:          evalKeysCommandConstructor,
		commandHelpGenerator: GenerateEvalKeysCommandHelpSections,
	},
	sleepCommand: {
		tokenTypes:           []ConfigTokenType{CtkWord},
		constructor:          sleepCommandConstructor,
		commandHelpGenerator: GenerateSleepCommandHelpSections,
	},
}

// GenerateConfigCommandHelpSections generates help documentation for all configuration commands
func GenerateConfigCommandHelpSections(config Config) (helpSections []*HelpSection) {
	helpSections = append(helpSections, &HelpSection{
		title: HelpSectionText{text: "Configuration Commands"},
		description: []HelpSectionText{
			{text: "The behaviour of GRV can be customised through the use of commands specified in a configuration file"},
			{text: "GRV will look for the following configuration files on start up:"},
			{},
			{text: " - $XDG_CONFIG_HOME/grv/grvrc", themeComponentID: CmpHelpViewSectionCodeBlock},
			{text: " - $HOME/.config/grv/grvrc", themeComponentID: CmpHelpViewSectionCodeBlock},
			{},
			{text: "GRV will attempt to process the first file which exists."},
			{text: "Commands can also be specified within GRV using the command prompt :"},
			{},
			{text: "Below are the set of configuration commands supported:"},
		},
	})

	commands := []string{}
	for command := range commandDescriptors {
		commands = append(commands, command)
	}

	slice.Sort(commands, func(i, j int) bool {
		return commands[i] < commands[j]
	})

	for _, command := range commands {
		commandDescriptor := commandDescriptors[command]
		if commandDescriptor.commandHelpGenerator != nil {
			helpSections = append(helpSections, commandDescriptor.commandHelpGenerator(config)...)
		}
	}

	return
}

// ConfigParser is a component capable of parsing config into commands
type ConfigParser struct {
	scanner     *ConfigScanner
	inputSource string
}

// NewConfigParser creates a new ConfigParser which will read input from the provided reader
func NewConfigParser(reader io.Reader, inputSource string) *ConfigParser {
	return &ConfigParser{
		scanner:     NewConfigScanner(reader),
		inputSource: inputSource,
	}
}

// Parse returns the next command from the input stream
// eof is set to true if the end of the input stream has been reached
func (parser *ConfigParser) Parse() (command ConfigCommand, eof bool, err error) {
	var token *ConfigToken

	for {
		token, err = parser.scan()
		if err != nil {
			return
		}

		switch token.tokenType {
		case CtkWord:
			command, eof, err = parser.parseCommand(token)
		case CtkShellCommand:
			command = parser.shellCommand(token)
		case CtkTerminator:
			continue
		case CtkEOF:
			eof = true
		case CtkOption:
			err = parser.generateParseError(token, "Unexpected Option \"%v\"", token.value)
		case CtkInvalid:
			err = parser.generateParseError(token, "Syntax Error")
		default:
			err = parser.generateParseError(token, "Unexpected token \"%v\"", token.value)
		}

		break
	}

	if err != nil {
		parser.discardTokensUntilNextCommand()
	}

	return
}

// InputSource returns the text description of the input source
func (parser *ConfigParser) InputSource() string {
	return parser.inputSource
}

func (parser *ConfigParser) scanAndIgnore(ignoreTokens map[ConfigTokenType]bool) (token *ConfigToken, err error) {
	for {
		token, err = parser.scanner.Scan()
		if err != nil {
			return
		} else if _, ignore := ignoreTokens[token.tokenType]; !ignore {
			break
		}
	}

	return
}

func (parser *ConfigParser) scan() (token *ConfigToken, err error) {
	return parser.scanAndIgnore(whiteSpaceTokens)
}

func (parser *ConfigParser) scanIgnoringTerminators() (token *ConfigToken, err error) {
	return parser.scanAndIgnore(whiteSpaceAndTerminatorTokens)
}

func (parser *ConfigParser) scanRaw() (token *ConfigToken, err error) {
	return parser.scanner.Scan()
}

func (parser *ConfigParser) generateParseError(token *ConfigToken, errorMessage string, args ...interface{}) error {
	return generateConfigError(parser.inputSource, token, errorMessage, args...)
}

func generateConfigError(inputSource string, token *ConfigToken, errorMessage string, args ...interface{}) error {
	var buffer bytes.Buffer

	if inputSource != "" {
		buffer.WriteString(inputSource)
		buffer.WriteRune(':')
		buffer.WriteString(fmt.Sprintf("%v:%v ", token.startPos.line, token.startPos.col))
	}

	buffer.WriteString(fmt.Sprintf(errorMessage, args...))

	if token.err != nil {
		buffer.WriteString(": ")
		buffer.WriteString(token.err.Error())
	}

	return errors.New(buffer.String())
}

func (parser *ConfigParser) discardTokensUntilNextCommand() {
	for {
		token, err := parser.scan()

		if err != nil ||
			token.tokenType == CtkTerminator ||
			token.tokenType == CtkEOF {
			return
		}
	}
}

func (parser *ConfigParser) parseCommand(commandToken *ConfigToken) (command ConfigCommand, eof bool, err error) {
	commandDescriptor, ok := commandDescriptors[commandToken.value]
	if !ok {
		err = parser.generateParseError(commandToken, "Invalid command \"%v\"", commandToken.value)
		return
	}

	var tokens []*ConfigToken

	if commandDescriptor.customParser != nil {
		if tokens, err = commandDescriptor.customParser(parser); err != nil {
			return
		}
	} else {
		for i := 0; i < len(commandDescriptor.tokenTypes); i++ {
			var token *ConfigToken
			token, err = parser.scan()
			expectedConfigTokenType := commandDescriptor.tokenTypes[i]

			switch {
			case err != nil:
				return
			case token.err != nil:
				err = parser.generateParseError(token, "Syntax Error when parsing %v command", commandToken.value)
				return
			case token.tokenType == CtkEOF:
				err = parser.generateParseError(token, "Unexpected EOF when parsing %v command", commandToken.value)
				eof = true
				return
			case (token.tokenType & expectedConfigTokenType) == 0:
				err = parser.generateParseError(token, "Invalid argument for %v command: Expected %v but got %v: \"%v\"",
					commandToken.value, ConfigTokenName(expectedConfigTokenType), ConfigTokenName(token.tokenType), token.value)
				return
			}

			tokens = append(tokens, token)
		}
	}

	command, err = commandDescriptor.constructor(parser, commandToken, tokens)

	return
}

func parseVarArgsCommand() commandCustomParser {
	return parseVarArgsParserGenerator(true)
}

func parseVarArgsParserGenerator(ignoreWhitespace bool) commandCustomParser {
	var ignoreTokens map[ConfigTokenType]bool
	if ignoreWhitespace {
		ignoreTokens = whiteSpaceTokens
	} else {
		ignoreTokens = commentTokens
	}

	return func(parser *ConfigParser) (tokens []*ConfigToken, err error) {
	OuterLoop:
		for {
			var token *ConfigToken
			token, err = parser.scanAndIgnore(ignoreTokens)

			switch {
			case err != nil:
				return
			case token.err != nil:
				err = parser.generateParseError(token, "Syntax Error")
				return
			case token.tokenType == CtkEOF:
				break OuterLoop
			case token.tokenType == CtkTerminator:
				break OuterLoop
			}

			tokens = append(tokens, token)
		}

		return
	}
}

func parseDefCommand(parser *ConfigParser) (tokens []*ConfigToken, err error) {
	commandNameToken, err := parser.scanIgnoringTerminators()
	if err != nil {
		return
	} else if err = commandNameToken.err; err != nil {
		return
	} else if commandNameToken.tokenType != CtkWord {
		return tokens, parser.generateParseError(commandNameToken, "Expected function name but found %v", commandNameToken.value)
	} else if !isIdentifier(commandNameToken.value) {
		return tokens, parser.generateParseError(commandNameToken, "Invalid function identifier %v", commandNameToken.value)
	}

	tokens = append(tokens, commandNameToken)

	openingBraceToken, err := parser.scanIgnoringTerminators()
	if err != nil {
		return
	} else if err = openingBraceToken.err; err != nil {
		return
	} else if openingBraceToken.tokenType != CtkWord || openingBraceToken.rawValue != openingBrace {
		return tokens, parser.generateParseError(openingBraceToken, "Expected %v but found %v", openingBrace, openingBraceToken.value)
	}

	tokens = append(tokens, openingBraceToken)

	closingBracesRemaining := 1
	wordsSinceTerminator := 0

	for closingBracesRemaining > 0 {
		var token *ConfigToken
		if token, err = parser.scanRaw(); err != nil {
			return
		} else if err = token.err; err != nil {
			return
		}

		tokens = append(tokens, token)

		switch token.tokenType {
		case CtkEOF:
			return nil, parser.generateParseError(token, "Expected %v but reached EOF", closingBrace)
		case CtkWord:
			if token.value == defCommand && wordsSinceTerminator == 0 {
				closingBracesRemaining++
			} else if token.rawValue == closingBrace {
				closingBracesRemaining--
			}

			wordsSinceTerminator++
		case CtkTerminator:
			wordsSinceTerminator = 0
		}
	}

	return
}

func (parser *ConfigParser) shellCommand(commandToken *ConfigToken) *ShellCommand {
	return &ShellCommand{
		command: commandToken,
	}
}

func setCommandConstructor(parser *ConfigParser, commandToken *ConfigToken, tokens []*ConfigToken) (ConfigCommand, error) {
	return &SetCommand{
		variable: tokens[0],
		value:    tokens[1],
	}, nil
}

func themeCommandConstructor(parser *ConfigParser, commandToken *ConfigToken, tokens []*ConfigToken) (ConfigCommand, error) {
	themeCommand := &ThemeCommand{}

	optionSetters := map[string]func(*ConfigToken){
		"--name":      func(name *ConfigToken) { themeCommand.name = name },
		"--component": func(component *ConfigToken) { themeCommand.component = component },
		"--bgcolor":   func(bgcolor *ConfigToken) { themeCommand.bgcolor = bgcolor },
		"--fgcolor":   func(fgcolor *ConfigToken) { themeCommand.fgcolor = fgcolor },
	}

	for i := 0; i+1 < len(tokens); i += 2 {
		optionToken := tokens[i]
		valueToken := tokens[i+1]

		optionSetter, ok := optionSetters[optionToken.value]
		if !ok {
			return nil, parser.generateParseError(optionToken, "Invalid option for theme command: \"%v\"", optionToken.value)
		}

		optionSetter(valueToken)
	}

	return themeCommand, nil
}

func mapCommandConstructor(parser *ConfigParser, commandToken *ConfigToken, tokens []*ConfigToken) (ConfigCommand, error) {
	return &MapCommand{
		view: tokens[0],
		from: tokens[1],
		to:   tokens[2],
	}, nil
}

func unmapCommandConstructor(parser *ConfigParser, commandToken *ConfigToken, tokens []*ConfigToken) (ConfigCommand, error) {
	return &UnmapCommand{
		view: tokens[0],
		from: tokens[1],
	}, nil
}

func quitCommandConstructor(parser *ConfigParser, commandToken *ConfigToken, tokens []*ConfigToken) (ConfigCommand, error) {
	return &QuitCommand{}, nil
}

func newTabCommandConstructor(parser *ConfigParser, commandToken *ConfigToken, tokens []*ConfigToken) (ConfigCommand, error) {
	return &NewTabCommand{
		tabName: tokens[0],
	}, nil
}

func newRemoveTabCommandConstructor(parser *ConfigParser, commandToken *ConfigToken, tokens []*ConfigToken) (ConfigCommand, error) {
	return &RemoveTabCommand{}, nil
}

func addViewCommandConstructor(parser *ConfigParser, commandToken *ConfigToken, tokens []*ConfigToken) (ConfigCommand, error) {
	if len(tokens) < 1 {
		addViewCommand := commandToken.value
		return nil, parser.generateParseError(commandToken, "Invalid %[1]v command. Usage: %[1]v [VIEW] [ARGS...]", addViewCommand)
	}

	return &AddViewCommand{
		view: tokens[0],
		args: tokens[1:],
	}, nil
}

func splitViewCommandConstructor(parser *ConfigParser, commandToken *ConfigToken, tokens []*ConfigToken) (ConfigCommand, error) {
	splitViewCommand := commandToken.value

	if len(tokens) < 1 {
		return nil, parser.generateParseError(commandToken, "Invalid %[1]v command. Usage: %[1]v [VIEW] [ARGS...]", splitViewCommand)
	}

	var orientation ContainerOrientation

	switch splitViewCommand {
	case splitCommand:
		orientation = CoDynamic
	case hsplitCommand:
		orientation = CoHorizontal
	case vsplitCommand:
		orientation = CoVertical
	default:
		return nil, parser.generateParseError(commandToken, "Unrecognised command: %v", splitViewCommand)
	}

	return &SplitViewCommand{
		orientation: orientation,
		view:        tokens[0],
		args:        tokens[1:],
	}, nil
}

func gitCommandConstructor(parser *ConfigParser, commandToken *ConfigToken, tokens []*ConfigToken) (ConfigCommand, error) {
	return &GitCommand{
		interactive: commandToken.value == gitInteractiveCommand,
		args:        tokens,
	}, nil
}

func helpCommandConstructor(parser *ConfigParser, commandToken *ConfigToken, tokens []*ConfigToken) (ConfigCommand, error) {
	var searchTerm string
	if len(tokens) > 0 {
		searchTerm = tokens[0].value
	}

	return &HelpCommand{
		searchTerm: searchTerm,
	}, nil
}

func defCommandConstructor(parser *ConfigParser, commandToken *ConfigToken, tokens []*ConfigToken) (configCommand ConfigCommand, err error) {
	if len(tokens) < 4 {
		err = parser.generateParseError(commandToken, "Too few tokens (%v) for function definition", len(tokens))
		return
	}

	commandName := tokens[0].value
	var functionBodyBuffer bytes.Buffer

	for i := 2; i < len(tokens)-1; i++ {
		functionBodyBuffer.WriteString(tokens[i].rawValue)
	}

	functionBody := functionBodyBuffer.String()

	configCommand = &DefCommand{
		commandName:  commandName,
		functionBody: functionBody,
	}

	return
}

func undefCommandConstructor(parser *ConfigParser, commandToken *ConfigToken, tokens []*ConfigToken) (configCommand ConfigCommand, err error) {
	return &UndefCommand{
		commandName: tokens[0],
	}, nil
}

func customCommandConstructor(parser *ConfigParser, commandToken *ConfigToken, tokens []*ConfigToken) (configCommand ConfigCommand, err error) {
	var args []string

	for _, token := range tokens {
		args = append(args, token.value)
	}

	return &CustomCommand{
		commandName: commandToken.value,
		args:        args,
	}, nil
}

func evalKeysCommandConstructor(parser *ConfigParser, commandToken *ConfigToken, tokens []*ConfigToken) (configCommand ConfigCommand, err error) {
	if len(tokens) == 0 {
		return nil, parser.generateParseError(commandToken, "No keys specified for %v command", evalkeysCommand)
	}

	var buffer bytes.Buffer

	for _, token := range tokens {
		buffer.WriteString(token.value)
	}

	keys := strings.TrimLeftFunc(buffer.String(), unicode.IsSpace)

	return &EvalKeysCommand{
		keys: keys,
	}, nil
}

func sleepCommandConstructor(parser *ConfigParser, commandToken *ConfigToken, tokens []*ConfigToken) (configCommand ConfigCommand, err error) {
	if len(tokens) < 1 {
		return nil, parser.generateParseError(commandToken, "No sleep time specified")
	}

	sleepToken := tokens[0]
	sleepSeconds, err := strconv.ParseFloat(sleepToken.value, 64)
	if err != nil || sleepSeconds <= 0.0 {
		return nil, parser.generateParseError(sleepToken, "Invalid sleep time: %v. Must be a positive integer", sleepToken.value)
	}

	return &SleepCommand{
		sleepSeconds: sleepSeconds,
	}, nil
}
