package main

import (
	"bytes"
	"errors"
	"fmt"
	glob "github.com/gobwas/glob"
	"reflect"
	"regexp"
	"strings"
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
	if logicalExpression, ok := expressionProcessor.expression.(LogicalExpression); ok {
		logicalExpression.ConvertTypes(expressionProcessor.fieldTypeDescriptor)
		errors = logicalExpression.Validate(expressionProcessor.fieldTypeDescriptor)
		expression = logicalExpression
	} else {
		errors = append(errors, fmt.Errorf("Expected logical expression but received expression of type %v",
			reflect.TypeOf(expressionProcessor.expression).Elem().Name()))
	}

	return
}

type BinaryOperatorPosition int

const (
	BOP_LEFT = iota
	BOP_RIGHT
)

var operatorAllowedOperandTypes = map[QueryTokenType]map[BinaryOperatorPosition]map[FieldType]bool{
	QTK_CMP_GLOB: map[BinaryOperatorPosition]map[FieldType]bool{
		BOP_LEFT: map[FieldType]bool{
			FT_STRING: true,
		},
		BOP_RIGHT: map[FieldType]bool{
			FT_GLOB: true,
		},
	},
	QTK_CMP_REGEXP: map[BinaryOperatorPosition]map[FieldType]bool{
		BOP_LEFT: map[FieldType]bool{
			FT_STRING: true,
		},
		BOP_RIGHT: map[FieldType]bool{
			FT_REGEX: true,
		},
	},
}

func (operator *Operator) IsOperandTypeRestricted() bool {
	_, isRestricted := operatorAllowedOperandTypes[operator.operator.tokenType]
	return isRestricted
}

func (operator *Operator) IsValidArgument(operatorPosition BinaryOperatorPosition, operandType FieldType) bool {
	allowedOperandTypes, ok := operatorAllowedOperandTypes[operator.operator.tokenType]
	if !ok {
		return true
	}

	allowedTypes, ok := allowedOperandTypes[operatorPosition]
	if !ok {
		return true
	}

	_, isAllowedType := allowedTypes[operandType]

	return isAllowedType
}

