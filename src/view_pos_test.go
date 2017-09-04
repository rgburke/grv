package main

import (
	"testing"
)

func (viewPos *ViewPosition) equal(other *ViewPosition) bool {
	return viewPos.activeRowIndex == other.activeRowIndex &&
		viewPos.viewStartRowIndex == other.viewStartRowIndex &&
		viewPos.viewStartColumn == other.viewStartColumn
}

func newViewPos(activeRowIndex, viewStartRowIndex, viewStartColumn uint) *ViewPosition {
	return &ViewPosition{
		activeRowIndex:    activeRowIndex,
		viewStartRowIndex: viewStartRowIndex,
		viewStartColumn:   viewStartColumn,
	}
}

func checkViewPos(expected, actual *ViewPosition, t *testing.T) {
	if !expected.equal(actual) {
		t.Errorf("ViewPos did not match expected value. Expected: %v. Actual: %v", *expected, *actual)
	}
}

func checkViewPosResult(expected, actual bool, t *testing.T) {
	if expected != actual {
		t.Errorf("ViewPos function result did not match expected value. Expected: %v. Actual: %v", expected, actual)
	}
}

func TestViewPosIsCreatedWithExpectedFieldValues(t *testing.T) {
	expected := newViewPos(0, 0, 1)
	actual := NewViewPosition()

	checkViewPos(expected, actual, t)
}

func TestMoveLineDownIncrementsActiveRowIndex(t *testing.T) {
	expected := newViewPos(1, 0, 1)

	actual := NewViewPosition()
	result := actual.MoveLineDown(5)

	checkViewPos(expected, actual, t)
	checkViewPosResult(true, result, t)
}

func TestMoveLineDownDoesNotIncrementsActiveRowIndexIfNoRowsLeft(t *testing.T) {
	expected := newViewPos(4, 0, 1)

	actual := newViewPos(4, 0, 1)
	result := actual.MoveLineDown(5)

	checkViewPos(expected, actual, t)
	checkViewPosResult(false, result, t)
}

func TestMoveLineUpDecrementsActiveRowIndex(t *testing.T) {
	expected := newViewPos(3, 0, 1)

	actual := newViewPos(4, 0, 1)
	result := actual.MoveLineUp()

	checkViewPos(expected, actual, t)
	checkViewPosResult(true, result, t)
}

func TestMoveLineUpDoesNotDecrementActiveRowIndexIfOnFirstRow(t *testing.T) {
	expected := newViewPos(0, 0, 1)

	actual := newViewPos(0, 0, 1)
	result := actual.MoveLineUp()

	checkViewPos(expected, actual, t)
	checkViewPosResult(false, result, t)
}

func TestMovePageDownUpdatesActiveRowIndexAndViewStartRowIndex(t *testing.T) {
	expected := newViewPos(7, 7, 1)

	actual := newViewPos(2, 1, 1)
	result := actual.MovePageDown(5, 10)

	checkViewPos(expected, actual, t)
	checkViewPosResult(true, result, t)
}

func TestMovePageDownUpdatesActiveRowIndexAndViewStartRowIndexWithAvailableRowNum(t *testing.T) {
	expected := newViewPos(9, 9, 1)

	actual := newViewPos(7, 7, 1)
	result := actual.MovePageDown(5, 10)

	checkViewPos(expected, actual, t)
	checkViewPosResult(true, result, t)
}

func TestMovePageDownDoesNotUpdateActiveRowIndexAndViewStartRowIndexIfNoRowsLeft(t *testing.T) {
	expected := newViewPos(9, 9, 1)

	actual := newViewPos(9, 9, 1)
	result := actual.MovePageDown(5, 10)

	checkViewPos(expected, actual, t)
	checkViewPosResult(false, result, t)
}

func TestMovePageUpUpdatesActiveRowIndexAndViewStartRowIndex(t *testing.T) {
	expected := newViewPos(2, 2, 1)

	actual := newViewPos(7, 5, 1)
	result := actual.MovePageUp(5)

	checkViewPos(expected, actual, t)
	checkViewPosResult(true, result, t)
}

func TestMovePageUpUpdatesActiveRowIndexAndViewStartRowIndexWithAvailableRowNum(t *testing.T) {
	expected := newViewPos(0, 0, 1)

	actual := newViewPos(2, 1, 1)
	result := actual.MovePageUp(5)

	checkViewPos(expected, actual, t)
	checkViewPosResult(true, result, t)
}

