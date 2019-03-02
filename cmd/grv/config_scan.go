package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode"
)

// ConfigTokenType is an enum of token types the config scanner can produce
type ConfigTokenType uint

// Token types produced by the config scanner
const (
	CtkInvalid ConfigTokenType = 1 << iota
	CtkWord
	CtkOption
	CtkWhiteSpace
	CtkComment
	CtkShellCommand
	CtkTerminator
	CtkEOF

	CtkCount
)

const (
	configScannerLookAhead = 2
)

var configTokenNames = map[ConfigTokenType]string{
	CtkInvalid:      "Invalid",
	CtkWord:         "Word",
	CtkOption:       "Option",
	CtkWhiteSpace:   "White Space",
	CtkComment:      "Comment",
	CtkShellCommand: "Shell Command",
	CtkTerminator:   "Terminator",
	CtkEOF:          "EOF",
}

type configReader struct {
	reader      *bufio.Reader
	buffer      []rune
	bufferIndex int
	unreadCount int
}

func newConfigReader(reader io.Reader, bufferSize int) *configReader {
	return &configReader{
		reader: bufio.NewReader(reader),
		buffer: make([]rune, bufferSize, bufferSize),
	}
}

func (configReader *configReader) readRune() (char rune, err error) {
	if configReader.unreadCount > 0 {
		char = configReader.buffer[configReader.bufferIndex]
		configReader.bufferIndex = (configReader.bufferIndex + 1) % len(configReader.buffer)
		configReader.unreadCount--
	} else if char, _, err = configReader.reader.ReadRune(); err == nil {
		configReader.buffer[configReader.bufferIndex] = char
		configReader.bufferIndex = (configReader.bufferIndex + 1) % len(configReader.buffer)
	}

	return
}

func (configReader *configReader) unreadRune() (err error) {
	if configReader.unreadCount+1 > len(configReader.buffer) {
		return fmt.Errorf("Only %v consecutive unreads can be performed", len(configReader.buffer))
	}

	configReader.unreadCount++

	if configReader.bufferIndex == 0 {
		configReader.bufferIndex = len(configReader.buffer) - 1
	} else {
		configReader.bufferIndex--
	}

	return
}

// ConfigScannerPos represents a position in the config scanner input stream
type ConfigScannerPos struct {
	line uint
	col  uint
}

// ConfigToken is a config token parsed from an input stream
// It contains position, error and value data
type ConfigToken struct {
	tokenType ConfigTokenType
	value     string
	rawValue  string
	startPos  ConfigScannerPos
	endPos    ConfigScannerPos
	err       error
}

// ConfigScanner scans an input stream and generates a stream of config tokens
type ConfigScanner struct {
	reader          *configReader
	pos             ConfigScannerPos
	lastCharLineEnd bool
	lastLineEndCol  uint
}

// Equal returns true if the other token is equal
func (token *ConfigToken) Equal(other *ConfigToken) bool {
	if other == nil {
		return false
	}

	return token.tokenType == other.tokenType &&
		token.value == other.value &&
		token.rawValue == other.rawValue &&
		token.startPos == other.startPos &&
		token.endPos == other.endPos &&
		((token.err == nil && other.err == nil) ||
			(token.err != nil && other.err != nil &&
				token.err.Error() == other.err.Error()))
}

// ConfigTokenName maps token types to human readable names
func ConfigTokenName(tokenType ConfigTokenType) string {
	tokens := []string{}

	for i := CtkInvalid; i < CtkCount; i <<= 1 {
		if (i & tokenType) != 0 {
			tokens = append(tokens, configTokenNames[i])
		}
	}

	return strings.Join(tokens, " or ")
}

// NewConfigScanner creates a new scanner which uses the provided reader
func NewConfigScanner(reader io.Reader) *ConfigScanner {
	return &ConfigScanner{
		reader: newConfigReader(reader, configScannerLookAhead),
		pos: ConfigScannerPos{
			line: 1,
			col:  0,
		},
	}
}

