package main

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	glob "github.com/gobwas/glob"
)

// FieldTypeDescriptor provides the type of the provided field (if it exists)
type FieldTypeDescriptor interface {
	FieldType(fieldName string) (fieldType FieldType, fieldExists bool)
}

// ExpressionProcessor takes the query expression that has been parsed and processes it further
// Type conversion and validation of the expression are performed
type ExpressionProcessor struct {
	expression          Expression
	fieldTypeDescriptor FieldTypeDescriptor
}

// NewExpressionProcessor creates an expression processor instance for the provided expression
func NewExpressionProcessor(expression Expression, fieldTypeDescriptor FieldTypeDescriptor) *ExpressionProcessor {
	return &ExpressionProcessor{
		expression:          expression,
		fieldTypeDescriptor: fieldTypeDescriptor,
	}
}

// Process performs type conversion and validates the expression
func (expressionProcessor *ExpressionProcessor) Process() (expression Expression, errors []error) {
	if logicalExpression, ok := expressionProcessor.expression.(LogicalExpression); ok {
		log.Debugf("Performing type conversion on expression comparisons")
		if errors = logicalExpression.ConvertTypes(expressionProcessor.fieldTypeDescriptor); len(errors) > 0 {
			log.Debugf("ConvertTypes returned errors")
			return
		}

		log.Debugf("Performing expression validation")
		if errors = logicalExpression.Validate(expressionProcessor.fieldTypeDescriptor); len(errors) > 0 {
			log.Debugf("Expression is not valid")
			return
		}

		log.Debugf("Finished processing expression")
		expression = logicalExpression
	} else {
		errors = append(errors, fmt.Errorf("Expected logical expression but received expression of type %v",
			reflect.TypeOf(expressionProcessor.expression).Elem().Name()))
	}

	return
}

type binaryOperatorPosition int

const (
	bopLeft = iota
	bopRight
)

var operatorAllowedOperandTypes = map[QueryTokenType]map[binaryOperatorPosition]map[FieldType]bool{
	QtkCmpGlob: {
		bopLeft: {
			FtString: true,
		},
		bopRight: {
			FtGlob: true,
		},
	},
	QtkCmpRegexp: {
		bopLeft: {
			FtString: true,
		},
		bopRight: {
			FtRegex: true,
		},
	},
}

func (operator *Operator) isOperandTypeRestricted() bool {
	_, isRestricted := operatorAllowedOperandTypes[operator.operator.tokenType]
	return isRestricted
}

