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
	WN_TAB_WIDTH        = 8
	WN_WINDOW_DUMP_FILE = "grv-window.log"
)

type RenderWindow interface {
	Id() string
	Rows() uint
	Cols() uint
	Clear()
	SetRow(rowIndex uint, format string, args ...interface{}) error
	SetSelectedRow(rowIndex uint, active bool) error
	DrawBorder()
}

type Line struct {
	cells []*Cell
}

type LineBuilder struct {
	line      *Line
	cellIndex uint
	column    uint
}

type CellStyle struct {
	attr     gc.Char
	acs_char gc.Char
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

func NewLineBuilder(line *Line) *LineBuilder {
	return &LineBuilder{
		line:   line,
		column: 1,
	}
}

func (lineBuilder *LineBuilder) Append(format string, args ...interface{}) {
	str := fmt.Sprintf(format, args...)
	line := lineBuilder.line

	for _, codePoint := range str {
		if lineBuilder.cellIndex > uint(len(line.cells)) {
			break
		} else if !unicode.IsPrint(codePoint) {
			if codePoint == '\t' {
				width := WN_TAB_WIDTH - ((lineBuilder.column - 1) % WN_TAB_WIDTH)

				for i := uint(0); i < width; i++ {
					lineBuilder.SetCellAndAdvanceIndex(' ', 1)
				}
			} else if codePoint != '\n' && (codePoint < 32 || codePoint == 127) {
				lineBuilder.SetCellAndAdvanceIndex('^', 1)

				if codePoint == 127 {
					lineBuilder.SetCellAndAdvanceIndex('?', 1)
				} else {
					lineBuilder.SetCellAndAdvanceIndex(codePoint+64, 1)
				}
			} else {
				lineBuilder.SetCellAndAdvanceIndex(codePoint, 1)
			}
		} else if width := uint(rw.RuneWidth(codePoint)); width == 0 {
			lineBuilder.AppendToPreviousCell(codePoint)
		} else if width > 1 {
			lineBuilder.SetCellAndAdvanceIndex(codePoint, width)
			lineBuilder.Clear(width - 1)
		} else {
			lineBuilder.SetCellAndAdvanceIndex(codePoint, width)
		}
	}
}

func (lineBuilder *LineBuilder) SetCellAndAdvanceIndex(codePoint rune, width uint) {
	line := lineBuilder.line

	if lineBuilder.cellIndex < uint(len(line.cells)) {
		cell := line.cells[lineBuilder.cellIndex]
		cell.codePoints.Reset()
		cell.codePoints.WriteRune(codePoint)
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

func (lineBuilder *LineBuilder) AppendToPreviousCell(codePoint rune) {
	if lineBuilder.cellIndex > 0 {
		cell := lineBuilder.line.cells[lineBuilder.cellIndex-1]
		cell.codePoints.WriteRune(codePoint)
	}
}

func NewWindow(id string) *Window {
	return &Window{
		id: id,
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
			cell.style.attr = gc.A_NORMAL
			cell.style.acs_char = 0
		}
	}
}

func (win *Window) LineBuilder(rowIndex uint) (*LineBuilder, error) {
	if rowIndex >= win.rows {
		return nil, fmt.Errorf("Invalid row index: %v >= %v rows", rowIndex, win.rows)
	}

	return NewLineBuilder(win.lines[rowIndex]), nil
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
	}

	return nil
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
