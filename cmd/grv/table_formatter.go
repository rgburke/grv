package main

import (
	"bytes"
	"fmt"
	"strings"
)

const (
	tfSeparator = " "
)

// TableCellText contains cell text and style data
type TableCellText struct {
	text             string
	themeComponentID ThemeComponentID
}

// TableCell represents a cell entry a table
// It contains all text and style data for a cell
type TableCell struct {
	textEntries []TableCellText
}

// TableHeader represents a column header
type TableHeader struct {
	text             string
	themeComponentID ThemeComponentID
}

// CellRendererListener is notified before and after a cell is rendered
// It has the option to modify the way the cell is rendered
type CellRendererListener interface {
	preRenderCell(rowIndex, colIndex uint, lineBuilder *LineBuilder, tableCell *TableCell) (err error)
	postRenderCell(rowIndex, colIndex uint, lineBuilder *LineBuilder, tableCell *TableCell) (err error)
}

// TableFormatter renders provided data in a tabular layout
type TableFormatter struct {
	config                Config
	maxColWidths          []uint
	headers               []TableHeader
	gridLines             bool
	borderColWidth        uint
	cells                 [][]TableCell
	cellRendererListeners map[uint]CellRendererListener
}

// NewTableFormatter creates a new instance of the table formatter supporting the specified number of columns
func NewTableFormatter(cols uint, config Config) *TableFormatter {
	return &TableFormatter{
		maxColWidths:          make([]uint, cols),
		cellRendererListeners: make(map[uint]CellRendererListener),
		borderColWidth:        1,
		config:                config,
	}
}

// NewTableFormatterWithHeaders creates a new instance of the table formatter using the provided headers
func NewTableFormatterWithHeaders(headers []TableHeader, config Config) *TableFormatter {
	tableFormatter := NewTableFormatter(uint(len(headers)), config)
	tableFormatter.SetHeaders(headers)
	return tableFormatter
}

// SetCellRendererListener sets the CellRendererListener for a column
func (tableFormatter *TableFormatter) SetCellRendererListener(colIndex uint, cellRendererListener CellRendererListener) (err error) {
	if colIndex >= tableFormatter.cols() {
		return fmt.Errorf("Cannot register CellRendererListener on column with index: %v", colIndex)
	}

	tableFormatter.cellRendererListeners[colIndex] = cellRendererListener

	return
}

// Rows returns the number of rows in the table formatter
func (tableFormatter *TableFormatter) Rows() uint {
	return uint(len(tableFormatter.cells))
}

// RenderedRows returns the number of rows that will actually be rendered
func (tableFormatter *TableFormatter) RenderedRows() uint {
	rows := tableFormatter.Rows()

	if tableFormatter.hasHeaders() {
		rows++

		if tableFormatter.gridLines {
			rows++
		}
	}

	return rows
}

// Cols returns the number of cols in the table formatter
func (tableFormatter *TableFormatter) Cols() uint {
	if tableFormatter.Rows() > 0 {
		return uint(len(tableFormatter.cells[0]))
	}

	return 0
}

func (tableFormatter *TableFormatter) cols() uint {
	return uint(len(tableFormatter.maxColWidths))
}

// SetHeaders sets the column headers
func (tableFormatter *TableFormatter) SetHeaders(headers []TableHeader) (err error) {
	headersCols := uint(len(headers))
	tableCols := tableFormatter.cols()

	if headersCols != tableCols {
		return fmt.Errorf("Headers is invalid size. Allowed %v but found %v", tableCols, headersCols)
	}

	tableFormatter.headers = headers

	return
}

func (tableFormatter *TableFormatter) hasHeaders() bool {
	return len(tableFormatter.headers) > 0
}

// SetGridLines sets whether gridlines should be rendered
func (tableFormatter *TableFormatter) SetGridLines(gridLines bool) {
	tableFormatter.gridLines = gridLines
}

// SetBorderColumnWidth sets the number of spaces padded to the left of the table
func (tableFormatter *TableFormatter) SetBorderColumnWidth(width uint) {
	tableFormatter.borderColWidth = width
}