func (scanner *ConfigScanner) read() (char rune, eof bool, err error) {
	char, err = scanner.reader.readRune()

	if err == io.EOF {
		eof = true
		err = nil

		if scanner.pos.col == 0 {
			scanner.pos.col = 1
		}
	} else if err == nil {
		if scanner.lastCharLineEnd {
			scanner.lastLineEndCol = scanner.pos.col
			scanner.pos.line++
			scanner.pos.col = 1
		} else {
			scanner.pos.col++
		}

		scanner.lastCharLineEnd = (char == '\n')
	}

	return
}

func (scanner *ConfigScanner) unreadRunes(runeNum int) (err error) {
	for i := 0; i < runeNum; i++ {
		if err = scanner.unread(); err != nil {
			return
		}
	}

	return
}

func (scanner *ConfigScanner) unread() (err error) {
	if err = scanner.reader.unreadRune(); err != nil {
		return
	}

	if scanner.pos.line > 1 && scanner.pos.col == 1 {
		scanner.pos.line--
		scanner.pos.col = scanner.lastLineEndCol
		scanner.lastCharLineEnd = true
	} else {
		scanner.pos.col--
		scanner.lastCharLineEnd = false
	}

	return
}

// Scan returns the next token from the input stream
func (scanner *ConfigScanner) Scan() (token *ConfigToken, err error) {
	char, eof, err := scanner.read()
	startPos := scanner.pos

	switch {
	case err != nil:
	case eof:
		token = &ConfigToken{
			tokenType: CtkEOF,
			endPos:    scanner.pos,
		}
	case char == '\n':
		token = &ConfigToken{
			tokenType: CtkTerminator,
			value:     string(char),
			rawValue:  string(char),
			endPos:    scanner.pos,
		}
	case unicode.IsSpace(char):
		if err = scanner.unread(); err != nil {
			break
		}

		token, err = scanner.scanWhiteSpace()
	case char == '#':
		if err = scanner.unread(); err != nil {
			break
		}

		token, err = scanner.scanComment()
	case char == '!' || char == '@':
		if err = scanner.unread(); err != nil {
			break
		}

		token, err = scanner.scanShellCommand()
	case char == '-':
		unreadCount := 2
		var nextChar rune
		nextChar, eof, err = scanner.read()

		if err != nil {
			break
		} else if eof {
			unreadCount = 1
		} else if nextChar == '-' {
			token, err = scanner.scanWord()

			if token != nil && token.tokenType != CtkInvalid {
				token.tokenType = CtkOption
				token.value = "--" + token.value
				token.rawValue = "--" + token.rawValue
			}

			break
		}

		if err = scanner.unreadRunes(unreadCount); err != nil {
			break
		}

		token, err = scanner.scanWord()
	case char == '"':
		if err = scanner.unread(); err != nil {
			break
		}

		token, err = scanner.scanStringWord()
	default:
		if err = scanner.unread(); err != nil {
			break
		}

		token, err = scanner.scanWord()
	}

	if token != nil {
		token.startPos = startPos
	}

	return
}

func (scanner *ConfigScanner) scanWhiteSpace() (token *ConfigToken, err error) {
	var buffer bytes.Buffer
	var rawBuffer bytes.Buffer
	var char rune
	var eof bool

	escape := false

OuterLoop:
	for {
		char, eof, err = scanner.read()

		switch {
		case err != nil:
			return
		case eof:
			break OuterLoop
		case char == '\\':
			unreadCount := 2
			var nextChar rune
			nextChar, eof, err = scanner.read()

			if err != nil {
				return
			} else if eof {
				unreadCount = 1
			} else if !eof && nextChar == '\n' {
				escape = true
				if err = scanner.unread(); err != nil {
					return
				}

				if _, err = rawBuffer.WriteRune(char); err != nil {
					return
				}

				continue
			}

			if err = scanner.unreadRunes(unreadCount); err != nil {
				return
			}

			break OuterLoop
		case char == '\n':
			if escape {
				if _, err = rawBuffer.WriteRune(char); err != nil {
					return
				}
			} else {
				if err = scanner.unread(); err != nil {
					return
				}

				break OuterLoop
			}
		case !unicode.IsSpace(char):
			if err = scanner.unread(); err != nil {
				return
			}

			break OuterLoop
		default:
			if _, err = rawBuffer.WriteRune(char); err != nil {
				return
			}

			if _, err = buffer.WriteRune(char); err != nil {
				return
			}
		}

		escape = false
	}

	token = &ConfigToken{
		tokenType: CtkWhiteSpace,
		value:     buffer.String(),
		rawValue:  rawBuffer.String(),
		endPos:    scanner.pos,
	}

	return
}

