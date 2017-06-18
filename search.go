package main

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"regexp"
)

type SearchDirection int

const (
	SD_FORWARD SearchDirection = iota
	SD_BACKWARD
)

var actionSearchDirection = map[ActionType]SearchDirection{
	ACTION_SEARCH:         SD_FORWARD,
	ACTION_REVERSE_SEARCH: SD_BACKWARD,
}

type SearchInputProvidor interface {
	Line(lineIndex uint) (line string, lineExists bool)
	LineNumber() (lineNumber uint)
}

type Search struct {
	direction     SearchDirection
	pattern       string
	regex         *regexp.Regexp
	inputProvidor SearchInputProvidor
}

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

func (search *Search) FindNext(startLineIndex uint) (matchedLineIndex uint, found bool) {
	switch search.direction {
	case SD_FORWARD:
		return search.findNext(startLineIndex)
	case SD_BACKWARD:
		return search.findPrev(startLineIndex)
	}

	panic(fmt.Sprintf("Invalid search direction: %v", search.direction))
}

func (search *Search) FindPrev(startLineIndex uint) (matchedLineIndex uint, found bool) {
	switch search.direction {
	case SD_FORWARD:
		return search.findPrev(startLineIndex)
	case SD_BACKWARD:
		return search.findNext(startLineIndex)
	}

	panic(fmt.Sprintf("Invalid search direction: %v", search.direction))
}

func (search *Search) findNext(startLineIndex uint) (matchedLineIndex uint, found bool) {
	currentLineIndex := startLineIndex + 1
	wrapped := false

	for !wrapped || currentLineIndex != startLineIndex {
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
	}

	return
}

func (search *Search) findPrev(startLineIndex uint) (matchedLineIndex uint, found bool) {
	currentLineIndex := startLineIndex
	wrapped := false

	for !wrapped || currentLineIndex != startLineIndex {
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
	}

	return
}
