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

func TestSelectableRowDecoratorProxiesCallToLine(t *testing.T) {
	selectableRowDecorator, decorated := setupSelectableRowDecorator()
	decorated.On("line", uint(5)).Return("Line Text")

	returnedLine := selectableRowDecorator.line(5)

	decorated.AssertCalled(t, "line", uint(5))
	assert.Equal(t, returnedLine, "Line Text", "Return value from line does not match injected value")
}

type selectableRowViewMocks struct {
	viewPos   *MockViewPos
	child     *MockSelectableRowChildWindowView
	channels  *MockChannels
	config    *MockConfig
	variables *MockGRVVariableSetter
	lock      *MockLock
}

func setupSelectableRowView() (*SelectableRowView, *selectableRowViewMocks) {
	mocks := &selectableRowViewMocks{
		viewPos:   &MockViewPos{},
		child:     &MockSelectableRowChildWindowView{},
		channels:  &MockChannels{},
		config:    &MockConfig{},
		variables: &MockGRVVariableSetter{},
		lock:      &MockLock{},
	}

	mocks.child.On("rows").Return(uint(100))
	mocks.child.On("viewPos").Return(mocks.viewPos)
	mocks.viewPos.On("ViewStartRowIndex").Return(uint(0))
	mocks.viewPos.On("ViewStartColumn").Return(uint(1))
	mocks.channels.On("UpdateDisplay").Return()
	mocks.variables.On("SetViewVariable", VarLineCount, "100", ViewStateInvisible).Return()
	mocks.variables.On("SetViewVariable", VarLineText, "", ViewStateInvisible).Return()
	mocks.lock.On("Lock").Return()
	mocks.lock.On("Unlock").Return()

	return NewSelectableRowView(mocks.child, mocks.channels, mocks.config, mocks.variables, mocks.lock, "test line"), mocks
}

func TestWhenActionIsNotHandledThenSelectableCheckIsNotDone(t *testing.T) {
	selectableRowView, mocks := setupSelectableRowView()
	mocks.viewPos.On("ActiveRowIndex").Return(uint(0))

	handled, _ := selectableRowView.HandleAction(Action{ActionType: ActionNone})

	mocks.child.AssertNotCalled(t, "isSelectableRow", uint(0))
	assert.False(t, handled, "ActionNone should not be handled")
}

func TestWhenActionIsHandledButActiveRowIndexDoesNotChangeThenSelectableCheckIsNotDone(t *testing.T) {
	selectableRowView, mocks := setupSelectableRowView()
	mocks.viewPos.On("ActiveRowIndex").Return(uint(0)).Times(4)
	mocks.viewPos.On("MoveLineUp").Return(false)

	handled, _ := selectableRowView.HandleAction(Action{ActionType: ActionPrevLine})

	mocks.child.AssertNotCalled(t, "isSelectableRow", uint(0))
	assert.True(t, handled, "ActionPrevLine should be handled")
}

func TestWhenActionResultsInErrorThenErrorIsReturnedAndSelectableCheckIsNotDone(t *testing.T) {
	selectableRowView, mocks := setupSelectableRowView()
	mocks.viewPos.On("ActiveRowIndex").Return(uint(0)).Times(4)

	_, err := selectableRowView.HandleAction(Action{ActionType: ActionMouseSelect})

	mocks.child.AssertNotCalled(t, "isSelectableRow", uint(0))
	assert.NotNil(t, err, "Error should be returned for invalid action")
}

func TestWhenActiveRowIndexDoesChangeAndRowIsSelectableThenChildIsNotifiedRowIsSelected(t *testing.T) {
	selectableRowView, mocks := setupSelectableRowView()
	mocks.viewPos.On("ActiveRowIndex").Return(uint(0)).Times(2)
	mocks.viewPos.On("ActiveRowIndex").Return(uint(1))
	mocks.viewPos.On("MoveLineDown", uint(100)).Return(true)
	mocks.child.On("isSelectableRow", uint(1)).Return(true)
	mocks.child.On("onRowSelected", uint(1)).Return(nil)
	mocks.child.On("line", uint(1)).Return("")
	mocks.variables.On("SetViewVariable", VarLineNumer, "2", ViewStateInvisible).Return()

	selectableRowView.HandleAction(Action{ActionType: ActionNextLine})

	mocks.child.AssertCalled(t, "isSelectableRow", uint(1))
	mocks.child.AssertCalled(t, "onRowSelected", uint(1))
}

