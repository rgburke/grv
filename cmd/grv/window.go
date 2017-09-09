package main

import (
	"bytes"
	"fmt"
	"unicode"

	log "github.com/Sirupsen/logrus"
	rw "github.com/mattn/go-runewidth"
	gc "github.com/rthornton128/goncurses"
)

// RenderWindow represents a window that will be drawn to the display
type RenderWindow interface {
	ID() string
	Rows() uint
	Cols() uint
	ViewDimensions() ViewDimension
	Clear()
	SetRow(rowIndex, startColumn uint, themeComponentID ThemeComponentID, format string, args ...interface{}) error
	SetSelectedRow(rowIndex uint, active bool) error
	SetCursor(rowIndex, colIndex uint) error
	SetTitle(themeComponentID ThemeComponentID, format string, args ...interface{}) error
	SetFooter(themeComponentID ThemeComponentID, format string, args ...interface{}) error
	ApplyStyle(themeComponentID ThemeComponentID)
	Highlight(pattern string, themeComponentID ThemeComponentID) error
	DrawBorder()
	LineBuilder(rowIndex, startColumn uint) (*LineBuilder, error)
}

// RenderedCodePoint contains the display values for a codepoint
type RenderedCodePoint struct {
	width     uint
	codePoint rune
}

type line struct {
	cells []*cell
}

// LineBuilder provides a way of drawing a single line to a window
type LineBuilder struct {
	line        *line
	cellIndex   uint
	column      uint
	startColumn uint
	config      Config
}

type cellStyle struct {
	themeComponentID ThemeComponentID
	attr             gc.Char
	acsChar          gc.Char
}

type cell struct {
	codePoints bytes.Buffer
	style      cellStyle
}

type cursor struct {
	row uint
	col uint
}

// Window implements the RenderWindow interface and contains all rendered data
type Window struct {
	id       string
	rows     uint
	cols     uint
	lines    []*line
	startRow uint
	startCol uint
	border   bool
	config   Config
	cursor   *cursor
}

func newLine(cols uint) *line {
	line := &line{
		cells: make([]*cell, cols),
	}

	for i := uint(0); i < cols; i++ {
		line.cells[i] = &cell{}
	}

	return line
}

// String returns the text contained in the line
func (line *line) String() string {
	var buf bytes.Buffer

	for _, cell := range line.cells {
		buf.Write(cell.codePoints.Bytes())
	}

	return buf.String()
}

func newLineBuilder(line *line, config Config, startColumn uint) *LineBuilder {
	return &LineBuilder{
		line:        line,
		column:      1,
		config:      config,
		startColumn: startColumn,
	}
}

// Append adds the provided text to the end of the line
func (lineBuilder *LineBuilder) Append(format string, args ...interface{}) *LineBuilder {
	return lineBuilder.AppendWithStyle(CmpNone, format, args...)
}

// AppendWithStyle adds the provided text with style information to the end of the line
func (lineBuilder *LineBuilder) AppendWithStyle(themeComponentID ThemeComponentID, format string, args ...interface{}) *LineBuilder {
	str := fmt.Sprintf(format, args...)
	line := lineBuilder.line

	for _, codePoint := range str {
		renderedCodePoints := DetermineRenderedCodePoint(codePoint, lineBuilder.column, lineBuilder.config)

		for _, renderedCodePoint := range renderedCodePoints {
			if lineBuilder.cellIndex > uint(len(line.cells)) {
				break
			}

			if renderedCodePoint.width > 1 {
				lineBuilder.setCellAndAdvanceIndex(renderedCodePoint.codePoint, renderedCodePoint.width, themeComponentID)
				lineBuilder.Clear(renderedCodePoint.width - 1)
			} else if renderedCodePoint.width > 0 {
				lineBuilder.setCellAndAdvanceIndex(renderedCodePoint.codePoint, renderedCodePoint.width, themeComponentID)
			} else {
				lineBuilder.appendToPreviousCell(renderedCodePoint.codePoint)
			}
		}
	}

	return lineBuilder
}

