package main

import (
	"strings"
	"testing"
	"time"
)

type TestRecord struct {
	id          int
	name        string
	lastUpdated time.Time
}

type TestRecordFieldDescriptor struct{}

func (testRecordFieldDescriptor *TestRecordFieldDescriptor) FieldValue(inputValue interface{}, fieldName string) interface{} {
	testRecord := inputValue.(*TestRecord)

	switch strings.ToLower(fieldName) {
	case "id":
		return float64(testRecord.id)
	case "name":
		return testRecord.name
	case "lastupdated":
		return testRecord.lastUpdated
	}

	panic("Invalid field")
}

func (testRecordFieldDescriptor *TestRecordFieldDescriptor) FieldType(fieldName string) (fieldType FieldType, fieldExists bool) {
	fieldExists = true

	switch strings.ToLower(fieldName) {
	case "id":
		fieldType = FT_NUMBER
	case "name":
		fieldType = FT_STRING
	case "lastupdated":
		fieldType = FT_DATE
	default:
		fieldExists = false
	}

	return
}

func TestValueComparators(t *testing.T) {
	var valueComparatorTests = []struct {
		inputQuery           string
		expectedFilterOutput bool
	}{
		// EQ
		{
			inputQuery:           "1 = 1",
			expectedFilterOutput: true,
		},
		{
			inputQuery:           "1 = 2",
			expectedFilterOutput: false,
		},
		{
			inputQuery:           `"test1" = "test1"`,
			expectedFilterOutput: true,
		},
		{
			inputQuery:           `"test1" = "test2"`,
			expectedFilterOutput: false,
		},
		{
			inputQuery:           `LastUpdated = "2017-07-16"`,
			expectedFilterOutput: true,
		},
		{
			inputQuery:           `LastUpdated = "2017-07-18"`,
			expectedFilterOutput: false,
		},
		{
			inputQuery:           `LastUpdated = "2017-07-16 00:00:00"`,
			expectedFilterOutput: true,
		},
		{
			inputQuery:           `LastUpdated = "2017-07-16 10:00:00"`,
			expectedFilterOutput: false,
		},
		// NE
		{
			inputQuery:           "1 != 2",
			expectedFilterOutput: true,
		},
		{
			inputQuery:           "1 != 1",
			expectedFilterOutput: false,
		},
		{
			inputQuery:           `"test1" != "test2"`,
			expectedFilterOutput: true,
		},
		{
			inputQuery:           `"test1" != "test1"`,
			expectedFilterOutput: false,
		},
		{
			inputQuery:           `LastUpdated != "2017-07-18"`,
			expectedFilterOutput: true,
		},
		{
			inputQuery:           `LastUpdated != "2017-07-16"`,
			expectedFilterOutput: false,
		},
		{
			inputQuery:           `LastUpdated != "2017-07-16 10:00:00"`,
			expectedFilterOutput: true,
		},
		{
			inputQuery:           `LastUpdated != "2017-07-16 00:00:00"`,
			expectedFilterOutput: false,
		},
		// GT
		{
			inputQuery:           "2 > 1",
			expectedFilterOutput: true,
		},
		{
			inputQuery:           "1 > 1",
			expectedFilterOutput: false,
		},
		{
			inputQuery:           `"test2" > "test1"`,
			expectedFilterOutput: true,
		},
		{
			inputQuery:           `"test1" > "test1"`,
			expectedFilterOutput: false,
		},
		{
			inputQuery:           `LastUpdated > "2017-07-15"`,
			expectedFilterOutput: true,
		},
		{
			inputQuery:           `LastUpdated > "2017-07-16"`,
			expectedFilterOutput: false,
		},
		{
			inputQuery:           `LastUpdated > "2017-07-15 23:59:59"`,
			expectedFilterOutput: true,
		},
		{
			inputQuery:           `LastUpdated > "2017-07-16 00:00:00"`,
			expectedFilterOutput: false,
		},
		// GE
		{
			inputQuery:           "1 >= 1",
			expectedFilterOutput: true,
		},
		{
			inputQuery:           "1 >= 2",
			expectedFilterOutput: false,
		},
		{
			inputQuery:           `"test1" >= "test1"`,
			expectedFilterOutput: true,
		},
		{
			inputQuery:           `"test1" >= "test2"`,
			expectedFilterOutput: false,
		},
		{
			inputQuery:           `LastUpdated >= "2017-07-16"`,
			expectedFilterOutput: true,
		},
		{
			inputQuery:           `LastUpdated >= "2017-07-18"`,
			expectedFilterOutput: false,
		},
		{
			inputQuery:           `LastUpdated >= "2017-07-16 00:00:00"`,
			expectedFilterOutput: true,
		},
		{
			inputQuery:           `LastUpdated >= "2017-07-16 10:00:00"`,
			expectedFilterOutput: false,
		},
		// LT
		{
			inputQuery:           "1 < 2",
			expectedFilterOutput: true,
		},
		{
			inputQuery:           "1 < 1",
			expectedFilterOutput: false,
		},
		{
			inputQuery:           `"test1" < "test2"`,
			expectedFilterOutput: true,
		},
		{
			inputQuery:           `"test1" < "test1"`,
			expectedFilterOutput: false,
		},
		{
			inputQuery:           `LastUpdated < "2017-07-18"`,
			expectedFilterOutput: true,
		},
		{
			inputQuery:           `LastUpdated < "2017-07-16"`,
			expectedFilterOutput: false,
		},
		{
			inputQuery:           `LastUpdated < "2017-07-16 10:00:00"`,
			expectedFilterOutput: true,
		},
		{
			inputQuery:           `LastUpdated < "2017-07-16 00:00:00"`,
			expectedFilterOutput: false,
		},
		// LE
		{
			inputQuery:           "1 <= 1",
			expectedFilterOutput: true,
		},
		{
			inputQuery:           "2 <= 1",
			expectedFilterOutput: false,
		},
		{
			inputQuery:           `"test1" <= "test1"`,
			expectedFilterOutput: true,
		},
		{
			inputQuery:           `"test2" <= "test1"`,
			expectedFilterOutput: false,
		},
		{
			inputQuery:           `LastUpdated <= "2017-07-16"`,
			expectedFilterOutput: true,
		},
		{
			inputQuery:           `LastUpdated <= "2017-07-15"`,
			expectedFilterOutput: false,
		},
		{
			inputQuery:           `LastUpdated <= "2017-07-16 00:00:00"`,
			expectedFilterOutput: true,
		},
		{
			inputQuery:           `LastUpdated <= "2017-07-15 00:00:00"`,
			expectedFilterOutput: false,
		},
	}

	testRecord := &TestRecord{
		lastUpdated: time.Date(2017, time.July, 16, 0, 0, 0, 0, time.Local),
	}

	for _, valueComparatorTest := range valueComparatorTests {
		inputQuery := valueComparatorTest.inputQuery
		expectedFilterOutput := valueComparatorTest.expectedFilterOutput

		filter, errors := CreateFilter(inputQuery, &TestRecordFieldDescriptor{})

		if len(errors) > 0 {
			t.Errorf("CreateFilter failed with errors %v", errors)
		} else {
			actualFilterOutput := filter(testRecord)

			if expectedFilterOutput != actualFilterOutput {
				t.Errorf("Filter ouput does not match expected value for query \"%v\". Expected: %v, Actual: %v", inputQuery, expectedFilterOutput, actualFilterOutput)
			}
		}
	}
}