func (operator *Operator) AllowedTypes(operatorPosition BinaryOperatorPosition) (fieldTypes []FieldType) {
	allowedOperandTypes, ok := operatorAllowedOperandTypes[operator.operator.tokenType]
	if !ok {
		return
	}

	allowedTypes, ok := allowedOperandTypes[operatorPosition]
	if !ok {
		return
	}

	for fieldType, _ := range allowedTypes {
		fieldTypes = append(fieldTypes, fieldType)
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
	FT_GLOB
	FT_REGEX
)

var fieldTypeNames = map[FieldType]string{
	FT_INVALID: "Invalid",
	FT_STRING:  "String",
	FT_NUMBER:  "Number",
	FT_DATE:    "Date",
	FT_GLOB:    "Glob",
	FT_REGEX:   "Regex",
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

type RegexLiteral struct {
	regex       *regexp.Regexp
	regexString *QueryToken
}

func (regexLiteral *RegexLiteral) Equal(expression Expression) bool {
	other, ok := expression.(*RegexLiteral)
	if !ok {
		return false
	}

	return regexLiteral.regex.String() == other.regex.String()
}

func (regexLiteral *RegexLiteral) String() string {
	return regexLiteral.String()
}

func (regexLiteral *RegexLiteral) Pos() QueryScannerPos {
	return regexLiteral.regexString.startPos
}

func (regexLiteral *RegexLiteral) FieldType(fieldTypeDescriptor FieldTypeDescriptor) FieldType {
	return FT_REGEX
}

type GlobLiteral struct {
	glob       glob.Glob
	globString *QueryToken
}

func (globLiteral *GlobLiteral) Equal(expression Expression) bool {
	other, ok := expression.(*GlobLiteral)
	if !ok {
		return false
	}

	return globLiteral.globString.value == other.globString.value
}

func (globLiteral *GlobLiteral) String() string {
	return globLiteral.globString.value
}

func (globLiteral *GlobLiteral) Pos() QueryScannerPos {
	return globLiteral.globString.startPos
}

func (globLiteral *GlobLiteral) FieldType(fieldTypeDescriptor FieldTypeDescriptor) FieldType {
	return FT_GLOB
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

type LogicalExpression interface {
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
	if logicalExpression, ok := parenExpression.expression.(LogicalExpression); ok {
		logicalExpression.ConvertTypes(fieldTypeDescriptor)
	}
}

func (parenExpression *ParenExpression) Validate(fieldTypeDescriptor FieldTypeDescriptor) (errors []error) {
	if _, ok := parenExpression.expression.(LogicalExpression); !ok {
		errors = append(errors, GenerateExpressionError(parenExpression, "Expression in parentheses must resolve to a boolean value"))
	}

	if validatableExpression, ok := parenExpression.expression.(ValidatableExpression); ok {
		errors = append(errors, validatableExpression.Validate(fieldTypeDescriptor)...)
	}

	return
}

func (unaryExpression *UnaryExpression) ConvertTypes(fieldTypeDescriptor FieldTypeDescriptor) {
	if logicalExpression, ok := unaryExpression.expression.(LogicalExpression); ok {
		logicalExpression.ConvertTypes(fieldTypeDescriptor)
	}
}

func (unaryExpression *UnaryExpression) Validate(fieldTypeDescriptor FieldTypeDescriptor) (errors []error) {
	if _, ok := unaryExpression.expression.(LogicalExpression); !ok {
		errors = append(errors, GenerateExpressionError(unaryExpression,
			"%v operator can only be applied to expressions that resolve to a boolean value",
			unaryExpression.operator.operator.value))
	}

	if validatableExpression, ok := unaryExpression.expression.(ValidatableExpression); ok {
		errors = append(errors, validatableExpression.Validate(fieldTypeDescriptor)...)
	}

	return
}

func (binaryExpression *BinaryExpression) ConvertTypes(fieldTypeDescriptor FieldTypeDescriptor) {
	if !binaryExpression.IsComparison() {
		if logicalExpression, ok := binaryExpression.lhs.(LogicalExpression); ok {
			logicalExpression.ConvertTypes(fieldTypeDescriptor)
		}

		if logicalExpression, ok := binaryExpression.rhs.(LogicalExpression); ok {
			logicalExpression.ConvertTypes(fieldTypeDescriptor)
		}

		return
	}

	binaryExpression.processDateComparison(fieldTypeDescriptor)
	binaryExpression.processGlobComparison(fieldTypeDescriptor)
	binaryExpression.processRegexComparison(fieldTypeDescriptor)
}

func (binaryExpression *BinaryExpression) processDateComparison(fieldTypeDescriptor FieldTypeDescriptor) {
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

func (binaryExpression *BinaryExpression) processGlobComparison(fieldTypeDescriptor FieldTypeDescriptor) {
	isGlobComparison, globString, globPtr := binaryExpression.isGlobComparison(fieldTypeDescriptor)
	if !isGlobComparison {
		return
	}

	glob, err := glob.Compile(globString.value.value)
	if err != nil {
		return
	}

	*globPtr = &GlobLiteral{
		glob:       glob,
		globString: globString.value,
	}
}

func (binaryExpression *BinaryExpression) isGlobComparison(fieldTypeDescriptor FieldTypeDescriptor) (isGlobComparison bool, globString *StringLiteral, globPtr *Expression) {
	if binaryExpression.operator.operator.tokenType != QTK_CMP_GLOB {
		return
	}

	identifier, ok := binaryExpression.lhs.(*Identifier)

	if ok {
		globString, _ = binaryExpression.rhs.(*StringLiteral)
		globPtr = &binaryExpression.rhs
	} else {
		globString, _ = binaryExpression.lhs.(*StringLiteral)
		identifier, _ = binaryExpression.rhs.(*Identifier)
		globPtr = &binaryExpression.lhs
	}

	if identifier == nil || globString == nil {
		return
	}

	fieldType, fieldExists := fieldTypeDescriptor.FieldType(identifier.identifier.value)
	if !fieldExists || fieldType != FT_STRING {
		return
	}

	isGlobComparison = true

	return
}

func (binaryExpression *BinaryExpression) processRegexComparison(fieldTypeDescriptor FieldTypeDescriptor) {
	isRegexComparison, regexString, regexPtr := binaryExpression.isRegexComparison(fieldTypeDescriptor)
	if !isRegexComparison {
		return
	}

	regex, err := regexp.Compile(regexString.value.value)
	if err != nil {
		return
	}

	*regexPtr = &RegexLiteral{
		regex:       regex,
		regexString: regexString.value,
	}
}

func (binaryExpression *BinaryExpression) isRegexComparison(fieldTypeDescriptor FieldTypeDescriptor) (isRegexComparison bool, regexString *StringLiteral, regexPtr *Expression) {
	if binaryExpression.operator.operator.tokenType != QTK_CMP_REGEXP {
		return
	}

	identifier, ok := binaryExpression.lhs.(*Identifier)

	if ok {
		regexString, _ = binaryExpression.rhs.(*StringLiteral)
		regexPtr = &binaryExpression.rhs
	} else {
		regexString, _ = binaryExpression.lhs.(*StringLiteral)
		identifier, _ = binaryExpression.rhs.(*Identifier)
		regexPtr = &binaryExpression.lhs
	}

	if identifier == nil || regexString == nil {
		return
	}

	fieldType, fieldExists := fieldTypeDescriptor.FieldType(identifier.identifier.value)
	if !fieldExists || fieldType != FT_STRING {
		return
	}

	isRegexComparison = true

	return
}

func (binaryExpression *BinaryExpression) Validate(fieldTypeDescriptor FieldTypeDescriptor) (errors []error) {
	if !binaryExpression.IsComparison() {
		if logicalExpression, ok := binaryExpression.lhs.(LogicalExpression); !ok {
			errors = append(errors, GenerateExpressionError(binaryExpression, "Operands of a logical operator must resolve to boolean values"))
		} else {
			errors = append(errors, logicalExpression.Validate(fieldTypeDescriptor)...)
		}

		if logicalExpression, ok := binaryExpression.rhs.(LogicalExpression); !ok {
			errors = append(errors, GenerateExpressionError(binaryExpression, "Operands of a logical operator must resolve to boolean values"))
		} else {
			errors = append(errors, logicalExpression.Validate(fieldTypeDescriptor)...)
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
	} else if binaryExpression.operator.IsOperandTypeRestricted() {
		if !(lhsType == FT_INVALID || rhsType == FT_INVALID) {
			if !binaryExpression.operator.IsValidArgument(BOP_LEFT, lhsType) {
				errors = append(errors, GenerateExpressionError(binaryExpression, "Argument on LHS has invalid type: %v. Allowed types are: %v",
					fieldTypeNames[lhsType], fieldTypeNamesString(binaryExpression.operator.AllowedTypes(BOP_LEFT))))
			}

			if !binaryExpression.operator.IsValidArgument(BOP_RIGHT, rhsType) {
				errors = append(errors, GenerateExpressionError(binaryExpression, "Argument on RHS has invalid type: %v. Allowed types are: %v",
					fieldTypeNames[rhsType], fieldTypeNamesString(binaryExpression.operator.AllowedTypes(BOP_RIGHT))))
			}
		}
	} else if lhsType != rhsType && !(lhsType == FT_INVALID || rhsType == FT_INVALID) {
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

func fieldTypeNamesString(fieldTypes []FieldType) string {
	var typeNames []string

	for _, fieldType := range fieldTypes {
		if fieldTypeName, ok := fieldTypeNames[fieldType]; ok {
			typeNames = append(typeNames, fieldTypeName)
		}
	}

	return strings.Join(typeNames, ", ")
}
