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

func (viewPos *ViewPos) DetermineViewStartRow(rows uint) {
	if viewPos.viewStartRowIndex > viewPos.activeRowIndex {
		viewPos.viewStartRowIndex = viewPos.activeRowIndex
	} else if rowDiff := viewPos.activeRowIndex - viewPos.viewStartRowIndex; rowDiff >= rows {
		viewPos.viewStartRowIndex += (rowDiff - rows) + 1
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