func TestWhenActiveRowIndexDoesChangeDownwardsAndRowIsNotSelectableThenNextSelectableRowDownIsFoundAndChildIsNotifiedRowIsSelected(t *testing.T) {
	selectableRowView, mocks := setupSelectableRowView()
	mocks.viewPos.On("ActiveRowIndex").Return(uint(0)).Times(2)
	mocks.viewPos.On("ActiveRowIndex").Return(uint(1)).Times(4)
	mocks.viewPos.On("ActiveRowIndex").Return(uint(2))
	mocks.viewPos.On("MoveLineDown", uint(100)).Return(true)
	mocks.child.On("isSelectableRow", uint(1)).Return(false)
	mocks.child.On("isSelectableRow", uint(2)).Return(true)
	mocks.viewPos.On("SetActiveRowIndex", uint(2)).Return()
	mocks.child.On("onRowSelected", uint(2)).Return(nil)
	mocks.child.On("line", uint(1)).Return("").Times(1)
	mocks.child.On("line", uint(2)).Return("")
	mocks.variables.On("SetViewVariable", VarLineNumer, "2", ViewStateInvisible).Return()
	mocks.variables.On("SetViewVariable", VarLineNumer, "3", ViewStateInvisible).Return()

	selectableRowView.HandleAction(Action{ActionType: ActionNextLine})

	mocks.child.AssertCalled(t, "isSelectableRow", uint(1))
	mocks.child.AssertCalled(t, "isSelectableRow", uint(2))
	mocks.viewPos.AssertCalled(t, "SetActiveRowIndex", uint(2))
	mocks.child.AssertCalled(t, "onRowSelected", uint(2))
}

func TestWhenActiveRowIndexDoesChangeUpwardsAndRowIsNotSelectableThenNextSelectableRowUpIsFoundAndChildIsNotifiedRowIsSelected(t *testing.T) {
	selectableRowView, mocks := setupSelectableRowView()
	mocks.viewPos.On("ActiveRowIndex").Return(uint(3)).Times(2)
	mocks.viewPos.On("ActiveRowIndex").Return(uint(2)).Times(4)
	mocks.viewPos.On("ActiveRowIndex").Return(uint(1))
	mocks.viewPos.On("MoveLineUp").Return(true)
	mocks.child.On("isSelectableRow", uint(2)).Return(false)
	mocks.child.On("isSelectableRow", uint(1)).Return(true)
	mocks.viewPos.On("SetActiveRowIndex", uint(1)).Return()
	mocks.child.On("onRowSelected", uint(1)).Return(nil)
	mocks.child.On("line", uint(2)).Return("")
	mocks.child.On("line", uint(1)).Return("")
	mocks.variables.On("SetViewVariable", VarLineNumer, "3", ViewStateInvisible).Return()
	mocks.variables.On("SetViewVariable", VarLineNumer, "2", ViewStateInvisible).Return()

	selectableRowView.HandleAction(Action{ActionType: ActionPrevLine})

	mocks.child.AssertCalled(t, "isSelectableRow", uint(2))
	mocks.child.AssertCalled(t, "isSelectableRow", uint(1))
	mocks.viewPos.AssertCalled(t, "SetActiveRowIndex", uint(1))
	mocks.child.AssertCalled(t, "onRowSelected", uint(1))
}

func TestWhenActiveRowIndexDoesChangeUpwardsAndRowIsNotSelectableThenNextSelectableRowDownIsFoundWhenNoUpwardsRowsAreAvailableAndChildIsNotifiedRowIsSelected(t *testing.T) {
	selectableRowView, mocks := setupSelectableRowView()
	mocks.viewPos.On("ActiveRowIndex").Return(uint(3)).Times(2)
	mocks.viewPos.On("ActiveRowIndex").Return(uint(2)).Times(4)
	mocks.viewPos.On("ActiveRowIndex").Return(uint(3))
	mocks.viewPos.On("MoveLineUp").Return(true)
	mocks.child.On("isSelectableRow", uint(2)).Return(false)
	mocks.child.On("isSelectableRow", uint(1)).Return(false)
	mocks.child.On("isSelectableRow", uint(0)).Return(false)
	mocks.child.On("isSelectableRow", uint(3)).Return(true)
	mocks.viewPos.On("SetActiveRowIndex", uint(3)).Return()
	mocks.child.On("onRowSelected", uint(3)).Return(nil)
	mocks.child.On("line", uint(2)).Return("")
	mocks.child.On("line", uint(3)).Return("")
	mocks.variables.On("SetViewVariable", VarLineNumer, "3", ViewStateInvisible).Return()
	mocks.variables.On("SetViewVariable", VarLineNumer, "4", ViewStateInvisible).Return()

	selectableRowView.HandleAction(Action{ActionType: ActionPrevLine})

	mocks.child.AssertCalled(t, "isSelectableRow", uint(2))
	mocks.child.AssertCalled(t, "isSelectableRow", uint(1))
	mocks.child.AssertCalled(t, "isSelectableRow", uint(0))
	mocks.child.AssertCalled(t, "isSelectableRow", uint(3))
	mocks.viewPos.AssertCalled(t, "SetActiveRowIndex", uint(3))
	mocks.child.AssertCalled(t, "onRowSelected", uint(3))
}

