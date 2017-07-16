package main

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"time"
)

type FieldTypeDescriptor interface {
	FieldType(fieldName string) (fieldType FieldType, fieldExists bool)
}

type ExpressionProcessor struct {
	expression          Expression
	fieldTypeDescriptor FieldTypeDescriptor
}

func NewExpressionProcessor(expression Expression, fieldTypeDescriptor FieldTypeDescriptor) *ExpressionProcessor {
	return &ExpressionProcessor{
		expression:          expression,
		fieldTypeDescriptor: fieldTypeDescriptor,
	}
}

func (expressionProcessor *ExpressionProcessor) Process() (expression Expression, errors []error) {
	if refinableExpression, ok := expressionProcessor.expression.(RefinableExpression); ok {
		refinableExpression.ConvertTypes(expressionProcessor.fieldTypeDescriptor)
		errors = refinableExpression.Validate(expressionProcessor.fieldTypeDescriptor)
		expression = refinableExpression
	} else {
		errors = append(errors, fmt.Errorf("Expected refinable expression but received expression of type %T", expressionProcessor.expression))
	}

	return
}

const (
	QUERY_DATE_FORMAT      = "2006-01-02"
	QUERY_DATE_TIME_FORMAT = "2006-01-02 15:04:05"
)

var dateFormatPattern = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
var dateTimeFormatPattern = regexp.MustCompile(`^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}$`)

type FieldType int

const (
	FT_INVALID = iota
	FT_STRING
	FT_NUMBER
	FT_DATE
)

var fieldTypeNames = map[FieldType]string{
	FT_INVALID: "Invalid",
	FT_STRING:  "String",
	FT_NUMBER:  "Number",
	FT_DATE:    "Date",
}

type TypeDescriptor interface {
	FieldType(fieldTypeDescriptor FieldTypeDescriptor) FieldType
}

type DateLiteral struct {
	dateTime   time.Time
	stringTime *QueryToken
}

func (dateLiteral *DateLiteral) Equal(expression Expression) bool {
	other, ok := expression.(*DateLiteral)
	if !ok {
		return false
	}

	return dateLiteral.dateTime.Equal(other.dateTime)
}

func (dateLiteral *DateLiteral) String() string {
	return dateLiteral.dateTime.Format(QUERY_DATE_TIME_FORMAT)
}

func (dateLiteral *DateLiteral) Pos() QueryScannerPos {
	return dateLiteral.stringTime.startPos
}

func (dateLiteral *DateLiteral) FieldType(fieldTypeDescriptor FieldTypeDescriptor) FieldType {
	return FT_DATE
}

func (stringLiteral *StringLiteral) FieldType(fieldTypeDescriptor FieldTypeDescriptor) FieldType {
	return FT_STRING
}

func (numberLiteral *NumberLiteral) FieldType(fieldTypeDescriptor FieldTypeDescriptor) FieldType {
	return FT_NUMBER
}

func (identifier *Identifier) FieldType(fieldTypeDescriptor FieldTypeDescriptor) FieldType {
	if fieldType, fieldExists := fieldTypeDescriptor.FieldType(identifier.identifier.value); fieldExists {
		return fieldType
	}

	return FT_INVALID
}

func (identifier *Identifier) Validate(fieldTypeDescriptor FieldTypeDescriptor) (errors []error) {
	if _, fieldExists := fieldTypeDescriptor.FieldType(identifier.identifier.value); !fieldExists {
		errors = append(errors, GenerateExpressionError(identifier, "Invalid field: %v", identifier.identifier.value))
	}

	return
}

type ValidatableExpression interface {
	Validate(FieldTypeDescriptor) []error
}

type RefinableExpression interface {
	Expression
	ValidatableExpression
	ConvertTypes(FieldTypeDescriptor)
}

func GenerateExpressionError(expression Expression, errorMessage string, args ...interface{}) error {
	var buffer bytes.Buffer

	buffer.WriteString(fmt.Sprintf("%v:%v: ", expression.Pos().line, expression.Pos().col))
	buffer.WriteString(fmt.Sprintf(errorMessage, args...))

	return errors.New(buffer.String())
}

func (parenExpression *ParenExpression) ConvertTypes(fieldTypeDescriptor FieldTypeDescriptor) {
	if refinableExpression, ok := parenExpression.expression.(RefinableExpression); ok {
		refinableExpression.ConvertTypes(fieldTypeDescriptor)
	}
}

func (parenExpression *ParenExpression) Validate(fieldTypeDescriptor FieldTypeDescriptor) (errors []error) {
	if _, ok := parenExpression.expression.(RefinableExpression); !ok {
		errors = append(errors, GenerateExpressionError(parenExpression, "Expression in parentheses must resolve to a boolean value"))
	}

	if validatableExpression, ok := parenExpression.expression.(ValidatableExpression); ok {
		errors = append(errors, validatableExpression.Validate(fieldTypeDescriptor)...)
	}

	return
}

