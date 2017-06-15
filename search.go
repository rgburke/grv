package main

import (
	log "github.com/Sirupsen/logrus"
	"regexp"
)

type SearchInputProvidor interface {
	Line(lineIndex uint) (line string, lineExists bool)
	LineNumber() (lineNumber uint)
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

func (search *Search) FindPrev(startLineIndex uint) (matchedLineIndex uint, found bool) {
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
