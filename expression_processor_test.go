package main

import (
	"errors"
	"reflect"
	"regexp"
	"testing"
	"time"
)

type TestFieldTypeDescriptor struct{}

func (testFieldTypeDescriptor *TestFieldTypeDescriptor) FieldType(fieldName string) (fieldType FieldType, fieldExists bool) {
	switch fieldName {
	case "AuthorName", "CommitterName", "Summary":
		fieldType = FT_STRING
		fieldExists = true
	case "AuthorDate", "CommitterDate":
		fieldType = FT_DATE
		fieldExists = true
	case "ParentCount":
		fieldType = FT_NUMBER
		fieldExists = true
	}

	return
}

func TestErrorReturnedIfExpressionNotRefinable(t *testing.T) {
	var expression Expression = &StringLiteral{}
	expectedErrorMessage := "Expected logical expression but received expression of type StringLiteral"

	expressionProcessor := NewExpressionProcessor(expression, &TestFieldTypeDescriptor{})

	_, errors := expressionProcessor.Process()

	if len(errors) != 1 {
		t.Errorf("Expected error but none returned for invalid expression type")
	} else if errors[0].Error() != expectedErrorMessage {
		t.Errorf("Returned error does not match expected error message. Expected: \"%v\". Actual: \"%v\"", expectedErrorMessage, errors[0])
	}
}

