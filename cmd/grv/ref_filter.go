package main

import (
	"strings"
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
