package main

import (
	"errors"
	"strings"
	"testing"
)

func TestScanSingleConfigToken(t *testing.T) {
	var singleTokenTests = []struct {
		input         string
		expectedToken ConfigToken
	}{
		{
			input: "-!\"word1世界",
			expectedToken: ConfigToken{
				tokenType: CTK_WORD,
				value:     "-!\"word1世界",
				startPos: ConfigScannerPos{
					line: 1,
					col:  1,
				},
				endPos: ConfigScannerPos{
					line: 1,
					col:  10,
				},
			},
		},
		{
			input: "\"word \\t\\\"with\\\"\\n spaces\"",
			expectedToken: ConfigToken{
				tokenType: CTK_WORD,
				value:     "word \t\"with\"\n spaces",
				startPos: ConfigScannerPos{
					line: 1,
					col:  1,
				},
				endPos: ConfigScannerPos{
					line: 1,
					col:  26,
				},
			},
		},
		{
			input: " \t\r\v\f",
			expectedToken: ConfigToken{
				tokenType: CTK_WHITE_SPACE,
				value:     " \t\r\v\f",
				startPos: ConfigScannerPos{
					line: 1,
					col:  1,
				},
				endPos: ConfigScannerPos{
					line: 1,
					col:  5,
				},
			},
		},
		{
			input: "# Comment",
			expectedToken: ConfigToken{
				tokenType: CTK_COMMENT,
				value:     "# Comment",
				startPos: ConfigScannerPos{
					line: 1,
					col:  1,
				},
				endPos: ConfigScannerPos{
					line: 1,
					col:  9,
				},
			},
		},
		{
			input: "--option",
			expectedToken: ConfigToken{
				tokenType: CTK_OPTION,
				value:     "--option",
				startPos: ConfigScannerPos{
					line: 1,
					col:  1,
				},
				endPos: ConfigScannerPos{
					line: 1,
					col:  8,
				},
			},
		},
		{
			input: "\n",
			expectedToken: ConfigToken{
				tokenType: CTK_TERMINATOR,
				value:     "\n",
				startPos: ConfigScannerPos{
					line: 1,
					col:  1,
				},
				endPos: ConfigScannerPos{
					line: 1,
					col:  1,
				},
			},
		},
		{
			input: "",
			expectedToken: ConfigToken{
				tokenType: CTK_EOF,
				startPos: ConfigScannerPos{
					line: 1,
					col:  1,
				},
				endPos: ConfigScannerPos{
					line: 1,
					col:  1,
				},
			},
		},
		{
			input: "\"Unterminated string",
			expectedToken: ConfigToken{
				tokenType: CTK_INVALID,
				value:     "\"Unterminated string",
				startPos: ConfigScannerPos{
					line: 1,
					col:  1,
				},
				endPos: ConfigScannerPos{
					line: 1,
					col:  20,
				},
				err: errors.New("Unterminated string"),
			},
		},
	}

	for _, singleTokenTest := range singleTokenTests {
		scanner := NewConfigScanner(strings.NewReader(singleTokenTest.input))
		token, err := scanner.Scan()

		if err != nil {
			t.Errorf("Scan failed with error %v", err)
		} else if !token.Equal(&singleTokenTest.expectedToken) {
			t.Errorf("ConfigToken does not match expected value. Expected %v, Actual %v", singleTokenTest.expectedToken, *token)
		}
	}
}

