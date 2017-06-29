package main

import (
	"bytes"
	"fmt"
	"strings"
)

const (
	TF_SEPARATOR = " "
)

type TableCellText struct {
	text             string
	themeComponentId ThemeComponentId
}

type TableCell struct {
	textEntries []TableCellText
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
			tableFormatter.cells[rowIndex][colIndex].textEntries = nil
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

	tableCell.textEntries = []TableCellText{
		TableCellText{
			text:             fmt.Sprintf(format, args...),
			themeComponentId: themeComponentId,
		},
	}

	return
}

func (tableFormatter *TableFormatter) AppendToCell(rowIndex, colIndex uint, format string, args ...interface{}) (err error) {
	return tableFormatter.AppendToCellWithStyle(rowIndex, colIndex, CMP_NONE, format, args...)
}

func (tableFormatter *TableFormatter) AppendToCellWithStyle(rowIndex, colIndex uint, themeComponentId ThemeComponentId, format string, args ...interface{}) (err error) {
	if !(rowIndex < tableFormatter.Rows() && colIndex < tableFormatter.Cols()) {
		return fmt.Errorf("Invalid rowIndex (%v), colIndex (%v) for dimensions rows (%v), cols (%v)",
			rowIndex, colIndex, tableFormatter.Rows(), tableFormatter.Cols())
	}

	tableCell := &tableFormatter.cells[rowIndex][colIndex]

	tableCell.textEntries = append(tableCell.textEntries, TableCellText{
		text:             fmt.Sprintf(format, args...),
		themeComponentId: themeComponentId,
	})

	return
}

func (tableFormatter *TableFormatter) RowString(rowIndex uint) (rowString string, err error) {
	if rowIndex >= tableFormatter.Rows() {
		err = fmt.Errorf("Invalid rowIndex: %v, total rows %v", rowIndex, tableFormatter.Rows())
		return
	}

	var buf bytes.Buffer

	for colIndex := range tableFormatter.cells[rowIndex] {
		for _, textEntry := range tableFormatter.cells[rowIndex][colIndex].textEntries {
			buf.WriteString(textEntry.text)
		}

		buf.WriteString(TF_SEPARATOR)
	}

	rowString = buf.String()
	return
}

func (tableFormatter *TableFormatter) Render(win RenderWindow, viewStartColumn uint, border bool) (err error) {
	tableFormatter.PadCells(border)

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

			for _, textEntry := range tableCell.textEntries {
				lineBuilder.AppendWithStyle(textEntry.themeComponentId, "%v", textEntry.text)
			}

			lineBuilder.Append(TF_SEPARATOR)
		}
	}

	if border {
		win.DrawBorder()
	}

	return
}

func (tableFormatter *TableFormatter) PadCells(border bool) {
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
				tableFormatter.AppendToCell(uint(rowIndex), uint(colIndex), strings.Repeat(" ", int(maxColWidth-width)))
			}

			column += maxColWidth + uint(len(TF_SEPARATOR))
		}
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
	textEntries := tableFormatter.cells[rowIndex][colIndex].textEntries

	for _, textEntry := range textEntries {
		for _, codePoint := range textEntry.text {
			renderedCodePoints := DetermineRenderedCodePoint(codePoint, column, tableFormatter.config)

			for _, renderedCodePoint := range renderedCodePoints {
				width += renderedCodePoint.width
				column += renderedCodePoint.width
			}
		}
	}

	return
}
