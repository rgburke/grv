package main

import (
	"bytes"
	"fmt"
	log "github.com/Sirupsen/logrus"
	rw "github.com/mattn/go-runewidth"
	gc "github.com/rthornton128/goncurses"
	"os"
	"time"
	"unicode"
)

const (
	WN_WINDOW_DUMP_FILE = "grv-window.log"
)

type RenderWindow interface {
	Id() string
	Rows() uint
	Cols() uint
	Clear()
	SetRow(rowIndex uint, format string, args ...interface{}) error
	SetSelectedRow(rowIndex uint, active bool) error
	SetTitle(themeComponentId ThemeComponentId, format string, args ...interface{}) error
	SetFooter(themeComponentId ThemeComponentId, format string, args ...interface{}) error
	DrawBorder()
	LineBuilder(rowIndex uint) (*LineBuilder, error)
}

type Line struct {
	cells []*Cell
}

type LineBuilder struct {
	line      *Line
	cellIndex uint
	column    uint
	config    Config
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

type Window struct {
	id       string
	rows     uint
	cols     uint
	lines    []*Line
	startRow uint
	startCol uint
	config   Config
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

func NewLineBuilder(line *Line, config Config) *LineBuilder {
	return &LineBuilder{
		line:   line,
		column: 1,
		config: config,
	}
}

func (lineBuilder *LineBuilder) Append(format string, args ...interface{}) *LineBuilder {
	return lineBuilder.AppendWithStyle(CMP_NONE, format, args...)
}

func (lineBuilder *LineBuilder) AppendWithStyle(componentId ThemeComponentId, format string, args ...interface{}) *LineBuilder {
	str := fmt.Sprintf(format, args...)
	line := lineBuilder.line
	tabWidth := uint(lineBuilder.config.GetInt(CV_TAB_WIDTH))

	for _, codePoint := range str {
		if lineBuilder.cellIndex > uint(len(line.cells)) {
			break
		} else if !unicode.IsPrint(codePoint) {
			if codePoint == '\t' {
				width := tabWidth - ((lineBuilder.column - 1) % tabWidth)

				for i := uint(0); i < width; i++ {
					lineBuilder.setCellAndAdvanceIndex(' ', 1, componentId)
				}
			} else if codePoint != '\n' && (codePoint < 32 || codePoint == 127) {
				lineBuilder.setCellAndAdvanceIndex('^', 1, componentId)

				if codePoint == 127 {
					lineBuilder.setCellAndAdvanceIndex('?', 1, componentId)
				} else {
					lineBuilder.setCellAndAdvanceIndex(codePoint+64, 1, componentId)
				}
			} else {
				lineBuilder.setCellAndAdvanceIndex(codePoint, 1, componentId)
			}
		} else if width := uint(rw.RuneWidth(codePoint)); width == 0 {
			lineBuilder.appendToPreviousCell(codePoint)
		} else if width > 1 {
			lineBuilder.setCellAndAdvanceIndex(codePoint, width, componentId)
			lineBuilder.Clear(width - 1)
		} else {
			lineBuilder.setCellAndAdvanceIndex(codePoint, width, componentId)
		}
	}

	return lineBuilder
}

func (lineBuilder *LineBuilder) setCellAndAdvanceIndex(codePoint rune, width uint, componentId ThemeComponentId) {
	line := lineBuilder.line

	if lineBuilder.cellIndex < uint(len(line.cells)) {
		cell := line.cells[lineBuilder.cellIndex]
		cell.codePoints.Reset()
		cell.codePoints.WriteRune(codePoint)
		cell.style.componentId = componentId
		cell.style.acs_char = 0
		lineBuilder.cellIndex++
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

func (win *Window) Id() string {
	return win.id
}

func (win *Window) Rows() uint {
	return win.rows
}

func (win *Window) Cols() uint {
	return win.cols
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
}

func (win *Window) LineBuilder(rowIndex uint) (*LineBuilder, error) {
	if rowIndex >= win.rows {
		return nil, fmt.Errorf("Invalid row index: %v >= %v rows", rowIndex, win.rows)
	}

	return NewLineBuilder(win.lines[rowIndex], win.config), nil
}

func (win *Window) SetRow(rowIndex uint, format string, args ...interface{}) error {
	lineBuilder, err := win.LineBuilder(rowIndex)
	if err != nil {
		return err
	}

	lineBuilder.Append(format, args...)

	return nil
}

func (win *Window) SetSelectedRow(rowIndex uint, active bool) error {
	log.Debugf("Set selected rowIndex for window %v to %v with active %v", win.id, rowIndex, active)

	if rowIndex >= win.rows {
		return fmt.Errorf("Invalid row index: %v >= %v rows", rowIndex, win.rows)
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
	lineBuilder, err = win.LineBuilder(rowIndex)

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
}

// For debugging
func (win *Window) DumpContent() error {
	borderMap := map[gc.Char]rune{
		gc.ACS_HLINE:    0x2500,
		gc.ACS_VLINE:    0x2502,
		gc.ACS_ULCORNER: 0x250C,
		gc.ACS_URCORNER: 0x2510,
		gc.ACS_LLCORNER: 0x2514,
		gc.ACS_LRCORNER: 0x2518,
	}
	var buffer bytes.Buffer

	buffer.WriteString(fmt.Sprintf("%v Dumping window %v\n", time.Now().Format("2006/01/02 15:04:05.000"), win.id))

	for _, line := range win.lines {
		for _, cell := range line.cells {
			if cell.style.acs_char != 0 {
				buffer.WriteRune(borderMap[cell.style.acs_char])
			} else if cell.codePoints.Len() > 0 {
				buffer.Write(cell.codePoints.Bytes())
			}
		}

		buffer.WriteString("\n")
	}

	buffer.WriteString("\n")

	file, err := os.OpenFile(WN_WINDOW_DUMP_FILE, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	defer file.Close()

	if err != nil {
		return err
	}

	buffer.WriteTo(file)

	return nil
}