func (operator *Operator) isValidArgument(operatorPosition binaryOperatorPosition, operandType FieldType) bool {
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

func (operator *Operator) allowedTypes(operatorPosition binaryOperatorPosition) (fieldTypes []FieldType) {
	allowedOperandTypes, ok := operatorAllowedOperandTypes[operator.operator.tokenType]
	if !ok {
		return
	}

	allowedTypes, ok := allowedOperandTypes[operatorPosition]
	if !ok {
		return
	}

	for fieldType := range allowedTypes {
		fieldTypes = append(fieldTypes, fieldType)
	}

	return
}

const (
	queryDateFormat     = "2006-01-02"
	queryDateTimeFormat = "2006-01-02 15:04:05"
)

var dateFormatPattern = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
var dateTimeFormatPattern = regexp.MustCompile(`^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}$`)

// FieldType represents the data type of a field
type FieldType int

// The set of supported field types
const (
	FtInvalid = iota
	FtString
	FtNumber
	FtDate
	FtGlob
	FtRegex
)

var fieldTypeNames = map[FieldType]string{
	FtInvalid: "Invalid",
	FtString:  "String",
	FtNumber:  "Number",
	FtDate:    "Date",
	FtGlob:    "Glob",
	FtRegex:   "Regex",
}

// FieldTypeName returns the name of the provided FieldType
func FieldTypeName(fieldType FieldType) (fieldTypeName string) {
	fieldTypeName, _ = fieldTypeNames[fieldType]
	return
}

// TypeDescriptor returns the type of a field or value
type TypeDescriptor interface {
	FieldType(fieldTypeDescriptor FieldTypeDescriptor) FieldType
}

// DateLiteral represents a date value
type DateLiteral struct {
	dateTime   time.Time
	stringTime *QueryToken
}

// Equal returns true if the provided expression is equal
func (dateLiteral *DateLiteral) Equal(expression Expression) bool {
	other, ok := expression.(*DateLiteral)
	if !ok {
		return false
	}

	return dateLiteral.dateTime.Equal(other.dateTime)
}

// String converts the date value into a string format
func (dateLiteral *DateLiteral) String() string {
	return "Date{" + dateLiteral.dateTime.Format(queryDateTimeFormat) + "}"
}

// Pos returns the position this date appeared at in the input stream
func (dateLiteral *DateLiteral) Pos() QueryScannerPos {
	return dateLiteral.stringTime.startPos
}

// FieldType retruns the data type of this value
func (dateLiteral *DateLiteral) FieldType(fieldTypeDescriptor FieldTypeDescriptor) FieldType {
	return FtDate
}

// RegexLiteral represents a regex value
type RegexLiteral struct {
	regex       *regexp.Regexp
	regexString *QueryToken
}

// Equal returns true if the provided expression is equal
func (regexLiteral *RegexLiteral) Equal(expression Expression) bool {
	other, ok := expression.(*RegexLiteral)
	if !ok {
		return false
	}

	return regexLiteral.regex.String() == other.regex.String()
}

// String returns the regex string used to construct this instance
func (regexLiteral *RegexLiteral) String() string {
	return "Regex{" + regexLiteral.regex.String() + "}"
}

// Pos returns the position this regex appeared in the input stream
func (regexLiteral *RegexLiteral) Pos() QueryScannerPos {
	return regexLiteral.regexString.startPos
}

// FieldType retruns the data type of this value
func (regexLiteral *RegexLiteral) FieldType(fieldTypeDescriptor FieldTypeDescriptor) FieldType {
	return FtRegex
}

// GlobLiteral represents a glob value
type GlobLiteral struct {
	glob       glob.Glob
	globString *QueryToken
}

// Equal returns true if the provided expression is equal
func (globLiteral *GlobLiteral) Equal(expression Expression) bool {
	other, ok := expression.(*GlobLiteral)
	if !ok {
		return false
	}

	return globLiteral.globString.value == other.globString.value
}

// String returns the string representation of the glob
func (globLiteral *GlobLiteral) String() string {
	return "Glob{" + globLiteral.globString.value + "}"
}

// Pos returns the position the glob appeared in the input stream
func (globLiteral *GlobLiteral) Pos() QueryScannerPos {
	return globLiteral.globString.startPos
}

// FieldType returns the data type of this value
func (globLiteral *GlobLiteral) FieldType(fieldTypeDescriptor FieldTypeDescriptor) FieldType {
	return FtGlob
}

// FieldType returns the data type of this value
func (stringLiteral *StringLiteral) FieldType(fieldTypeDescriptor FieldTypeDescriptor) FieldType {
	return FtString
}

// FieldType returns the data type of this value
func (numberLiteral *NumberLiteral) FieldType(fieldTypeDescriptor FieldTypeDescriptor) FieldType {
	return FtNumber
}

// FieldType returns the data type of the field represented by this identifier
func (identifier *Identifier) FieldType(fieldTypeDescriptor FieldTypeDescriptor) FieldType {
	if fieldType, fieldExists := fieldTypeDescriptor.FieldType(identifier.identifier.value); fieldExists {
		return fieldType
	}

	return FtInvalid
}

// Validate that this identifier represents a valid field
func (identifier *Identifier) Validate(fieldTypeDescriptor FieldTypeDescriptor) (errors []error) {
	if _, fieldExists := fieldTypeDescriptor.FieldType(identifier.identifier.value); !fieldExists {
		errors = append(errors, GenerateExpressionError(identifier, "Invalid field: %v", identifier.identifier.value))
	}

	return
}

// ValidatableExpression is an expression which can be validated for correctness
type ValidatableExpression interface {
	Validate(FieldTypeDescriptor) []error
}

// LogicalExpression is an expression which resolves to a boolean value and is composed of child expressions
type LogicalExpression interface {
	Expression
	ValidatableExpression
	ConvertTypes(FieldTypeDescriptor) []error
}

// GenerateExpressionError generates an error with expression position information included
func GenerateExpressionError(expression Expression, errorMessage string, args ...interface{}) error {
	var buffer bytes.Buffer

	buffer.WriteString(fmt.Sprintf("%v:%v: ", expression.Pos().line, expression.Pos().col))
	buffer.WriteString(fmt.Sprintf(errorMessage, args...))

	return errors.New(buffer.String())
}

// ConvertTypes defers the call to the child expression if it is a logical expression
func (parenExpression *ParenExpression) ConvertTypes(fieldTypeDescriptor FieldTypeDescriptor) (errors []error) {
	if logicalExpression, ok := parenExpression.expression.(LogicalExpression); ok {
		errors = append(errors, logicalExpression.ConvertTypes(fieldTypeDescriptor)...)
	}

	return
}

// Validate checks the child expression is valid
func (parenExpression *ParenExpression) Validate(fieldTypeDescriptor FieldTypeDescriptor) (errors []error) {
	if _, ok := parenExpression.expression.(LogicalExpression); !ok {
		errors = append(errors, GenerateExpressionError(parenExpression, "Expression in parentheses must resolve to a boolean value"))
	}

	if validatableExpression, ok := parenExpression.expression.(ValidatableExpression); ok {
		errors = append(errors, validatableExpression.Validate(fieldTypeDescriptor)...)
	}

	return
}

// ConvertTypes defers the call to the child expression
func (unaryExpression *UnaryExpression) ConvertTypes(fieldTypeDescriptor FieldTypeDescriptor) (errors []error) {
	if logicalExpression, ok := unaryExpression.expression.(LogicalExpression); ok {
		errors = append(errors, logicalExpression.ConvertTypes(fieldTypeDescriptor)...)
	}

	return
}

// Validate checks the child expression is valid
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

// ConvertTypes defers the call to the child expressions if they're logical
// Otherwise performs type conversion on the child expressions if necessary
func (binaryExpression *BinaryExpression) ConvertTypes(fieldTypeDescriptor FieldTypeDescriptor) (errors []error) {
	if !binaryExpression.IsComparison() {
		if logicalExpression, ok := binaryExpression.lhs.(LogicalExpression); ok {
			errors = append(errors, logicalExpression.ConvertTypes(fieldTypeDescriptor)...)
		}

		if logicalExpression, ok := binaryExpression.rhs.(LogicalExpression); ok {
			errors = append(errors, logicalExpression.ConvertTypes(fieldTypeDescriptor)...)
		}

		return
	}

	if err := binaryExpression.processDateComparison(fieldTypeDescriptor); err != nil {
		errors = append(errors, err)
	} else if err := binaryExpression.processGlobComparison(fieldTypeDescriptor); err != nil {
		errors = append(errors, err)
	} else if err := binaryExpression.processRegexComparison(fieldTypeDescriptor); err != nil {
		errors = append(errors, err)
	}

	return
}

func (binaryExpression *BinaryExpression) processDateComparison(fieldTypeDescriptor FieldTypeDescriptor) (err error) {
	isDateComparison, dateString, datePtr := binaryExpression.isDateComparison(fieldTypeDescriptor)
	if !isDateComparison {
		return
	}

	var dateFormat string

	switch {
	case dateFormatPattern.MatchString(dateString.value.value):
		dateFormat = queryDateFormat
	case dateTimeFormatPattern.MatchString(dateString.value.value):
		dateFormat = queryDateTimeFormat
	default:
		return GenerateExpressionError(dateString, "Invalid date: %v. Format must be either %v or %v",
			dateString.value.value, queryDateFormat, queryDateTimeFormat)
	}

	utcDateTime, err := time.Parse(dateFormat, dateString.value.value)
	if err != nil {
		return GenerateExpressionError(dateString, "Unable to parse date %v: %v", dateString.value.value, err)
	}

	dateTime := TimeWithLocation(utcDateTime, time.Local)

	*datePtr = &DateLiteral{
		dateTime:   dateTime,
		stringTime: dateString.value,
	}

	return
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
	if !fieldExists || fieldType != FtDate {
		return
	}

	isDateComparison = true

	return
}

func (binaryExpression *BinaryExpression) processGlobComparison(fieldTypeDescriptor FieldTypeDescriptor) (err error) {
	isGlobComparison, globString, globPtr := binaryExpression.isGlobComparison(fieldTypeDescriptor)
	if !isGlobComparison {
		return
	}

	glob, err := glob.Compile(globString.value.value)
	if err != nil {
		return GenerateExpressionError(globString, "Invalid glob %v: %v", globString.value.value, err)
	}

	*globPtr = &GlobLiteral{
		glob:       glob,
		globString: globString.value,
	}

	return
}

func (binaryExpression *BinaryExpression) isGlobComparison(fieldTypeDescriptor FieldTypeDescriptor) (isGlobComparison bool, globString *StringLiteral, globPtr *Expression) {
	if binaryExpression.operator.operator.tokenType != QtkCmpGlob {
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
	if !fieldExists || fieldType != FtString {
		return
	}

	isGlobComparison = true

	return
}

func (binaryExpression *BinaryExpression) processRegexComparison(fieldTypeDescriptor FieldTypeDescriptor) (err error) {
	isRegexComparison, regexString, regexPtr := binaryExpression.isRegexComparison(fieldTypeDescriptor)
	if !isRegexComparison {
		return
	}

	regex, err := regexp.Compile(regexString.value.value)
	if err != nil {
		return GenerateExpressionError(regexString, "Invalid regex %v: %v", regexString.value.value, err)
	}

	*regexPtr = &RegexLiteral{
		regex:       regex,
		regexString: regexString.value,
	}

	return
}

func (binaryExpression *BinaryExpression) isRegexComparison(fieldTypeDescriptor FieldTypeDescriptor) (isRegexComparison bool, regexString *StringLiteral, regexPtr *Expression) {
	if binaryExpression.operator.operator.tokenType != QtkCmpRegexp {
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
	if !fieldExists || fieldType != FtString {
		return
	}

	isRegexComparison = true

	return
}

// Validate the child expressions and operator are valid
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

	lhsType, isLHSValueType := determineFieldType(binaryExpression.lhs, fieldTypeDescriptor)
	rhsType, isRHSValueType := determineFieldType(binaryExpression.rhs, fieldTypeDescriptor)

	if !(isLHSValueType && isRHSValueType) {
		errors = append(errors, GenerateExpressionError(binaryExpression, "Comparison expressions must compare value types"))
	} else if binaryExpression.operator.isOperandTypeRestricted() {
		if !(lhsType == FtInvalid || rhsType == FtInvalid) {
			if !binaryExpression.operator.isValidArgument(bopLeft, lhsType) {
				errors = append(errors, GenerateExpressionError(binaryExpression, "Argument on LHS has invalid type: %v. Allowed types are: %v",
					fieldTypeNames[lhsType], fieldTypeNamesString(binaryExpression.operator.allowedTypes(bopLeft))))
			}

			if !binaryExpression.operator.isValidArgument(bopRight, rhsType) {
				errors = append(errors, GenerateExpressionError(binaryExpression, "Argument on RHS has invalid type: %v. Allowed types are: %v",
					fieldTypeNames[rhsType], fieldTypeNamesString(binaryExpression.operator.allowedTypes(bopRight))))
			}
		}
	} else if lhsType != rhsType && !(lhsType == FtInvalid || rhsType == FtInvalid) {
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
