package main

import (
	"errors"
	"strings"
	"testing"
)

func TestScanSingleToken(t *testing.T) {
	var singleTokenTests = []struct {
		input         string
		expectedToken Token
	}{
		{
			input: "-!\"word1世界",
			expectedToken: Token{
				tokenType: TK_WORD,
				value:     "-!\"word1世界",
				startPos: ScannerPos{
					line: 1,
					col:  1,
				},
				endPos: ScannerPos{
					line: 1,
					col:  10,
				},
			},
		},
		{
			input: "\"word \\\"with\\\" spaces\"",
			expectedToken: Token{
				tokenType: TK_WORD,
				value:     "\"word \\\"with\\\" spaces\"",
				startPos: ScannerPos{
					line: 1,
					col:  1,
				},
				endPos: ScannerPos{
					line: 1,
					col:  22,
				},
			},
		},
		{
			input: " \t\r\v\f",
			expectedToken: Token{
				tokenType: TK_WHITE_SPACE,
				value:     " \t\r\v\f",
				startPos: ScannerPos{
					line: 1,
					col:  1,
				},
				endPos: ScannerPos{
					line: 1,
					col:  5,
				},
			},
		},
		{
			input: "--option",
			expectedToken: Token{
				tokenType: TK_OPTION,
				value:     "--option",
				startPos: ScannerPos{
					line: 1,
					col:  1,
				},
				endPos: ScannerPos{
					line: 1,
					col:  8,
				},
			},
		},
		{
			input: "\n",
			expectedToken: Token{
				tokenType: TK_TERMINATOR,
				value:     "\n",
				startPos: ScannerPos{
					line: 1,
					col:  1,
				},
				endPos: ScannerPos{
					line: 1,
					col:  1,
				},
			},
		},
		{
			input: "",
			expectedToken: Token{
				tokenType: TK_EOF,
				startPos: ScannerPos{
					line: 1,
					col:  1,
				},
				endPos: ScannerPos{
					line: 1,
					col:  1,
				},
			},
		},
		{
			input: "\"Unterminated string",
			expectedToken: Token{
				tokenType: TK_INVALID,
				value:     "\"Unterminated string",
				startPos: ScannerPos{
					line: 1,
					col:  1,
				},
				endPos: ScannerPos{
					line: 1,
					col:  20,
				},
				err: errors.New("Unterminated string"),
			},
		},
	}

	for _, singleTokenTest := range singleTokenTests {
		scanner := NewScanner(strings.NewReader(singleTokenTest.input))
		token, err := scanner.Scan()

		if err != nil {
			t.Errorf("Scan failed with error %v", err)
		} else if !token.Equal(&singleTokenTest.expectedToken) {
			t.Errorf("Token does not match expected value. Expected %v, Actual %v", singleTokenTest.expectedToken, *token)
		}
	}
}