func TestMovePageUpDoesNotUpdateActiveRowIndexAndViewStartRowIndexIfNoRowsLeft(t *testing.T) {
	expected := newViewPos(0, 0, 1)

	actual := newViewPos(0, 0, 1)
	result := actual.MovePageUp(5)

	checkViewPos(expected, actual, t)
	checkViewPosResult(false, result, t)
}

func TestMovePageRightIncreasesViewStartColumnByHalfPageSize(t *testing.T) {
	expected := newViewPos(0, 0, 6)

	actual := newViewPos(0, 0, 1)
	actual.MovePageRight(10)

	checkViewPos(expected, actual, t)
}

func TestMovePageLeftDecreasesViewStartColumnByHalfPage(t *testing.T) {
	expected := newViewPos(0, 0, 1)

	actual := newViewPos(0, 0, 6)
	result := actual.MovePageLeft(10)

	checkViewPos(expected, actual, t)
	checkViewPosResult(true, result, t)
}

func TestMovePageLeftDecreasesViewStartColumnByRemainingColumns(t *testing.T) {
	expected := newViewPos(0, 0, 1)

	actual := newViewPos(0, 0, 3)
	result := actual.MovePageLeft(10)

	checkViewPos(expected, actual, t)
	checkViewPosResult(true, result, t)
}

func TestMovePageLeftDoesNotDecreaseViewStartColumnIfNoColumnsRemain(t *testing.T) {
	expected := newViewPos(0, 0, 1)

	actual := newViewPos(0, 0, 1)
	result := actual.MovePageLeft(10)

	checkViewPos(expected, actual, t)
	checkViewPosResult(false, result, t)
}

func TestMoveToFirstLineUpdatesActiveRowIndex(t *testing.T) {
	expected := newViewPos(0, 5, 1)

	actual := newViewPos(20, 5, 1)
	result := actual.MoveToFirstLine()

	checkViewPos(expected, actual, t)
	checkViewPosResult(true, result, t)
}

func TestMoveToFirstLineDoesNotUpdateActiveRowIndexIfAlreadyOnFirstLine(t *testing.T) {
	expected := newViewPos(0, 0, 1)

	actual := newViewPos(0, 0, 1)
	result := actual.MoveToFirstLine()

	checkViewPos(expected, actual, t)
	checkViewPosResult(false, result, t)
}

func TestMoveToLastLineUpdatesActiveRowIndex(t *testing.T) {
	expected := newViewPos(9, 0, 1)

	actual := newViewPos(0, 0, 1)
	result := actual.MoveToLastLine(10)

	checkViewPos(expected, actual, t)
	checkViewPosResult(true, result, t)
}

func TestMoveToLastLineDoesNotUpdateActiveRowIndexIfAlreadyOnLastLine(t *testing.T) {
	expected := newViewPos(9, 9, 1)

	actual := newViewPos(9, 9, 1)
	result := actual.MoveToLastLine(10)

	checkViewPos(expected, actual, t)
	checkViewPosResult(false, result, t)
}

func TestDetermineViewStartRowSetsViewStartRowIndexToActiveRowIndexIfGreater(t *testing.T) {
	expected := newViewPos(5, 5, 1)

	actual := newViewPos(5, 9, 1)
	actual.DetermineViewStartRow(10, 20)

	checkViewPos(expected, actual, t)
}

func TestDetermineViewStartRowIncreasesViewStartRowIndexSoThatActiveRowIndexIsVisible(t *testing.T) {
	expected := newViewPos(15, 6, 1)

	actual := newViewPos(15, 2, 1)
	actual.DetermineViewStartRow(10, 20)

	checkViewPos(expected, actual, t)
}

func TestDetermineViewStartRowDecreasesViewStartRowIndexSoThatAsManyRowsAreVisibleAsPossible(t *testing.T) {
	expected := newViewPos(15, 10, 1)

	actual := newViewPos(15, 13, 1)
	actual.DetermineViewStartRow(10, 20)

	checkViewPos(expected, actual, t)
}

func TestDetermineViewStartRowDecreasesActiveRowIndexIfItIsGreaterThanTheTotalNumberOfRows(t *testing.T) {
	expected := newViewPos(19, 10, 1)

	actual := newViewPos(20, 10, 1)
	actual.DetermineViewStartRow(10, 20)

	checkViewPos(expected, actual, t)
}
