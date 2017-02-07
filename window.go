package main

import (
	"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
	gc "github.com/rthornton128/goncurses"
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

type CellStyle struct {
	attr     gc.Char
	acs_char gc.Char
}

type Cell struct {
	codePoint rune
	style     CellStyle
}

type Window struct {
	id       string
	rows     uint
	cols     uint
	cells    [][]Cell
	startRow uint
	startCol uint
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

	win.cells = make([][]Cell, win.rows)

	for i := uint(0); i < win.rows; i++ {
		win.cells[i] = make([]Cell, win.cols)
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

	for i := uint(0); i < win.rows; i++ {
		for j := uint(0); j < win.cols; j++ {
			win.cells[i][j].codePoint = ' '
			win.cells[i][j].style.attr = gc.A_NORMAL
			win.cells[i][j].style.acs_char = 0
		}
	}
}

func (win *Window) SetRow(rowIndex uint, format string, args ...interface{}) error {
	if rowIndex >= win.rows {
		return errors.New(fmt.Sprintf("Invalid row index: %v >= %v rows", rowIndex, win.rows))
	}

	str := fmt.Sprintf(format, args...)

	colIndex := uint(0)
	rowCells := win.cells[rowIndex]

	for _, codePoint := range str {
		rowCells[colIndex].codePoint = codePoint
		colIndex++

		if colIndex >= win.cols {
			break
		}
	}

	for colIndex < win.cols {
		rowCells[colIndex].codePoint = ' '
		colIndex++
	}

	return nil
}

func (win *Window) SetSelectedRow(rowIndex uint, active bool) error {
	log.Debugf("Set selected rowIndex for window %v to %v with active %v", win.id, rowIndex, active)

	if rowIndex >= win.rows {
		return errors.New(fmt.Sprintf("Invalid row index: %v >= %v rows", rowIndex, win.rows))
	}

	var attr gc.Char = gc.A_REVERSE

	if !active {
		attr |= gc.A_DIM
	}

	rowCells := win.cells[rowIndex]

	for colIndex, _ := range rowCells {
		rowCells[colIndex].style.attr |= attr
	}

	return nil
}

func (win *Window) DrawBorder() {
	if win.rows < 3 || win.cols < 3 {
		return
	}

	firstRow := win.cells[0]
	firstRow[0].style.acs_char = gc.ACS_ULCORNER

	for i := uint(1); i < win.cols-1; i++ {
		firstRow[i].style.acs_char = gc.ACS_HLINE
	}

	firstRow[win.cols-1].style.acs_char = gc.ACS_URCORNER

	for i := uint(1); i < win.rows-1; i++ {
		row := win.cells[i]
		row[0].style.acs_char = gc.ACS_VLINE
		row[win.cols-1].style.acs_char = gc.ACS_VLINE
	}

	lastRow := win.cells[win.rows-1]
	lastRow[0].style.acs_char = gc.ACS_LLCORNER

	for i := uint(1); i < win.cols-1; i++ {
		lastRow[i].style.acs_char = gc.ACS_HLINE
	}

	lastRow[win.cols-1].style.acs_char = gc.ACS_LRCORNER
}
