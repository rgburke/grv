package main

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	glob "github.com/gobwas/glob"
)

// Filter is a function which returns true if the argument matches the filter and false otherwise
type Filter func(interface{}) bool

// FieldDescriptor describes field data for an object type
type FieldDescriptor interface {
	FieldTypeDescriptor
	FieldValue(inputValue interface{}, fieldName string) interface{}
}

// CreateFilter constructs a filter instance from the provided query and field information
func CreateFilter(query string, fieldDescriptor FieldDescriptor) (filter Filter, errors []error) {
	queryParser := NewQueryParser(strings.NewReader(query))

	expression, eof, err := queryParser.Parse()
	if err != nil {
		log.Debugf("Errors encountered when parsing query")
		errors = append(errors, err)
		return
	} else if eof {
		return
	}

	log.Debugf("Received query: %v", expression)

	expressionProcessor := NewExpressionProcessor(expression, fieldDescriptor)

	if expression, errors = expressionProcessor.Process(); len(errors) > 0 {
		log.Debugf("Errors encountered when processing query")
		return
	}

	log.Infof("Creating filter for processed expression: %v", expression)

	filterGenerator := NewFilterGenerator(expression, fieldDescriptor)

	filter, err = filterGenerator.GenerateFilter()
	if err != nil {
		log.Debugf("Errors encountered when generating filter from expression")
		errors = append(errors, err)
		return
	}

	return
}

// FilterGenerator generates a filter from a parsed and processed expression
type FilterGenerator struct {
	expression      Expression
	fieldDescriptor FieldDescriptor
}

// NewFilterGenerator creates a FilterGenerator instance
func NewFilterGenerator(expression Expression, fieldDescriptor FieldDescriptor) *FilterGenerator {
	return &FilterGenerator{
		expression:      expression,
		fieldDescriptor: fieldDescriptor,
	}
}

// GenerateFilter generates a filter from the provided expression and field information
func (filterGenerator *FilterGenerator) GenerateFilter() (filter Filter, err error) {
	if filterGeneratorExpression, ok := filterGenerator.expression.(filterGeneratorExpression); ok {
		filter = filterGeneratorExpression.generateFilter(filterGenerator.fieldDescriptor)
	} else {
		err = fmt.Errorf("Expected filter generator expression but received expression of type %T", filterGenerator.expression)
	}

	return
}

type filterGeneratorExpression interface {
	generateFilter(FieldDescriptor) Filter
}

func (parenExpression *ParenExpression) generateFilter(fieldDescriptor FieldDescriptor) Filter {
	filterGeneratorExpression := parenExpression.expression.(filterGeneratorExpression)
	return filterGeneratorExpression.generateFilter(fieldDescriptor)
}

func (unaryExpression *UnaryExpression) generateFilter(fieldDescriptor FieldDescriptor) Filter {
	filterGeneratorExpression := unaryExpression.expression.(filterGeneratorExpression)
	filter := filterGeneratorExpression.generateFilter(fieldDescriptor)

	switch unaryExpression.operator.operator.tokenType {
	case QtkNot:
		return func(inputValue interface{}) bool {
			return !filter(inputValue)
		}
	}

	panic(fmt.Sprintf("Encountered invalid operator: %v", unaryExpression.operator.operator.value))
}

func (binaryExpression *BinaryExpression) generateFilter(fieldDescriptor FieldDescriptor) Filter {
	if !binaryExpression.IsComparison() {
		lhs := binaryExpression.lhs.(filterGeneratorExpression).generateFilter(fieldDescriptor)
		rhs := binaryExpression.rhs.(filterGeneratorExpression).generateFilter(fieldDescriptor)

		switch binaryExpression.operator.operator.tokenType {
		case QtkAnd:
			return func(inputValue interface{}) bool {
				return lhs(inputValue) && rhs(inputValue)
			}
		case QtkOr:
			return func(inputValue interface{}) bool {
				return lhs(inputValue) || rhs(inputValue)
			}
		default:
			panic(fmt.Sprintf("Encountered invalid operator: %v", binaryExpression.operator.operator.value))
		}
	}

	lhs := binaryExpression.lhs.(valueType)
	rhs := binaryExpression.rhs.(valueType)

	var comparator fieldComparator

	switch binaryExpression.operator.operator.tokenType {
	case QtkCmpGlob:
		comparator = globComparator
	case QtkCmpRegexp:
		comparator = regexpComparator
	default:
		comparator = basicFieldComparators[binaryExpression.operator.operator.tokenType][lhs.FieldType(fieldDescriptor)]
	}

	return func(inputValue interface{}) bool {
		return comparator(lhs.getValue(inputValue, fieldDescriptor), rhs.getValue(inputValue, fieldDescriptor))
	}
}

