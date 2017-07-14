package main

import (
	"strings"
	"testing"
)

func TestParseQuery(t *testing.T) {
	var queryTests = []struct {
		input              string
		expectedExpression Expression
	}{
		{
			input: "AuthorName = \"Test Author\"",
			expectedExpression: &BinaryExpression{
				operator: &Operator{
					operator: &QueryToken{
						value: "=",
					},
					precedence: 3,
				},
				lhs: &Identifier{
					identifier: &QueryToken{
						value: "AuthorName",
					},
				},
				rhs: &StringLiteral{
					value: &QueryToken{
						value: "Test Author",
					},
				},
			},
		},
		{
			input: "AuthorDate >= \"2017-07-10\" AND (AuthorName = \"Test Author\" OR CommitterName = \"Test Author\")",
			expectedExpression: &BinaryExpression{
				operator: &Operator{
					operator: &QueryToken{
						value: "AND",
					},
					precedence: 2,
				},
				lhs: &BinaryExpression{
					operator: &Operator{
						operator: &QueryToken{
							value: ">=",
						},
						precedence: 3,
					},
					lhs: &Identifier{
						identifier: &QueryToken{
							value: "AuthorDate",
						},
					},
					rhs: &StringLiteral{
						value: &QueryToken{
							value: "2017-07-10",
						},
					},
				},
				rhs: &ParenExpression{
					expression: &BinaryExpression{
						operator: &Operator{
							operator: &QueryToken{
								value: "OR",
							},
							precedence: 1,
						},
						lhs: &BinaryExpression{
							operator: &Operator{
								operator: &QueryToken{
									value: "=",
								},
								precedence: 3,
							},
							lhs: &Identifier{
								identifier: &QueryToken{
									value: "AuthorName",
								},
							},
							rhs: &StringLiteral{
								value: &QueryToken{
									value: "Test Author",
								},
							},
						},
						rhs: &BinaryExpression{
							operator: &Operator{
								operator: &QueryToken{
									value: "=",
								},
								precedence: 3,
							},
							lhs: &Identifier{
								identifier: &QueryToken{
									value: "CommitterName",
								},
							},
							rhs: &StringLiteral{
								value: &QueryToken{
									value: "Test Author",
								},
							},
						},
					},
				},
			},
		},
	}

	for _, queryTest := range queryTests {
		expectedExpression := queryTest.expectedExpression
		parser := NewQueryParser(strings.NewReader(queryTest.input))
		expression, _, err := parser.Parse()

		if err != nil {
			t.Errorf("Parse failed with error %v", err)
		} else if !expectedExpression.Equal(expression) {
			t.Errorf("Expression does not match expected value. Expected: %v, Actual: %v", expectedExpression, expression)
		}
	}
}

func TestEOFIsSetByQueryParser(t *testing.T) {
	var eofTests = []struct {
		input string
	}{
		{
			input: "AuthorName = \"Test Author\"",
		},
		{
			input: "(AuthorName = \"Test Author\")",
		},
	}

	for _, eofTest := range eofTests {
		parser := NewQueryParser(strings.NewReader(eofTest.input))
		_, _, err := parser.Parse()

		if err != nil {
			t.Errorf("Parse failed with error %v", err)
		}

		_, eof, err := parser.Parse()

		if err != nil {
			t.Errorf("Parse failed with error %v", err)
		} else if !eof {
			t.Errorf("Expected EOF after calling parser twice for expression: %v", eofTest.input)
		}
	}

}

func TestErrorsAreReceivedForInvalidQueryTokenSequences(t *testing.T) {
	var errorTests = []struct {
		input                string
		expectedErrorMessage string
	}{
		{
			input:                "AuthorName \"Test\"",
			expectedErrorMessage: "1:12: Expected operator but found: Test",
		},
		{
			input:                "(AuthorName = \"Test\"",
			expectedErrorMessage: "1:20: Expected ')' but found: EOF",
		},
		{
			input:                "= \"Test\"",
			expectedErrorMessage: "1:1: Expected Identifier, String or Number but found: =",
		},
	}

	for _, errorTest := range errorTests {
		parser := NewQueryParser(strings.NewReader(errorTest.input))
		_, _, err := parser.Parse()

		if err == nil {
			t.Errorf("Expected Parse to return error: %v", errorTest.expectedErrorMessage)
		} else if err.Error() != errorTest.expectedErrorMessage {
			t.Errorf("Error message does not match expected value. Expected %v, Actual %v", errorTest.expectedErrorMessage, err.Error())
		}
	}
}

func TestOperatorPrecedenceIsRespected(t *testing.T) {
	var queryTests = []struct {
		input              string
		expectedExpression Expression
	}{
		{
			input: "AuthorDate >= \"2017-07-10\" AND AuthorName = \"Test Author\" OR CommitterName = \"Test Author\"",
			expectedExpression: &BinaryExpression{
				operator: &Operator{
					operator: &QueryToken{
						value: "OR",
					},
					precedence: 1,
				},
				lhs: &BinaryExpression{
					operator: &Operator{
						operator: &QueryToken{
							value: "AND",
						},
						precedence: 2,
					},
					lhs: &BinaryExpression{
						operator: &Operator{
							operator: &QueryToken{
								value: ">=",
							},
							precedence: 3,
						},
						lhs: &Identifier{
							identifier: &QueryToken{
								value: "AuthorDate",
							},
						},
						rhs: &StringLiteral{
							value: &QueryToken{
								value: "2017-07-10",
							},
						},
					},
					rhs: &BinaryExpression{
						operator: &Operator{
							operator: &QueryToken{
								value: "=",
							},
							precedence: 3,
						},
						lhs: &Identifier{
							identifier: &QueryToken{
								value: "AuthorName",
							},
						},
						rhs: &StringLiteral{
							value: &QueryToken{
								value: "Test Author",
							},
						},
					},
				},
				rhs: &BinaryExpression{
					operator: &Operator{
						operator: &QueryToken{
							value: "=",
						},
						precedence: 3,
					},
					lhs: &Identifier{
						identifier: &QueryToken{
							value: "CommitterName",
						},
					},
					rhs: &StringLiteral{
						value: &QueryToken{
							value: "Test Author",
						},
					},
				},
			},
		},
	}

	for _, queryTest := range queryTests {
		expectedExpression := queryTest.expectedExpression
		parser := NewQueryParser(strings.NewReader(queryTest.input))
		expression, _, err := parser.Parse()

		if err != nil {
			t.Errorf("Parse failed with error %v", err)
		} else if !expectedExpression.Equal(expression) {
			t.Errorf("Expression does not match expected value. Expected: %v, Actual: %v", expectedExpression, expression)
		}
	}
}