func TestWhenActiveRowIndexDoesChangeDownwardsAndRowIsNotSelectableThenNextSelectableRowUpIsFoundAndChildIsNotifiedRowIsSelectedIfNoDownwardsRowsAreAvailable(t *testing.T) {
	selectableRowView, mocks := setupSelectableRowView()
	mocks.viewPos.On("ActiveRowIndex").Return(uint(98)).Times(2)
	mocks.viewPos.On("ActiveRowIndex").Return(uint(99)).Times(4)
	mocks.viewPos.On("ActiveRowIndex").Return(uint(97))
	mocks.viewPos.On("MoveLineDown", uint(100)).Return(true)
	mocks.child.On("isSelectableRow", uint(99)).Return(false)
	mocks.child.On("isSelectableRow", uint(98)).Return(false)
	mocks.child.On("isSelectableRow", uint(97)).Return(true)
	mocks.viewPos.On("SetActiveRowIndex", uint(97)).Return()
	mocks.child.On("onRowSelected", uint(97)).Return(nil)
	mocks.child.On("line", uint(99)).Return("")
	mocks.child.On("line", uint(97)).Return("")
	mocks.variables.On("SetViewVariable", VarLineNumer, "100", ViewStateInvisible).Return()
	mocks.variables.On("SetViewVariable", VarLineNumer, "98", ViewStateInvisible).Return()

	selectableRowView.HandleAction(Action{ActionType: ActionNextLine})

	mocks.child.AssertCalled(t, "isSelectableRow", uint(99))
	mocks.child.AssertCalled(t, "isSelectableRow", uint(98))
	mocks.child.AssertCalled(t, "isSelectableRow", uint(97))
	mocks.viewPos.AssertCalled(t, "SetActiveRowIndex", uint(97))
	mocks.child.AssertCalled(t, "onRowSelected", uint(97))
}

func TestWhenRowCountIsZeroThenSelectNearestSelectableRowDoesNothing(t *testing.T) {
	selectableRowView, mocks := setupSelectableRowView()

	mocks.child.On("rows").Return(uint(0))
	mocks.child.ExpectedCalls = mocks.child.ExpectedCalls[1:]

	selectableRowView.selectNearestSelectableRow()
}

func TestWhenCurrentRowIsSelectableThenSelectNearestSelectableRowDoesNothing(t *testing.T) {
	selectableRowView, mocks := setupSelectableRowView()

	mocks.viewPos.On("ActiveRowIndex").Return(uint(0))
	mocks.child.On("isSelectableRow", uint(0)).Return(true)

	selectableRowView.selectNearestSelectableRow()
}

func TestWhenCurrentRowIsNotSelectableThenSelectNearestSelectableRowSetsNearestSelectableRow(t *testing.T) {
	selectableRowView, mocks := setupSelectableRowView()

	mocks.viewPos.On("ActiveRowIndex").Return(uint(0))
	mocks.child.On("isSelectableRow", uint(0)).Return(false)
	mocks.child.On("isSelectableRow", uint(1)).Return(true)
	mocks.viewPos.On("SetActiveRowIndex", uint(1)).Return()
	mocks.child.On("onRowSelected", uint(1)).Return(nil)
	mocks.child.On("line", uint(0)).Return("")
	mocks.child.On("line", uint(1)).Return("")
	mocks.variables.On("SetViewVariable", VarLineNumer, "1", ViewStateInvisible).Return()
	mocks.variables.On("SetViewVariable", VarLineNumer, "2", ViewStateInvisible).Return()

	selectableRowView.selectNearestSelectableRow()

	mocks.viewPos.AssertCalled(t, "SetActiveRowIndex", uint(1))
	mocks.child.AssertCalled(t, "onRowSelected", uint(1))
}

func TestErrorReturnedByOnRowSelectedIsReturnedBySelectNeartestSelectableRow(t *testing.T) {
	selectableRowView, mocks := setupSelectableRowView()

	mocks.viewPos.On("ActiveRowIndex").Return(uint(0))
	mocks.child.On("isSelectableRow", uint(0)).Return(false)
	mocks.child.On("isSelectableRow", uint(1)).Return(true)
	mocks.viewPos.On("SetActiveRowIndex", uint(1)).Return()
	mocks.child.On("onRowSelected", uint(1)).Return(errors.New("Test error"))
	mocks.child.On("line", uint(0)).Return("")
	mocks.child.On("line", uint(1)).Return("")
	mocks.variables.On("SetViewVariable", VarLineNumer, "1", ViewStateInvisible).Return()
	mocks.variables.On("SetViewVariable", VarLineNumer, "2", ViewStateInvisible).Return()

	returnedError := selectableRowView.selectNearestSelectableRow()

	assert.EqualError(t, returnedError, "Test error", "SelectNearestSelectableRow should return error returned by onRowSelected")
}

