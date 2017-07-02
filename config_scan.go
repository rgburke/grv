package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"unicode"
)

type ConfigTokenType int

const (
	CTK_INVALID ConfigTokenType = iota
	CTK_WORD
	CTK_OPTION
	CTK_WHITE_SPACE
	CTK_COMMENT
	CTK_TERMINATOR
	CTK_EOF
)

var configTokenNames = map[ConfigTokenType]string{
	CTK_INVALID:     "Invalid",
	CTK_WORD:        "Word",
	CTK_OPTION:      "Option",
	CTK_WHITE_SPACE: "White Space",
	CTK_COMMENT:     "Comment",
	CTK_TERMINATOR:  "Terminator",
	CTK_EOF:         "EOF",
}

type ConfigScannerPos struct {
	line uint
	col  uint
}

type ConfigToken struct {
	tokenType ConfigTokenType
	value     string
	startPos  ConfigScannerPos
	endPos    ConfigScannerPos
	err       error
}

type ConfigScanner struct {
	reader          *bufio.Reader
	pos             ConfigScannerPos
	lastCharLineEnd bool
	lastLineEndCol  uint
}

func (token *ConfigToken) Equal(other *ConfigToken) bool {
	if other == nil {
		return false
	}

	return token.tokenType == other.tokenType &&
		token.value == other.value &&
		token.startPos == other.startPos &&
		token.endPos == other.endPos &&
		((token.err == nil && other.err == nil) ||
			(token.err != nil && other.err != nil &&
				token.err.Error() == other.err.Error()))
}

func ConfigTokenName(tokenType ConfigTokenType) string {
	return configTokenNames[tokenType]
}

func NewConfigScanner(reader io.Reader) *ConfigScanner {
	return &ConfigScanner{
		reader: bufio.NewReader(reader),
		pos: ConfigScannerPos{
			line: 1,
			col:  0,
		},
	}
}

func (scanner *ConfigScanner) read() (char rune, eof bool, err error) {
	char, _, err = scanner.reader.ReadRune()

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

func (scanner *ConfigScanner) unread() (err error) {
	if err = scanner.reader.UnreadRune(); err != nil {
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

func (scanner *ConfigScanner) Scan() (token *ConfigToken, err error) {
	char, eof, err := scanner.read()
	startPos := scanner.pos

	switch {
	case err != nil:
	case eof:
		token = &ConfigToken{
			tokenType: CTK_EOF,
			endPos:    scanner.pos,
		}
	case char == '\n':
		token = &ConfigToken{
			tokenType: CTK_TERMINATOR,
			value:     string(char),
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
	case char == '-':
		var nextBytes []byte
		nextBytes, err = scanner.reader.Peek(1)

		if err != nil {
			break
		} else if len(nextBytes) == 1 && nextBytes[0] == '-' {
			token, err = scanner.scanWord()

			if token != nil && token.tokenType != CTK_INVALID {
				token.tokenType = CTK_OPTION
				token.value = "-" + token.value
			}

			break
		}

		if err = scanner.unread(); err != nil {
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
			var nextBytes []byte
			nextBytes, err = scanner.reader.Peek(1)

			if err != nil {
				return
			} else if len(nextBytes) == 1 && nextBytes[0] == '\n' {
				escape = true
				continue
			}

			if err = scanner.unread(); err != nil {
				return
			}

			break OuterLoop
		case char == '\n':
			if !escape {
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
			if _, err = buffer.WriteRune(char); err != nil {
				return
			}
		}

		escape = false
	}

	token = &ConfigToken{
		tokenType: CTK_WHITE_SPACE,
		value:     buffer.String(),
		endPos:    scanner.pos,
	}

	return
}

func (scanner *ConfigScanner) scanComment() (token *ConfigToken, err error) {
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

	token = &ConfigToken{
		tokenType: CTK_COMMENT,
		value:     buffer.String(),
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

	token = &ConfigToken{
		tokenType: CTK_WORD,
		value:     buffer.String(),
		endPos:    scanner.pos,
	}

	return
}

func (scanner *ConfigScanner) scanStringWord() (token *ConfigToken, err error) {
	var buffer bytes.Buffer
	var char rune
	var eof bool

	char, eof, err = scanner.read()
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

	if closingQuoteFound {
		var word string
		word, err = scanner.processStringWord(buffer.String())
		if err != nil {
			return
		}

		token = &ConfigToken{
			tokenType: CTK_WORD,
			value:     word,
			endPos:    scanner.pos,
		}
	} else {
		token = &ConfigToken{
			tokenType: CTK_INVALID,
			value:     buffer.String(),
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