func TestDateStringsAreConvertedToDateLiteralsInDateFieldContext(t *testing.T) {
	var typeConversionTests = []struct {
		inputExpression    Expression
		expectedExpression Expression
	}{
		{
			inputExpression: &BinaryExpression{
				operator: &Operator{
					operator: &QueryToken{
						value:     "=",
						tokenType: QTK_CMP_EQ,
					},
				},
				lhs: &Identifier{
					identifier: &QueryToken{
						value: "AuthorDate",
					},
				},
				rhs: &StringLiteral{
					value: &QueryToken{
						value: "2017-07-16",
					},
				},
			},
			expectedExpression: &BinaryExpression{
				operator: &Operator{
					operator: &QueryToken{
						value:     "=",
						tokenType: QTK_CMP_EQ,
					},
				},
				lhs: &Identifier{
					identifier: &QueryToken{
						value: "AuthorDate",
					},
				},
				rhs: &DateLiteral{
					dateTime: time.Date(2017, time.July, 16, 0, 0, 0, 0, time.Local),
				},
			},
		},
		{
			inputExpression: &UnaryExpression{
				operator: &Operator{
					operator: &QueryToken{
						value:     "NOT",
						tokenType: QTK_NOT,
					},
				},
				expression: &BinaryExpression{
					operator: &Operator{
						operator: &QueryToken{
							value:     "=",
							tokenType: QTK_CMP_EQ,
						},
					},
					lhs: &Identifier{
						identifier: &QueryToken{
							value: "AuthorDate",
						},
					},
					rhs: &StringLiteral{
						value: &QueryToken{
							value: "2017-07-16",
						},
					},
				},
			},
			expectedExpression: &UnaryExpression{
				operator: &Operator{
					operator: &QueryToken{
						value:     "NOT",
						tokenType: QTK_NOT,
					},
				},
				expression: &BinaryExpression{
					operator: &Operator{
						operator: &QueryToken{
							value:     "=",
							tokenType: QTK_CMP_EQ,
						},
					},
					lhs: &Identifier{
						identifier: &QueryToken{
							value: "AuthorDate",
						},
					},
					rhs: &DateLiteral{
						dateTime: time.Date(2017, time.July, 16, 0, 0, 0, 0, time.Local),
					},
				},
			},
		},
		{
			inputExpression: &BinaryExpression{
				operator: &Operator{
					operator: &QueryToken{
						value:     "AND",
						tokenType: QTK_AND,
					},
				},
				lhs: &BinaryExpression{
					operator: &Operator{
						operator: &QueryToken{
							value:     "=",
							tokenType: QTK_CMP_EQ,
						},
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
				rhs: &ParenExpression{
					expression: &BinaryExpression{
						operator: &Operator{
							operator: &QueryToken{
								value:     "OR",
								tokenType: QTK_OR,
							},
						},
						lhs: &BinaryExpression{
							operator: &Operator{
								operator: &QueryToken{
									value:     "<=",
									tokenType: QTK_CMP_LE,
								},
							},
							lhs: &Identifier{
								identifier: &QueryToken{
									value: "AuthorDate",
								},
							},
							rhs: &StringLiteral{
								value: &QueryToken{
									value: "2017-07-16 23:59:59",
								},
							},
						},
						rhs: &BinaryExpression{
							operator: &Operator{
								operator: &QueryToken{
									value:     ">=",
									tokenType: QTK_CMP_GE,
								},
							},
							lhs: &Identifier{
								identifier: &QueryToken{
									value: "CommitterDate",
								},
							},
							rhs: &StringLiteral{
								value: &QueryToken{
									value: "2017-07-16",
								},
							},
						},
					},
				},
			},
			expectedExpression: &BinaryExpression{
				operator: &Operator{
					operator: &QueryToken{
						value:     "AND",
						tokenType: QTK_AND,
					},
				},
				lhs: &BinaryExpression{
					operator: &Operator{
						operator: &QueryToken{
							value:     "=",
							tokenType: QTK_CMP_EQ,
						},
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
				rhs: &ParenExpression{
					expression: &BinaryExpression{
						operator: &Operator{
							operator: &QueryToken{
								value:     "OR",
								tokenType: QTK_OR,
							},
						},
						lhs: &BinaryExpression{
							operator: &Operator{
								operator: &QueryToken{
									value:     "<=",
									tokenType: QTK_CMP_LE,
								},
							},
							lhs: &Identifier{
								identifier: &QueryToken{
									value: "AuthorDate",
								},
							},
							rhs: &DateLiteral{
								dateTime: time.Date(2017, time.July, 16, 23, 59, 59, 0, time.Local),
							},
						},
						rhs: &BinaryExpression{
							operator: &Operator{
								operator: &QueryToken{
									value:     ">=",
									tokenType: QTK_CMP_GE,
								},
							},
							lhs: &Identifier{
								identifier: &QueryToken{
									value: "CommitterDate",
								},
							},
							rhs: &DateLiteral{
								dateTime: time.Date(2017, time.July, 16, 0, 0, 0, 0, time.Local),
							},
						},
					},
				},
			},
		},
	}

	for _, typeConversionTest := range typeConversionTests {
		inputExpression := typeConversionTest.inputExpression
		expectedExpression := typeConversionTest.expectedExpression

		expressionProcessor := NewExpressionProcessor(inputExpression, &TestFieldTypeDescriptor{})
		actualExpression, errors := expressionProcessor.Process()

		if len(errors) > 0 {
			t.Errorf("Process failed with errors: %v", errors)
		} else if !expectedExpression.Equal(actualExpression) {
			t.Errorf("Expression does not match expected value. Expected: %v, Actual: %v", expectedExpression, actualExpression)
		}
	}
}

func TestGlobStringsAreConvertedToGlobLiteralsInGlobFieldContext(t *testing.T) {
	var typeConversionTests = []struct {
		inputExpression    Expression
		expectedExpression Expression
	}{
		{
			inputExpression: &BinaryExpression{
				operator: &Operator{
					operator: &QueryToken{
						value:     "GLOB",
						tokenType: QTK_CMP_GLOB,
					},
				},
				lhs: &Identifier{
					identifier: &QueryToken{
						value: "Summary",
					},
				},
				rhs: &StringLiteral{
					value: &QueryToken{
						value: "Added*",
					},
				},
			},
			expectedExpression: &BinaryExpression{
				operator: &Operator{
					operator: &QueryToken{
						value:     "GLOB",
						tokenType: QTK_CMP_GLOB,
					},
				},
				lhs: &Identifier{
					identifier: &QueryToken{
						value: "Summary",
					},
				},
				rhs: &GlobLiteral{
					globString: &QueryToken{
						value: "Added*",
					},
				},
			},
		},
	}

	for _, typeConversionTest := range typeConversionTests {
		inputExpression := typeConversionTest.inputExpression
		expectedExpression := typeConversionTest.expectedExpression

		expressionProcessor := NewExpressionProcessor(inputExpression, &TestFieldTypeDescriptor{})
		actualExpression, errors := expressionProcessor.Process()

		if len(errors) > 0 {
			t.Errorf("Process failed with errors: %v", errors)
		} else if !expectedExpression.Equal(actualExpression) {
			t.Errorf("Expression does not match expected value. Expected: %v, Actual: %v", expectedExpression, actualExpression)
		}
	}
}

func TestRegexStringsAreConvertedToRegexLiteralsInRegexFieldContext(t *testing.T) {
	var typeConversionTests = []struct {
		inputExpression    Expression
		expectedExpression Expression
	}{
		{
			inputExpression: &BinaryExpression{
				operator: &Operator{
					operator: &QueryToken{
						value:     "REGEXP",
						tokenType: QTK_CMP_REGEXP,
					},
				},
				lhs: &Identifier{
					identifier: &QueryToken{
						value: "Summary",
					},
				},
				rhs: &StringLiteral{
					value: &QueryToken{
						value: `^Added\s+.*$`,
					},
				},
			},
			expectedExpression: &BinaryExpression{
				operator: &Operator{
					operator: &QueryToken{
						value:     "REGEXP",
						tokenType: QTK_CMP_REGEXP,
					},
				},
				lhs: &Identifier{
					identifier: &QueryToken{
						value: "Summary",
					},
				},
				rhs: &RegexLiteral{
					regex: regexp.MustCompile(`^Added\s+.*$`),
				},
			},
		},
	}

	for _, typeConversionTest := range typeConversionTests {
		inputExpression := typeConversionTest.inputExpression
		expectedExpression := typeConversionTest.expectedExpression

		expressionProcessor := NewExpressionProcessor(inputExpression, &TestFieldTypeDescriptor{})
		actualExpression, errors := expressionProcessor.Process()

		if len(errors) > 0 {
			t.Errorf("Process failed with errors: %v", errors)
		} else if !expectedExpression.Equal(actualExpression) {
			t.Errorf("Expression does not match expected value. Expected: %v, Actual: %v", expectedExpression, actualExpression)
		}
	}
}

func TestExpressionsAreValid(t *testing.T) {
	var validationTests = []struct {
		inputExpression Expression
		expectedErrors  []error
	}{
		{
			inputExpression: &BinaryExpression{
				operator: &Operator{
					operator: &QueryToken{
						value:     "=",
						tokenType: QTK_CMP_EQ,
					},
				},
				lhs: &Identifier{
					identifier: &QueryToken{
						value: "AuthorDate",
					},
				},
				rhs: &StringLiteral{
					value: &QueryToken{
						value: "2017-07-16",
					},
				},
			},
		},
		{
			inputExpression: &BinaryExpression{
				operator: &Operator{
					operator: &QueryToken{
						value:     "AND",
						tokenType: QTK_AND,
						startPos: QueryScannerPos{
							line: 1,
							col:  10,
						},
					},
				},
				lhs: &Identifier{
					identifier: &QueryToken{
						value: "AuthorDate",
					},
				},
				rhs: &BinaryExpression{
					operator: &Operator{
						operator: &QueryToken{
							value:     "=",
							tokenType: QTK_CMP_EQ,
						},
					},
					lhs: &Identifier{
						identifier: &QueryToken{
							value: "AuthorDate",
						},
					},
					rhs: &StringLiteral{
						value: &QueryToken{
							value: "2017-07-16",
						},
					},
				},
			},
			expectedErrors: []error{
				errors.New("1:10: Operands of a logical operator must resolve to boolean values"),
			},
		},
		{
			inputExpression: &BinaryExpression{
				operator: &Operator{
					operator: &QueryToken{
						value:     "=",
						tokenType: QTK_CMP_EQ,
						startPos: QueryScannerPos{
							line: 1,
							col:  5,
						},
					},
				},
				lhs: &Identifier{
					identifier: &QueryToken{
						value: "AuthorName",
					},
				},
				rhs: &ParenExpression{
					expression: &Identifier{
						identifier: &QueryToken{
							value: "AuthorDate",
							startPos: QueryScannerPos{
								line: 1,
								col:  14,
							},
						},
					},
				},
			},
			expectedErrors: []error{
				errors.New("1:14: Expression in parentheses must resolve to a boolean value"),
				errors.New("1:5: Comparison expressions must compare value types"),
			},
		},
		{
			inputExpression: &BinaryExpression{
				operator: &Operator{
					operator: &QueryToken{
						value:     "=",
						tokenType: QTK_CMP_EQ,
						startPos: QueryScannerPos{
							line: 1,
							col:  8,
						},
					},
				},
				lhs: &Identifier{
					identifier: &QueryToken{
						value: "AuthorDate",
					},
				},
				rhs: &StringLiteral{
					value: &QueryToken{
						value: "Invalid Date",
					},
				},
			},
			expectedErrors: []error{
				errors.New("1:8: Attempting to compare different types - LHS Type: Date vs RHS Type: String"),
			},
		},
		{
			inputExpression: &BinaryExpression{
				operator: &Operator{
					operator: &QueryToken{
						value:     "=",
						tokenType: QTK_CMP_EQ,
						startPos: QueryScannerPos{
							line: 1,
							col:  15,
						},
					},
				},
				lhs: &Identifier{
					identifier: &QueryToken{
						value: "AuthorNamey",
						startPos: QueryScannerPos{
							line: 1,
							col:  1,
						},
					},
				},
				rhs: &StringLiteral{
					value: &QueryToken{
						value: "Test Author",
					},
				},
			},
			expectedErrors: []error{
				errors.New("1:1: Invalid field: AuthorNamey"),
			},
		},
		{
			inputExpression: &BinaryExpression{
				operator: &Operator{
					operator: &QueryToken{
						value:     "GLOB",
						tokenType: QTK_CMP_GLOB,
						startPos: QueryScannerPos{
							line: 1,
							col:  15,
						},
					},
				},
				lhs: &Identifier{
					identifier: &QueryToken{
						value: "ParentCount",
						startPos: QueryScannerPos{
							line: 1,
							col:  1,
						},
					},
				},
				rhs: &GlobLiteral{
					globString: &QueryToken{
						value: "Test",
					},
				},
			},
			expectedErrors: []error{
				errors.New("1:15: Argument on LHS has invalid type: Number. Allowed types are: String"),
			},
		},
		{
			inputExpression: &BinaryExpression{
				operator: &Operator{
					operator: &QueryToken{
						value:     "REGEXP",
						tokenType: QTK_CMP_REGEXP,
						startPos: QueryScannerPos{
							line: 1,
							col:  15,
						},
					},
				},
				lhs: &Identifier{
					identifier: &QueryToken{
						value: "AuthorName",
						startPos: QueryScannerPos{
							line: 1,
							col:  1,
						},
					},
				},
				rhs: &StringLiteral{
					value: &QueryToken{
						value: "[Invalid Regex",
					},
				},
			},
			expectedErrors: []error{
				errors.New("1:15: Argument on RHS has invalid type: String. Allowed types are: Regex"),
			},
		},
		{
			inputExpression: &UnaryExpression{
				operator: &Operator{
					operator: &QueryToken{
						value:     "NOT",
						tokenType: QTK_NOT,
						startPos: QueryScannerPos{
							line: 1,
							col:  18,
						},
					},
				},
				expression: &StringLiteral{
					value: &QueryToken{
						value: "Test",
					},
				},
			},
			expectedErrors: []error{
				errors.New("1:18: NOT operator can only be applied to expressions that resolve to a boolean value"),
			},
		},
	}

	for _, validationTest := range validationTests {
		inputExpression := validationTest.inputExpression
		expectedErrors := validationTest.expectedErrors

		expressionProcessor := NewExpressionProcessor(inputExpression, &TestFieldTypeDescriptor{})
		_, actualErrors := expressionProcessor.Process()

		if !reflect.DeepEqual(expectedErrors, actualErrors) {
			t.Errorf("Returned errors do not match expected errors. Expected: %v, Actual: %v", expectedErrors, actualErrors)
		}
	}
}
