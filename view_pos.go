package main

type ViewPos struct {
	activeRowIndex    uint
	viewStartRowIndex uint
	viewStartColumn   uint
}

func NewViewPos() *ViewPos {
	return &ViewPos{
		activeRowIndex:    0,
		viewStartRowIndex: 0,
		viewStartColumn:   1,
	}
}

func (viewPos *ViewPos) DetermineViewStartRow(viewRows, rows uint) {
	if viewPos.viewStartRowIndex > viewPos.activeRowIndex {
		viewPos.viewStartRowIndex = viewPos.activeRowIndex
	} else if rowDiff := viewPos.activeRowIndex - viewPos.viewStartRowIndex; rowDiff >= viewRows {
		viewPos.viewStartRowIndex += (rowDiff - viewRows) + 1
	} else if visibleRows := rows - (viewPos.viewStartRowIndex + 1); visibleRows < viewRows && viewPos.viewStartRowIndex > 0 {
		viewPos.viewStartRowIndex -= Min(viewPos.viewStartRowIndex, (viewRows-visibleRows)-1)
	}
}

func (viewPos *ViewPos) MoveLineDown(rows uint) (changed bool) {
	if viewPos.activeRowIndex+1 < rows {
		viewPos.activeRowIndex++
		changed = true
	}

	return
}

func (viewPos *ViewPos) MoveLineUp() (changed bool) {
	if viewPos.activeRowIndex > 0 {
		viewPos.activeRowIndex--
		changed = true
	}

	return
}

func (viewPos *ViewPos) MovePageRight(cols uint) {
	halfPage := cols / 2
	viewPos.viewStartColumn += halfPage
}

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

func (viewPos *ViewPos) MoveToFirstLine() (changed bool) {
	if viewPos.activeRowIndex > 0 {
		viewPos.activeRowIndex = 0
		changed = true
	}

	return
}

func (viewPos *ViewPos) MoveToLastLine(rows uint) (changed bool) {
	if rows > 0 && viewPos.activeRowIndex+1 != rows {
		viewPos.activeRowIndex = rows - 1
		changed = true
	}

	return
}
