package main

import (
	"fmt"
	glob "github.com/gobwas/glob"
	"regexp"
	"strings"
	"time"
)

type Filter func(interface{}) bool

type FieldDescriptor interface {
	FieldTypeDescriptor
	FieldValue(inputValue interface{}, fieldName string) interface{}
}

func CreateFilter(query string, fieldDescriptor FieldDescriptor) (filter Filter, errors []error) {
	queryParser := NewQueryParser(strings.NewReader(query))

	expression, _, err := queryParser.Parse()
	if err != nil {
		errors = append(errors, err)
		return
	}

	expressionProcessor := NewExpressionProcessor(expression, fieldDescriptor)

	if expression, errors = expressionProcessor.Process(); len(errors) > 0 {
		return
	}

	filterGenerator := NewFilterGenerator(expression, fieldDescriptor)

	filter, err = filterGenerator.GenerateFilter()
	if err != nil {
		errors = append(errors, err)
		return
	}

	return
}

type FilterGenerator struct {
	expression      Expression
	fieldDescriptor FieldDescriptor
}

func NewFilterGenerator(expression Expression, fieldDescriptor FieldDescriptor) *FilterGenerator {
	return &FilterGenerator{
		expression:      expression,
		fieldDescriptor: fieldDescriptor,
	}
}

func (filterGenerator *FilterGenerator) GenerateFilter() (filter Filter, err error) {
	if filterGeneratorExpression, ok := filterGenerator.expression.(FilterGeneratorExpression); ok {
		filter = filterGeneratorExpression.GenerateFilter(filterGenerator.fieldDescriptor)
	} else {
		err = fmt.Errorf("Expected filter generator expression but received expression of type %T", filterGenerator.expression)
	}

	return
}

type FilterGeneratorExpression interface {
	GenerateFilter(FieldDescriptor) Filter
}

func (parenExpression *ParenExpression) GenerateFilter(fieldDescriptor FieldDescriptor) Filter {
	filterGeneratorExpression := parenExpression.expression.(FilterGeneratorExpression)
	return filterGeneratorExpression.GenerateFilter(fieldDescriptor)
}

func (unaryExpression *UnaryExpression) GenerateFilter(fieldDescriptor FieldDescriptor) Filter {
	filterGeneratorExpression := unaryExpression.expression.(FilterGeneratorExpression)
	filter := filterGeneratorExpression.GenerateFilter(fieldDescriptor)

	switch unaryExpression.operator.operator.tokenType {
	case QTK_NOT:
		return func(inputValue interface{}) bool {
			return !filter(inputValue)
		}
	}

	panic(fmt.Sprintf("Encountered invalid operator: %v", unaryExpression.operator.operator.value))
}

func (binaryExpression *BinaryExpression) GenerateFilter(fieldDescriptor FieldDescriptor) Filter {
	if !binaryExpression.IsComparison() {
		lhs := binaryExpression.lhs.(FilterGeneratorExpression).GenerateFilter(fieldDescriptor)
		rhs := binaryExpression.rhs.(FilterGeneratorExpression).GenerateFilter(fieldDescriptor)

		switch binaryExpression.operator.operator.tokenType {
		case QTK_AND:
			return func(inputValue interface{}) bool {
				return lhs(inputValue) && rhs(inputValue)
			}
		case QTK_OR:
			return func(inputValue interface{}) bool {
				return lhs(inputValue) || rhs(inputValue)
			}
		default:
			panic(fmt.Sprintf("Encountered invalid operator: %v", binaryExpression.operator.operator.value))
		}
	}

	lhs := binaryExpression.lhs.(ValueType)
	rhs := binaryExpression.rhs.(ValueType)

	var fieldComparator FieldComparator

	switch binaryExpression.operator.operator.tokenType {
	case QTK_CMP_GLOB:
		fieldComparator = globComparator
	case QTK_CMP_REGEXP:
		fieldComparator = regexpComparator
	default:
		fieldComparator = basicFieldComparators[binaryExpression.operator.operator.tokenType][lhs.FieldType(fieldDescriptor)]
	}

	return func(inputValue interface{}) bool {
		return fieldComparator(lhs.Value(inputValue, fieldDescriptor), rhs.Value(inputValue, fieldDescriptor))
	}
}

type ValueType interface {
	TypeDescriptor
	Value(inputValue interface{}, fieldDescriptor FieldDescriptor) interface{}
}

func (stringLiteral *StringLiteral) Value(inputValue interface{}, fieldDescriptor FieldDescriptor) interface{} {
	return stringLiteral.value.value
}

