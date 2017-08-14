package main

import (
	"bytes"
	"fmt"
	log "github.com/Sirupsen/logrus"
	rw "github.com/mattn/go-runewidth"
	gc "github.com/rthornton128/goncurses"
	"unicode"
)

type RenderWindow interface {
	Id() string
	Rows() uint
	Cols() uint
	ViewDimensions() ViewDimension
	Clear()
	SetRow(rowIndex, startColumn uint, themeComponentId ThemeComponentId, format string, args ...interface{}) error
	SetSelectedRow(rowIndex uint, active bool) error
	SetCursor(rowIndex, colIndex uint) error
	SetTitle(themeComponentId ThemeComponentId, format string, args ...interface{}) error
	SetFooter(themeComponentId ThemeComponentId, format string, args ...interface{}) error
	ApplyStyle(themeComponentId ThemeComponentId)
	Highlight(pattern string, themeComponentId ThemeComponentId) error
	DrawBorder()
	LineBuilder(rowIndex, startColumn uint) (*LineBuilder, error)
}

type RenderedCodePoint struct {
	width     uint
	codePoint rune
}

type Line struct {
	cells []*Cell
}

type LineBuilder struct {
	line        *Line
	cellIndex   uint
	column      uint
	startColumn uint
	config      Config
}

type CellStyle struct {
	componentId ThemeComponentId
	attr        gc.Char
	acs_char    gc.Char
}

type Cell struct {
	codePoints bytes.Buffer
	style      CellStyle
}

type Cursor struct {
	row uint
	col uint
}

type Window struct {
	id       string
	rows     uint
	cols     uint
	lines    []*Line
	startRow uint
	startCol uint
	border   bool
	config   Config
	cursor   *Cursor
}

func NewLine(cols uint) *Line {
	line := &Line{
		cells: make([]*Cell, cols),
	}

	for i := uint(0); i < cols; i++ {
		line.cells[i] = &Cell{}
	}

	return line
}

func (line *Line) String() string {
	var buf bytes.Buffer

	for _, cell := range line.cells {
		buf.Write(cell.codePoints.Bytes())
	}

	return buf.String()
}

func NewLineBuilder(line *Line, config Config, startColumn uint) *LineBuilder {
	return &LineBuilder{
		line:        line,
		column:      1,
		config:      config,
		startColumn: startColumn,
	}
}

func (lineBuilder *LineBuilder) Append(format string, args ...interface{}) *LineBuilder {
	return lineBuilder.AppendWithStyle(CMP_NONE, format, args...)
}

func (lineBuilder *LineBuilder) AppendWithStyle(componentId ThemeComponentId, format string, args ...interface{}) *LineBuilder {
	str := fmt.Sprintf(format, args...)
	line := lineBuilder.line

	for _, codePoint := range str {
		renderedCodePoints := DetermineRenderedCodePoint(codePoint, lineBuilder.column, lineBuilder.config)

		for _, renderedCodePoint := range renderedCodePoints {
			if lineBuilder.cellIndex > uint(len(line.cells)) {
				break
			}

			if renderedCodePoint.width > 1 {
				lineBuilder.setCellAndAdvanceIndex(renderedCodePoint.codePoint, renderedCodePoint.width, componentId)
				lineBuilder.Clear(renderedCodePoint.width - 1)
			} else if renderedCodePoint.width > 0 {
				lineBuilder.setCellAndAdvanceIndex(renderedCodePoint.codePoint, renderedCodePoint.width, componentId)
			} else {
				lineBuilder.appendToPreviousCell(renderedCodePoint.codePoint)
			}
		}
	}

	return lineBuilder
}

func (lineBuilder *LineBuilder) setCellAndAdvanceIndex(codePoint rune, width uint, componentId ThemeComponentId) {
	line := lineBuilder.line

	if lineBuilder.cellIndex < uint(len(line.cells)) {
		if lineBuilder.column >= lineBuilder.startColumn {
			cell := line.cells[lineBuilder.cellIndex]
			cell.codePoints.Reset()
			cell.codePoints.WriteRune(codePoint)
			cell.style.componentId = componentId
			cell.style.acs_char = 0
			lineBuilder.cellIndex++
		}

		lineBuilder.column += width
	}
}

func (lineBuilder *LineBuilder) Clear(cellNum uint) {
	line := lineBuilder.line

	for i := uint(0); i < cellNum && lineBuilder.cellIndex < uint(len(line.cells)); i++ {
		line.cells[lineBuilder.cellIndex].codePoints.Reset()
		lineBuilder.cellIndex++
	}
}

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

func NewWindow(id string, config Config) *Window {
	return &Window{
		id:     id,
		config: config,
	}
}

func (win *Window) Resize(viewDimension ViewDimension) {
	if win.rows == viewDimension.rows && win.cols == viewDimension.cols {
		return
	}

	log.Debugf("Resizing window %v from rows:%v,cols:%v to %v", win.id, win.rows, win.cols, viewDimension)

	win.rows = viewDimension.rows
	win.cols = viewDimension.cols

	win.lines = make([]*Line, win.rows)

	for i := uint(0); i < win.rows; i++ {
		win.lines[i] = NewLine(win.cols)
	}
}

func (win *Window) SetPosition(startRow, startCol uint) {
	win.startRow = startRow
	win.startCol = startCol
}

func (win *Window) OffsetPosition(rowOffset, colOffset int) {
	win.startRow = applyOffset(win.startRow, rowOffset)
	win.startCol = applyOffset(win.startCol, colOffset)
}

func applyOffset(value uint, offset int) uint {
	if value < 0 {
		return value - Min(value, Abs(offset))
	}

	return value + uint(offset)
}

