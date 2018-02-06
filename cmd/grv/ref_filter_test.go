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

	fieldDescriptor := &refFieldDescriptor{}

	for _, fieldExistsTest := range fieldExistsTests {
		fieldName := fieldExistsTest.fieldName
		expectedExists := fieldExistsTest.expectedExists

		_, actualExists := fieldDescriptor.FieldType(fieldName)

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

	fieldDescriptor := &refFieldDescriptor{}

	for _, field := range fields {
		_, exists := fieldDescriptor.FieldType(field)

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
			expectedFieldType: FtString,
		},
	}

	fieldDescriptor := &refFieldDescriptor{}

	for _, refFieldTypeTest := range refFieldTypeTests {
		fieldName := refFieldTypeTest.fieldName
		expectedFieldType := refFieldTypeTest.expectedFieldType

		actualFieldType, fieldExists := fieldDescriptor.FieldType(fieldName)

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
		renderedRefType: RvLocalBranch,
		value:           "Test",
	}

	fieldDescriptor := &refFieldDescriptor{}

	for _, refFieldValueTest := range refFieldValueTests {
		fieldName := refFieldValueTest.fieldName
		expectedValue := refFieldValueTest.expectedValue

		actualValue := fieldDescriptor.FieldValue(renderedRef, fieldName)

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
			renderedRefType:      RvLocalBranchGroup,
			expectedFilterOutput: true,
		},
		{
			renderedRefType:      RvRemoteBranchGroup,
			expectedFilterOutput: true,
		},
		{
			renderedRefType:      RvTagGroup,
			expectedFilterOutput: true,
		},
		{
			renderedRefType:      RvSpace,
			expectedFilterOutput: true,
		},
		{
			renderedRefType:      RvLoading,
			expectedFilterOutput: true,
		},
		{
			renderedRefType:      RvLocalBranch,
			expectedFilterOutput: false,
		},
		{
			renderedRefType:      RvRemoteBranch,
			expectedFilterOutput: false,
		},
		{
			renderedRefType:      RvTag,
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

func TestNilRefFilterIsReturnedIfQueryDoesNotDefineFilter(t *testing.T) {
	query := "         "
	refFilter, errors := CreateRefFilter(query)

	if len(errors) > 0 {
		t.Errorf("CreateRefFilter failed with errors %v", errors)
	} else if refFilter != nil {
		t.Errorf("Expected returned filter to be nil but found: %[1]v of type %[1]T", refFilter)
	}
}
