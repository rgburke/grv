package main

import (
	"regexp"
)

type SearchInputProvidor interface {
	Line(lineIndex uint) (line string, lineExists bool)
}

type Search struct {
	pattern       string
	regex         *regexp.Regexp
	inputProvidor SearchInputProvidor
}

func NewSearch(pattern string, inputProvidor SearchInputProvidor) (search *Search, err error) {
	search = &Search{
		pattern:       pattern,
		inputProvidor: inputProvidor,
	}

	if search.regex, err = regexp.Compile(pattern); err != nil {
		return
	}

	return
}

func (search *Search) FindNext(startLineIndex uint) (matchedLineIndex uint, found bool) {
	currentLineIndex := startLineIndex
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