func (lineBuilder *LineBuilder) setCellAndAdvanceIndex(codePoint rune, width uint, themeComponentID ThemeComponentID) {
	line := lineBuilder.line

	if lineBuilder.cellIndex < uint(len(line.cells)) {
		if lineBuilder.column >= lineBuilder.startColumn {
			cell := line.cells[lineBuilder.cellIndex]
			cell.codePoints.Reset()
			cell.codePoints.WriteRune(codePoint)
			cell.style.themeComponentID = themeComponentID
			cell.style.acsChar = 0
			lineBuilder.cellIndex++
		}

		lineBuilder.column += width
	}
}

// Clear resets the next cellNum cells in the line
func (lineBuilder *LineBuilder) Clear(cellNum uint) {
	line := lineBuilder.line

	for i := uint(0); i < cellNum && lineBuilder.cellIndex < uint(len(line.cells)); i++ {
		line.cells[lineBuilder.cellIndex].codePoints.Reset()
		lineBuilder.cellIndex++
	}
}

// ToLineStart moves the draw position to the start of the line
func (lineBuilder *LineBuilder) ToLineStart() {
	lineBuilder.cellIndex = 0
	lineBuilder.startColumn = 1
}

func (lineBuilder *LineBuilder) appendToPreviousCell(codePoint rune) {
	if lineBuilder.cellIndex > 0 {
		cell := lineBuilder.line.cells[lineBuilder.cellIndex-1]
		cell.codePoints.WriteRune(codePoint)
	}
}

// NewWindow creates a new instance
func NewWindow(id string, config Config) *Window {
	return &Window{
		id:     id,
		config: config,
	}
}

// Resize updates the windows internal storage capacity
func (win *Window) Resize(viewDimension ViewDimension) {
	if win.rows == viewDimension.rows && win.cols == viewDimension.cols {
		return
	}

	log.Debugf("Resizing window %v from rows:%v,cols:%v to %v", win.id, win.rows, win.cols, viewDimension)

	win.rows = viewDimension.rows
	win.cols = viewDimension.cols

	win.lines = make([]*line, win.rows)

	for i := uint(0); i < win.rows; i++ {
		win.lines[i] = newLine(win.cols)
	}
}

// SetPosition sets the coordintates the window should appear on the display
func (win *Window) SetPosition(startRow, startCol uint) {
	win.startRow = startRow
	win.startCol = startCol
}

// OffsetPosition applies the provided offsets to the windows position
func (win *Window) OffsetPosition(rowOffset, colOffset int) {
	win.startRow = applyOffset(win.startRow, rowOffset)
	win.startCol = applyOffset(win.startCol, colOffset)
}

func applyOffset(value uint, offset int) uint {
	if offset < 0 {
		return value - Min(value, Abs(offset))
	}

	return value + uint(offset)
}

// ID returns the window ID
func (win *Window) ID() string {
	return win.id
}

// Rows returns the number of rows in this window
func (win *Window) Rows() uint {
	return win.rows
}

// Cols returns the number of cols in this window
func (win *Window) Cols() uint {
	return win.cols
}

// ViewDimensions returns the dimensions of the window
func (win *Window) ViewDimensions() ViewDimension {
	return ViewDimension{
		rows: win.rows,
		cols: win.cols,
	}
}

// Clear resets all cells in the window
func (win *Window) Clear() {
	log.Debugf("Clearing window %v", win.id)

	for _, line := range win.lines {
		for _, cell := range line.cells {
			cell.codePoints.Reset()
			cell.codePoints.WriteRune(' ')
			cell.style.themeComponentID = CmpNone
			cell.style.attr = gc.A_NORMAL
			cell.style.acsChar = 0
		}
	}

	win.cursor = nil
	win.border = false
}

// LineBuilder returns a line builder instance for the provided line index
func (win *Window) LineBuilder(rowIndex, startColumn uint) (*LineBuilder, error) {
	if rowIndex >= win.rows {
		return nil, fmt.Errorf("LineBuilder: Invalid row index: %v >= %v rows", rowIndex, win.rows)
	} else if startColumn == 0 {
		return nil, fmt.Errorf("Column must be positive")
	}

	return newLineBuilder(win.lines[rowIndex], win.config, startColumn), nil
}