func (binaryExpression *BinaryExpression) ConvertTypes(fieldTypeDescriptor FieldTypeDescriptor) {
	if !binaryExpression.IsComparison() {
		if refinableExpression, ok := binaryExpression.lhs.(RefinableExpression); ok {
			refinableExpression.ConvertTypes(fieldTypeDescriptor)
		}

		if refinableExpression, ok := binaryExpression.rhs.(RefinableExpression); ok {
			refinableExpression.ConvertTypes(fieldTypeDescriptor)
		}

		return
	}

	isDateComparison, dateString, datePtr := binaryExpression.isDateComparison(fieldTypeDescriptor)
	if !isDateComparison {
		return
	}

	var dateFormat string

	switch {
	case dateFormatPattern.MatchString(dateString.value.value):
		dateFormat = QUERY_DATE_FORMAT
	case dateTimeFormatPattern.MatchString(dateString.value.value):
		dateFormat = QUERY_DATE_TIME_FORMAT
	default:
		return
	}

	utcDateTime, err := time.Parse(dateFormat, dateString.value.value)
	if err != nil {
		return
	}

	dateTime := time.Date(utcDateTime.Year(), utcDateTime.Month(), utcDateTime.Day(), utcDateTime.Hour(),
		utcDateTime.Minute(), utcDateTime.Second(), utcDateTime.Nanosecond(), time.Local)

	*datePtr = &DateLiteral{
		dateTime:   dateTime,
		stringTime: dateString.value,
	}
}

func (binaryExpression *BinaryExpression) isDateComparison(fieldTypeDescriptor FieldTypeDescriptor) (isDateComparison bool, dateString *StringLiteral, datePtr *Expression) {
	var identifier *Identifier
	var ok bool

	identifier, ok = binaryExpression.lhs.(*Identifier)

	if ok {
		dateString, _ = binaryExpression.rhs.(*StringLiteral)
		datePtr = &binaryExpression.rhs
	} else {
		dateString, _ = binaryExpression.lhs.(*StringLiteral)
		identifier, _ = binaryExpression.rhs.(*Identifier)
		datePtr = &binaryExpression.lhs
	}

	if identifier == nil || dateString == nil {
		return
	}

	fieldType, fieldExists := fieldTypeDescriptor.FieldType(identifier.identifier.value)
	if !fieldExists || fieldType != FT_DATE {
		return
	}

	isDateComparison = true

	return
}

func (binaryExpression *BinaryExpression) Validate(fieldTypeDescriptor FieldTypeDescriptor) (errors []error) {
	if !binaryExpression.IsComparison() {
		if refinableExpression, ok := binaryExpression.lhs.(RefinableExpression); !ok {
			errors = append(errors, GenerateExpressionError(binaryExpression, "Operands of a logical operator must resolve to boolean values"))
		} else {
			errors = append(errors, refinableExpression.Validate(fieldTypeDescriptor)...)
		}

		if refinableExpression, ok := binaryExpression.rhs.(RefinableExpression); !ok {
			errors = append(errors, GenerateExpressionError(binaryExpression, "Operands of a logical operator must resolve to boolean values"))
		} else {
			errors = append(errors, refinableExpression.Validate(fieldTypeDescriptor)...)
		}

		return
	}

	if validatableExpression, ok := binaryExpression.lhs.(ValidatableExpression); ok {
		errors = append(errors, validatableExpression.Validate(fieldTypeDescriptor)...)
	}

	if validatableExpression, ok := binaryExpression.rhs.(ValidatableExpression); ok {
		errors = append(errors, validatableExpression.Validate(fieldTypeDescriptor)...)
	}

	lhsType, isLhsValueType := determineFieldType(binaryExpression.lhs, fieldTypeDescriptor)
	rhsType, isRhsValueType := determineFieldType(binaryExpression.rhs, fieldTypeDescriptor)

	if !(isLhsValueType && isRhsValueType) {
		errors = append(errors, GenerateExpressionError(binaryExpression, "Comparison expressions must compare value types"))
	} else if lhsType != rhsType {
		errors = append(errors, GenerateExpressionError(binaryExpression, "Attempting to compare different types - LHS Type: %v vs RHS Type: %v",
			fieldTypeNames[lhsType], fieldTypeNames[rhsType]))
	}

	return
}

func determineFieldType(expression Expression, fieldTypeDescriptor FieldTypeDescriptor) (fieldType FieldType, isValueType bool) {
	if typeDescriptor, ok := expression.(TypeDescriptor); ok {
		fieldType = typeDescriptor.FieldType(fieldTypeDescriptor)
		isValueType = true
	}

	return
}
