package main

// ViewPos manages the display positioning of a view
type ViewPos interface {
	ActiveRowIndex() uint
	SetActiveRowIndex(activeRowIndex uint)
	ViewStartRowIndex() uint
	ViewStartColumn() uint
	SelectedRowIndex() uint
	DetermineViewStartRow(viewRows, rows uint)
	MoveLineDown(rows uint) (changed bool)
	MoveLineUp() (changed bool)
	MovePageDown(pageRows, rows uint) (changed bool)
	MovePageUp(pageRows uint) (changed bool)
	MovePageRight(cols uint)
	MovePageLeft(cols uint) (changed bool)
	MoveToFirstLine() (changed bool)
	MoveToLastLine(rows uint) (changed bool)
	CenterActiveRow(pageRows uint) (changed bool)
	ScrollActiveRowTop() (changed bool)
	ScrollActiveRowBottom(pageRows uint) (changed bool)
	MoveCursorTopPage() (changed bool)
	MoveCursorMiddlePage(pageRows, rows uint) (changed bool)
	MoveCursorBottomPage(pageRows, rows uint) (changed bool)
	ScrollDown(rows, pageRows, scrollRows uint) (changed bool)
	ScrollUp(pageRows, scrollRows uint) (changed bool)
}

// ViewPosition implements the ViewPos interface
type ViewPosition struct {
	activeRowIndex    uint
	viewStartRowIndex uint
	viewStartColumn   uint
}

// NewViewPosition creates a new instance
func NewViewPosition() *ViewPosition {
	return &ViewPosition{
		activeRowIndex:    0,
		viewStartRowIndex: 0,
		viewStartColumn:   1,
	}
}

// ActiveRowIndex returns the row index the curosr is on
func (viewPos *ViewPosition) ActiveRowIndex() uint {
	return viewPos.activeRowIndex
}

// SetActiveRowIndex sets the row index the cursor is on
func (viewPos *ViewPosition) SetActiveRowIndex(activeRowIndex uint) {
	viewPos.activeRowIndex = activeRowIndex
}

// ViewStartRowIndex returns the row index the view should be drawn from
func (viewPos *ViewPosition) ViewStartRowIndex() uint {
	return viewPos.viewStartRowIndex
}

// ViewStartColumn returns the column the display should be drawn from
func (viewPos *ViewPosition) ViewStartColumn() uint {
	return viewPos.viewStartColumn
}

// SelectedRowIndex calculates the offset of the active row
func (viewPos *ViewPosition) SelectedRowIndex() uint {
	return viewPos.activeRowIndex - viewPos.viewStartRowIndex
}

// DetermineViewStartRow determines the row the view should start displaying from based on the current cursor position
func (viewPos *ViewPosition) DetermineViewStartRow(viewRows, rows uint) {
	if rows > 0 && viewPos.activeRowIndex >= rows {
		viewPos.activeRowIndex = rows - 1
	}

	if viewPos.viewStartRowIndex > viewPos.activeRowIndex {
		viewPos.viewStartRowIndex = viewPos.activeRowIndex
	} else if rowDiff := viewPos.activeRowIndex - viewPos.viewStartRowIndex; rowDiff >= viewRows {
		viewPos.viewStartRowIndex += (rowDiff - viewRows) + 1
	} else if visibleRows := rows - (viewPos.viewStartRowIndex + 1); visibleRows < viewRows && viewPos.viewStartRowIndex > 0 {
		viewPos.viewStartRowIndex -= MinUInt(viewPos.viewStartRowIndex, (viewRows-visibleRows)-1)
	}
}

// MoveLineDown moves the cursor down one line
func (viewPos *ViewPosition) MoveLineDown(rows uint) (changed bool) {
	if viewPos.activeRowIndex+1 < rows {
		viewPos.activeRowIndex++
		changed = true
	}

	return
}

// MoveLineUp moves the cursor up one line
func (viewPos *ViewPosition) MoveLineUp() (changed bool) {
	if viewPos.activeRowIndex > 0 {
		viewPos.activeRowIndex--
		changed = true
	}

	return
}

// MovePageDown moves the cursor and display down a page
func (viewPos *ViewPosition) MovePageDown(pageRows, rows uint) (changed bool) {
	if viewPos.activeRowIndex+1 < rows {
		viewPos.activeRowIndex += MinUInt(pageRows, rows-(viewPos.activeRowIndex+1))
		viewPos.viewStartRowIndex = viewPos.activeRowIndex
		changed = true
	}

	return
}

// MovePageUp moves the cursor and display up a page
func (viewPos *ViewPosition) MovePageUp(pageRows uint) (changed bool) {
	if viewPos.activeRowIndex > 0 {
		viewPos.activeRowIndex -= MinUInt(pageRows, viewPos.activeRowIndex)
		viewPos.viewStartRowIndex = viewPos.activeRowIndex
		changed = true
	}

	return
}

// MovePageRight scrolls the view right a page (half the available view width)
func (viewPos *ViewPosition) MovePageRight(cols uint) {
	halfPage := cols / 2
	viewPos.viewStartColumn += halfPage
}