func TestPatternComparators(t *testing.T) {
	var patternComparatorTests = []struct {
		inputQuery           string
		expectedFilterOutput bool
	}{
		// GLOB
		{
			inputQuery:           `Name GLOB "John*"`,
			expectedFilterOutput: true,
		},
		{
			inputQuery:           `Name GLOB "Johny*"`,
			expectedFilterOutput: false,
		},
		{
			inputQuery:           `Name GLOB "John?Smith"`,
			expectedFilterOutput: true,
		},
		{
			inputQuery:           `Name GLOB "?John?Smith"`,
			expectedFilterOutput: false,
		},
		{
			inputQuery:           `Name GLOB "[IJK]ohn*"`,
			expectedFilterOutput: true,
		},
		{
			inputQuery:           `Name GLOB "[IK]ohn*"`,
			expectedFilterOutput: false,
		},
		{
			inputQuery:           `Name GLOB "[IK]ohn*"`,
			expectedFilterOutput: false,
		},
		{
			inputQuery:           `Name GLOB "[I-K]ohn*"`,
			expectedFilterOutput: true,
		},
		{
			inputQuery:           `Name GLOB "[A-C]ohn*"`,
			expectedFilterOutput: false,
		},
		{
			inputQuery:           `Name GLOB "[!A-C]ohn*"`,
			expectedFilterOutput: true,
		},
		{
			inputQuery:           `Name GLOB "[!I-K]ohn*"`,
			expectedFilterOutput: false,
		},
		{
			inputQuery:           `Name GLOB "{Jo,Bo}hn*"`,
			expectedFilterOutput: true,
		},
		{
			inputQuery:           `Name GLOB "{Ho,Bo}hn*"`,
			expectedFilterOutput: false,
		},
		// REGEX
		{
			inputQuery:           `Name REGEXP "^[Jj]ohn\\s\\w+$"`,
			expectedFilterOutput: true,
		},
		{
			inputQuery:           `Name REGEXP "\w+\s+\w+\s+\w+"`,
			expectedFilterOutput: false,
		},
	}

	testRecord := &TestRecord{
		name: "John Smith",
	}

	for _, patternComparatorTest := range patternComparatorTests {
		inputQuery := patternComparatorTest.inputQuery
		expectedFilterOutput := patternComparatorTest.expectedFilterOutput

		filter, errors := CreateFilter(inputQuery, &TestRecordFieldDescriptor{})

		if len(errors) > 0 {
			t.Errorf("CreateFilter failed with errors %v", errors)
		} else {
			actualFilterOutput := filter(testRecord)

			if expectedFilterOutput != actualFilterOutput {
				t.Errorf("Filter ouput does not match expected value for query \"%v\". Expected: %v, Actual: %v", inputQuery, expectedFilterOutput, actualFilterOutput)
			}
		}
	}
}

