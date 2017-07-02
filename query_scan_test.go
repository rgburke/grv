package main

import (
	"errors"
	"strings"
	"testing"
)

func TestScanSingleQueryToken(t *testing.T) {
	var singleTokenTests = []struct {
		input         string
		expectedToken QueryToken
	}{
		{
			input: "\n \t\r\v\f",
			expectedToken: QueryToken{
				tokenType: QTK_WHITE_SPACE,
				value:     "\n \t\r\v\f",
				startPos: QueryScannerPos{
					line: 1,
					col:  1,
				},
				endPos: QueryScannerPos{
					line: 2,
					col:  5,
				},
			},
		},
		{
			input: "",
			expectedToken: QueryToken{
				tokenType: QTK_EOF,
				startPos: QueryScannerPos{
					line: 1,
					col:  1,
				},
				endPos: QueryScannerPos{
					line: 1,
					col:  1,
				},
			},
		},
		{
			input: "authorName",
			expectedToken: QueryToken{
				tokenType: QTK_IDENTIFIER,
				value:     "authorName",
				startPos: QueryScannerPos{
					line: 1,
					col:  1,
				},
				endPos: QueryScannerPos{
					line: 1,
					col:  10,
				},
			},
		},
		{
			input: "1234",
			expectedToken: QueryToken{
				tokenType: QTK_NUMBER,
				value:     "1234",
				startPos: QueryScannerPos{
					line: 1,
					col:  1,
				},
				endPos: QueryScannerPos{
					line: 1,
					col:  4,
				},
			},
		},
		{
			input: "12.34",
			expectedToken: QueryToken{
				tokenType: QTK_NUMBER,
				value:     "12.34",
				startPos: QueryScannerPos{
					line: 1,
					col:  1,
				},
				endPos: QueryScannerPos{
					line: 1,
					col:  5,
				},
			},
		},
		{
			input: ".1234",
			expectedToken: QueryToken{
				tokenType: QTK_NUMBER,
				value:     ".1234",
				startPos: QueryScannerPos{
					line: 1,
					col:  1,
				},
				endPos: QueryScannerPos{
					line: 1,
					col:  5,
				},
			},
		},
		{
			input: "-1234",
			expectedToken: QueryToken{
				tokenType: QTK_NUMBER,
				value:     "-1234",
				startPos: QueryScannerPos{
					line: 1,
					col:  1,
				},
				endPos: QueryScannerPos{
					line: 1,
					col:  5,
				},
			},
		},
		{
			input: "-12.34",
			expectedToken: QueryToken{
				tokenType: QTK_NUMBER,
				value:     "-12.34",
				startPos: QueryScannerPos{
					line: 1,
					col:  1,
				},
				endPos: QueryScannerPos{
					line: 1,
					col:  6,
				},
			},
		},
		{
			input: "-.1234",
			expectedToken: QueryToken{
				tokenType: QTK_NUMBER,
				value:     "-.1234",
				startPos: QueryScannerPos{
					line: 1,
					col:  1,
				},
				endPos: QueryScannerPos{
					line: 1,
					col:  6,
				},
			},
		},
		{
			input: "12.3.4",
			expectedToken: QueryToken{
				tokenType: QTK_INVALID,
				value:     "12.3.",
				startPos: QueryScannerPos{
					line: 1,
					col:  1,
				},
				endPos: QueryScannerPos{
					line: 1,
					col:  5,
				},
				err: errors.New("Unexpected '.' character in number"),
			},
		},
		{
			input: "-1234-",
			expectedToken: QueryToken{
				tokenType: QTK_INVALID,
				value:     "-1234-",
				startPos: QueryScannerPos{
					line: 1,
					col:  1,
				},
				endPos: QueryScannerPos{
					line: 1,
					col:  6,
				},
				err: errors.New("Unexpected '-' character in number"),
			},
		},
		{
			input: "\"Bug fix\"",
			expectedToken: QueryToken{
				tokenType: QTK_STRING,
				value:     "Bug fix",
				startPos: QueryScannerPos{
					line: 1,
					col:  1,
				},
				endPos: QueryScannerPos{
					line: 1,
					col:  9,
				},
			},
		},
		{
			input: "\"\\tBug\\n\\tfix\"",
			expectedToken: QueryToken{
				tokenType: QTK_STRING,
				value:     "\tBug\n\tfix",
				startPos: QueryScannerPos{
					line: 1,
					col:  1,
				},
				endPos: QueryScannerPos{
					line: 1,
					col:  14,
				},
			},
		},
		{
			input: "\"Unterminated string",
			expectedToken: QueryToken{
				tokenType: QTK_INVALID,
				value:     "\"Unterminated string",
				startPos: QueryScannerPos{
					line: 1,
					col:  1,
				},
				endPos: QueryScannerPos{
					line: 1,
					col:  20,
				},
				err: errors.New("Unterminated string"),
			},
		},
		{
			input: "AND",
			expectedToken: QueryToken{
				tokenType: QTK_AND,
				value:     "AND",
				startPos: QueryScannerPos{
					line: 1,
					col:  1,
				},
				endPos: QueryScannerPos{
					line: 1,
					col:  3,
				},
			},
		},
		{
			input: "OR",
			expectedToken: QueryToken{
				tokenType: QTK_OR,
				value:     "OR",
				startPos: QueryScannerPos{
					line: 1,
					col:  1,
				},
				endPos: QueryScannerPos{
					line: 1,
					col:  2,
				},
			},
		},
		{
			input: "NOT",
			expectedToken: QueryToken{
				tokenType: QTK_NOT,
				value:     "NOT",
				startPos: QueryScannerPos{
					line: 1,
					col:  1,
				},
				endPos: QueryScannerPos{
					line: 1,
					col:  3,
				},
			},
		},
		{
			input: "=",
			expectedToken: QueryToken{
				tokenType: QTK_CMP_EQ,
				value:     "=",
				startPos: QueryScannerPos{
					line: 1,
					col:  1,
				},
				endPos: QueryScannerPos{
					line: 1,
					col:  1,
				},
			},
		},
		{
			input: "!=",
			expectedToken: QueryToken{
				tokenType: QTK_CMP_NE,
				value:     "!=",
				startPos: QueryScannerPos{
					line: 1,
					col:  1,
				},
				endPos: QueryScannerPos{
					line: 1,
					col:  2,
				},
			},
		},
		{
			input: "!>",
			expectedToken: QueryToken{
				tokenType: QTK_INVALID,
				value:     "!>",
				startPos: QueryScannerPos{
					line: 1,
					col:  1,
				},
				endPos: QueryScannerPos{
					line: 1,
					col:  2,
				},
				err: errors.New("Expected '=' character after '!'"),
			},
		},
		{
			input: ">",
			expectedToken: QueryToken{
				tokenType: QTK_CMP_GT,
				value:     ">",
				startPos: QueryScannerPos{
					line: 1,
					col:  1,
				},
				endPos: QueryScannerPos{
					line: 1,
					col:  1,
				},
			},
		},
		{
			input: ">=",
			expectedToken: QueryToken{
				tokenType: QTK_CMP_GE,
				value:     ">=",
				startPos: QueryScannerPos{
					line: 1,
					col:  1,
				},
				endPos: QueryScannerPos{
					line: 1,
					col:  2,
				},
			},
		},
		{
			input: "<",
			expectedToken: QueryToken{
				tokenType: QTK_CMP_LT,
				value:     "<",
				startPos: QueryScannerPos{
					line: 1,
					col:  1,
				},
				endPos: QueryScannerPos{
					line: 1,
					col:  1,
				},
			},
		},
		{
			input: "<=",
			expectedToken: QueryToken{
				tokenType: QTK_CMP_LE,
				value:     "<=",
				startPos: QueryScannerPos{
					line: 1,
					col:  1,
				},
				endPos: QueryScannerPos{
					line: 1,
					col:  2,
				},
			},
		},
		{
			input: "(",
			expectedToken: QueryToken{
				tokenType: QTK_LPAREN,
				value:     "(",
				startPos: QueryScannerPos{
					line: 1,
					col:  1,
				},
				endPos: QueryScannerPos{
					line: 1,
					col:  1,
				},
			},
		},
		{
			input: ")",
			expectedToken: QueryToken{
				tokenType: QTK_RPAREN,
				value:     ")",
				startPos: QueryScannerPos{
					line: 1,
					col:  1,
				},
				endPos: QueryScannerPos{
					line: 1,
					col:  1,
				},
			},
		},
	}

	for _, singleTokenTest := range singleTokenTests {
		scanner := NewQueryScanner(strings.NewReader(singleTokenTest.input))
		token, err := scanner.Scan()

		if err != nil {
			t.Errorf("Scan failed with error %v", err)
		} else if !token.Equal(&singleTokenTest.expectedToken) {
			t.Errorf("QueryToken does not match expected value. Expected %v, Actual %v", singleTokenTest.expectedToken, *token)
		}
	}
}

