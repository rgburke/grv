package main

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

type MockSelectableRowChildWindowView struct {
	MockChildWindowView
}

func (selectableRowChildWindowView *MockSelectableRowChildWindowView) isSelectableRow(rowIndex uint) bool {
	args := selectableRowChildWindowView.Called(rowIndex)
	return args.Bool(0)
}

func setupSelectableRowDecorator() (*selectableRowDecorator, *MockSelectableRowChildWindowView) {
	child := &MockSelectableRowChildWindowView{}
	return newSelectableRowDecorator(child), child
}

func TestSelectableRowDecoratorProxiesCallToViewPos(t *testing.T) {
	selectableRowDecorator, decorated := setupSelectableRowDecorator()
	viewPos := NewViewPosition()
	decorated.On("viewPos").Return(viewPos)

	returnedViewPos := selectableRowDecorator.viewPos()

	decorated.AssertCalled(t, "viewPos")
	assert.Equal(t, viewPos, returnedViewPos, "Returned ViewPos should match injected value")
}

func TestSelectableRowDecoratorProxiesCallToRows(t *testing.T) {
	selectableRowDecorator, decorated := setupSelectableRowDecorator()
	decorated.On("rows").Return(uint(5))

	returnedRows := selectableRowDecorator.rows()

	decorated.AssertCalled(t, "rows")
	assert.Equal(t, uint(5), returnedRows, "Returned rows should be 5")
}

func TestSelectableRowDecoratorProxiesCallToViewDimension(t *testing.T) {
	selectableRowDecorator, decorated := setupSelectableRowDecorator()
	decorated.On("viewDimension").Return(ViewDimension{rows: 24, cols: 80})

	returnedViewDimension := selectableRowDecorator.viewDimension()

	decorated.AssertCalled(t, "viewDimension")
	assert.Equal(t, ViewDimension{rows: 24, cols: 80}, returnedViewDimension, "Returned ViewDimension should match injected value")
}

func TestSelectableRowDecoratorProxiesCallToIsSelectableRow(t *testing.T) {
	selectableRowDecorator, decorated := setupSelectableRowDecorator()
	decorated.On("isSelectableRow", uint(8)).Return(true)

	returnedIsSelectableRow := selectableRowDecorator.isSelectableRow(8)

	decorated.AssertCalled(t, "isSelectableRow", uint(8))
	assert.True(t, returnedIsSelectableRow, "Return value from isSelectableRow match injected value")
}

func TestSelectableRowDecoratorDoesNotProxyCallToOnRowSelected(t *testing.T) {
	selectableRowDecorator, decorated := setupSelectableRowDecorator()

	returnedError := selectableRowDecorator.onRowSelected(4)

	decorated.AssertNotCalled(t, "onRowSelected", uint(4))
	assert.NoError(t, returnedError, "Returned error should be nil")
}

func TestSelectableRowDecoratorCallesOnRowSelectedWhennotifyChildRowSelectedIsCalled(t *testing.T) {
	selectableRowDecorator, decorated := setupSelectableRowDecorator()
	decorated.On("onRowSelected", uint(8)).Return(errors.New("Test error"))

	returnedError := selectableRowDecorator.notifyChildRowSelected(uint(8))

	decorated.AssertCalled(t, "onRowSelected", uint(8))
	assert.EqualError(t, returnedError, "Test error", "notifyChildRowSelected should returned error from onRowSelected")
}