// SetRow sets the text and style information for a line
func (win *Window) SetRow(rowIndex, startColumn uint, themeComponentID ThemeComponentID, format string, args ...interface{}) error {
	lineBuilder, err := win.LineBuilder(rowIndex, startColumn)
	if err != nil {
		return err
	}

	lineBuilder.AppendWithStyle(themeComponentID, format, args...)

	return nil
}

// SetSelectedRow sets the row to be highlighted as the selected row
func (win *Window) SetSelectedRow(rowIndex uint, active bool) error {
	log.Debugf("Set selected rowIndex for window %v to %v with active %v", win.id, rowIndex, active)

	if rowIndex >= win.rows {
		return fmt.Errorf("SetSelectedRow: Invalid row index: %v >= %v rows", rowIndex, win.rows)
	}

	var attr gc.Char = gc.A_REVERSE

	if !active {
		attr |= gc.A_DIM
	}

	line := win.lines[rowIndex]

	for _, cell := range line.cells {
		cell.style.attr |= attr
		cell.style.themeComponentID = CmpNone
	}

	return nil
}

// IsCursorSet returns true if a cursor position has been set
func (win *Window) IsCursorSet() bool {
	return win.cursor != nil
}

// SetCursor sets a cursor position on the window
// If this is set then a cursor will be displayed in this window
func (win *Window) SetCursor(rowIndex, colIndex uint) (err error) {
	if rowIndex >= win.rows {
		return fmt.Errorf("SetCursor: Invalid row index: %v >= %v rows", rowIndex, win.rows)
	} else if colIndex >= win.cols {
		return fmt.Errorf("Invalid col index: %v >= %v cols", colIndex, win.cols)
	}

	win.cursor = &cursor{
		row: rowIndex,
		col: colIndex,
	}

	return
}

// SetTitle sets the title to display for the window
func (win *Window) SetTitle(themeComponentID ThemeComponentID, format string, args ...interface{}) (err error) {
	return win.setHeader(0, false, themeComponentID, format, args...)
}

// SetFooter sets the footer to display for thw window
func (win *Window) SetFooter(themeComponentID ThemeComponentID, format string, args ...interface{}) (err error) {
	if win.rows < 1 {
		log.Errorf("Can't set footer on window %v with %v rows", win.id, win.rows)
		return
	}

	return win.setHeader(win.rows-1, true, themeComponentID, format, args...)
}

func (win *Window) setHeader(rowIndex uint, rightJustified bool, themeComponentID ThemeComponentID, format string, args ...interface{}) (err error) {
	if win.rows < 3 || win.cols < 3 {
		log.Errorf("Can't set header on window %v with %v rows and %v cols", win.id, win.rows, win.cols)
		return
	}

	var lineBuilder *LineBuilder
	lineBuilder, err = win.LineBuilder(rowIndex, 1)

	if err != nil {
		return
	}

	format = " " + format + " "

	if rightJustified {
		// Assume only ascii alphanumeric characters and space character
		// present in footer text
		formattedLen := uint(len([]rune(fmt.Sprintf(format, args...))))
		if formattedLen > win.cols+2 {
			return
		}

		lineBuilder.cellIndex = win.cols - (2 + formattedLen)
	} else {
		lineBuilder.cellIndex = 2
	}

	lineBuilder.column = lineBuilder.cellIndex + 1

	lineBuilder.AppendWithStyle(themeComponentID, format, args...)

	return
}