type valueType interface {
	TypeDescriptor
	getValue(inputValue interface{}, fieldDescriptor FieldDescriptor) interface{}
}

func (stringLiteral *StringLiteral) getValue(inputValue interface{}, fieldDescriptor FieldDescriptor) interface{} {
	return stringLiteral.value.value
}

func (numberLiteral *NumberLiteral) getValue(inputValue interface{}, fieldDescriptor FieldDescriptor) interface{} {
	return numberLiteral.number
}

func (dateLiteral *DateLiteral) getValue(inputValue interface{}, fieldDescriptor FieldDescriptor) interface{} {
	return dateLiteral.dateTime
}

func (globLiteral *GlobLiteral) getValue(inputValue interface{}, fieldDescriptor FieldDescriptor) interface{} {
	return globLiteral.glob
}

func (regexLiteral *RegexLiteral) getValue(inputValue interface{}, fieldDescriptor FieldDescriptor) interface{} {
	return regexLiteral.regex
}

func (identifier *Identifier) getValue(inputValue interface{}, fieldDescriptor FieldDescriptor) interface{} {
	return fieldDescriptor.FieldValue(inputValue, identifier.identifier.value)
}

type fieldComparator func(interface{}, interface{}) bool

var basicFieldComparators = map[QueryTokenType]map[FieldType]fieldComparator{
	QtkCmpEq: {
		FtNumber: func(value1 interface{}, value2 interface{}) bool {
			num1 := value1.(float64)
			num2 := value2.(float64)

			return num1 == num2
		},
		FtString: func(value1 interface{}, value2 interface{}) bool {
			str1 := value1.(string)
			str2 := value2.(string)

			return str1 == str2
		},
		FtDate: func(value1 interface{}, value2 interface{}) bool {
			time1 := value1.(time.Time)
			time2 := value2.(time.Time)

			return time1.Equal(time2)
		},
	},
	QtkCmpNe: {
		FtNumber: func(value1 interface{}, value2 interface{}) bool {
			num1 := value1.(float64)
			num2 := value2.(float64)

			return num1 != num2
		},
		FtString: func(value1 interface{}, value2 interface{}) bool {
			str1 := value1.(string)
			str2 := value2.(string)

			return str1 != str2
		},
		FtDate: func(value1 interface{}, value2 interface{}) bool {
			time1 := value1.(time.Time)
			time2 := value2.(time.Time)

			return !time1.Equal(time2)
		},
	},
	QtkCmpGt: {
		FtNumber: func(value1 interface{}, value2 interface{}) bool {
			num1 := value1.(float64)
			num2 := value2.(float64)

			return num1 > num2
		},
		FtString: func(value1 interface{}, value2 interface{}) bool {
			str1 := value1.(string)
			str2 := value2.(string)

			return str1 > str2
		},
		FtDate: func(value1 interface{}, value2 interface{}) bool {
			time1 := value1.(time.Time)
			time2 := value2.(time.Time)

			return time1.After(time2)
		},
	},
	QtkCmpGe: {
		FtNumber: func(value1 interface{}, value2 interface{}) bool {
			num1 := value1.(float64)
			num2 := value2.(float64)

			return num1 >= num2
		},
		FtString: func(value1 interface{}, value2 interface{}) bool {
			str1 := value1.(string)
			str2 := value2.(string)

			return str1 >= str2
		},
		FtDate: func(value1 interface{}, value2 interface{}) bool {
			time1 := value1.(time.Time)
			time2 := value2.(time.Time)

			return time1.After(time2) || time1.Equal(time2)
		},
	},
	QtkCmpLt: {
		FtNumber: func(value1 interface{}, value2 interface{}) bool {
			num1 := value1.(float64)
			num2 := value2.(float64)

			return num1 < num2
		},
		FtString: func(value1 interface{}, value2 interface{}) bool {
			str1 := value1.(string)
			str2 := value2.(string)

			return str1 < str2
		},
		FtDate: func(value1 interface{}, value2 interface{}) bool {
			time1 := value1.(time.Time)
			time2 := value2.(time.Time)

			return time1.Before(time2)
		},
	},
	QtkCmpLe: {
		FtNumber: func(value1 interface{}, value2 interface{}) bool {
			num1 := value1.(float64)
			num2 := value2.(float64)

			return num1 <= num2
		},
		FtString: func(value1 interface{}, value2 interface{}) bool {
			str1 := value1.(string)
			str2 := value2.(string)

			return str1 <= str2
		},
		FtDate: func(value1 interface{}, value2 interface{}) bool {
			time1 := value1.(time.Time)
			time2 := value2.(time.Time)

			return time1.Before(time2) || time1.Equal(time2)
		},
	},
}

