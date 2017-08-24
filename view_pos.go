package main

// ViewPos manages the display positioning of a view
type ViewPos struct {
	activeRowIndex    uint
	viewStartRowIndex uint
	viewStartColumn   uint
}

// NewViewPos creates a new instance
func NewViewPos() *ViewPos {
	return &ViewPos{
		activeRowIndex:    0,
		viewStartRowIndex: 0,
		viewStartColumn:   1,
	}
}

// DetermineViewStartRow determines the row the view should start displaying from based on the current cursor position
func (viewPos *ViewPos) DetermineViewStartRow(viewRows, rows uint) {
	if rows > 0 && viewPos.activeRowIndex >= rows {
		viewPos.activeRowIndex = rows - 1
	}

	if viewPos.viewStartRowIndex > viewPos.activeRowIndex {
		viewPos.viewStartRowIndex = viewPos.activeRowIndex
	} else if rowDiff := viewPos.activeRowIndex - viewPos.viewStartRowIndex; rowDiff >= viewRows {
		viewPos.viewStartRowIndex += (rowDiff - viewRows) + 1
	} else if visibleRows := rows - (viewPos.viewStartRowIndex + 1); visibleRows < viewRows && viewPos.viewStartRowIndex > 0 {
		viewPos.viewStartRowIndex -= Min(viewPos.viewStartRowIndex, (viewRows-visibleRows)-1)
	}
}

// MoveLineDown moves the cursor down one line
func (viewPos *ViewPos) MoveLineDown(rows uint) (changed bool) {
	if viewPos.activeRowIndex+1 < rows {
		viewPos.activeRowIndex++
		changed = true
	}

	return
}

// MoveLineUp moves the cursor up one line
func (viewPos *ViewPos) MoveLineUp() (changed bool) {
	if viewPos.activeRowIndex > 0 {
		viewPos.activeRowIndex--
		changed = true
	}

	return
}

// MovePageDown moves the cursor and display down a page
func (viewPos *ViewPos) MovePageDown(pageRows, rows uint) (changed bool) {
	if viewPos.activeRowIndex+1 < rows {
		viewPos.activeRowIndex += Min(pageRows, rows-(viewPos.activeRowIndex+1))
		viewPos.viewStartRowIndex = viewPos.activeRowIndex
		changed = true
	}

	return
}

// MovePageUp moves the cursor and display up a page
func (viewPos *ViewPos) MovePageUp(pageRows uint) (changed bool) {
	if viewPos.activeRowIndex > 0 {
		viewPos.activeRowIndex -= Min(pageRows, viewPos.activeRowIndex)
		viewPos.viewStartRowIndex = viewPos.activeRowIndex
		changed = true
	}

	return
}

// MovePageRight scrolls the view right a page (half the available view width)
func (viewPos *ViewPos) MovePageRight(cols uint) {
	halfPage := cols / 2
	viewPos.viewStartColumn += halfPage
}

// MovePageLeft scrolls the view left a page (half the available view width)
func (viewPos *ViewPos) MovePageLeft(cols uint) (changed bool) {
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
func (viewPos *ViewPos) MoveToFirstLine() (changed bool) {
	if viewPos.activeRowIndex > 0 {
		viewPos.activeRowIndex = 0
		changed = true
	}

	return
}

// MoveToLastLine moves the cursor to the last line of the view
func (viewPos *ViewPos) MoveToLastLine(rows uint) (changed bool) {
	if rows > 0 && viewPos.activeRowIndex+1 != rows {
		viewPos.activeRowIndex = rows - 1
		changed = true
	}

	return
}