// Resize updates the number of rows the tableformatter can store
func (tableFormatter *TableFormatter) Resize(newRows uint) {
	rows := tableFormatter.Rows()

	if rows == newRows {
		return
	}

	cols := tableFormatter.cols()

	tableFormatter.cells = make([][]TableCell, newRows)
	for rowIndex := range tableFormatter.cells {
		tableFormatter.cells[rowIndex] = make([]TableCell, cols)
	}
}

// Clear text in all cells
func (tableFormatter *TableFormatter) Clear() {
	for rowIndex := range tableFormatter.cells {
		for colIndex := range tableFormatter.cells[rowIndex] {
			tableFormatter.cells[rowIndex][colIndex].textEntries = nil
		}
	}
}

// SetCell sets the text value of the cell at the specified coordinates
func (tableFormatter *TableFormatter) SetCell(rowIndex, colIndex uint, format string, args ...interface{}) (err error) {
	return tableFormatter.SetCellWithStyle(rowIndex, colIndex, CmpNone, format, args...)
}

// SetCellWithStyle sets text with style information value of the cell at the specified coordinates
func (tableFormatter *TableFormatter) SetCellWithStyle(rowIndex, colIndex uint, themeComponentID ThemeComponentID, format string, args ...interface{}) (err error) {
	if !(rowIndex < tableFormatter.Rows() && colIndex < tableFormatter.Cols()) {
		return fmt.Errorf("Invalid rowIndex (%v), colIndex (%v) for dimensions rows (%v), cols (%v)",
			rowIndex, colIndex, tableFormatter.Rows(), tableFormatter.Cols())
	}

	tableCell := &tableFormatter.cells[rowIndex][colIndex]

	tableCell.textEntries = []TableCellText{
		{
			text:             fmt.Sprintf(format, args...),
			themeComponentID: themeComponentID,
		},
	}

	return
}

// AppendToCell appends text to the specified cell
func (tableFormatter *TableFormatter) AppendToCell(rowIndex, colIndex uint, format string, args ...interface{}) (err error) {
	return tableFormatter.AppendToCellWithStyle(rowIndex, colIndex, CmpNone, format, args...)
}

// AppendToCellWithStyle appends text with style information to the specified cell
func (tableFormatter *TableFormatter) AppendToCellWithStyle(rowIndex, colIndex uint, themeComponentID ThemeComponentID, format string, args ...interface{}) (err error) {
	if !(rowIndex < tableFormatter.Rows() && colIndex < tableFormatter.Cols()) {
		return fmt.Errorf("Invalid rowIndex (%v), colIndex (%v) for dimensions rows (%v), cols (%v)",
			rowIndex, colIndex, tableFormatter.Rows(), tableFormatter.Cols())
	}

	tableCell := &tableFormatter.cells[rowIndex][colIndex]

	tableCell.textEntries = append(tableCell.textEntries, TableCellText{
		text:             fmt.Sprintf(format, args...),
		themeComponentID: themeComponentID,
	})

	return
}

// RowString returns the string representation of the row at the specified index
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

		buf.WriteString(tfSeparator)
	}

	rowString = buf.String()
	return
}

// Render pads the content of the table formatter and writes it to the provided window
func (tableFormatter *TableFormatter) Render(win RenderWindow, viewStartColumn uint, border bool) (err error) {
	if err = tableFormatter.PadCells(border); err != nil {
		return
	}

	winStartRowIndex := uint(0)
	if border {
		winStartRowIndex++
	}

	for rowIndex := uint(0); rowIndex < tableFormatter.RenderedRows(); rowIndex++ {
		if err = tableFormatter.RenderRow(win, winStartRowIndex, rowIndex, viewStartColumn, border); err != nil {
			return
		}
	}

	return
}

