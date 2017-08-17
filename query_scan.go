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

type TokenTypeEvaluator func(char rune) (isTokenType bool)

type QueryTokenType int

const (
	QTK_INVALID QueryTokenType = iota

	QTK_WHITE_SPACE
	QTK_EOF

	QTK_IDENTIFIER
	QTK_NUMBER
	QTK_STRING

	QTK_AND
	QTK_OR

	QTK_NOT

	QTK_CMP_EQ
	QTK_CMP_NE
	QTK_CMP_GT
	QTK_CMP_GE
	QTK_CMP_LT
	QTK_CMP_LE

	QTK_CMP_GLOB
	QTK_CMP_REGEXP

	QTK_LPAREN
	QTK_RPAREN
)

type QueryScannerPos struct {
	line uint
	col  uint
}

type QueryToken struct {
	tokenType QueryTokenType
	value     string
	startPos  QueryScannerPos
	endPos    QueryScannerPos
	err       error
}

type QueryScanner struct {
	reader          *bufio.Reader
	pos             QueryScannerPos
	lastCharLineEnd bool
	lastLineEndCol  uint
}

func (token *QueryToken) Equal(other *QueryToken) bool {
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

func (token *QueryToken) Value() string {
	if token.tokenType == QTK_EOF {
		return "EOF"
	}

	return token.value
}

func NewQueryScanner(reader io.Reader) *QueryScanner {
	return &QueryScanner{
		reader: bufio.NewReader(reader),
		pos: QueryScannerPos{
			line: 1,
			col:  0,
		},
	}
}

func (scanner *QueryScanner) read() (char rune, eof bool, err error) {
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

func (scanner *QueryScanner) unread() (err error) {
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

func (scanner *QueryScanner) Scan() (token *QueryToken, err error) {
	char, eof, err := scanner.read()
	startPos := scanner.pos

	switch {
	case err != nil:
	case eof:
		token = &QueryToken{
			tokenType: QTK_EOF,
			endPos:    scanner.pos,
		}
	case unicode.IsSpace(char):
		if err = scanner.unread(); err != nil {
			break
		}

		token, err = scanner.scanWhiteSpace()
	case unicode.IsLetter(char):
		if err = scanner.unread(); err != nil {
			break
		}

		token, err = scanner.scanIdentifier()

		if err != nil || token.err != nil {
			break
		}

		switch strings.ToUpper(token.value) {
		case "AND":
			token.tokenType = QTK_AND
		case "OR":
			token.tokenType = QTK_OR
		case "NOT":
			token.tokenType = QTK_NOT
		case "GLOB":
			token.tokenType = QTK_CMP_GLOB
		case "REGEXP":
			token.tokenType = QTK_CMP_REGEXP
		}
	case char == '"':
		if err = scanner.unread(); err != nil {
			break
		}

		token, err = scanner.scanString()
	case char == '-' || char == '.' || unicode.IsNumber(char):
		if err = scanner.unread(); err != nil {
			break
		}

		token, err = scanner.scanNumber()
	case char == '=':
		token = &QueryToken{
			tokenType: QTK_CMP_EQ,
			value:     "=",
			endPos:    scanner.pos,
		}
	case char == '!':
		char, eof, err = scanner.read()

		if err != nil {
			break
		}

		if char == '=' {
			token = &QueryToken{
				tokenType: QTK_CMP_NE,
				value:     "!=",
				endPos:    scanner.pos,
			}
		} else {
			token = &QueryToken{
				tokenType: QTK_INVALID,
				value:     fmt.Sprintf("!%c", char),
				endPos:    scanner.pos,
				err:       errors.New("Expected '=' character after '!'"),
			}
		}
	case char == '>' || char == '<':
		nextChar, eof, err := scanner.read()

		if err != nil {
			break
		}

		var tokenType QueryTokenType
		var tokenValue string

		if nextChar == '=' {
			if char == '>' {
				tokenType = QTK_CMP_GE
			} else {
				tokenType = QTK_CMP_LE
			}

			tokenValue = fmt.Sprintf("%c=", char)
		} else {
			if char == '>' {
				tokenType = QTK_CMP_GT
			} else {
				tokenType = QTK_CMP_LT
			}

			tokenValue = fmt.Sprintf("%c", char)

			if !eof {
				if err = scanner.unread(); err != nil {
					break
				}
			}
		}

		token = &QueryToken{
			tokenType: tokenType,
			value:     tokenValue,
			endPos:    scanner.pos,
		}
	case char == '(':
		token = &QueryToken{
			tokenType: QTK_LPAREN,
			value:     "(",
			endPos:    scanner.pos,
		}
	case char == ')':
		token = &QueryToken{
			tokenType: QTK_RPAREN,
			value:     ")",
			endPos:    scanner.pos,
		}
	default:
		token = &QueryToken{
			tokenType: QTK_INVALID,
			value:     fmt.Sprintf("%c", char),
			endPos:    scanner.pos,
			err:       fmt.Errorf("Unexpected character %c", char),
		}
	}

	if token != nil {
		token.startPos = startPos
	}

	return
}

func (scanner *QueryScanner) scanToken(tokenType QueryTokenType, evaluator TokenTypeEvaluator) (token *QueryToken, err error) {
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
		case !evaluator(char):
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

	token = &QueryToken{
		tokenType: tokenType,
		value:     buffer.String(),
		endPos:    scanner.pos,
	}

	return
}

func isLetterOrNumber(char rune) bool {
	return unicode.IsLetter(char) || unicode.IsNumber(char)
}

func (scanner *QueryScanner) scanWhiteSpace() (token *QueryToken, err error) {
	return scanner.scanToken(QTK_WHITE_SPACE, unicode.IsSpace)
}

func (scanner *QueryScanner) scanIdentifier() (token *QueryToken, err error) {
	return scanner.scanToken(QTK_IDENTIFIER, isLetterOrNumber)
}

func (scanner *QueryScanner) scanNumber() (token *QueryToken, err error) {
	var buffer bytes.Buffer
	var char rune
	var eof bool
	dotSeen := false

OuterLoop:
	for {
		char, eof, err = scanner.read()

		switch {
		case err != nil:
			return
		case eof:
			break OuterLoop
		case char == '-':
			offset := buffer.Len()

			if _, err = buffer.WriteRune(char); err != nil {
				return
			}

			if offset != 0 {
				token = &QueryToken{
					tokenType: QTK_INVALID,
					value:     buffer.String(),
					endPos:    scanner.pos,
					err:       errors.New("Unexpected '-' character in number"),
				}

				return
			}
		case char == '.':
			if _, err = buffer.WriteRune(char); err != nil {
				return
			}

			if dotSeen {
				token = &QueryToken{
					tokenType: QTK_INVALID,
					value:     buffer.String(),
					endPos:    scanner.pos,
					err:       errors.New("Unexpected '.' character in number"),
				}

				return
			}

			dotSeen = true
		case !unicode.IsNumber(char):
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

	token = &QueryToken{
		tokenType: QTK_NUMBER,
		value:     buffer.String(),
		endPos:    scanner.pos,
	}

	return
}

func (scanner *QueryScanner) scanString() (token *QueryToken, err error) {
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
		word, err = scanner.processString(buffer.String())
		if err != nil {
			return
		}

		token = &QueryToken{
			tokenType: QTK_STRING,
			value:     word,
			endPos:    scanner.pos,
		}
	} else {
		token = &QueryToken{
			tokenType: QTK_INVALID,
			value:     buffer.String(),
			endPos:    scanner.pos,
			err:       errors.New("Unterminated string"),
		}
	}

	return
}

func (scanner *QueryScanner) processString(str string) (string, error) {
	var buffer bytes.Buffer
	chars := []rune(str)

	if len(chars) < 2 || chars[0] != '"' || chars[len(chars)-1] != '"' {
		return str, fmt.Errorf("Invalid string: %v", str)
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
