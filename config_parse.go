package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
)

type CommandConstructor func(*Parser, []*ConfigToken) (Command, error)

type Command interface {
	Equal(Command) bool
}

type SetCommand struct {
	variable *ConfigToken
	value    *ConfigToken
}

func (setCommand *SetCommand) Equal(command Command) bool {
	other, ok := command.(*SetCommand)
	if !ok {
		return false
	}

	return ((setCommand.variable != nil && setCommand.variable.Equal(other.variable)) ||
		(setCommand.variable == nil && other.variable == nil)) &&
		((setCommand.value != nil && setCommand.value.Equal(other.value)) ||
			(setCommand.value == nil && other.value == nil))
}

type ThemeCommand struct {
	name      *ConfigToken
	component *ConfigToken
	bgcolor   *ConfigToken
	fgcolor   *ConfigToken
}

func (themeCommand *ThemeCommand) Equal(command Command) bool {
	other, ok := command.(*ThemeCommand)
	if !ok {
		return false
	}

	return ((themeCommand.name != nil && themeCommand.name.Equal(other.name)) ||
		(themeCommand.name == nil && other.name == nil)) &&
		((themeCommand.component != nil && themeCommand.component.Equal(other.component)) ||
			(themeCommand.component == nil && other.component == nil)) &&
		((themeCommand.bgcolor != nil && themeCommand.bgcolor.Equal(other.bgcolor)) ||
			(themeCommand.bgcolor == nil && other.bgcolor == nil)) &&
		((themeCommand.fgcolor != nil && themeCommand.fgcolor.Equal(other.fgcolor)) ||
			(themeCommand.fgcolor == nil && other.fgcolor == nil))
}

type MapCommand struct {
	view *ConfigToken
	from *ConfigToken
	to   *ConfigToken
}

func (mapCommand *MapCommand) Equal(command Command) bool {
	other, ok := command.(*MapCommand)
	if !ok {
		return false
	}

	return ((mapCommand.from != nil && mapCommand.from.Equal(other.from)) ||
		(mapCommand.from == nil && other.from == nil)) &&
		((mapCommand.to != nil && mapCommand.to.Equal(other.to)) ||
			(mapCommand.to == nil && other.to == nil)) &&
		((mapCommand.view != nil && mapCommand.view.Equal(other.view)) ||
			(mapCommand.view == nil && other.view == nil))
}

type QuitCommand struct{}

func (quitCommand *QuitCommand) Equal(command Command) bool {
	_, ok := command.(*QuitCommand)
	return ok
}

type CommandDescriptor struct {
	tokenTypes  []ConfigTokenType
	constructor CommandConstructor
}

var commandDescriptors = map[string]*CommandDescriptor{
	"set": &CommandDescriptor{
		tokenTypes:  []ConfigTokenType{CTK_WORD, CTK_WORD},
		constructor: setCommandConstructor,
	},
	"theme": &CommandDescriptor{
		tokenTypes:  []ConfigTokenType{CTK_OPTION, CTK_WORD, CTK_OPTION, CTK_WORD, CTK_OPTION, CTK_WORD, CTK_OPTION, CTK_WORD},
		constructor: themeCommandConstructor,
	},
	"map": &CommandDescriptor{
		tokenTypes:  []ConfigTokenType{CTK_WORD, CTK_WORD, CTK_WORD},
		constructor: mapCommandConstructor,
	},
	"q": &CommandDescriptor{
		tokenTypes:  []ConfigTokenType{},
		constructor: quitCommandConstructor,
	},
}

type Parser struct {
	scanner     *ConfigScanner
	inputSource string
}

func NewParser(reader io.Reader, inputSource string) *Parser {
	return &Parser{
		scanner:     NewConfigScanner(reader),
		inputSource: inputSource,
	}
}

func (parser *Parser) Parse() (command Command, eof bool, err error) {
	var token *ConfigToken

	for {
		token, err = parser.scan()
		if err != nil {
			return
		}

		switch token.tokenType {
		case CTK_WORD:
			command, eof, err = parser.parseCommand(token)
		case CTK_TERMINATOR:
			continue
		case CTK_EOF:
			eof = true
		case CTK_OPTION:
			err = parser.generateParseError(token, "Unexpected Option \"%v\"", token.value)
		case CTK_INVALID:
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

func (parser *Parser) InputSource() string {
	return parser.inputSource
}

func (parser *Parser) scan() (token *ConfigToken, err error) {
	for {
		token, err = parser.scanner.Scan()
		if err != nil {
			return
		}

		if token.tokenType != CTK_WHITE_SPACE && token.tokenType != CTK_COMMENT {
			break
		}
	}

	return
}

func (parser *Parser) generateParseError(token *ConfigToken, errorMessage string, args ...interface{}) error {
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

func (parser *Parser) discardTokensUntilNextCommand() {
	for {
		token, err := parser.scan()

		if err != nil ||
			token.tokenType == CTK_TERMINATOR ||
			token.tokenType == CTK_EOF {
			return
		}
	}
}

func (parser *Parser) parseCommand(token *ConfigToken) (command Command, eof bool, err error) {
	commandDescriptor, ok := commandDescriptors[token.value]
	if !ok {
		err = parser.generateParseError(token, "Invalid command \"%v\"", token.value)
		return
	}

	var tokens []*ConfigToken

	for i := 0; i < len(commandDescriptor.tokenTypes); i++ {
		token, err = parser.scan()
		expectedConfigTokenType := commandDescriptor.tokenTypes[i]

		switch {
		case err != nil:
			return
		case token.err != nil:
			err = parser.generateParseError(token, "Syntax Error")
			return
		case token.tokenType == CTK_EOF:
			err = parser.generateParseError(token, "Unexpected EOF")
			eof = true
			return
		case token.tokenType != expectedConfigTokenType:
			err = parser.generateParseError(token, "Expected %v but got %v: \"%v\"",
				ConfigTokenName(expectedConfigTokenType), ConfigTokenName(token.tokenType), token.value)
			return
		}

		tokens = append(tokens, token)
	}

	command, err = commandDescriptor.constructor(parser, tokens)

	return
}

func setCommandConstructor(parser *Parser, tokens []*ConfigToken) (Command, error) {
	return &SetCommand{
		variable: tokens[0],
		value:    tokens[1],
	}, nil
}

func themeCommandConstructor(parser *Parser, tokens []*ConfigToken) (Command, error) {
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

		if optionSetter, ok := optionSetters[optionToken.value]; !ok {
			return nil, parser.generateParseError(optionToken, "Invalid option for theme command: \"%v\"", optionToken.value)
		} else {
			optionSetter(valueToken)
		}
	}

	return themeCommand, nil
}

func mapCommandConstructor(parser *Parser, tokens []*ConfigToken) (Command, error) {
	return &MapCommand{
		view: tokens[0],
		from: tokens[1],
		to:   tokens[2],
	}, nil
}

func quitCommandConstructor(parser *Parser, tokens []*ConfigToken) (Command, error) {
	return &QuitCommand{}, nil
}