// RenderRow renders the specified row of the table to the provided window
func (tableFormatter *TableFormatter) RenderRow(win RenderWindow, winStartRowIndex, renderedRowIndex, viewStartColumn uint, border bool) (err error) {
	lineIndex := renderedRowIndex + winStartRowIndex

	lineBuilder, err := win.LineBuilder(lineIndex, viewStartColumn)
	if err != nil {
		return
	}

	if border {
		lineBuilder.Append(strings.Repeat(" ", int(tableFormatter.borderColWidth)))
	}

	rowIndex := renderedRowIndex

	if tableFormatter.hasHeaders() {
		if renderedRowIndex == 0 {
			tableFormatter.renderHeaders(lineBuilder)
			return
		} else if tableFormatter.gridLines && renderedRowIndex == 1 {
			tableFormatter.renderHeaderGridLines(lineBuilder)
			return
		} else {
			rowIndex--

			if tableFormatter.gridLines {
				rowIndex--
			}
		}
	}

	for colIndex := range tableFormatter.cells[rowIndex] {
		tableCell := &tableFormatter.cells[rowIndex][colIndex]

		if err = tableFormatter.firePreCellRenderListener(rowIndex, uint(colIndex), lineBuilder, tableCell); err != nil {
			return
		}

		for _, textEntry := range tableCell.textEntries {
			lineBuilder.AppendWithStyle(textEntry.themeComponentID, "%v", textEntry.text)
		}

		if err = tableFormatter.firePostCellRenderListener(rowIndex, uint(colIndex), lineBuilder, tableCell); err != nil {
			return
		}

		tableFormatter.appendSeparator(lineBuilder, uint(colIndex))
	}

	return
}

func (tableFormatter *TableFormatter) firePreCellRenderListener(rowIndex, colIndex uint, lineBuilder *LineBuilder, tableCell *TableCell) (err error) {
	cellRendererListener, exists := tableFormatter.cellRendererListeners[colIndex]

	if exists {
		err = cellRendererListener.preRenderCell(rowIndex, colIndex, lineBuilder, tableCell)
	}

	return
}

func (tableFormatter *TableFormatter) firePostCellRenderListener(rowIndex, colIndex uint, lineBuilder *LineBuilder, tableCell *TableCell) (err error) {
	cellRendererListener, exists := tableFormatter.cellRendererListeners[colIndex]

	if exists {
		err = cellRendererListener.postRenderCell(rowIndex, colIndex, lineBuilder, tableCell)
	}

	return
}

func (tableFormatter *TableFormatter) renderHeaders(lineBuilder *LineBuilder) {
	for colIndex, header := range tableFormatter.headers {
		lineBuilder.AppendWithStyle(header.themeComponentID, "%v", header.text)
		tableFormatter.appendSeparator(lineBuilder, uint(colIndex))
	}
}

func (tableFormatter *TableFormatter) renderHeaderGridLines(lineBuilder *LineBuilder) {
	for maxColIndex, maxColWidth := range tableFormatter.maxColWidths {
		for i := uint(0); i < maxColWidth+1; i++ {
			lineBuilder.AppendACSChar(AcsHline, CmpNone)
		}

		if uint(maxColIndex) != tableFormatter.cols()-1 {
			lineBuilder.AppendACSChar(AcsPlus, CmpNone)
			lineBuilder.AppendACSChar(AcsHline, CmpNone)
		}
	}
}

// RenderRowText generates a textual representation of a rendered row
func (tableFormatter *TableFormatter) RenderRowText(renderedRowIndex uint) string {
	rowIndex := renderedRowIndex
	var buffer bytes.Buffer

	if tableFormatter.hasHeaders() {
		if renderedRowIndex == 0 {
			for colIndex, header := range tableFormatter.headers {
				buffer.WriteString(header.text)

				if tableFormatter.gridLines && colIndex != len(tableFormatter.headers)-1 {
					buffer.WriteString(" | ")
				}
			}

			return buffer.String()
		} else if tableFormatter.gridLines && renderedRowIndex == 1 {
			for maxColIndex, maxColWidth := range tableFormatter.maxColWidths {
				for i := uint(0); i < maxColWidth+1; i++ {
					buffer.WriteRune('-')
				}

				if uint(maxColIndex) != tableFormatter.cols()-1 {
					buffer.WriteRune('+')
					buffer.WriteRune('-')
				}
			}

			return buffer.String()
		} else {
			rowIndex--

			if tableFormatter.gridLines {
				rowIndex--
			}
		}
	}

	for colIndex := range tableFormatter.cells[rowIndex] {
		tableCell := &tableFormatter.cells[rowIndex][colIndex]

		for _, textEntry := range tableCell.textEntries {
			buffer.WriteString(textEntry.text)
		}

		if colIndex != len(tableFormatter.cells[rowIndex])-1 {
			buffer.WriteString(" | ")
		}
	}

	return buffer.String()
}

