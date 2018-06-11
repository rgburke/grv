package main

import (
	"strings"

	slice "github.com/bradfitz/slice"
)

// CreateRefFilter creates a ref filter from the provided query
func CreateRefFilter(query string) (refFilter *RefFilter, errors []error) {
	filter, errors := CreateFilter(query, &refFieldDescriptor{})
	if len(errors) > 0 || filter == nil {
		return
	}

	refFilter = NewRefFilter(filter)
	return
}

// RefFilter is a wrapper around the raw filter to provide type safety
type RefFilter struct {
	filter Filter
}

// NewRefFilter creates a new instance of the wrapper
func NewRefFilter(filter Filter) *RefFilter {
	return &RefFilter{
		filter: filter,
	}
}

// MatchesFilter returns true if the ref matches the filter
func (refFilter *RefFilter) MatchesFilter(renderedRef *RenderedRef) bool {
	switch renderedRef.renderedRefType {
	case RvLocalBranchGroup, RvRemoteBranchGroup, RvTagGroup, RvSpace, RvLoading:
		return true
	default:
		return refFilter.filter(renderedRef)
	}
}

type refFieldDescriptor struct{}

func (fieldDescriptor *refFieldDescriptor) FieldType(fieldName string) (fieldType FieldType, fieldExists bool) {
	if field, ok := refFields[strings.ToLower(fieldName)]; ok {
		fieldType = field.fieldType
		fieldExists = true
	}

	return
}

func (fieldDescriptor *refFieldDescriptor) FieldValue(inputValue interface{}, fieldName string) interface{} {
	renderedRef := inputValue.(*RenderedRef)
	refField := refFields[strings.ToLower(fieldName)]

	return refField.value(renderedRef)
}

type refFieldValue func(*RenderedRef) interface{}

type refField struct {
	fieldType FieldType
	value     refFieldValue
}

var refFields = map[string]refField{
	"name": {
		fieldType: FtString,
		value: func(renderedRef *RenderedRef) interface{} {
			return strings.TrimLeft(renderedRef.value, " ")
		},
	},
}

// GenerateRefFieldHelpSection generates help documentation for the ref fields available
func GenerateRefFieldHelpSection(config Config) *HelpSection {
	headers := []TableHeader{
		{text: "Field", themeComponentID: CmpHelpViewSectionTableHeader},
		{text: "Type", themeComponentID: CmpHelpViewSectionTableHeader},
	}

	tableFormatter := NewTableFormatterWithHeaders(headers, config)
	tableFormatter.SetGridLines(true)

	refFieldNames := []string{}
	for refFieldName := range refFields {
		refFieldNames = append(refFieldNames, refFieldName)
	}

	slice.Sort(refFieldNames, func(i, j int) bool {
		return refFieldNames[i] < refFieldNames[j]
	})

	tableFormatter.Resize(uint(len(refFieldNames)))

	for rowIndex, refFieldName := range refFieldNames {
		refField := refFields[refFieldName]
		tableFormatter.SetCellWithStyle(uint(rowIndex), 0, CmpHelpViewSectionTableRow, "%v", refFieldName)
		tableFormatter.SetCellWithStyle(uint(rowIndex), 1, CmpHelpViewSectionTableRow, "%v", FieldTypeName(refField.fieldType))
	}

	return &HelpSection{
		tableFormatter: tableFormatter,
	}
}