func TestScanMultipleConfigTokens(t *testing.T) {
	var multiTokenTests = []struct {
		input          string
		expectedTokens []ConfigToken
	}{
		{
			input: "theme --create \"my theme\"\n",
			expectedTokens: []ConfigToken{
				ConfigToken{
					tokenType: CTK_WORD,
					value:     "theme",
					startPos: ConfigScannerPos{
						line: 1,
						col:  1,
					},
					endPos: ConfigScannerPos{
						line: 1,
						col:  5,
					},
				},
				ConfigToken{
					tokenType: CTK_WHITE_SPACE,
					value:     " ",
					startPos: ConfigScannerPos{
						line: 1,
						col:  6,
					},
					endPos: ConfigScannerPos{
						line: 1,
						col:  6,
					},
				},
				ConfigToken{
					tokenType: CTK_OPTION,
					value:     "--create",
					startPos: ConfigScannerPos{
						line: 1,
						col:  7,
					},
					endPos: ConfigScannerPos{
						line: 1,
						col:  14,
					},
				},
				ConfigToken{
					tokenType: CTK_WHITE_SPACE,
					value:     " ",
					startPos: ConfigScannerPos{
						line: 1,
						col:  15,
					},
					endPos: ConfigScannerPos{
						line: 1,
						col:  15,
					},
				},
				ConfigToken{
					tokenType: CTK_WORD,
					value:     "my theme",
					startPos: ConfigScannerPos{
						line: 1,
						col:  16,
					},
					endPos: ConfigScannerPos{
						line: 1,
						col:  25,
					},
				},
				ConfigToken{
					tokenType: CTK_TERMINATOR,
					value:     "\n",
					startPos: ConfigScannerPos{
						line: 1,
						col:  26,
					},
					endPos: ConfigScannerPos{
						line: 1,
						col:  26,
					},
				},
				ConfigToken{
					tokenType: CTK_EOF,
					startPos: ConfigScannerPos{
						line: 1,
						col:  26,
					},
					endPos: ConfigScannerPos{
						line: 1,
						col:  26,
					},
				},
			},
		},
		{
			input: "set theme mytheme\nset\tCommitView.dateformat \"%yyyy-mm-dd HH:MM\"\n",
			expectedTokens: []ConfigToken{
				ConfigToken{
					tokenType: CTK_WORD,
					value:     "set",
					startPos: ConfigScannerPos{
						line: 1,
						col:  1,
					},
					endPos: ConfigScannerPos{
						line: 1,
						col:  3,
					},
				},
				ConfigToken{
					tokenType: CTK_WHITE_SPACE,
					value:     " ",
					startPos: ConfigScannerPos{
						line: 1,
						col:  4,
					},
					endPos: ConfigScannerPos{
						line: 1,
						col:  4,
					},
				},
				ConfigToken{
					tokenType: CTK_WORD,
					value:     "theme",
					startPos: ConfigScannerPos{
						line: 1,
						col:  5,
					},
					endPos: ConfigScannerPos{
						line: 1,
						col:  9,
					},
				},
				ConfigToken{
					tokenType: CTK_WHITE_SPACE,
					value:     " ",
					startPos: ConfigScannerPos{
						line: 1,
						col:  10,
					},
					endPos: ConfigScannerPos{
						line: 1,
						col:  10,
					},
				},
				ConfigToken{
					tokenType: CTK_WORD,
					value:     "mytheme",
					startPos: ConfigScannerPos{
						line: 1,
						col:  11,
					},
					endPos: ConfigScannerPos{
						line: 1,
						col:  17,
					},
				},
				ConfigToken{
					tokenType: CTK_TERMINATOR,
					value:     "\n",
					startPos: ConfigScannerPos{
						line: 1,
						col:  18,
					},
					endPos: ConfigScannerPos{
						line: 1,
						col:  18,
					},
				},
				ConfigToken{
					tokenType: CTK_WORD,
					value:     "set",
					startPos: ConfigScannerPos{
						line: 2,
						col:  1,
					},
					endPos: ConfigScannerPos{
						line: 2,
						col:  3,
					},
				},
				ConfigToken{
					tokenType: CTK_WHITE_SPACE,
					value:     "\t",
					startPos: ConfigScannerPos{
						line: 2,
						col:  4,
					},
					endPos: ConfigScannerPos{
						line: 2,
						col:  4,
					},
				},
				ConfigToken{
					tokenType: CTK_WORD,
					value:     "CommitView.dateformat",
					startPos: ConfigScannerPos{
						line: 2,
						col:  5,
					},
					endPos: ConfigScannerPos{
						line: 2,
						col:  25,
					},
				},
				ConfigToken{
					tokenType: CTK_WHITE_SPACE,
					value:     " ",
					startPos: ConfigScannerPos{
						line: 2,
						col:  26,
					},
					endPos: ConfigScannerPos{
						line: 2,
						col:  26,
					},
				},
				ConfigToken{
					tokenType: CTK_WORD,
					value:     "%yyyy-mm-dd HH:MM",
					startPos: ConfigScannerPos{
						line: 2,
						col:  27,
					},
					endPos: ConfigScannerPos{
						line: 2,
						col:  45,
					},
				},
				ConfigToken{
					tokenType: CTK_TERMINATOR,
					value:     "\n",
					startPos: ConfigScannerPos{
						line: 2,
						col:  46,
					},
					endPos: ConfigScannerPos{
						line: 2,
						col:  46,
					},
				},
				ConfigToken{
					tokenType: CTK_EOF,
					startPos: ConfigScannerPos{
						line: 2,
						col:  46,
					},
					endPos: ConfigScannerPos{
						line: 2,
						col:  46,
					},
				},
			},
		},
		{
			input: "theme --create \\\n\tmytheme\n",
			expectedTokens: []ConfigToken{
				ConfigToken{
					tokenType: CTK_WORD,
					value:     "theme",
					startPos: ConfigScannerPos{
						line: 1,
						col:  1,
					},
					endPos: ConfigScannerPos{
						line: 1,
						col:  5,
					},
				},
				ConfigToken{
					tokenType: CTK_WHITE_SPACE,
					value:     " ",
					startPos: ConfigScannerPos{
						line: 1,
						col:  6,
					},
					endPos: ConfigScannerPos{
						line: 1,
						col:  6,
					},
				},
				ConfigToken{
					tokenType: CTK_OPTION,
					value:     "--create",
					startPos: ConfigScannerPos{
						line: 1,
						col:  7,
					},
					endPos: ConfigScannerPos{
						line: 1,
						col:  14,
					},
				},
				ConfigToken{
					tokenType: CTK_WHITE_SPACE,
					value:     " \t",
					startPos: ConfigScannerPos{
						line: 1,
						col:  15,
					},
					endPos: ConfigScannerPos{
						line: 2,
						col:  1,
					},
				},
				ConfigToken{
					tokenType: CTK_WORD,
					value:     "mytheme",
					startPos: ConfigScannerPos{
						line: 2,
						col:  2,
					},
					endPos: ConfigScannerPos{
						line: 2,
						col:  8,
					},
				},
				ConfigToken{
					tokenType: CTK_TERMINATOR,
					value:     "\n",
					startPos: ConfigScannerPos{
						line: 2,
						col:  9,
					},
					endPos: ConfigScannerPos{
						line: 2,
						col:  9,
					},
				},
				ConfigToken{
					tokenType: CTK_EOF,
					startPos: ConfigScannerPos{
						line: 2,
						col:  9,
					},
					endPos: ConfigScannerPos{
						line: 2,
						col:  9,
					},
				},
			},
		},
		{
			input: "theme --create \"my theme\nset theme mytheme\n",
			expectedTokens: []ConfigToken{
				ConfigToken{
					tokenType: CTK_WORD,
					value:     "theme",
					startPos: ConfigScannerPos{
						line: 1,
						col:  1,
					},
					endPos: ConfigScannerPos{
						line: 1,
						col:  5,
					},
				},
				ConfigToken{
					tokenType: CTK_WHITE_SPACE,
					value:     " ",
					startPos: ConfigScannerPos{
						line: 1,
						col:  6,
					},
					endPos: ConfigScannerPos{
						line: 1,
						col:  6,
					},
				},
				ConfigToken{
					tokenType: CTK_OPTION,
					value:     "--create",
					startPos: ConfigScannerPos{
						line: 1,
						col:  7,
					},
					endPos: ConfigScannerPos{
						line: 1,
						col:  14,
					},
				},
				ConfigToken{
					tokenType: CTK_WHITE_SPACE,
					value:     " ",
					startPos: ConfigScannerPos{
						line: 1,
						col:  15,
					},
					endPos: ConfigScannerPos{
						line: 1,
						col:  15,
					},
				},
				ConfigToken{
					tokenType: CTK_INVALID,
					value:     "\"my theme\nset theme mytheme\n",
					startPos: ConfigScannerPos{
						line: 1,
						col:  16,
					},
					endPos: ConfigScannerPos{
						line: 2,
						col:  18,
					},
					err: errors.New("Unterminated string"),
				},
				ConfigToken{
					tokenType: CTK_EOF,
					startPos: ConfigScannerPos{
						line: 2,
						col:  18,
					},
					endPos: ConfigScannerPos{
						line: 2,
						col:  18,
					},
				},
			},
		},
		{
			input: "set theme mytheme # Set theme \n # set theme again\nset theme mytheme #EOF",
			expectedTokens: []ConfigToken{
				ConfigToken{
					tokenType: CTK_WORD,
					value:     "set",
					startPos: ConfigScannerPos{
						line: 1,
						col:  1,
					},
					endPos: ConfigScannerPos{
						line: 1,
						col:  3,
					},
				},
				ConfigToken{
					tokenType: CTK_WHITE_SPACE,
					value:     " ",
					startPos: ConfigScannerPos{
						line: 1,
						col:  4,
					},
					endPos: ConfigScannerPos{
						line: 1,
						col:  4,
					},
				},
				ConfigToken{
					tokenType: CTK_WORD,
					value:     "theme",
					startPos: ConfigScannerPos{
						line: 1,
						col:  5,
					},
					endPos: ConfigScannerPos{
						line: 1,
						col:  9,
					},
				},
				ConfigToken{
					tokenType: CTK_WHITE_SPACE,
					value:     " ",
					startPos: ConfigScannerPos{
						line: 1,
						col:  10,
					},
					endPos: ConfigScannerPos{
						line: 1,
						col:  10,
					},
				},
				ConfigToken{
					tokenType: CTK_WORD,
					value:     "mytheme",
					startPos: ConfigScannerPos{
						line: 1,
						col:  11,
					},
					endPos: ConfigScannerPos{
						line: 1,
						col:  17,
					},
				},
				ConfigToken{
					tokenType: CTK_WHITE_SPACE,
					value:     " ",
					startPos: ConfigScannerPos{
						line: 1,
						col:  18,
					},
					endPos: ConfigScannerPos{
						line: 1,
						col:  18,
					},
				},
				ConfigToken{
					tokenType: CTK_COMMENT,
					value:     "# Set theme ",
					startPos: ConfigScannerPos{
						line: 1,
						col:  19,
					},
					endPos: ConfigScannerPos{
						line: 1,
						col:  30,
					},
				},
				ConfigToken{
					tokenType: CTK_TERMINATOR,
					value:     "\n",
					startPos: ConfigScannerPos{
						line: 1,
						col:  31,
					},
					endPos: ConfigScannerPos{
						line: 1,
						col:  31,
					},
				},
				ConfigToken{
					tokenType: CTK_WHITE_SPACE,
					value:     " ",
					startPos: ConfigScannerPos{
						line: 2,
						col:  1,
					},
					endPos: ConfigScannerPos{
						line: 2,
						col:  1,
					},
				},
				ConfigToken{
					tokenType: CTK_COMMENT,
					value:     "# set theme again",
					startPos: ConfigScannerPos{
						line: 2,
						col:  2,
					},
					endPos: ConfigScannerPos{
						line: 2,
						col:  18,
					},
				},
				ConfigToken{
					tokenType: CTK_TERMINATOR,
					value:     "\n",
					startPos: ConfigScannerPos{
						line: 2,
						col:  19,
					},
					endPos: ConfigScannerPos{
						line: 2,
						col:  19,
					},
				},
				ConfigToken{
					tokenType: CTK_WORD,
					value:     "set",
					startPos: ConfigScannerPos{
						line: 3,
						col:  1,
					},
					endPos: ConfigScannerPos{
						line: 3,
						col:  3,
					},
				},
				ConfigToken{
					tokenType: CTK_WHITE_SPACE,
					value:     " ",
					startPos: ConfigScannerPos{
						line: 3,
						col:  4,
					},
					endPos: ConfigScannerPos{
						line: 3,
						col:  4,
					},
				},
				ConfigToken{
					tokenType: CTK_WORD,
					value:     "theme",
					startPos: ConfigScannerPos{
						line: 3,
						col:  5,
					},
					endPos: ConfigScannerPos{
						line: 3,
						col:  9,
					},
				},
				ConfigToken{
					tokenType: CTK_WHITE_SPACE,
					value:     " ",
					startPos: ConfigScannerPos{
						line: 3,
						col:  10,
					},
					endPos: ConfigScannerPos{
						line: 3,
						col:  10,
					},
				},
				ConfigToken{
					tokenType: CTK_WORD,
					value:     "mytheme",
					startPos: ConfigScannerPos{
						line: 3,
						col:  11,
					},
					endPos: ConfigScannerPos{
						line: 3,
						col:  17,
					},
				},
				ConfigToken{
					tokenType: CTK_WHITE_SPACE,
					value:     " ",
					startPos: ConfigScannerPos{
						line: 3,
						col:  18,
					},
					endPos: ConfigScannerPos{
						line: 3,
						col:  18,
					},
				},
				ConfigToken{
					tokenType: CTK_COMMENT,
					value:     "#EOF",
					startPos: ConfigScannerPos{
						line: 3,
						col:  19,
					},
					endPos: ConfigScannerPos{
						line: 3,
						col:  22,
					},
				},
				ConfigToken{
					tokenType: CTK_EOF,
					startPos: ConfigScannerPos{
						line: 3,
						col:  22,
					},
					endPos: ConfigScannerPos{
						line: 3,
						col:  22,
					},
				},
			},
		},
	}

	for _, multiTokenTest := range multiTokenTests {
		scanner := NewConfigScanner(strings.NewReader(multiTokenTest.input))

		for _, expectedToken := range multiTokenTest.expectedTokens {
			token, err := scanner.Scan()

			if err != nil {
				t.Errorf("Scan failed with error %v", err)
			} else if !token.Equal(&expectedToken) {
				t.Errorf("ConfigToken does not match expected value. Expected %v, Actual %v", expectedToken, *token)
			}
		}
	}
}