func TestNotifyChildRowSelectedCallsSetVariables(t *testing.T) {
	selectableRowView, mocks := setupSelectableRowView()

	mocks.child.On("line", uint(5)).Return("Line Text")
	mocks.viewPos.On("ActiveRowIndex").Return(uint(5))
	mocks.child.On("onRowSelected", uint(5)).Return(nil)
	mocks.variables.On("SetViewVariable", VarLineNumer, "6", ViewStateInvisible).Return()
	mocks.variables.On("SetViewVariable", VarLineCount, "100", ViewStateInvisible).Return()
	mocks.variables.On("SetViewVariable", VarLineText, "Line Text", ViewStateInvisible).Return()

	selectableRowView.notifyChildRowSelected(5)

	mocks.variables.AssertCalled(t, "SetViewVariable", VarLineNumer, "6", ViewStateInvisible)
	mocks.variables.AssertCalled(t, "SetViewVariable", VarLineCount, "100", ViewStateInvisible)
	mocks.variables.AssertCalled(t, "SetViewVariable", VarLineText, "Line Text", ViewStateInvisible)
}

func TestOnSearchMatchUpdatesTheViewPosAndNotifiesTheSelectableChild(t *testing.T) {
	selectableRowView, mocks := setupSelectableRowView()

	mocks.viewPos.On("SetActiveRowIndex", uint(10)).Return()
	mocks.viewPos.On("ActiveRowIndex").Return(uint(10))
	mocks.child.On("isSelectableRow", uint(10)).Return(true)
	mocks.child.On("onRowSelected", uint(10)).Return(nil)
	mocks.child.On("line", uint(10)).Return("Line Text")

	mocks.variables.On("SetViewVariable", VarLineNumer, "11", ViewStateInvisible).Return()
	mocks.variables.On("SetViewVariable", VarLineCount, "100", ViewStateInvisible).Return()
	mocks.variables.On("SetViewVariable", VarLineText, "Line Text", ViewStateInvisible).Return()

	selectableRowView.OnSearchMatch(mocks.viewPos, 10)

	mocks.lock.AssertCalled(t, "Lock")
	mocks.lock.AssertCalled(t, "Unlock")

	mocks.viewPos.AssertCalled(t, "SetActiveRowIndex", uint(10))
	mocks.child.AssertCalled(t, "onRowSelected", uint(10))
}

func TestOnSearchMatchDoesNotUpdateViewPosAndNotifyTheSelectableChildWhenTheRowIsNotSelectable(t *testing.T) {
	selectableRowView, mocks := setupSelectableRowView()

	mocks.child.On("isSelectableRow", uint(10)).Return(false)

	selectableRowView.OnSearchMatch(mocks.viewPos, 10)

	mocks.lock.AssertCalled(t, "Lock")
	mocks.lock.AssertCalled(t, "Unlock")

	mocks.viewPos.AssertNotCalled(t, "SetActiveRowIndex", uint(10))
	mocks.child.AssertNotCalled(t, "onRowSelected", uint(10))
}

func TestOnSearchMatchDoesNotUpdateViewPosAndNotifySelectableChildWhenViewPosHasChanged(t *testing.T) {
	selectableRowView, mocks := setupSelectableRowView()

	selectableRowView.OnSearchMatch(NewViewPosition(), 10)

	mocks.lock.AssertCalled(t, "Lock")
	mocks.lock.AssertCalled(t, "Unlock")

	mocks.viewPos.AssertNotCalled(t, "SetActiveRowIndex", uint(10))
	mocks.child.AssertNotCalled(t, "onRowSelected", uint(10))
}

func TestOnSearchMatchDoesNotUpdateViewPosAndNotifySelectableChildWhenMatchedLineIndexIsTooLarge(t *testing.T) {
	selectableRowView, mocks := setupSelectableRowView()

	selectableRowView.OnSearchMatch(mocks.viewPos, 200)

	mocks.lock.AssertCalled(t, "Lock")
	mocks.lock.AssertCalled(t, "Unlock")

	mocks.viewPos.AssertNotCalled(t, "SetActiveRowIndex", uint(200))
	mocks.child.AssertNotCalled(t, "onRowSelected", uint(200))
}