// MovePageLeft scrolls the view left a page (half the available view width)
func (viewPos *ViewPosition) MovePageLeft(cols uint) (changed bool) {
	if viewPos.viewStartColumn > 1 {
		halfPage := cols / 2

		if halfPage > viewPos.viewStartColumn {
			viewPos.viewStartColumn = 1
		} else {
			viewPos.viewStartColumn -= halfPage
		}

		changed = true
	}

	return
}

// MoveToFirstLine moves the cursor to the first line of the view
func (viewPos *ViewPosition) MoveToFirstLine() (changed bool) {
	if viewPos.activeRowIndex > 0 {
		viewPos.activeRowIndex = 0
		changed = true
	}

	return
}

// MoveToLastLine moves the cursor to the last line of the view
func (viewPos *ViewPosition) MoveToLastLine(rows uint) (changed bool) {
	if rows > 0 && viewPos.activeRowIndex+1 != rows {
		viewPos.activeRowIndex = rows - 1
		changed = true
	}

	return
}

// CenterActiveRow updates the view start position to center the cursor
func (viewPos *ViewPosition) CenterActiveRow(pageRows uint) (changed bool) {
	selectedRow := viewPos.SelectedRowIndex()
	centerRow := (pageRows / 2)

	if selectedRow > centerRow {
		viewPos.viewStartRowIndex += selectedRow - centerRow
		changed = true
	} else if centerRow > selectedRow {
		viewPos.viewStartRowIndex -= MinUInt(centerRow-selectedRow, viewPos.viewStartRowIndex)
		changed = true
	}

	return
}

// ScrollActiveRowTop updates the view start position to the cursor
func (viewPos *ViewPosition) ScrollActiveRowTop() (changed bool) {
	selectedRow := viewPos.SelectedRowIndex()

	if selectedRow != 0 {
		viewPos.viewStartRowIndex = viewPos.activeRowIndex
		changed = true
	}

	return
}

// ScrollActiveRowBottom updates the view bottom position to the cursor
func (viewPos *ViewPosition) ScrollActiveRowBottom(pageRows uint) (changed bool) {
	selectedRow := viewPos.SelectedRowIndex()

	if selectedRow != pageRows-1 {
		viewPos.viewStartRowIndex = uint(MaxInt(int(viewPos.activeRowIndex-(pageRows-1)), 0))
		changed = true
	}

	return
}

// MoveCursorTopPage moves the cursor to top of the page
func (viewPos *ViewPosition) MoveCursorTopPage() (changed bool) {
	firstRowInPage := viewPos.viewStartRowIndex

	if viewPos.activeRowIndex != firstRowInPage {
		viewPos.activeRowIndex = firstRowInPage
		changed = true
	}

	return
}

// MoveCursorMiddlePage moves the cursor to middle of the page
func (viewPos *ViewPosition) MoveCursorMiddlePage(pageRows, rows uint) (changed bool) {
	middleRowInPage := viewPos.viewStartRowIndex + MinUInt(pageRows, rows-viewPos.viewStartRowIndex)/2

	if viewPos.activeRowIndex != middleRowInPage {
		viewPos.activeRowIndex = middleRowInPage
		changed = true
	}

	return
}

// MoveCursorBottomPage moves the cursor to bottom of the page
func (viewPos *ViewPosition) MoveCursorBottomPage(pageRows, rows uint) (changed bool) {
	lastRowInPage := MinUInt(viewPos.viewStartRowIndex+pageRows-1, rows-1)

	if viewPos.activeRowIndex != lastRowInPage {
		viewPos.activeRowIndex = lastRowInPage
		changed = true
	}

	return
}

// ScrollDown scrolls the view down by scrollRows
func (viewPos *ViewPosition) ScrollDown(rows, pageRows, scrollRows uint) (changed bool) {
	if viewPos.viewStartRowIndex+scrollRows < rows {
		viewPos.viewStartRowIndex += scrollRows
		changed = true
	}

	if changed {
		if viewPos.activeRowIndex < viewPos.viewStartRowIndex {
			viewPos.activeRowIndex = viewPos.viewStartRowIndex

			viewPos.DetermineViewStartRow(pageRows, rows)

			if viewPos.activeRowIndex > viewPos.viewStartRowIndex {
				viewPos.activeRowIndex = viewPos.viewStartRowIndex
			}
		}
	}

	return
}

// ScrollUp scrolls the view up by scrollRows
func (viewPos *ViewPosition) ScrollUp(pageRows, scrollRows uint) (changed bool) {
	if viewPos.viewStartRowIndex >= scrollRows {
		viewPos.viewStartRowIndex -= scrollRows
		changed = true
	} else if viewPos.viewStartRowIndex != 0 {
		viewPos.viewStartRowIndex = 0
		changed = true
	}

	if changed {
		if viewPos.activeRowIndex >= viewPos.viewStartRowIndex+pageRows {
			viewPos.activeRowIndex = (viewPos.viewStartRowIndex + pageRows) - 1
		}
	}

	return
}
