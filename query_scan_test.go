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
				tokenType: QtkWhiteSpace,
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
				tokenType: QtkEOF,
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
			input: "AuthorName",
			expectedToken: QueryToken{
				tokenType: QtkIdentifier,
				value:     "AuthorName",
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
				tokenType: QtkNumber,
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
				tokenType: QtkNumber,
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
				tokenType: QtkNumber,
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
				tokenType: QtkNumber,
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
				tokenType: QtkNumber,
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
				tokenType: QtkNumber,
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
				tokenType: QtkInvalid,
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
				tokenType: QtkInvalid,
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
				tokenType: QtkString,
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
				tokenType: QtkString,
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
				tokenType: QtkInvalid,
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
				tokenType: QtkAnd,
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
				tokenType: QtkOr,
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
				tokenType: QtkNot,
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
				tokenType: QtkCmpEq,
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
				tokenType: QtkCmpNe,
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
				tokenType: QtkInvalid,
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
				tokenType: QtkCmpGt,
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
				tokenType: QtkCmpGe,
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
				tokenType: QtkCmpLt,
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
				tokenType: QtkCmpLe,
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
				tokenType: QtkLparen,
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
				tokenType: QtkRparen,
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
		{
			input: "GLOB",
			expectedToken: QueryToken{
				tokenType: QtkCmpGlob,
				value:     "GLOB",
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
			input: "REGEXP",
			expectedToken: QueryToken{
				tokenType: QtkCmpRegexp,
				value:     "REGEXP",
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

func TestOperatorsAreCaseInsensitive(t *testing.T) {
	var operatorTokenTests = []struct {
		input         string
		expectedToken QueryToken
	}{
		{
			input: "and",
			expectedToken: QueryToken{
				tokenType: QtkAnd,
				value:     "and",
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
			input: "Or",
			expectedToken: QueryToken{
				tokenType: QtkOr,
				value:     "Or",
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
			input: "nOT",
			expectedToken: QueryToken{
				tokenType: QtkNot,
				value:     "nOT",
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
			input: "GlOb",
			expectedToken: QueryToken{
				tokenType: QtkCmpGlob,
				value:     "GlOb",
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
			input: "ReGeXp",
			expectedToken: QueryToken{
				tokenType: QtkCmpRegexp,
				value:     "ReGeXp",
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
	}

	for _, operatorTokenTest := range operatorTokenTests {
		scanner := NewQueryScanner(strings.NewReader(operatorTokenTest.input))
		token, err := scanner.Scan()

		if err != nil {
			t.Errorf("Scan failed with error %v", err)
		} else if !token.Equal(&operatorTokenTest.expectedToken) {
			t.Errorf("QueryToken does not match expected value. Expected %v, Actual %v", operatorTokenTest.expectedToken, *token)
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
				{
					tokenType: QtkIdentifier,
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
				{
					tokenType: QtkCmpEq,
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
				{
					tokenType: QtkString,
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
				{
					tokenType: QtkEOF,
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
				{
					tokenType: QtkIdentifier,
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
				{
					tokenType: QtkWhiteSpace,
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
				{
					tokenType: QtkCmpGe,
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
				{
					tokenType: QtkWhiteSpace,
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
				{
					tokenType: QtkString,
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
				{
					tokenType: QtkWhiteSpace,
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
				{
					tokenType: QtkAnd,
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
				{
					tokenType: QtkWhiteSpace,
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
				{
					tokenType: QtkLparen,
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
				{
					tokenType: QtkIdentifier,
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
				{
					tokenType: QtkWhiteSpace,
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
				{
					tokenType: QtkCmpEq,
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
				{
					tokenType: QtkWhiteSpace,
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
				{
					tokenType: QtkString,
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
				{
					tokenType: QtkWhiteSpace,
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
				{
					tokenType: QtkOr,
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
				{
					tokenType: QtkWhiteSpace,
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
				{
					tokenType: QtkIdentifier,
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
				{
					tokenType: QtkWhiteSpace,
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
				{
					tokenType: QtkCmpEq,
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
				{
					tokenType: QtkWhiteSpace,
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
				{
					tokenType: QtkString,
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
				{
					tokenType: QtkRparen,
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
				{
					tokenType: QtkEOF,
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
		{
			input: "authorName GLOB \"%John%\"",
			expectedTokens: []QueryToken{
				{
					tokenType: QtkIdentifier,
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
				{
					tokenType: QtkWhiteSpace,
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
				{
					tokenType: QtkCmpGlob,
					value:     "GLOB",
					startPos: QueryScannerPos{
						line: 1,
						col:  12,
					},
					endPos: QueryScannerPos{
						line: 1,
						col:  15,
					},
				},
				{
					tokenType: QtkWhiteSpace,
					value:     " ",
					startPos: QueryScannerPos{
						line: 1,
						col:  16,
					},
					endPos: QueryScannerPos{
						line: 1,
						col:  16,
					},
				},
				{
					tokenType: QtkString,
					value:     "%John%",
					startPos: QueryScannerPos{
						line: 1,
						col:  17,
					},
					endPos: QueryScannerPos{
						line: 1,
						col:  24,
					},
				},
				{
					tokenType: QtkEOF,
					startPos: QueryScannerPos{
						line: 1,
						col:  24,
					},
					endPos: QueryScannerPos{
						line: 1,
						col:  24,
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