func TestScanMultipleTokens(t *testing.T) {
	var multiTokenTests = []struct {
		input          string
		expectedTokens []Token
	}{
		{
			input: "theme --create \"my theme\"\n",
			expectedTokens: []Token{
				Token{
					tokenType: TK_WORD,
					value:     "theme",
					startPos: ScannerPos{
						line: 1,
						col:  1,
					},
					endPos: ScannerPos{
						line: 1,
						col:  5,
					},
				},
				Token{
					tokenType: TK_WHITE_SPACE,
					value:     " ",
					startPos: ScannerPos{
						line: 1,
						col:  6,
					},
					endPos: ScannerPos{
						line: 1,
						col:  6,
					},
				},
				Token{
					tokenType: TK_OPTION,
					value:     "--create",
					startPos: ScannerPos{
						line: 1,
						col:  7,
					},
					endPos: ScannerPos{
						line: 1,
						col:  14,
					},
				},
				Token{
					tokenType: TK_WHITE_SPACE,
					value:     " ",
					startPos: ScannerPos{
						line: 1,
						col:  15,
					},
					endPos: ScannerPos{
						line: 1,
						col:  15,
					},
				},
				Token{
					tokenType: TK_WORD,
					value:     "\"my theme\"",
					startPos: ScannerPos{
						line: 1,
						col:  16,
					},
					endPos: ScannerPos{
						line: 1,
						col:  25,
					},
				},
				Token{
					tokenType: TK_TERMINATOR,
					value:     "\n",
					startPos: ScannerPos{
						line: 1,
						col:  26,
					},
					endPos: ScannerPos{
						line: 1,
						col:  26,
					},
				},
				Token{
					tokenType: TK_EOF,
					startPos: ScannerPos{
						line: 1,
						col:  26,
					},
					endPos: ScannerPos{
						line: 1,
						col:  26,
					},
				},
			},
		},
		{
			input: "set theme mytheme\nset\tCommitView.dateformat \"%yyyy-mm-dd HH:MM\"\n",
			expectedTokens: []Token{
				Token{
					tokenType: TK_WORD,
					value:     "set",
					startPos: ScannerPos{
						line: 1,
						col:  1,
					},
					endPos: ScannerPos{
						line: 1,
						col:  3,
					},
				},
				Token{
					tokenType: TK_WHITE_SPACE,
					value:     " ",
					startPos: ScannerPos{
						line: 1,
						col:  4,
					},
					endPos: ScannerPos{
						line: 1,
						col:  4,
					},
				},
				Token{
					tokenType: TK_WORD,
					value:     "theme",
					startPos: ScannerPos{
						line: 1,
						col:  5,
					},
					endPos: ScannerPos{
						line: 1,
						col:  9,
					},
				},
				Token{
					tokenType: TK_WHITE_SPACE,
					value:     " ",
					startPos: ScannerPos{
						line: 1,
						col:  10,
					},
					endPos: ScannerPos{
						line: 1,
						col:  10,
					},
				},
				Token{
					tokenType: TK_WORD,
					value:     "mytheme",
					startPos: ScannerPos{
						line: 1,
						col:  11,
					},
					endPos: ScannerPos{
						line: 1,
						col:  17,
					},
				},
				Token{
					tokenType: TK_TERMINATOR,
					value:     "\n",
					startPos: ScannerPos{
						line: 1,
						col:  18,
					},
					endPos: ScannerPos{
						line: 1,
						col:  18,
					},
				},
				Token{
					tokenType: TK_WORD,
					value:     "set",
					startPos: ScannerPos{
						line: 2,
						col:  1,
					},
					endPos: ScannerPos{
						line: 2,
						col:  3,
					},
				},
				Token{
					tokenType: TK_WHITE_SPACE,
					value:     "\t",
					startPos: ScannerPos{
						line: 2,
						col:  4,
					},
					endPos: ScannerPos{
						line: 2,
						col:  4,
					},
				},
				Token{
					tokenType: TK_WORD,
					value:     "CommitView.dateformat",
					startPos: ScannerPos{
						line: 2,
						col:  5,
					},
					endPos: ScannerPos{
						line: 2,
						col:  25,
					},
				},
				Token{
					tokenType: TK_WHITE_SPACE,
					value:     " ",
					startPos: ScannerPos{
						line: 2,
						col:  26,
					},
					endPos: ScannerPos{
						line: 2,
						col:  26,
					},
				},
				Token{
					tokenType: TK_WORD,
					value:     "\"%yyyy-mm-dd HH:MM\"",
					startPos: ScannerPos{
						line: 2,
						col:  27,
					},
					endPos: ScannerPos{
						line: 2,
						col:  45,
					},
				},
				Token{
					tokenType: TK_TERMINATOR,
					value:     "\n",
					startPos: ScannerPos{
						line: 2,
						col:  46,
					},
					endPos: ScannerPos{
						line: 2,
						col:  46,
					},
				},
				Token{
					tokenType: TK_EOF,
					startPos: ScannerPos{
						line: 2,
						col:  46,
					},
					endPos: ScannerPos{
						line: 2,
						col:  46,
					},
				},
			},
		},
		{
			input: "theme --create \\\n\tmytheme\n",
			expectedTokens: []Token{
				Token{
					tokenType: TK_WORD,
					value:     "theme",
					startPos: ScannerPos{
						line: 1,
						col:  1,
					},
					endPos: ScannerPos{
						line: 1,
						col:  5,
					},
				},
				Token{
					tokenType: TK_WHITE_SPACE,
					value:     " ",
					startPos: ScannerPos{
						line: 1,
						col:  6,
					},
					endPos: ScannerPos{
						line: 1,
						col:  6,
					},
				},
				Token{
					tokenType: TK_OPTION,
					value:     "--create",
					startPos: ScannerPos{
						line: 1,
						col:  7,
					},
					endPos: ScannerPos{
						line: 1,
						col:  14,
					},
				},
				Token{
					tokenType: TK_WHITE_SPACE,
					value:     " \t",
					startPos: ScannerPos{
						line: 1,
						col:  15,
					},
					endPos: ScannerPos{
						line: 2,
						col:  1,
					},
				},
				Token{
					tokenType: TK_WORD,
					value:     "mytheme",
					startPos: ScannerPos{
						line: 2,
						col:  2,
					},
					endPos: ScannerPos{
						line: 2,
						col:  8,
					},
				},
				Token{
					tokenType: TK_TERMINATOR,
					value:     "\n",
					startPos: ScannerPos{
						line: 2,
						col:  9,
					},
					endPos: ScannerPos{
						line: 2,
						col:  9,
					},
				},
				Token{
					tokenType: TK_EOF,
					startPos: ScannerPos{
						line: 2,
						col:  9,
					},
					endPos: ScannerPos{
						line: 2,
						col:  9,
					},
				},
			},
		},
		{
			input: "theme --create \"my theme\nset theme mytheme\n",
			expectedTokens: []Token{
				Token{
					tokenType: TK_WORD,
					value:     "theme",
					startPos: ScannerPos{
						line: 1,
						col:  1,
					},
					endPos: ScannerPos{
						line: 1,
						col:  5,
					},
				},
				Token{
					tokenType: TK_WHITE_SPACE,
					value:     " ",
					startPos: ScannerPos{
						line: 1,
						col:  6,
					},
					endPos: ScannerPos{
						line: 1,
						col:  6,
					},
				},
				Token{
					tokenType: TK_OPTION,
					value:     "--create",
					startPos: ScannerPos{
						line: 1,
						col:  7,
					},
					endPos: ScannerPos{
						line: 1,
						col:  14,
					},
				},
				Token{
					tokenType: TK_WHITE_SPACE,
					value:     " ",
					startPos: ScannerPos{
						line: 1,
						col:  15,
					},
					endPos: ScannerPos{
						line: 1,
						col:  15,
					},
				},
				Token{
					tokenType: TK_INVALID,
					value:     "\"my theme\nset theme mytheme\n",
					startPos: ScannerPos{
						line: 1,
						col:  16,
					},
					endPos: ScannerPos{
						line: 2,
						col:  18,
					},
					err: errors.New("Unterminated string"),
				},
				Token{
					tokenType: TK_EOF,
					startPos: ScannerPos{
						line: 2,
						col:  18,
					},
					endPos: ScannerPos{
						line: 2,
						col:  18,
					},
				},
			},
		},
	}

	for _, multiTokenTest := range multiTokenTests {
		scanner := NewScanner(strings.NewReader(multiTokenTest.input))

		for _, expectedToken := range multiTokenTest.expectedTokens {
			token, err := scanner.Scan()

			if err != nil {
				t.Errorf("Scan failed with error %v", err)
			} else if !token.Equal(&expectedToken) {
				t.Errorf("Token does not match expected value. Expected %v, Actual %v", expectedToken, *token)
			}
		}
	}
}