func TestScanMultipleTokens(t *testing.T) {
	var multiTokenTests = []struct {
		input          string
		expectedTokens []QueryToken
	}{
		{
			input: "authorName=\"John Smith\"",
			expectedTokens: []QueryToken{
				QueryToken{
					tokenType: QTK_IDENTIFIER,
					value:     "authorName",
					startPos: QueryScannerPos{
						line: 1,
						col:  1,
					},
					endPos: QueryScannerPos{
						line: 1,
						col:  10,
					},
				},
				QueryToken{
					tokenType: QTK_CMP_EQ,
					value:     "=",
					startPos: QueryScannerPos{
						line: 1,
						col:  11,
					},
					endPos: QueryScannerPos{
						line: 1,
						col:  11,
					},
				},
				QueryToken{
					tokenType: QTK_STRING,
					value:     "John Smith",
					startPos: QueryScannerPos{
						line: 1,
						col:  12,
					},
					endPos: QueryScannerPos{
						line: 1,
						col:  23,
					},
				},
				QueryToken{
					tokenType: QTK_EOF,
					startPos: QueryScannerPos{
						line: 1,
						col:  23,
					},
					endPos: QueryScannerPos{
						line: 1,
						col:  23,
					},
				},
			},
		},
		{
			input: "authorDate >= \"2017-07-02 00:00:00\" AND (authorName = \"John Smith\" OR committerName = \"John Smith\")",
			expectedTokens: []QueryToken{
				QueryToken{
					tokenType: QTK_IDENTIFIER,
					value:     "authorDate",
					startPos: QueryScannerPos{
						line: 1,
						col:  1,
					},
					endPos: QueryScannerPos{
						line: 1,
						col:  10,
					},
				},
				QueryToken{
					tokenType: QTK_WHITE_SPACE,
					value:     " ",
					startPos: QueryScannerPos{
						line: 1,
						col:  11,
					},
					endPos: QueryScannerPos{
						line: 1,
						col:  11,
					},
				},
				QueryToken{
					tokenType: QTK_CMP_GE,
					value:     ">=",
					startPos: QueryScannerPos{
						line: 1,
						col:  12,
					},
					endPos: QueryScannerPos{
						line: 1,
						col:  13,
					},
				},
				QueryToken{
					tokenType: QTK_WHITE_SPACE,
					value:     " ",
					startPos: QueryScannerPos{
						line: 1,
						col:  14,
					},
					endPos: QueryScannerPos{
						line: 1,
						col:  14,
					},
				},
				QueryToken{
					tokenType: QTK_STRING,
					value:     "2017-07-02 00:00:00",
					startPos: QueryScannerPos{
						line: 1,
						col:  15,
					},
					endPos: QueryScannerPos{
						line: 1,
						col:  35,
					},
				},
				QueryToken{
					tokenType: QTK_WHITE_SPACE,
					value:     " ",
					startPos: QueryScannerPos{
						line: 1,
						col:  36,
					},
					endPos: QueryScannerPos{
						line: 1,
						col:  36,
					},
				},
				QueryToken{
					tokenType: QTK_AND,
					value:     "AND",
					startPos: QueryScannerPos{
						line: 1,
						col:  37,
					},
					endPos: QueryScannerPos{
						line: 1,
						col:  39,
					},
				},
				QueryToken{
					tokenType: QTK_WHITE_SPACE,
					value:     " ",
					startPos: QueryScannerPos{
						line: 1,
						col:  40,
					},
					endPos: QueryScannerPos{
						line: 1,
						col:  40,
					},
				},
				QueryToken{
					tokenType: QTK_LPAREN,
					value:     "(",
					startPos: QueryScannerPos{
						line: 1,
						col:  41,
					},
					endPos: QueryScannerPos{
						line: 1,
						col:  41,
					},
				},
				QueryToken{
					tokenType: QTK_IDENTIFIER,
					value:     "authorName",
					startPos: QueryScannerPos{
						line: 1,
						col:  42,
					},
					endPos: QueryScannerPos{
						line: 1,
						col:  51,
					},
				},
				QueryToken{
					tokenType: QTK_WHITE_SPACE,
					value:     " ",
					startPos: QueryScannerPos{
						line: 1,
						col:  52,
					},
					endPos: QueryScannerPos{
						line: 1,
						col:  52,
					},
				},
				QueryToken{
					tokenType: QTK_CMP_EQ,
					value:     "=",
					startPos: QueryScannerPos{
						line: 1,
						col:  53,
					},
					endPos: QueryScannerPos{
						line: 1,
						col:  53,
					},
				},
				QueryToken{
					tokenType: QTK_WHITE_SPACE,
					value:     " ",
					startPos: QueryScannerPos{
						line: 1,
						col:  54,
					},
					endPos: QueryScannerPos{
						line: 1,
						col:  54,
					},
				},
				QueryToken{
					tokenType: QTK_STRING,
					value:     "John Smith",
					startPos: QueryScannerPos{
						line: 1,
						col:  55,
					},
					endPos: QueryScannerPos{
						line: 1,
						col:  66,
					},
				},
				QueryToken{
					tokenType: QTK_WHITE_SPACE,
					value:     " ",
					startPos: QueryScannerPos{
						line: 1,
						col:  67,
					},
					endPos: QueryScannerPos{
						line: 1,
						col:  67,
					},
				},
				QueryToken{
					tokenType: QTK_OR,
					value:     "OR",
					startPos: QueryScannerPos{
						line: 1,
						col:  68,
					},
					endPos: QueryScannerPos{
						line: 1,
						col:  69,
					},
				},
				QueryToken{
					tokenType: QTK_WHITE_SPACE,
					value:     " ",
					startPos: QueryScannerPos{
						line: 1,
						col:  70,
					},
					endPos: QueryScannerPos{
						line: 1,
						col:  70,
					},
				},
				QueryToken{
					tokenType: QTK_IDENTIFIER,
					value:     "committerName",
					startPos: QueryScannerPos{
						line: 1,
						col:  71,
					},
					endPos: QueryScannerPos{
						line: 1,
						col:  83,
					},
				},
				QueryToken{
					tokenType: QTK_WHITE_SPACE,
					value:     " ",
					startPos: QueryScannerPos{
						line: 1,
						col:  84,
					},
					endPos: QueryScannerPos{
						line: 1,
						col:  84,
					},
				},
				QueryToken{
					tokenType: QTK_CMP_EQ,
					value:     "=",
					startPos: QueryScannerPos{
						line: 1,
						col:  85,
					},
					endPos: QueryScannerPos{
						line: 1,
						col:  85,
					},
				},
				QueryToken{
					tokenType: QTK_WHITE_SPACE,
					value:     " ",
					startPos: QueryScannerPos{
						line: 1,
						col:  86,
					},
					endPos: QueryScannerPos{
						line: 1,
						col:  86,
					},
				},
				QueryToken{
					tokenType: QTK_STRING,
					value:     "John Smith",
					startPos: QueryScannerPos{
						line: 1,
						col:  87,
					},
					endPos: QueryScannerPos{
						line: 1,
						col:  98,
					},
				},
				QueryToken{
					tokenType: QTK_RPAREN,
					value:     ")",
					startPos: QueryScannerPos{
						line: 1,
						col:  99,
					},
					endPos: QueryScannerPos{
						line: 1,
						col:  99,
					},
				},
				QueryToken{
					tokenType: QTK_EOF,
					startPos: QueryScannerPos{
						line: 1,
						col:  99,
					},
					endPos: QueryScannerPos{
						line: 1,
						col:  99,
					},
				},
			},
		},
	}

	for _, multiTokenTest := range multiTokenTests {
		scanner := NewQueryScanner(strings.NewReader(multiTokenTest.input))

		for _, expectedToken := range multiTokenTest.expectedTokens {
			token, err := scanner.Scan()

			if err != nil {
				t.Errorf("Scan failed with error %v", err)
			} else if !token.Equal(&expectedToken) {
				t.Errorf("QueryToken does not match expected value. Expected %v, Actual %v", expectedToken, *token)
			}
		}
	}
}
