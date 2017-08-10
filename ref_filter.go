package main

import (
	"strings"
)

func CreateRefFilter(query string) (refFilter *RefFilter, errors []error) {
	filter, errors := CreateFilter(query, &RefFieldDescriptor{})
	if len(errors) > 0 {
		return
	}

	refFilter = NewRefFilter(filter)
	return
}

type RefFilter struct {
	filter Filter
}

func NewRefFilter(filter Filter) *RefFilter {
	return &RefFilter{
		filter: filter,
	}
}

func (refFilter *RefFilter) MatchesFilter(renderedRef *RenderedRef) bool {
	switch renderedRef.renderedRefType {
	case RV_LOCAL_BRANCH_GROUP, RV_REMOTE_BRANCH_GROUP, RV_TAG_GROUP, RV_SPACE, RV_LOADING:
		return true
	default:
		return refFilter.filter(renderedRef)
	}
}

type RefFieldDescriptor struct{}

func (refFieldDescriptor *RefFieldDescriptor) FieldType(fieldName string) (fieldType FieldType, fieldExists bool) {
	if refField, ok := refFields[strings.ToLower(fieldName)]; ok {
		fieldType = refField.fieldType
		fieldExists = true
	}

	return
}

func (refFieldDescriptor *RefFieldDescriptor) FieldValue(inputValue interface{}, fieldName string) interface{} {
	renderedRef := inputValue.(*RenderedRef)
	refField := refFields[strings.ToLower(fieldName)]

	return refField.value(renderedRef)
}

type RefFieldValue func(*RenderedRef) interface{}

type RefField struct {
	fieldType FieldType
	value     RefFieldValue
}

var refFields = map[string]RefField{
	"name": RefField{
		fieldType: FT_STRING,
		value: func(renderedRef *RenderedRef) interface{} {
			return strings.TrimLeft(renderedRef.value, " ")
		},
	},
}