func (win *Window) Id() string {
	return win.id
}

func (win *Window) Rows() uint {
	return win.rows
}

func (win *Window) Cols() uint {
	return win.cols
}

func (win *Window) ViewDimensions() ViewDimension {
	return ViewDimension{
		rows: win.rows,
		cols: win.cols,
	}
}

func (win *Window) Clear() {
	log.Debugf("Clearing window %v", win.id)

	for _, line := range win.lines {
		for _, cell := range line.cells {
			cell.codePoints.Reset()
			cell.codePoints.WriteRune(' ')
			cell.style.componentId = CMP_NONE
			cell.style.attr = gc.A_NORMAL
			cell.style.acs_char = 0
		}
	}

	win.cursor = nil
	win.border = false
}

func (win *Window) LineBuilder(rowIndex, startColumn uint) (*LineBuilder, error) {
	if rowIndex >= win.rows {
		return nil, fmt.Errorf("LineBuilder: Invalid row index: %v >= %v rows", rowIndex, win.rows)
	} else if startColumn == 0 {
		return nil, fmt.Errorf("Column must be postive")
	}

	return NewLineBuilder(win.lines[rowIndex], win.config, startColumn), nil
}

func (win *Window) SetRow(rowIndex, startColumn uint, themeComponentId ThemeComponentId, format string, args ...interface{}) error {
	lineBuilder, err := win.LineBuilder(rowIndex, startColumn)
	if err != nil {
		return err
	}

	lineBuilder.AppendWithStyle(themeComponentId, format, args...)

	return nil
}

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
		cell.style.componentId = CMP_NONE
	}

	return nil
}

func (win *Window) IsCursorSet() bool {
	return win.cursor != nil
}

func (win *Window) SetCursor(rowIndex, colIndex uint) (err error) {
	if rowIndex >= win.rows {
		return fmt.Errorf("SetCursor: Invalid row index: %v >= %v rows", rowIndex, win.rows)
	} else if colIndex >= win.cols {
		return fmt.Errorf("Invalid col index: %v >= %v cols", colIndex, win.cols)
	}

	win.cursor = &Cursor{
		row: rowIndex,
		col: colIndex,
	}

	return
}

func (win *Window) SetTitle(componentId ThemeComponentId, format string, args ...interface{}) (err error) {
	return win.setHeader(0, false, componentId, format, args...)
}

func (win *Window) SetFooter(componentId ThemeComponentId, format string, args ...interface{}) (err error) {
	if win.rows < 1 {
		log.Errorf("Can't set footer on window %v with %v rows", win.id, win.rows)
		return
	}

	return win.setHeader(win.rows-1, true, componentId, format, args...)
}

func (win *Window) setHeader(rowIndex uint, rightJustified bool, componentId ThemeComponentId, format string, args ...interface{}) (err error) {
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

	lineBuilder.AppendWithStyle(componentId, format, args...)

	return
}

func (win *Window) DrawBorder() {
	if win.rows < 3 || win.cols < 3 {
		return
	}

	firstLine := win.lines[0]
	firstLine.cells[0].style.acs_char = gc.ACS_ULCORNER

	for i := uint(1); i < win.cols-1; i++ {
		firstLine.cells[i].style.acs_char = gc.ACS_HLINE
	}

	firstLine.cells[win.cols-1].style.acs_char = gc.ACS_URCORNER

	for i := uint(1); i < win.rows-1; i++ {
		line := win.lines[i]
		line.cells[0].style.acs_char = gc.ACS_VLINE
		line.cells[win.cols-1].style.acs_char = gc.ACS_VLINE
	}

	lastLine := win.lines[win.rows-1]
	lastLine.cells[0].style.acs_char = gc.ACS_LLCORNER

	for i := uint(1); i < win.cols-1; i++ {
		lastLine.cells[i].style.acs_char = gc.ACS_HLINE
	}

	lastLine.cells[win.cols-1].style.acs_char = gc.ACS_LRCORNER

	win.border = true
}

func (win *Window) ApplyStyle(themeComponentId ThemeComponentId) {
	for _, line := range win.lines {
		for _, cell := range line.cells {
			cell.style.componentId = themeComponentId
		}
	}
}

func DetermineRenderedCodePoint(codePoint rune, column uint, config Config) (renderedCodePoints []RenderedCodePoint) {
	if !unicode.IsPrint(codePoint) {
		if codePoint == '\t' {
			tabWidth := uint(config.GetInt(CV_TAB_WIDTH))
			width := tabWidth - ((column - 1) % tabWidth)

			for i := uint(0); i < width; i++ {
				renderedCodePoints = append(renderedCodePoints, RenderedCodePoint{
					width:     1,
					codePoint: ' ',
				})
			}
		} else if codePoint != '\n' && (codePoint < 32 || codePoint == 127) {
			for _, char := range nonPrintableCharString(codePoint) {
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

func (win *Window) Line(lineIndex uint) (line string, lineExists bool) {
	if lineIndex < win.rows {
		if win.border && lineIndex == 0 || lineIndex+1 == win.rows {
			lineExists = true
			return
		}

		line = win.lines[lineIndex].String()

		if win.border && len(line) > 0 {
			line = line[1:]
		}

		lineExists = true
	}

	return
}

func (win *Window) LineNumber() (lineNumber uint) {
	return win.rows
}

func (win *Window) Highlight(pattern string, themeComponentId ThemeComponentId) (err error) {
	search, err := NewSearch(SD_FORWARD, pattern, win)
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
				cell.style.componentId = themeComponentId
			}

			bytes += uint(cell.codePoints.Len())
			cellIndex++
		}
	}

	return
}
