package main

import (
	"reflect"
	"testing"
)

var testText = []string{
	"Test line 1: test",
	" Test line 2: ",
	"Test line 3:  test",
	"Test line 4: tst",
}

type TestInputProvidor struct{}

func (inputProvidor *TestInputProvidor) Line(lineIndex uint) (line string, lineExists bool) {
	if lineIndex < inputProvidor.LineNumber() {
		line = testText[lineIndex]
		lineExists = true
	}

	return
}

func (inputProvidor *TestInputProvidor) LineNumber() (lineNumber uint) {
	return uint(len(testText))
}

func createSearch(direction SearchDirection, pattern string, t *testing.T) *Search {
	search, err := NewSearch(direction, pattern, &TestInputProvidor{})
	if err != nil {
		t.Fatalf("Failed to create search instance: %v", err)
	}

	return search
}

func checkResult(expectedMatchLineIndex uint, expectedFound bool, actualMatchLineIndex uint, actualFound bool, t *testing.T) {
	if expectedFound != actualFound {
		t.Errorf("Search found result did not match expected value. Expected: %v. Actual: %v", expectedFound, actualFound)
	} else if expectedMatchLineIndex != actualMatchLineIndex {
		t.Errorf("Search line index result did not match expected value. Expected: %v. Actual: %v",
			expectedMatchLineIndex, actualMatchLineIndex)
	}
}

func TestSearchForwardFindNextFindsMatch(t *testing.T) {
	search := createSearch(SdForward, "line 2", t)

	lineIndex, found := search.FindNext(0)

	checkResult(1, true, lineIndex, found, t)
}

func TestSearchForwardFindNextFindsNoMatch(t *testing.T) {
	search := createSearch(SdForward, "non-existent text", t)

	lineIndex, found := search.FindNext(0)

	checkResult(0, false, lineIndex, found, t)
}

func TestSearchForwardFindNextWraps(t *testing.T) {
	search := createSearch(SdForward, "line 2", t)

	lineIndex, found := search.FindNext(2)

	checkResult(1, true, lineIndex, found, t)
}

func TestSearchForwardFindNextStartsFromLineBelowProvided(t *testing.T) {
	search := createSearch(SdForward, `line \d`, t)

	lineIndex, found := search.FindNext(0)

	checkResult(1, true, lineIndex, found, t)
}

func TestSearchForwardFindPrevFindsMatch(t *testing.T) {
	search := createSearch(SdForward, "line 1", t)

	lineIndex, found := search.FindPrev(2)

	checkResult(0, true, lineIndex, found, t)
}

func TestSearchForwardFindPrevFindsNoMatch(t *testing.T) {
	search := createSearch(SdForward, "non-existent text", t)

	lineIndex, found := search.FindPrev(3)

	checkResult(0, false, lineIndex, found, t)
}

func TestSearchForwardFindPrevWraps(t *testing.T) {
	search := createSearch(SdForward, "line 2", t)

	lineIndex, found := search.FindPrev(0)

	checkResult(1, true, lineIndex, found, t)
}

func TestSearchForwardFindPrevStartsFromLineBelowProvided(t *testing.T) {
	search := createSearch(SdForward, `line \d`, t)

	lineIndex, found := search.FindPrev(3)

	checkResult(2, true, lineIndex, found, t)
}

func TestSearchBackwardFindNextFindsMatch(t *testing.T) {
	search := createSearch(SdBackward, "line 2", t)

	lineIndex, found := search.FindNext(2)

	checkResult(1, true, lineIndex, found, t)
}

func TestSearchBackwardFindNextFindsNoMatch(t *testing.T) {
	search := createSearch(SdBackward, "non-existent text", t)

	lineIndex, found := search.FindNext(0)

	checkResult(0, false, lineIndex, found, t)
}

func TestSearchBackwardFindNextWraps(t *testing.T) {
	search := createSearch(SdBackward, "line 2", t)

	lineIndex, found := search.FindNext(0)

	checkResult(1, true, lineIndex, found, t)
}

func TestSearchBackwardFindNextStartsFromLineAboveProvided(t *testing.T) {
	search := createSearch(SdBackward, `line \d`, t)

	lineIndex, found := search.FindNext(1)

	checkResult(0, true, lineIndex, found, t)
}

func TestSearchBackwardFindPrevFindsMatch(t *testing.T) {
	search := createSearch(SdBackward, "line 2", t)

	lineIndex, found := search.FindPrev(0)

	checkResult(1, true, lineIndex, found, t)
}

func TestSearchBackwardFindPrevFindsNoMatch(t *testing.T) {
	search := createSearch(SdBackward, "non-existent text", t)

	lineIndex, found := search.FindPrev(0)

	checkResult(0, false, lineIndex, found, t)
}

func TestSearchBackwardFindPrevWraps(t *testing.T) {
	search := createSearch(SdBackward, "line 2", t)

	lineIndex, found := search.FindPrev(2)

	checkResult(1, true, lineIndex, found, t)
}

func TestSearchBackwardFindPrevStartsFromLineBelowProvided(t *testing.T) {
	search := createSearch(SdBackward, `line \d`, t)

	lineIndex, found := search.FindPrev(1)

	checkResult(2, true, lineIndex, found, t)
}

func TestSearchFindAll(t *testing.T) {
	search := createSearch(SdForward, `[Tt]est`, t)

	expectedMatches := []SearchMatch{
		SearchMatch{
			RowIndex: 0,
			MatchIndexes: []SearchMatchIndex{
				SearchMatchIndex{
					ByteStartIndex: 0,
					ByteEndIndex:   4,
				},
				SearchMatchIndex{
					ByteStartIndex: 13,
					ByteEndIndex:   17,
				},
			},
		},
		SearchMatch{
			RowIndex: 1,
			MatchIndexes: []SearchMatchIndex{
				SearchMatchIndex{
					ByteStartIndex: 1,
					ByteEndIndex:   5,
				},
			},
		},
		SearchMatch{
			RowIndex: 2,
			MatchIndexes: []SearchMatchIndex{
				SearchMatchIndex{
					ByteStartIndex: 0,
					ByteEndIndex:   4,
				},
				SearchMatchIndex{
					ByteStartIndex: 14,
					ByteEndIndex:   18,
				},
			},
		},
		SearchMatch{
			RowIndex: 3,
			MatchIndexes: []SearchMatchIndex{
				SearchMatchIndex{
					ByteStartIndex: 0,
					ByteEndIndex:   4,
				},
			},
		},
	}

	actualMatches := search.FindAll()

	if !reflect.DeepEqual(expectedMatches, actualMatches) {
		t.Errorf("FindAll did not return expected matches. Expected: %v. Actual %v", expectedMatches, actualMatches)
	}
}
