package main

import (
	"fmt"
	"regexp"
	"runtime"

	log "github.com/Sirupsen/logrus"
)

const (
	searchMaxIterationsBeforeYeild = 1000
)

// SearchDirection describes the direction the search should be performed in
type SearchDirection int

// The set of search search directions
const (
	SdForward SearchDirection = iota
	SdBackward
)

var actionSearchDirection = map[ActionType]SearchDirection{
	ActionSearch:        SdForward,
	ActionReverseSearch: SdBackward,
}

// SearchInputProvidor provides input to the search alorithm
// This abstracts the source of the data from the search logic
type SearchInputProvidor interface {
	Line(lineIndex uint) (line string, lineExists bool)
	LineNumber() (lineNumber uint)
}

// SearchMatchIndex describes the byte range of a match on a line
type SearchMatchIndex struct {
	ByteStartIndex uint
	ByteEndIndex   uint
}

// SearchMatch contains all search match positions for a line
type SearchMatch struct {
	RowIndex     uint
	MatchIndexes []SearchMatchIndex
}

// Search manages performing a search on an input providor
type Search struct {
	direction     SearchDirection
	pattern       string
	regex         *regexp.Regexp
	inputProvidor SearchInputProvidor
}

// CreateSearchFromAction is a utility method to create a search configured based on the action that triggered it
func CreateSearchFromAction(action Action, inputProvidor SearchInputProvidor) (search *Search, err error) {
	direction, ok := actionSearchDirection[action.ActionType]
	if !ok {
		return search, fmt.Errorf("Invalid ActionType: %v", action.ActionType)
	}

	if !(len(action.Args) > 0) {
		return search, fmt.Errorf("Expected search pattern")
	}

	pattern, ok := action.Args[0].(string)
	if !ok {
		return search, fmt.Errorf("Expected search pattern")
	}

	return NewSearch(direction, pattern, inputProvidor)
}

// NewSearch creates a new search instance
func NewSearch(direction SearchDirection, pattern string, inputProvidor SearchInputProvidor) (search *Search, err error) {
	search = &Search{
		direction:     direction,
		pattern:       pattern,
		inputProvidor: inputProvidor,
	}

	if search.regex, err = regexp.Compile(pattern); err != nil {
		return
	}

	return
}

// FindNext looks for the next match starting from the line index provided
func (search *Search) FindNext(startLineIndex uint) (matchedLineIndex uint, found bool) {
	switch search.direction {
	case SdForward:
		return search.findNext(startLineIndex)
	case SdBackward:
		return search.findPrev(startLineIndex)
	}

	panic(fmt.Sprintf("Invalid search direction: %v", search.direction))
}

// FindPrev looks for the next match in the reverse direction starting from the line index provided
func (search *Search) FindPrev(startLineIndex uint) (matchedLineIndex uint, found bool) {
	switch search.direction {
	case SdForward:
		return search.findPrev(startLineIndex)
	case SdBackward:
		return search.findNext(startLineIndex)
	}

	panic(fmt.Sprintf("Invalid search direction: %v", search.direction))
}

func (search *Search) findNext(startLineIndex uint) (matchedLineIndex uint, found bool) {
	currentLineIndex := startLineIndex + 1
	wrapped := false

	for !wrapped || currentLineIndex <= startLineIndex {
		line, lineExists := search.inputProvidor.Line(currentLineIndex)
		if !lineExists {
			currentLineIndex = 0
			wrapped = true
			continue
		}

		if search.regex.MatchString(line) {
			matchedLineIndex = currentLineIndex
			found = true
			break
		}

		currentLineIndex++

		if currentLineIndex%searchMaxIterationsBeforeYeild == 0 {
			runtime.Gosched()
		}
	}

	return
}

func (search *Search) findPrev(startLineIndex uint) (matchedLineIndex uint, found bool) {
	currentLineIndex := startLineIndex
	wrapped := false

	for !wrapped || currentLineIndex >= startLineIndex {
		if currentLineIndex == 0 {
			currentLineIndex = search.inputProvidor.LineNumber()

			if currentLineIndex == 0 {
				break
			}

			wrapped = true
		}

		currentLineIndex--

		line, lineExists := search.inputProvidor.Line(currentLineIndex)
		if !lineExists {
			log.Errorf("Attempted to fetch non-existent line %v in reverse search", currentLineIndex)
			break
		}

		if search.regex.MatchString(line) {
			matchedLineIndex = currentLineIndex
			found = true
			break
		}

		if currentLineIndex%searchMaxIterationsBeforeYeild == 0 {
			runtime.Gosched()
		}
	}

	return
}

// FindAll find all matches across the entire input provided
func (search *Search) FindAll() (matches []SearchMatch) {
	lineIndex := uint(0)

	for {
		line, lineExists := search.inputProvidor.Line(lineIndex)
		if !lineExists {
			break
		}

		lineMatches := search.regex.FindAllStringIndex(line, -1)

		if len(lineMatches) > 0 {
			searchMatch := SearchMatch{
				RowIndex: lineIndex,
			}

			for _, lineMatch := range lineMatches {
				searchMatch.MatchIndexes = append(searchMatch.MatchIndexes, SearchMatchIndex{
					ByteStartIndex: uint(lineMatch[0]),
					ByteEndIndex:   uint(lineMatch[1]),
				})
			}

			matches = append(matches, searchMatch)
		}

		lineIndex++

		if lineIndex%searchMaxIterationsBeforeYeild == 0 {
			runtime.Gosched()
		}
	}

	return
}
