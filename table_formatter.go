package main

import (
	"fmt"
	"strings"
)

const (
	TF_SEPARATOR = " "
)

type TableCell struct {
	text             string
	themeComponentId ThemeComponentId
}

type TableFormatter struct {
	config       Config
	maxColWidths []uint
	cells        [][]TableCell
}

func NewTableFormatter(cols uint) *TableFormatter {
	return &TableFormatter{
		maxColWidths: make([]uint, cols),
	}
}

func (tableFormatter *TableFormatter) Rows() uint {
	return uint(len(tableFormatter.cells))
}

func (tableFormatter *TableFormatter) Cols() uint {
	if tableFormatter.Rows() > 0 {
		return uint(len(tableFormatter.cells[0]))
	}

	return 0
}

func (tableFormatter *TableFormatter) Resize(newRows uint) {
	rows := tableFormatter.Rows()

	if rows == newRows {
		return
	}

	cols := len(tableFormatter.maxColWidths)

	tableFormatter.cells = make([][]TableCell, newRows)
	for rowIndex := range tableFormatter.cells {
		tableFormatter.cells[rowIndex] = make([]TableCell, cols)
	}
}

func (tableFormatter *TableFormatter) Clear() {
	for rowIndex := range tableFormatter.cells {
		for colIndex := range tableFormatter.cells[rowIndex] {
			tableFormatter.cells[rowIndex][colIndex].text = ""
			tableFormatter.cells[rowIndex][colIndex].themeComponentId = CMP_NONE
		}
	}
}

func (tableFormatter *TableFormatter) SetCell(rowIndex, colIndex uint, format string, args ...interface{}) (err error) {
	return tableFormatter.SetCellWithStyle(rowIndex, colIndex, CMP_NONE, format, args...)
}

func (tableFormatter *TableFormatter) SetCellWithStyle(rowIndex, colIndex uint, themeComponentId ThemeComponentId, format string, args ...interface{}) (err error) {
	if !(rowIndex < tableFormatter.Rows() && colIndex < tableFormatter.Cols()) {
		return fmt.Errorf("Invalid rowIndex (%v), colIndex (%v) for dimensions rows (%v), cols (%v)",
			rowIndex, colIndex, tableFormatter.Rows(), tableFormatter.Cols())
	}

	tableCell := &tableFormatter.cells[rowIndex][colIndex]

	tableCell.text = fmt.Sprintf(format, args...)
	tableCell.themeComponentId = themeComponentId

	return
}

func (tableFormatter *TableFormatter) Render(win RenderWindow, viewStartColumn uint, border bool) (err error) {
	tableFormatter.padCells(border)

	var lineBuilder *LineBuilder

	for rowIndex := range tableFormatter.cells {
		adjustedRowIndex := uint(rowIndex)
		if border {
			adjustedRowIndex++
		}

		if lineBuilder, err = win.LineBuilder(adjustedRowIndex, viewStartColumn); err != nil {
			return
		}

		if border {
			lineBuilder.Append(" ")
		}

		for colIndex := range tableFormatter.cells[rowIndex] {
			tableCell := &tableFormatter.cells[rowIndex][colIndex]
			lineBuilder.
				AppendWithStyle(tableCell.themeComponentId, "%v", tableCell.text).
				Append(TF_SEPARATOR)
		}
	}

	if border {
		win.DrawBorder()
	}

	return
}

func (tableFormatter *TableFormatter) padCells(border bool) {
	tableFormatter.determineMaxColWidths(border)

	for rowIndex := range tableFormatter.cells {
		column := uint(1)

		if border {
			column++
		}

		for colIndex := range tableFormatter.cells[rowIndex] {
			width := tableFormatter.textWidth(rowIndex, colIndex, column)
			maxColWidth := tableFormatter.maxColWidths[colIndex]

			if width < maxColWidth {
				tableCell := &tableFormatter.cells[rowIndex][colIndex]
				tableCell.text += strings.Repeat(" ", int(maxColWidth-width))
			}

			column += maxColWidth
		}

		column += uint(len(TF_SEPARATOR))
	}
}

func (tableFormatter *TableFormatter) determineMaxColWidths(border bool) {
	for colIndex := 0; colIndex < len(tableFormatter.maxColWidths); colIndex++ {
		column := uint(1)

		if border {
			column++
		}

		for doneColIndex := 0; doneColIndex < colIndex; doneColIndex++ {
			column += tableFormatter.maxColWidths[doneColIndex]
			column += uint(len(TF_SEPARATOR))
		}

		for rowIndex := range tableFormatter.cells {
			width := tableFormatter.textWidth(rowIndex, colIndex, column)

			if width > tableFormatter.maxColWidths[colIndex] {
				tableFormatter.maxColWidths[colIndex] = width
			}
		}
	}

}

func (tableFormatter *TableFormatter) textWidth(rowIndex, colIndex int, column uint) (width uint) {
	text := tableFormatter.cells[rowIndex][colIndex].text

	for _, codePoint := range text {
		renderedCodePoints := DetermineRenderedCodePoint(codePoint, column, tableFormatter.config)

		for _, renderedCodePoint := range renderedCodePoints {
			width += renderedCodePoint.width
			column += renderedCodePoint.width
		}
	}

	return
}
