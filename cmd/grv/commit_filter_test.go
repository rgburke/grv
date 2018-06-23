package main

import (
	"os"
	"reflect"
	"testing"

	git "gopkg.in/libgit2/git2go.v27"
)

func TestCommitFieldExistence(t *testing.T) {
	var fieldExistsTests = []struct {
		fieldName      string
		expectedExists bool
	}{
		{
			fieldName:      "authorname",
			expectedExists: true,
		},
		{
			fieldName:      "invalidfield",
			expectedExists: false,
		},
		{
			fieldName:      "authordate",
			expectedExists: true,
		},
		{
			fieldName:      "comittername",
			expectedExists: false,
		},
	}

	commitFieldDescriptor := &CommitFieldDescriptor{}

	for _, fieldExistsTest := range fieldExistsTests {
		fieldName := fieldExistsTest.fieldName
		expectedExists := fieldExistsTest.expectedExists

		_, actualExists := commitFieldDescriptor.FieldType(fieldName)

		if expectedExists != actualExists {
			t.Errorf("Field existence does not match expected value for field %v. Expected: %v, Actual: %v", fieldName, expectedExists, actualExists)
		}
	}
}

func TestCommitFieldsAreCaseInsenstive(t *testing.T) {
	fields := []string{
		"authorname",
		"AuthorName",
		"AuThOrNaMe",
		"AUTHORNAME",
	}

	commitFieldDescriptor := &CommitFieldDescriptor{}

	for _, field := range fields {
		_, exists := commitFieldDescriptor.FieldType(field)

		if !exists {
			t.Errorf("Field does not exist: %v", field)
		}
	}
}

func TestCommitFieldTypes(t *testing.T) {
	var commitFieldTypeTests = []struct {
		fieldName         string
		expectedFieldType FieldType
	}{
		{
			fieldName:         "authorname",
			expectedFieldType: FtString,
		},
		{
			fieldName:         "authoremail",
			expectedFieldType: FtString,
		},
		{
			fieldName:         "authordate",
			expectedFieldType: FtDate,
		},
		{
			fieldName:         "committername",
			expectedFieldType: FtString,
		},
		{
			fieldName:         "committeremail",
			expectedFieldType: FtString,
		},
		{
			fieldName:         "committerdate",
			expectedFieldType: FtDate,
		},
		{
			fieldName:         "id",
			expectedFieldType: FtString,
		},
		{
			fieldName:         "summary",
			expectedFieldType: FtString,
		},
		{
			fieldName:         "parentcount",
			expectedFieldType: FtNumber,
		},
	}

	commitFieldDescriptor := &CommitFieldDescriptor{}

	for _, commitFieldTypeTest := range commitFieldTypeTests {
		fieldName := commitFieldTypeTest.fieldName
		expectedFieldType := commitFieldTypeTest.expectedFieldType

		actualFieldType, fieldExists := commitFieldDescriptor.FieldType(fieldName)

		if !fieldExists {
			t.Errorf("Expected field %v to exist", fieldName)
		} else if expectedFieldType != actualFieldType {
			t.Errorf("Field type does not match expected value for field %v. Expected: %v, Actual: %v", fieldName, expectedFieldType, actualFieldType)
		}
	}
}

func TestCommitFieldValuesAreExtracted(t *testing.T) {
	t.SkipNow()

	folderPath, err := os.Getwd()
	if err != nil {
		t.Fatalf("Unable to determine working directory: %v", err)
	}

	folderPath += "/.."

	repo, err := git.OpenRepository(folderPath)
	if err != nil {
		t.Fatalf("Unable to open repo: %v", err)
	}

	commitID := "300dc7fd7cf162e89136ad3b1b8b4b7bd7dd13a5"
	oid, err := git.NewOid(commitID)
	if err != nil {
		t.Fatalf("Unable to create oid with Id %v: %v", commitID, err)
	}

	rawCommit, err := repo.LookupCommit(oid)
	if err != nil {
		t.Fatalf("Unable to load commit with Id %v: %v", commitID, err)
	}

	var commitFieldValueTests = []struct {
		fieldName     string
		expectedValue interface{}
	}{
		{
			fieldName:     "authorname",
			expectedValue: rawCommit.Author().Name,
		},
		{
			fieldName:     "authoremail",
			expectedValue: rawCommit.Author().Email,
		},
		{
			fieldName:     "authordate",
			expectedValue: rawCommit.Author().When,
		},
		{
			fieldName:     "committername",
			expectedValue: rawCommit.Committer().Name,
		},
		{
			fieldName:     "committeremail",
			expectedValue: rawCommit.Committer().Email,
		},
		{
			fieldName:     "committerdate",
			expectedValue: rawCommit.Committer().When,
		},
		{
			fieldName:     "id",
			expectedValue: rawCommit.Id().String(),
		},
		{
			fieldName:     "summary",
			expectedValue: rawCommit.Summary(),
		},
		{
			fieldName:     "parentcount",
			expectedValue: float64(rawCommit.ParentCount()),
		},
	}

	commit := &Commit{
		commit: rawCommit,
	}

	commitFieldDescriptor := &CommitFieldDescriptor{}

	for _, commitFieldValueTest := range commitFieldValueTests {
		fieldName := commitFieldValueTest.fieldName
		expectedValue := commitFieldValueTest.expectedValue

		actualValue := commitFieldDescriptor.FieldValue(commit, fieldName)

		if !reflect.DeepEqual(expectedValue, actualValue) {
			t.Errorf("Field value does not match expected value for field %v. Expected: %v, Actual: %v", fieldName, expectedValue, actualValue)
		}
	}
}

func TestNilCommitFilterIsReturnedIfQueryDoesNotDefineFilter(t *testing.T) {
	query := " \t\v\r\n"
	commitFilter, errors := CreateCommitFilter(query)

	if len(errors) > 0 {
		t.Errorf("CreateCommitFilter failed with errors %v", errors)
	} else if commitFilter != nil {
		t.Errorf("Expected returned filter to be nil but found: %[1]v of type %[1]T", commitFilter)
	}
}
