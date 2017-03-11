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
	variable string
	value    string
}

func (setCommand *SetCommand) Equal(command Command) bool {
	other, ok := command.(*SetCommand)
	if !ok {
		return false
	}

	return setCommand.variable == other.variable &&
		setCommand.value == other.value
}

type ThemeCommand struct {
	name      string
	component string
	bgcolor   string
	fgcolour  string
}

func (themeCommand *ThemeCommand) Equal(command Command) bool {
	other, ok := command.(*ThemeCommand)
	if !ok {
		return false
	}

	return themeCommand.name == other.name &&
		themeCommand.component == other.component &&
		themeCommand.bgcolor == other.bgcolor &&
		themeCommand.fgcolour == other.fgcolour
}

type CommandDescriptor struct {
	tokenTypes  []TokenType
	constructor CommandConstructor
}

var commandDescriptors = map[string]CommandDescriptor{
	"set": CommandDescriptor{
		tokenTypes:  []TokenType{TK_WORD, TK_WORD},
		constructor: setCommandConstructor,
	},
	"theme": CommandDescriptor{
		tokenTypes:  []TokenType{TK_OPTION, TK_WORD, TK_OPTION, TK_WORD, TK_OPTION, TK_WORD, TK_OPTION, TK_WORD},
		constructor: themeCommandConstructor,
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

func (parser *Parser) scan() (token *Token, err error) {
	token, err = parser.scanner.Scan()
	if err != nil {
		return
	}

	if token.tokenType == TK_WHITE_SPACE {
		token, err = parser.scan()
	}

	return
}

func (parser *Parser) generateParseError(token *Token, errorMessage string, args ...interface{}) error {
	var buffer bytes.Buffer

	if parser.inputSource != "" {
		buffer.WriteString(parser.inputSource)
		buffer.WriteRune(':')
	}

	buffer.WriteString(fmt.Sprintf("%v:%v ", token.startPos.line, token.startPos.col))
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
		variable: tokens[0].value,
		value:    tokens[1].value,
	}, nil
}

func themeCommandConstructor(parser *Parser, tokens []*Token) (Command, error) {
	themeCommand := &ThemeCommand{}

	optionSetters := map[string]func(string){
		"--name":      func(name string) { themeCommand.name = name },
		"--component": func(component string) { themeCommand.component = component },
		"--bgcolor":   func(bgcolor string) { themeCommand.bgcolor = bgcolor },
		"--fgcolor":   func(fgcolour string) { themeCommand.fgcolour = fgcolour },
	}

	for i := 0; i+1 < len(tokens); i += 2 {
		optionToken := tokens[i]
		valueToken := tokens[i+1]

		if optionSetter, ok := optionSetters[optionToken.value]; !ok {
			return nil, parser.generateParseError(optionToken, "Invalid option for theme command: \"%v\"", optionToken.value)
		} else {
			optionSetter(valueToken.value)
		}
	}

	return themeCommand, nil
}
