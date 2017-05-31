package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
)

type CommandConstructor func(*Parser, []*Token) (Command, error)

type Command interface {
	Equal(Command) bool
}

type SetCommand struct {
	variable *Token
	value    *Token
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
	name      *Token
	component *Token
	bgcolor   *Token
	fgcolor   *Token
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
	view *Token
	from *Token
	to   *Token
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
	tokenTypes  []TokenType
	constructor CommandConstructor
}

var commandDescriptors = map[string]*CommandDescriptor{
	"set": &CommandDescriptor{
		tokenTypes:  []TokenType{TK_WORD, TK_WORD},
		constructor: setCommandConstructor,
	},
	"theme": &CommandDescriptor{
		tokenTypes:  []TokenType{TK_OPTION, TK_WORD, TK_OPTION, TK_WORD, TK_OPTION, TK_WORD, TK_OPTION, TK_WORD},
		constructor: themeCommandConstructor,
	},
	"map": &CommandDescriptor{
		tokenTypes:  []TokenType{TK_WORD, TK_WORD, TK_WORD},
		constructor: mapCommandConstructor,
	},
	"q": &CommandDescriptor{
		tokenTypes:  []TokenType{},
		constructor: quitCommandConstructor,
	},
}

type Parser struct {
	scanner     *Scanner
	inputSource string
}

func NewParser(reader io.Reader, inputSource string) *Parser {
	return &Parser{
		scanner:     NewScanner(reader),
		inputSource: inputSource,
	}
}

func (parser *Parser) Parse() (command Command, eof bool, err error) {
	var token *Token

	for {
		token, err = parser.scan()
		if err != nil {
			return
		}

		switch token.tokenType {
		case TK_WORD:
			command, eof, err = parser.parseCommand(token)
		case TK_TERMINATOR:
			continue
		case TK_EOF:
			eof = true
		case TK_OPTION:
			err = parser.generateParseError(token, "Unexpected Option \"%v\"", token.value)
		case TK_INVALID:
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

func (parser *Parser) scan() (token *Token, err error) {
	for {
		token, err = parser.scanner.Scan()
		if err != nil {
			return
		}

		if token.tokenType != TK_WHITE_SPACE && token.tokenType != TK_COMMENT {
			break
		}
	}

	return
}

func (parser *Parser) generateParseError(token *Token, errorMessage string, args ...interface{}) error {
	return generateConfigError(parser.inputSource, token, errorMessage, args...)
}

func generateConfigError(inputSource string, token *Token, errorMessage string, args ...interface{}) error {
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
			token.tokenType == TK_TERMINATOR ||
			token.tokenType == TK_EOF {
			return
		}
	}
}

func (parser *Parser) parseCommand(token *Token) (command Command, eof bool, err error) {
	commandDescriptor, ok := commandDescriptors[token.value]
	if !ok {
		err = parser.generateParseError(token, "Invalid command \"%v\"", token.value)
		return
	}

	var tokens []*Token

	for i := 0; i < len(commandDescriptor.tokenTypes); i++ {
		token, err = parser.scan()
		expectedTokenType := commandDescriptor.tokenTypes[i]

		switch {
		case err != nil:
			return
		case token.err != nil:
			err = parser.generateParseError(token, "Syntax Error")
			return
		case token.tokenType == TK_EOF:
			err = parser.generateParseError(token, "Unexpected EOF")
			eof = true
			return
		case token.tokenType != expectedTokenType:
			err = parser.generateParseError(token, "Expected %v but got %v: \"%v\"",
				TokenName(expectedTokenType), TokenName(token.tokenType), token.value)
			return
		}

		tokens = append(tokens, token)
	}

	command, err = commandDescriptor.constructor(parser, tokens)

	return
}

func setCommandConstructor(parser *Parser, tokens []*Token) (Command, error) {
	return &SetCommand{
		variable: tokens[0],
		value:    tokens[1],
	}, nil
}

func themeCommandConstructor(parser *Parser, tokens []*Token) (Command, error) {
	themeCommand := &ThemeCommand{}

	optionSetters := map[string]func(*Token){
		"--name":      func(name *Token) { themeCommand.name = name },
		"--component": func(component *Token) { themeCommand.component = component },
		"--bgcolor":   func(bgcolor *Token) { themeCommand.bgcolor = bgcolor },
		"--fgcolor":   func(fgcolor *Token) { themeCommand.fgcolor = fgcolor },
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

func mapCommandConstructor(parser *Parser, tokens []*Token) (Command, error) {
	return &MapCommand{
		view: tokens[0],
		from: tokens[1],
		to:   tokens[2],
	}, nil
}

func quitCommandConstructor(parser *Parser, tokens []*Token) (Command, error) {
	return &QuitCommand{}, nil
}
