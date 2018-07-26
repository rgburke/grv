package main

import (
	log "github.com/Sirupsen/logrus"
)

// SelectableRowChildWindowView extends ChildWindowView to handle
// views with rows that aren't selectable
type SelectableRowChildWindowView interface {
	ChildWindowView
	isSelectableRow(rowIndex uint) bool
}

// SelectableRowView extends AbstractWindowView and supports
// views with rows that aren't selectable
type SelectableRowView struct {
	*AbstractWindowView
	child *selectableRowDecorator
}

// NewSelectableRowView creates a new instance
func NewSelectableRowView(child SelectableRowChildWindowView, channels Channels, config Config, variables GRVVariableSetter, lock Lock, rowDescriptor string) *SelectableRowView {
	decoratedChild := newSelectableRowDecorator(child)
	return &SelectableRowView{
		AbstractWindowView: NewAbstractWindowView(decoratedChild, channels, config, variables, lock, rowDescriptor),
		child:              decoratedChild,
	}
}

// HandleAction proxies the call down to the underlying AbstractWindowView.
// If the ActiveRowIndex is not on a non-selectable row then the nearest selectable
// row is selected and the child view is notified
func (selectableRowView *SelectableRowView) HandleAction(action Action) (handled bool, err error) {
	activeRowIndexStart := selectableRowView.child.viewPos().ActiveRowIndex()

	handled, err = selectableRowView.AbstractWindowView.HandleAction(action)
	if err != nil || !handled {
		return
	}

	activeRowIndexEnd := selectableRowView.child.viewPos().ActiveRowIndex()

	if activeRowIndexStart == activeRowIndexEnd {
		log.Debugf("activeRowIndexStart (%v) == activeRowIndexEnd (%v)", activeRowIndexStart, activeRowIndexEnd)
		return
	}

	var selectedRowIndex uint
	if selectableRowView.child.isSelectableRow(activeRowIndexEnd) {
		selectedRowIndex = activeRowIndexEnd
	} else {
		selectedRowIndex = selectableRowView.findNearestSelectableRow(activeRowIndexEnd, activeRowIndexEnd > activeRowIndexStart)
		log.Debugf("Nearest selectable row index: %v", selectedRowIndex)
		selectableRowView.child.viewPos().SetActiveRowIndex(selectedRowIndex)
	}

	log.Debugf("Notifying child view of selected row index: %v", selectedRowIndex)
	err = selectableRowView.notifyChildRowSelected(selectedRowIndex)
	return
}

func (selectableRowView *SelectableRowView) findNearestSelectableRow(startRowIndex uint, searchDownwardsFirst bool) uint {
	if searchDownwardsFirst {
		if rowIndex, found := selectableRowView.searchDownwards(startRowIndex); found {
			return rowIndex
		} else if rowIndex, found := selectableRowView.searchUpwards(startRowIndex); found {
			return rowIndex
		}
	} else {
		if rowIndex, found := selectableRowView.searchUpwards(startRowIndex); found {
			return rowIndex
		} else if rowIndex, found := selectableRowView.searchDownwards(startRowIndex); found {
			return rowIndex
		}
	}

	return 0
}

func (selectableRowView *SelectableRowView) searchDownwards(startRowIndex uint) (rowIndex uint, found bool) {
	rows := selectableRowView.child.rows()
	for rowIndex = startRowIndex + 1; rowIndex < rows; rowIndex++ {
		if selectableRowView.child.isSelectableRow(rowIndex) {
			found = true
			break
		}
	}

	return
}

func (selectableRowView *SelectableRowView) searchUpwards(startRowIndex uint) (rowIndex uint, found bool) {
	if startRowIndex == 0 {
		return
	}

	for rowIndex = startRowIndex - 1; ; rowIndex-- {
		if selectableRowView.child.isSelectableRow(rowIndex) {
			found = true
			break
		} else if rowIndex == 0 {
			break
		}
	}

	return
}

func (selectableRowView *SelectableRowView) notifyChildRowSelected(rowIndex uint) (err error) {
	err = selectableRowView.child.notifyChildRowSelected(rowIndex)
	selectableRowView.setVariables()

	return
}

// SelectNearestSelectableRow selects the nearest selectable row
// if the current row is not selectable
func (selectableRowView *SelectableRowView) SelectNearestSelectableRow() (err error) {
	if selectableRowView.child.rows() == 0 {
		return
	}

	currentRowIndex := selectableRowView.child.viewPos().ActiveRowIndex()
	if selectableRowView.child.isSelectableRow(currentRowIndex) {
		return
	}

	nearestSelectableRow := selectableRowView.findNearestSelectableRow(currentRowIndex, true)
	selectableRowView.child.viewPos().SetActiveRowIndex(nearestSelectableRow)
	return selectableRowView.notifyChildRowSelected(nearestSelectableRow)
}

type selectableRowDecorator struct {
	child SelectableRowChildWindowView
}

func newSelectableRowDecorator(child SelectableRowChildWindowView) *selectableRowDecorator {
	return &selectableRowDecorator{
		child: child,
	}
}

func (selectableRowDecorator *selectableRowDecorator) viewPos() ViewPos {
	return selectableRowDecorator.child.viewPos()
}

func (selectableRowDecorator *selectableRowDecorator) rows() uint {
	return selectableRowDecorator.child.rows()
}

func (selectableRowDecorator *selectableRowDecorator) line(lineIndex uint) string {
	return selectableRowDecorator.child.line(lineIndex)
}

func (selectableRowDecorator *selectableRowDecorator) viewDimension() ViewDimension {
	return selectableRowDecorator.child.viewDimension()
}

func (selectableRowDecorator *selectableRowDecorator) isSelectableRow(rowIndex uint) bool {
	return selectableRowDecorator.child.isSelectableRow(rowIndex)
}

func (selectableRowDecorator *selectableRowDecorator) onRowSelected(rowIndex uint) (err error) {
	// At this point the current row may not be selectable.
	// Therefore intercept this call and do nothing.
	// The child will be notified using notifyChildRowSelected when
	// any changes to the ActiveRowIndex have been made to ensure the active row is selectable.
	return
}

func (selectableRowDecorator *selectableRowDecorator) notifyChildRowSelected(rowIndex uint) error {
	return selectableRowDecorator.child.onRowSelected(rowIndex)
}