// DrawBorder draws a line of a single cells width around the edge of the window
func (win *Window) DrawBorder() {
	if win.rows < 3 || win.cols < 3 {
		return
	}

	firstLine := win.lines[0]
	firstLine.cells[0].style.acsChar = gc.ACS_ULCORNER

	for i := uint(1); i < win.cols-1; i++ {
		firstLine.cells[i].style.acsChar = gc.ACS_HLINE
	}

	firstLine.cells[win.cols-1].style.acsChar = gc.ACS_URCORNER

	for i := uint(1); i < win.rows-1; i++ {
		line := win.lines[i]
		line.cells[0].style.acsChar = gc.ACS_VLINE
		line.cells[win.cols-1].style.acsChar = gc.ACS_VLINE
	}

	lastLine := win.lines[win.rows-1]
	lastLine.cells[0].style.acsChar = gc.ACS_LLCORNER

	for i := uint(1); i < win.cols-1; i++ {
		lastLine.cells[i].style.acsChar = gc.ACS_HLINE
	}

	lastLine.cells[win.cols-1].style.acsChar = gc.ACS_LRCORNER

	win.border = true
}

// ApplyStyle sets a single style for all cells in the window
func (win *Window) ApplyStyle(themeComponentID ThemeComponentID) {
	for _, line := range win.lines {
		for _, cell := range line.cells {
			cell.style.themeComponentID = themeComponentID
		}
	}
}

// DetermineRenderedCodePoint converts a code point into its rendered representation
func DetermineRenderedCodePoint(codePoint rune, column uint, config Config) (renderedCodePoints []RenderedCodePoint) {
	if !unicode.IsPrint(codePoint) {
		if codePoint == '\t' {
			tabWidth := uint(config.GetInt(CfTabWidth))
			width := tabWidth - ((column - 1) % tabWidth)

			for i := uint(0); i < width; i++ {
				renderedCodePoints = append(renderedCodePoints, RenderedCodePoint{
					width:     1,
					codePoint: ' ',
				})
			}
		} else if codePoint != '\n' && (codePoint < 32 || codePoint == 127) {
			for _, char := range NonPrintableCharString(codePoint) {
				renderedCodePoints = append(renderedCodePoints, RenderedCodePoint{
					width:     1,
					codePoint: char,
				})
			}
		} else {
			renderedCodePoints = append(renderedCodePoints, RenderedCodePoint{
				width:     1,
				codePoint: codePoint,
			})
		}
	} else {
		renderedCodePoints = append(renderedCodePoints, RenderedCodePoint{
			width:     uint(rw.RuneWidth(codePoint)),
			codePoint: codePoint,
		})
	}

	return
}

// Line returns the text contained on the specified line index
func (win *Window) Line(lineIndex uint) (line string) {
	if lineIndex >= win.rows {
		log.Errorf("Invalid lineIndex: %v", lineIndex)
		return
	}

	if win.border && lineIndex == 0 || lineIndex+1 == win.rows {
		return
	}

	line = win.lines[lineIndex].String()

	if win.border && len(line) > 0 {
		line = line[1:]
	}

	return
}

// LineNumber returns the number of lines in the window
func (win *Window) LineNumber() (lineNumber uint) {
	return win.rows
}

// Highlight searches the window for all occurrences of the specified pattern.
// Each match then has the provided style applied to it
func (win *Window) Highlight(pattern string, themeComponentID ThemeComponentID) (err error) {
	search, err := NewSearch(SdForward, pattern, win)
	if err != nil {
		return
	}

	lineMatches := search.FindAll()

	for _, lineMatch := range lineMatches {
		line := win.lines[lineMatch.RowIndex]
		bytes := uint(0)
		index := 0
		lineMatchIndex := lineMatch.MatchIndexes[index]
		cellIndex := 0

		if win.border {
			cellIndex++
		}

		for cellIndex < len(line.cells) {
			cell := line.cells[cellIndex]

			if bytes >= lineMatchIndex.ByteEndIndex {
				if index++; index < len(lineMatch.MatchIndexes) {
					lineMatchIndex = lineMatch.MatchIndexes[index]
				} else {
					break
				}
			}

			if bytes >= lineMatchIndex.ByteStartIndex {
				attr := int(cell.style.attr)
				attr &= ^gc.A_REVERSE
				cell.style.attr = gc.Char(attr)
				cell.style.themeComponentID = themeComponentID
			}

			bytes += uint(cell.codePoints.Len())
			cellIndex++
		}
	}

	return
}