// PadCells pads each cell with whitespace so that the text in each column is of uniform width
func (tableFormatter *TableFormatter) PadCells(border bool) (err error) {
	tableFormatter.determineMaxColWidths(border)

	column := uint(1)

	for colIndex, header := range tableFormatter.headers {
		if border {
			column += tableFormatter.borderColWidth
		}

		width := tableFormatter.textWidth(header.text, column)
		maxColWidth := tableFormatter.maxColWidths[colIndex]

		if width < maxColWidth {
			tableFormatter.headers[colIndex].text = header.text + strings.Repeat(" ", int(maxColWidth-width))
		}

		column += maxColWidth + uint(len(tfSeparator))
	}

	for rowIndex := range tableFormatter.cells {
		column = uint(1)

		if border {
			column += tableFormatter.borderColWidth
		}

		for colIndex := range tableFormatter.cells[rowIndex] {
			width := tableFormatter.cellTextWidth(uint(rowIndex), uint(colIndex), column)
			maxColWidth := tableFormatter.maxColWidths[colIndex]

			if width < maxColWidth {
				if err = tableFormatter.AppendToCell(uint(rowIndex), uint(colIndex), strings.Repeat(" ", int(maxColWidth-width))); err != nil {
					return
				}
			}

			column += maxColWidth + tableFormatter.separatorWidth()
		}
	}

	return
}

func (tableFormatter *TableFormatter) determineMaxColWidths(border bool) {
	for colIndex := uint(0); colIndex < tableFormatter.cols(); colIndex++ {
		column := uint(1)

		if border {
			column += tableFormatter.borderColWidth
		}

		for doneColIndex := uint(0); doneColIndex < colIndex; doneColIndex++ {
			column += tableFormatter.maxColWidths[doneColIndex]
			column += tableFormatter.separatorWidth()
		}

		if tableFormatter.hasHeaders() {
			width := tableFormatter.textWidth(tableFormatter.headers[colIndex].text, column)

			if width > tableFormatter.maxColWidths[colIndex] {
				tableFormatter.maxColWidths[colIndex] = width
			}
		}

		for rowIndex := range tableFormatter.cells {
			width := tableFormatter.cellTextWidth(uint(rowIndex), colIndex, column)

			if width > tableFormatter.maxColWidths[colIndex] {
				tableFormatter.maxColWidths[colIndex] = width
			}
		}
	}

}

func (tableFormatter *TableFormatter) cellTextWidth(rowIndex, colIndex uint, column uint) (width uint) {
	textEntries := tableFormatter.cells[rowIndex][colIndex].textEntries

	for _, textEntry := range textEntries {
		textWidth := tableFormatter.textWidth(textEntry.text, column)
		width += textWidth
		column += textWidth
	}

	return
}

func (tableFormatter *TableFormatter) textWidth(text string, column uint) (width uint) {
	for _, codePoint := range text {
		renderedCodePoints := DetermineRenderedCodePoint(codePoint, column, tableFormatter.config)

		for _, renderedCodePoint := range renderedCodePoints {
			width += renderedCodePoint.width
			column += renderedCodePoint.width
		}
	}

	return
}

func (tableFormatter *TableFormatter) separatorWidth() uint {
	width := uint(len(tfSeparator))

	if tableFormatter.gridLines {
		width *= 2
		width++
	}

	return width
}

func (tableFormatter *TableFormatter) appendSeparator(lineBuilder *LineBuilder, colIndex uint) {
	lineBuilder.Append(tfSeparator)

	if tableFormatter.gridLines && colIndex != tableFormatter.cols()-1 {
		lineBuilder.AppendACSChar(AcsVline, CmpNone)
		lineBuilder.Append(tfSeparator)
	}
}