func (scanner *ConfigScanner) scanComment() (token *ConfigToken, err error) {
	return scanner.scanToEndOfLine(CtkComment)
}

func (scanner *ConfigScanner) scanShellCommand() (token *ConfigToken, err error) {
	return scanner.scanToEndOfLine(CtkShellCommand)
}

func (scanner *ConfigScanner) scanToEndOfLine(tokenType ConfigTokenType) (token *ConfigToken, err error) {
	var buffer bytes.Buffer
	var char rune
	var eof bool

OuterLoop:
	for {
		char, eof, err = scanner.read()

		switch {
		case err != nil:
			return
		case eof:
			break OuterLoop
		case char == '\n':
			if err = scanner.unread(); err != nil {
				return
			}

			break OuterLoop
		default:
			if _, err = buffer.WriteRune(char); err != nil {
				return
			}
		}
	}

	value := buffer.String()

	token = &ConfigToken{
		tokenType: tokenType,
		value:     value,
		rawValue:  value,
		endPos:    scanner.pos,
	}

	return
}

func (scanner *ConfigScanner) scanWord() (token *ConfigToken, err error) {
	var buffer bytes.Buffer
	var char rune
	var eof bool

OuterLoop:
	for {
		char, eof, err = scanner.read()

		switch {
		case err != nil:
			return
		case eof:
			break OuterLoop
		case unicode.IsSpace(char):
			if err = scanner.unread(); err != nil {
				return
			}

			break OuterLoop
		default:
			if _, err = buffer.WriteRune(char); err != nil {
				return
			}
		}
	}

	value := buffer.String()

	token = &ConfigToken{
		tokenType: CtkWord,
		value:     value,
		rawValue:  value,
		endPos:    scanner.pos,
	}

	return
}

func (scanner *ConfigScanner) scanStringWord() (token *ConfigToken, err error) {
	var buffer bytes.Buffer
	var char rune
	var eof bool

	char, eof, err = scanner.read()
	if err != nil || eof {
		return
	}

	if _, err = buffer.WriteRune(char); err != nil {
		return
	}

	closingQuoteFound := false
	escape := false

OuterLoop:
	for {
		char, eof, err = scanner.read()

		switch {
		case err != nil:
			return
		case eof:
			break OuterLoop
		case char == '\\':
			if _, err = buffer.WriteRune(char); err != nil {
				return
			}

			if !escape {
				escape = true
				continue
			}
		case char == '"':
			if _, err = buffer.WriteRune(char); err != nil {
				return
			}

			if !escape {
				closingQuoteFound = true
				break OuterLoop
			}
		default:
			if _, err = buffer.WriteRune(char); err != nil {
				return
			}
		}

		escape = false
	}

	rawValue := buffer.String()

	if closingQuoteFound {
		var value string
		value, err = scanner.processStringWord(rawValue)
		if err != nil {
			return
		}

		token = &ConfigToken{
			tokenType: CtkWord,
			value:     value,
			rawValue:  rawValue,
			endPos:    scanner.pos,
		}
	} else {
		token = &ConfigToken{
			tokenType: CtkInvalid,
			value:     rawValue,
			rawValue:  rawValue,
			endPos:    scanner.pos,
			err:       errors.New("Unterminated string"),
		}
	}

	return
}

func (scanner *ConfigScanner) processStringWord(str string) (string, error) {
	var buffer bytes.Buffer
	chars := []rune(str)

	if len(chars) < 2 || chars[0] != '"' || chars[len(chars)-1] != '"' {
		return str, fmt.Errorf("Invalid string word: %v", str)
	}

	chars = chars[1 : len(chars)-1]
	escape := false

	for _, char := range chars {
		switch {
		case escape:
			switch char {
			case 'n':
				buffer.WriteRune('\n')
			case 't':
				buffer.WriteRune('\t')
			default:
				buffer.WriteRune(char)
			}

			escape = false
		case char == '\\':
			escape = true
		default:
			buffer.WriteRune(char)
		}
	}

	return buffer.String(), nil
}
