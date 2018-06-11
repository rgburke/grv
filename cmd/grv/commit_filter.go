package main

import (
	"strings"

	slice "github.com/bradfitz/slice"
)

// CreateCommitFilter constructs a commit filter from the provided query
func CreateCommitFilter(query string) (commitFilter *CommitFilter, errors []error) {
	filter, errors := CreateFilter(query, &CommitFieldDescriptor{})
	if len(errors) > 0 || filter == nil {
		return
	}

	commitFilter = NewCommitFilter(filter)
	return
}

// CommitFilter is a wrapper around the raw commit filter
// Used for filter argument type safety
type CommitFilter struct {
	filter Filter
}

// NewCommitFilter creates a wrapper instance around a commit filter
func NewCommitFilter(filter Filter) *CommitFilter {
	return &CommitFilter{
		filter: filter,
	}
}

// MatchesFilter tests if the provided commit matches this filter
func (commitFilter *CommitFilter) MatchesFilter(commit *Commit) bool {
	return commitFilter.filter(commit)
}

// CommitFieldDescriptor exposes functions describing commit field properties
type CommitFieldDescriptor struct{}

// FieldType returns the type of the provided field (if it exists)
func (commitFieldDescriptor *CommitFieldDescriptor) FieldType(fieldName string) (fieldType FieldType, fieldExists bool) {
	if commitField, ok := commitFields[strings.ToLower(fieldName)]; ok {
		fieldType = commitField.fieldType
		fieldExists = true
	}

	return
}

// FieldValue extracts a field value from a commit object
func (commitFieldDescriptor *CommitFieldDescriptor) FieldValue(inputValue interface{}, fieldName string) interface{} {
	commit := inputValue.(*Commit)
	commitField := commitFields[strings.ToLower(fieldName)]

	return commitField.value(commit)
}

// CommitFieldValue accepts a commit and returns a field value of that commit
type CommitFieldValue func(*Commit) interface{}

// CommitField provides data for a commit field
type CommitField struct {
	fieldType FieldType
	value     CommitFieldValue
}

var commitFields = map[string]CommitField{
	"authorname": {
		fieldType: FtString,
		value: func(commit *Commit) interface{} {
			return commit.commit.Author().Name
		},
	},
	"authoremail": {
		fieldType: FtString,
		value: func(commit *Commit) interface{} {
			return commit.commit.Author().Email
		},
	},
	"authordate": {
		fieldType: FtDate,
		value: func(commit *Commit) interface{} {
			return commit.commit.Author().When
		},
	},
	"committername": {
		fieldType: FtString,
		value: func(commit *Commit) interface{} {
			return commit.commit.Committer().Name
		},
	},
	"committeremail": {
		fieldType: FtString,
		value: func(commit *Commit) interface{} {
			return commit.commit.Committer().Email
		},
	},
	"committerdate": {
		fieldType: FtDate,
		value: func(commit *Commit) interface{} {
			return commit.commit.Committer().When
		},
	},
	"id": {
		fieldType: FtString,
		value: func(commit *Commit) interface{} {
			return commit.commit.Id().String()
		},
	},
	"summary": {
		fieldType: FtString,
		value: func(commit *Commit) interface{} {
			return commit.commit.Summary()
		},
	},
	"message": {
		fieldType: FtString,
		value: func(commit *Commit) interface{} {
			return commit.commit.Message()
		},
	},
	"parentcount": {
		fieldType: FtNumber,
		value: func(commit *Commit) interface{} {
			return float64(commit.commit.ParentCount())
		},
	},
}

// GenerateCommitFieldHelpSection generates documentation for the commit fields available
func GenerateCommitFieldHelpSection(config Config) *HelpSection {
	headers := []TableHeader{
		{text: "Field", themeComponentID: CmpHelpViewSectionTableHeader},
		{text: "Type", themeComponentID: CmpHelpViewSectionTableHeader},
	}

	tableFormatter := NewTableFormatterWithHeaders(headers, config)
	tableFormatter.SetGridLines(true)

	commitFieldNames := []string{}
	for commitFieldName := range commitFields {
		commitFieldNames = append(commitFieldNames, commitFieldName)
	}

	slice.Sort(commitFieldNames, func(i, j int) bool {
		return commitFieldNames[i] < commitFieldNames[j]
	})

	tableFormatter.Resize(uint(len(commitFieldNames)))

	for rowIndex, commitFieldName := range commitFieldNames {
		commitField := commitFields[commitFieldName]
		tableFormatter.SetCellWithStyle(uint(rowIndex), 0, CmpHelpViewSectionTableRow, "%v", commitFieldName)
		tableFormatter.SetCellWithStyle(uint(rowIndex), 1, CmpHelpViewSectionTableRow, "%v", FieldTypeName(commitField.fieldType))
	}

	return &HelpSection{
		tableFormatter: tableFormatter,
	}
}