func TestLogicalComparators(t *testing.T) {
	var logicalComparatorTests = []struct {
		inputQuery           string
		expectedFilterOutput bool
	}{
		{
			inputQuery:           "1 = 1 AND 2 = 2",
			expectedFilterOutput: true,
		},
		{
			inputQuery:           "1 = 1 AND 2 = 3",
			expectedFilterOutput: false,
		},
		{
			inputQuery:           "1 = 2 AND 2 = 2",
			expectedFilterOutput: false,
		},
		{
			inputQuery:           "1 = 2 AND 2 = 3",
			expectedFilterOutput: false,
		},
		{
			inputQuery:           "1 = 1 OR 2 = 2",
			expectedFilterOutput: true,
		},
		{
			inputQuery:           "1 = 1 OR 2 = 3",
			expectedFilterOutput: true,
		},
		{
			inputQuery:           "1 = 2 OR 2 = 2",
			expectedFilterOutput: true,
		},
		{
			inputQuery:           "1 = 2 OR 2 = 3",
			expectedFilterOutput: false,
		},
		{
			inputQuery:           "NOT 1 = 1",
			expectedFilterOutput: false,
		},
		{
			inputQuery:           "NOT 1 = 2",
			expectedFilterOutput: true,
		},
	}

	for _, logicalComparatorTest := range logicalComparatorTests {
		inputQuery := logicalComparatorTest.inputQuery
		expectedFilterOutput := logicalComparatorTest.expectedFilterOutput

		filter, errors := CreateFilter(inputQuery, &TestRecordFieldDescriptor{})

		if len(errors) > 0 {
			t.Errorf("CreateFilter failed with errors %v", errors)
		} else {
			actualFilterOutput := filter(nil)

			if expectedFilterOutput != actualFilterOutput {
				t.Errorf("Filter ouput does not match expected value for query \"%v\". Expected: %v, Actual: %v", inputQuery, expectedFilterOutput, actualFilterOutput)
			}
		}
	}
}

func TestFieldValuesAreRetrievedFromInput(t *testing.T) {
	var fieldValueTestQueries = []string{
		"Id = 1",
		`Name = "test1"`,
		`LastUpdated = "2017-07-16"`,
		`Id < 3 AND Name <= "test1"`,
		`Id < 3 AND (Name >= "test3" OR LastUpdated < "2017-07-18")`,
	}

	matchRecord := &TestRecord{
		id:          1,
		name:        "test1",
		lastUpdated: time.Date(2017, time.July, 16, 0, 0, 0, 0, time.Local),
	}

	nonMatchingRecord := &TestRecord{
		id:          2,
		name:        "test2",
		lastUpdated: time.Date(2017, time.July, 18, 0, 0, 0, 0, time.Local),
	}

	for _, query := range fieldValueTestQueries {
		filter, errors := CreateFilter(query, &TestRecordFieldDescriptor{})

		if len(errors) > 0 {
			t.Errorf("CreateFilter failed with errors %v", errors)
		} else {
			actualMatchRecordFilterOutput := filter(matchRecord)
			actualNonMatchingRecordFilterOutput := filter(nonMatchingRecord)

			if !actualMatchRecordFilterOutput || actualNonMatchingRecordFilterOutput {
				t.Errorf("Filter output does not match expected value for query \"%v\". MatchRecord: %v, NonMatchingRecord: %v",
					query, actualMatchRecordFilterOutput, actualNonMatchingRecordFilterOutput)
			}
		}
	}
}
