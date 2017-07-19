package main

import (
	"strings"
)

func CreateCommitFilter(query string) (commitFilter *CommitFilter, errors []error) {
	filter, errors := CreateFilter(query, &CommitFieldDescriptor{})
	if len(errors) > 0 {
		return
	}

	commitFilter = NewCommitFilter(filter)
	return
}

type CommitFilter struct {
	filter Filter
}

func NewCommitFilter(filter Filter) *CommitFilter {
	return &CommitFilter{
		filter: filter,
	}
}

func (commitFilter *CommitFilter) MatchesFilter(commit *Commit) bool {
	return commitFilter.filter(commit)
}

type CommitFieldDescriptor struct{}

func (commitFieldDescriptor *CommitFieldDescriptor) FieldType(fieldName string) (fieldType FieldType, fieldExists bool) {
	if commitField, ok := commitFields[strings.ToLower(fieldName)]; ok {
		fieldType = commitField.fieldType
		fieldExists = true
	}

	return
}

func (commitFieldDescriptor *CommitFieldDescriptor) FieldValue(inputValue interface{}, fieldName string) interface{} {
	commit := inputValue.(*Commit)
	commitField := commitFields[strings.ToLower(fieldName)]

	return commitField.value(commit)
}

type CommitFieldValue func(*Commit) interface{}

type CommitField struct {
	fieldType FieldType
	value     CommitFieldValue
}

var commitFields = map[string]CommitField{
	"authorname": CommitField{
		fieldType: FT_STRING,
		value: func(commit *Commit) interface{} {
			return commit.commit.Author().Name
		},
	},
	"authoremail": CommitField{
		fieldType: FT_STRING,
		value: func(commit *Commit) interface{} {
			return commit.commit.Author().Email
		},
	},
	"authordate": CommitField{
		fieldType: FT_DATE,
		value: func(commit *Commit) interface{} {
			return commit.commit.Author().When
		},
	},
	"committername": CommitField{
		fieldType: FT_STRING,
		value: func(commit *Commit) interface{} {
			return commit.commit.Committer().Name
		},
	},
	"committeremail": CommitField{
		fieldType: FT_STRING,
		value: func(commit *Commit) interface{} {
			return commit.commit.Committer().Email
		},
	},
	"committerdate": CommitField{
		fieldType: FT_DATE,
		value: func(commit *Commit) interface{} {
			return commit.commit.Committer().When
		},
	},
	"id": CommitField{
		fieldType: FT_STRING,
		value: func(commit *Commit) interface{} {
			return commit.commit.Id().String()
		},
	},
	"summary": CommitField{
		fieldType: FT_STRING,
		value: func(commit *Commit) interface{} {
			return commit.commit.Summary()
		},
	},
	"parentcount": CommitField{
		fieldType: FT_NUMBER,
		value: func(commit *Commit) interface{} {
			return float64(commit.commit.ParentCount())
		},
	},
}