func globComparator(value1 interface{}, value2 interface{}) bool {
	input := value1.(string)
	glob := value2.(glob.Glob)

	return glob.Match(input)
}

func regexpComparator(value1 interface{}, value2 interface{}) bool {
	input := value1.(string)
	regex := value2.(*regexp.Regexp)

	return regex.MatchString(input)
}

// GenerateFilterQueryLanguageHelpSections generates help documentation for the Filter Query Language
func GenerateFilterQueryLanguageHelpSections(config Config) (helpSections []*HelpSection) {
	description := []HelpSectionText{
		{text: "GRV has a built in query language which can be used to filter the content of the Ref and Commit views."},
		{text: "All queries resolve to boolean values which are tested against each item listed in the view."},
		{text: "A query is composed of at least one comparison:"},
		{},
		{text: "field CMP value", themeComponentID: CmpHelpViewSectionCodeBlock},
		{},
		{text: "CMP can be any of the following comparison operators, which are case-insensitive:"},
		{},
		{text: "=, !=, >, >=, <, <=, GLOB, REGEXP", themeComponentID: CmpHelpViewSectionCodeBlock},
		{},
		{text: "Value is one of the following types:"},
		{},
		{text: "string          (e.g. \"test\")", themeComponentID: CmpHelpViewSectionCodeBlock},
		{text: "number          (e.g. 123 or 123.0)", themeComponentID: CmpHelpViewSectionCodeBlock},
		{text: "date            (e.g. \"2017-09-05 10:05:25\" or \"2017-09-05\")", themeComponentID: CmpHelpViewSectionCodeBlock},
		{},
		{text: "Field is specific to the view that is being filtered."},
		{text: "For example, to filter commits to those whose commit messages start with \"Bug Fix:\":"},
		{},
		{text: "summary GLOB \"Bug Fix:*\"", themeComponentID: CmpHelpViewSectionCodeBlock},
		{},
		{text: "Or equivalently:"},
		{},
		{text: "summary REGEXP \"^Bug Fix:.*\"", themeComponentID: CmpHelpViewSectionCodeBlock},
		{},
		{text: "For more inforation about the supported GLOB syntax see: https://github.com/gobwas/glob"},
		{},
		{text: "For more information about the supported regex syntax see: https://golang.org/s/re2syntax"},
		{},
		{text: "Comparisons can be composed together using the following logical operators, which are case-insensitive:"},
		{},
		{text: "AND, OR, NOT", themeComponentID: CmpHelpViewSectionCodeBlock},
		{},
		{text: "For example, to filter commits to those authored by John Smith or Jane Roe in September 2017, ignoring merge commits:"},
		{},
		{text: `authordate >= "2017-09-01" AND authordate < "2017-10-01" AND (authorname = "John Smith" OR authorname = "Jane Roe") AND parentcount < 2`, themeComponentID: CmpHelpViewSectionCodeBlock},
		{},
		{text: "As shown above, expressions can be grouped using parentheses."},
		{},
		{text: "The list of (case-insensitive) fields that can be used in the Commit View is:"},
	}

	helpSections = append(helpSections, &HelpSection{
		title:       HelpSectionText{text: "Filter Query Language"},
		description: description,
	})

	helpSections = append(helpSections, GenerateCommitFieldHelpSection(config))

	helpSections = append(helpSections, &HelpSection{
		description: []HelpSectionText{
			{text: "The list of (case-insensitive) fields that can be used in the Ref View is:"},
		},
	})

	helpSections = append(helpSections, GenerateRefFieldHelpSection(config))

	return
}
