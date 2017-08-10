package main

import (
	"reflect"
	"testing"
)

func TestRefFieldExistence(t *testing.T) {
	var fieldExistsTests = []struct {
		fieldName      string
		expectedExists bool
	}{
		{
			fieldName:      "Name",
			expectedExists: true,
		},
		{
			fieldName:      "invalidfield",
			expectedExists: false,
		},
	}

	refFieldDescriptor := &RefFieldDescriptor{}

	for _, fieldExistsTest := range fieldExistsTests {
		fieldName := fieldExistsTest.fieldName
		expectedExists := fieldExistsTest.expectedExists

		_, actualExists := refFieldDescriptor.FieldType(fieldName)

		if expectedExists != actualExists {
			t.Errorf("Field existence does not match expected value for field %v. Expected: %v, Actual: %v", fieldName, expectedExists, actualExists)
		}
	}
}

func TestRefFieldsAreCaseInsenstive(t *testing.T) {
	fields := []string{
		"name",
		"Name",
		"NAME",
	}

	refFieldDescriptor := &RefFieldDescriptor{}

	for _, field := range fields {
		_, exists := refFieldDescriptor.FieldType(field)

		if !exists {
			t.Errorf("Field does not exist: %v", field)
		}
	}
}

func TestRefFieldTypes(t *testing.T) {
	var refFieldTypeTests = []struct {
		fieldName         string
		expectedFieldType FieldType
	}{
		{
			fieldName:         "name",
			expectedFieldType: FT_STRING,
		},
	}

	refFieldDescriptor := &RefFieldDescriptor{}

	for _, refFieldTypeTest := range refFieldTypeTests {
		fieldName := refFieldTypeTest.fieldName
		expectedFieldType := refFieldTypeTest.expectedFieldType

		actualFieldType, fieldExists := refFieldDescriptor.FieldType(fieldName)

		if !fieldExists {
			t.Errorf("Expected field %v to exist", fieldName)
		} else if expectedFieldType != actualFieldType {
			t.Errorf("Field type does not match expected value for field %v. Expected: %v, Actual: %v", fieldName, expectedFieldType, actualFieldType)
		}
	}
}

func TestRefFieldValuesAreExtracted(t *testing.T) {
	var refFieldValueTests = []struct {
		fieldName     string
		expectedValue interface{}
	}{
		{
			fieldName:     "Name",
			expectedValue: "Test",
		},
	}

	renderedRef := &RenderedRef{
		renderedRefType: RV_LOCAL_BRANCH,
		value:           "Test",
	}

	refFieldDescriptor := &RefFieldDescriptor{}

	for _, refFieldValueTest := range refFieldValueTests {
		fieldName := refFieldValueTest.fieldName
		expectedValue := refFieldValueTest.expectedValue

		actualValue := refFieldDescriptor.FieldValue(renderedRef, fieldName)

		if !reflect.DeepEqual(expectedValue, actualValue) {
			t.Errorf("Field value does not match expected value for field %v. Expected: %v, Actual: %v", fieldName, expectedValue, actualValue)
		}
	}
}

func TestCertainRenderedRefTypesAlwaysMatchFilter(t *testing.T) {
	var renderedRefValueTests = []struct {
		renderedRefType      RenderedRefType
		expectedFilterOutput bool
	}{
		{
			renderedRefType:      RV_LOCAL_BRANCH_GROUP,
			expectedFilterOutput: true,
		},
		{
			renderedRefType:      RV_REMOTE_BRANCH_GROUP,
			expectedFilterOutput: true,
		},
		{
			renderedRefType:      RV_TAG_GROUP,
			expectedFilterOutput: true,
		},
		{
			renderedRefType:      RV_SPACE,
			expectedFilterOutput: true,
		},
		{
			renderedRefType:      RV_LOADING,
			expectedFilterOutput: true,
		},
		{
			renderedRefType:      RV_LOCAL_BRANCH,
			expectedFilterOutput: false,
		},
		{
			renderedRefType:      RV_REMOTE_BRANCH,
			expectedFilterOutput: false,
		},
		{
			renderedRefType:      RV_TAG,
			expectedFilterOutput: false,
		},
	}

	refFilter, errors := CreateRefFilter(`Name = "Test"`)
	if len(errors) > 0 {
		t.Errorf("Unexpected errors when creating filter: %v", errors)
		return
	}

	for _, renderedRefValueTest := range renderedRefValueTests {
		renderedRef := &RenderedRef{
			renderedRefType: renderedRefValueTest.renderedRefType,
		}

		expectedValue := renderedRefValueTest.expectedFilterOutput
		actualValue := refFilter.MatchesFilter(renderedRef)

		if actualValue != expectedValue {
			t.Errorf("Filter output does not match expected value for RenderedRefType %v. Expected: %v, Actual: %v",
				renderedRef.renderedRefType, expectedValue, actualValue)
		}
	}
}