func (numberLiteral *NumberLiteral) Value(inputValue interface{}, fieldDescriptor FieldDescriptor) interface{} {
	return numberLiteral.number
}

func (dateLiteral *DateLiteral) Value(inputValue interface{}, fieldDescriptor FieldDescriptor) interface{} {
	return dateLiteral.dateTime
}

func (globLiteral *GlobLiteral) Value(inputValue interface{}, fieldDescriptor FieldDescriptor) interface{} {
	return globLiteral.glob
}

func (regexLiteral *RegexLiteral) Value(inputValue interface{}, fieldDescriptor FieldDescriptor) interface{} {
	return regexLiteral.regex
}

func (identifier *Identifier) Value(inputValue interface{}, fieldDescriptor FieldDescriptor) interface{} {
	return fieldDescriptor.FieldValue(inputValue, identifier.identifier.value)
}

type FieldComparator func(interface{}, interface{}) bool

var basicFieldComparators = map[QueryTokenType]map[FieldType]FieldComparator{
	QTK_CMP_EQ: map[FieldType]FieldComparator{
		FT_NUMBER: func(value1 interface{}, value2 interface{}) bool {
			num1 := value1.(float64)
			num2 := value2.(float64)

			return num1 == num2
		},
		FT_STRING: func(value1 interface{}, value2 interface{}) bool {
			str1 := value1.(string)
			str2 := value2.(string)

			return str1 == str2
		},
		FT_DATE: func(value1 interface{}, value2 interface{}) bool {
			time1 := value1.(time.Time)
			time2 := value2.(time.Time)

			return time1.Equal(time2)
		},
	},
	QTK_CMP_NE: map[FieldType]FieldComparator{
		FT_NUMBER: func(value1 interface{}, value2 interface{}) bool {
			num1 := value1.(float64)
			num2 := value2.(float64)

			return num1 != num2
		},
		FT_STRING: func(value1 interface{}, value2 interface{}) bool {
			str1 := value1.(string)
			str2 := value2.(string)

			return str1 != str2
		},
		FT_DATE: func(value1 interface{}, value2 interface{}) bool {
			time1 := value1.(time.Time)
			time2 := value2.(time.Time)

			return !time1.Equal(time2)
		},
	},
	QTK_CMP_GT: map[FieldType]FieldComparator{
		FT_NUMBER: func(value1 interface{}, value2 interface{}) bool {
			num1 := value1.(float64)
			num2 := value2.(float64)

			return num1 > num2
		},
		FT_STRING: func(value1 interface{}, value2 interface{}) bool {
			str1 := value1.(string)
			str2 := value2.(string)

			return str1 > str2
		},
		FT_DATE: func(value1 interface{}, value2 interface{}) bool {
			time1 := value1.(time.Time)
			time2 := value2.(time.Time)

			return time1.After(time2)
		},
	},
	QTK_CMP_GE: map[FieldType]FieldComparator{
		FT_NUMBER: func(value1 interface{}, value2 interface{}) bool {
			num1 := value1.(float64)
			num2 := value2.(float64)

			return num1 >= num2
		},
		FT_STRING: func(value1 interface{}, value2 interface{}) bool {
			str1 := value1.(string)
			str2 := value2.(string)

			return str1 >= str2
		},
		FT_DATE: func(value1 interface{}, value2 interface{}) bool {
			time1 := value1.(time.Time)
			time2 := value2.(time.Time)

			return time1.After(time2) || time1.Equal(time2)
		},
	},
	QTK_CMP_LT: map[FieldType]FieldComparator{
		FT_NUMBER: func(value1 interface{}, value2 interface{}) bool {
			num1 := value1.(float64)
			num2 := value2.(float64)

			return num1 < num2
		},
		FT_STRING: func(value1 interface{}, value2 interface{}) bool {
			str1 := value1.(string)
			str2 := value2.(string)

			return str1 < str2
		},
		FT_DATE: func(value1 interface{}, value2 interface{}) bool {
			time1 := value1.(time.Time)
			time2 := value2.(time.Time)

			return time1.Before(time2)
		},
	},
	QTK_CMP_LE: map[FieldType]FieldComparator{
		FT_NUMBER: func(value1 interface{}, value2 interface{}) bool {
			num1 := value1.(float64)
			num2 := value2.(float64)

			return num1 <= num2
		},
		FT_STRING: func(value1 interface{}, value2 interface{}) bool {
			str1 := value1.(string)
			str2 := value2.(string)

			return str1 <= str2
		},
		FT_DATE: func(value1 interface{}, value2 interface{}) bool {
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
