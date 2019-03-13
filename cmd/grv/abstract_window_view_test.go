package main

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockViewPos struct {
	mock.Mock
}

func (viewPos *MockViewPos) ActiveRowIndex() uint {
	args := viewPos.Called()
	return args.Get(0).(uint)
}

func (viewPos *MockViewPos) SetActiveRowIndex(activeRowIndex uint) {
	viewPos.Called(activeRowIndex)
}

func (viewPos *MockViewPos) ViewStartRowIndex() uint {
	args := viewPos.Called()
	return args.Get(0).(uint)
}

func (viewPos *MockViewPos) ViewStartColumn() uint {
	args := viewPos.Called()
	return args.Get(0).(uint)
}

func (viewPos *MockViewPos) SelectedRowIndex() uint {
	args := viewPos.Called()
	return args.Get(0).(uint)
}

func (viewPos *MockViewPos) DetermineViewStartRow(viewRows, rows uint) {
	viewPos.Called(viewRows, rows)
}

func (viewPos *MockViewPos) MoveLineDown(rows uint) (changed bool) {
	args := viewPos.Called(rows)
	return args.Bool(0)
}

func (viewPos *MockViewPos) MoveLineUp() (changed bool) {
	args := viewPos.Called()
	return args.Bool(0)
}

func (viewPos *MockViewPos) MovePageDown(pageRows, rows uint) (changed bool) {
	args := viewPos.Called(pageRows, rows)
	return args.Bool(0)
}

func (viewPos *MockViewPos) MovePageUp(pageRows uint) (changed bool) {
	args := viewPos.Called(pageRows)
	return args.Bool(0)
}

func (viewPos *MockViewPos) MovePageRight(cols uint) {
	viewPos.Called(cols)
}

func (viewPos *MockViewPos) MovePageLeft(cols uint) (changed bool) {
	args := viewPos.Called(cols)
	return args.Bool(0)
}

func (viewPos *MockViewPos) MoveToFirstLine() (changed bool) {
	args := viewPos.Called()
	return args.Bool(0)
}

func (viewPos *MockViewPos) MoveToLastLine(rows uint) (changed bool) {
	args := viewPos.Called(rows)
	return args.Bool(0)
}

func (viewPos *MockViewPos) CenterActiveRow(pageRows uint) (changed bool) {
	args := viewPos.Called(pageRows)
	return args.Bool(0)
}

func (viewPos *MockViewPos) ScrollActiveRowTop() (changed bool) {
	args := viewPos.Called()
	return args.Bool(0)
}

func (viewPos *MockViewPos) ScrollActiveRowBottom(pageRows uint) (changed bool) {
	args := viewPos.Called(pageRows)
	return args.Bool(0)
}

func (viewPos *MockViewPos) MoveCursorTopPage() (changed bool) {
	args := viewPos.Called()
	return args.Bool(0)
}

func (viewPos *MockViewPos) MoveCursorMiddlePage(pageRows, rows uint) (changed bool) {
	args := viewPos.Called(pageRows, rows)
	return args.Bool(0)
}

func (viewPos *MockViewPos) MoveCursorBottomPage(pageRows, rows uint) (changed bool) {
	args := viewPos.Called(pageRows, rows)
	return args.Bool(0)
}

func (viewPos *MockViewPos) ScrollDown(rows, pageRows, scrollRows uint) (changed bool) {
	args := viewPos.Called(rows, pageRows, scrollRows)
	return args.Bool(0)
}

func (viewPos *MockViewPos) ScrollUp(pageRows, scrollRows uint) (changed bool) {
	args := viewPos.Called(pageRows, scrollRows)
	return args.Bool(0)
}

type MockChildWindowView struct {
	mock.Mock
}

func (childWindowView *MockChildWindowView) viewPos() ViewPos {
	args := childWindowView.Called()
	return args.Get(0).(ViewPos)
}

func (childWindowView *MockChildWindowView) rows() uint {
	args := childWindowView.Called()
	return args.Get(0).(uint)
}

func (childWindowView *MockChildWindowView) viewDimension() ViewDimension {
	args := childWindowView.Called()
	return args.Get(0).(ViewDimension)
}

func (childWindowView *MockChildWindowView) onRowSelected(rowIndex uint) error {
	args := childWindowView.Called(rowIndex)
	return args.Error(0)
}

func (childWindowView *MockChildWindowView) line(lineIndex uint) string {
	args := childWindowView.Called(lineIndex)
	return args.String(0)
}

type MockChannels struct {
	mock.Mock
}

func (channels *MockChannels) UpdateDisplay() {
	channels.Called()
}

func (channels *MockChannels) Exit() bool {
	args := channels.Called()
	return args.Bool(0)
}

func (channels *MockChannels) ReportError(err error) {
	channels.Called(err)
}

func (channels *MockChannels) ReportErrors(errors []error) {
	channels.Called(errors)
}

func (channels *MockChannels) DoAction(action Action) {
	channels.Called(action)
}

func (channels *MockChannels) ReportEvent(event Event) {
	channels.Called(event)
}

func (channels *MockChannels) ReportStatus(format string, args ...interface{}) {
	channels.Called(format, args)
}

type MockConfig struct {
	mock.Mock
}

func (config *MockConfig) GetBool(configVariable ConfigVariable) bool {
	args := config.Called(configVariable)
	return args.Bool(0)
}

func (config *MockConfig) GetString(configVariable ConfigVariable) string {
	args := config.Called(configVariable)
	return args.String(0)
}

func (config *MockConfig) GetInt(configVariable ConfigVariable) int {
	args := config.Called(configVariable)
	return args.Int(0)
}

func (config *MockConfig) GetFloat(configVariable ConfigVariable) float64 {
	args := config.Called(configVariable)
	return args.Get(0).(float64)
}

func (config *MockConfig) GetTheme() Theme {
	args := config.Called()
	return args.Get(0).(Theme)
}

func (config *MockConfig) AddOnChangeListener(configVariable ConfigVariable, configVariableOnChangeListener ConfigVariableOnChangeListener) {
	config.Called(configVariable, configVariableOnChangeListener)
}

func (config *MockConfig) ConfigDir() string {
	args := config.Called()
	return args.String(0)
}

func (config *MockConfig) KeyStrings(actionType ActionType, viewHierarchy ViewHierarchy) []BoundKeyString {
	args := config.Called(actionType, viewHierarchy)
	return args.Get(0).([]BoundKeyString)
}

func (config *MockConfig) GenerateHelpSections() []*HelpSection {
	args := config.Called()
	return args.Get(0).([]*HelpSection)
}

type MockGRVVariableSetter struct {
	mock.Mock
}

func (variables *MockGRVVariableSetter) SetViewVariable(variable GRVVariable, value string, viewState ViewState) {
	variables.Called(variable, value, viewState)
}

func (variables *MockGRVVariableSetter) ClearViewVariable(variable GRVVariable, viewState ViewState) {
	variables.Called(variable, viewState)
}

func (variables *MockGRVVariableSetter) VariableValues() map[GRVVariable]string {
	args := variables.Called()
	return args.Get(0).(map[GRVVariable]string)
}

func (variables *MockGRVVariableSetter) VariableValue(variable GRVVariable) (value string, isSet bool) {
	args := variables.Called(variable)
	return args.String(0), args.Bool(1)
}

type MockLock struct {
	mock.Mock
}

func (lock *MockLock) Lock() {
	lock.Called()
}

func (lock *MockLock) Unlock() {
	lock.Called()
}

type abstractWindowViewMocks struct {
	viewPos   *MockViewPos
	child     *MockChildWindowView
	channels  *MockChannels
	config    *MockConfig
	variables *MockGRVVariableSetter
	lock      *MockLock
}

func setupAbstractWindowView() (*AbstractWindowView, *abstractWindowViewMocks) {
	mocks := &abstractWindowViewMocks{
		viewPos:   &MockViewPos{},
		child:     &MockChildWindowView{},
		channels:  &MockChannels{},
		config:    &MockConfig{},
		variables: &MockGRVVariableSetter{},
		lock:      &MockLock{},
	}

	mocks.child.On("viewPos").Return(mocks.viewPos)
	mocks.child.On("onRowSelected", uint(0)).Return(nil)
	mocks.child.On("viewDimension").Return(ViewDimension{rows: 24, cols: 80})
	mocks.child.On("rows").Return(uint(24))
	mocks.child.On("line", uint(0)).Return("")

	mocks.viewPos.On("ActiveRowIndex").Return(uint(0))
	mocks.viewPos.On("ViewStartRowIndex").Return(uint(0))
	mocks.viewPos.On("ViewStartColumn").Return(uint(1))

	mocks.channels.On("UpdateDisplay").Return()

	mocks.config.On("GetInt", CfMouseScrollRows).Return(3)

	mocks.variables.On("SetViewVariable", VarLineNumer, "1", ViewStateInvisible)
	mocks.variables.On("SetViewVariable", VarLineCount, "24", ViewStateInvisible)
	mocks.variables.On("SetViewVariable", VarLineText, "", ViewStateInvisible)

	mocks.lock.On("Lock").Return()
	mocks.lock.On("Unlock").Return()

	return NewAbstractWindowView(mocks.child, mocks.channels, mocks.config, mocks.variables, mocks.lock, "test line"), mocks
}

func assertChildViewAndDisplayUpdated(t *testing.T, mocks *abstractWindowViewMocks) {
	mocks.child.AssertCalled(t, "onRowSelected", uint(0))
	mocks.channels.AssertCalled(t, "UpdateDisplay")
}

func assertChildViewAndDisplayNotUpdated(t *testing.T, mocks *abstractWindowViewMocks) {
	mocks.child.AssertNotCalled(t, "onRowSelected", uint(0))
	mocks.channels.AssertNotCalled(t, "UpdateDisplay")
}

func TestBorderWidthDefaultsToTwo(t *testing.T) {
	abstractWindowView, _ := setupAbstractWindowView()

	if abstractWindowView.borderWidth != 2 {
		t.Errorf("Expected borderWidth to be 2 but found: %v", abstractWindowView.borderWidth)
	}
}

func TestHandleActionReturnsHandledTrueWhenActionIsHandled(t *testing.T) {
	abstractWindowView, mocks := setupAbstractWindowView()
	mocks.viewPos.On("MoveLineUp").Return(false)

	handled, _ := abstractWindowView.HandleAction(Action{ActionType: ActionPrevLine})

	if !handled {
		t.Errorf("Expected handled to be true")
	}
}

func TestHandleActionReturnsHandledFalseWhenActionIsHandled(t *testing.T) {
	abstractWindowView, _ := setupAbstractWindowView()

	handled, _ := abstractWindowView.HandleAction(Action{ActionType: ActionNone})

	if handled {
		t.Errorf("Expected handled to be false")
	}
}

func TestErrorByActionHandlerIsReturned(t *testing.T) {
	abstractWindowView, mocks := setupAbstractWindowView()
	errTest := errors.New("Test error")

	mocks.viewPos.On("MoveLineUp").Return(true)
	mocks.child = &MockChildWindowView{}
	mocks.child.On("viewPos").Return(mocks.viewPos)
	mocks.child.On("rows").Return(uint(24))
	mocks.child.On("line", uint(0)).Return("")
	abstractWindowView.child = mocks.child
	mocks.child.On("onRowSelected", uint(0)).Return(errTest)

	_, err := abstractWindowView.HandleAction(Action{ActionType: ActionPrevLine})

	if errTest != err {
		t.Errorf(`Expected error returned to be "%v" but found "%v"`, errTest, err)
	}
}

func TestViewPosIsLoggedBeforeAndAfterActionIsHandled(t *testing.T) {
	abstractWindowView, mocks := setupAbstractWindowView()
	mocks.viewPos.On("MoveLineUp").Return(false)

	abstractWindowView.HandleAction(Action{ActionType: ActionPrevLine})

	mocks.viewPos.AssertNumberOfCalls(t, "ActiveRowIndex", 2)
	mocks.viewPos.AssertNumberOfCalls(t, "ViewStartRowIndex", 2)
	mocks.viewPos.AssertNumberOfCalls(t, "ViewStartColumn", 2)
}

func TestActionPrevLineIsHandledAndNoUpdatesResultWhenMoveLineUpReturnsFalse(t *testing.T) {
	abstractWindowView, mocks := setupAbstractWindowView()
	mocks.viewPos.On("MoveLineUp").Return(false)

	abstractWindowView.HandleAction(Action{ActionType: ActionPrevLine})

	mocks.viewPos.AssertCalled(t, "MoveLineUp")
	assertChildViewAndDisplayNotUpdated(t, mocks)
}

func TestActionPrevLineIsHandledAndUpdatesResultWhenMoveLineUpReturnsTrue(t *testing.T) {
	abstractWindowView, mocks := setupAbstractWindowView()
	mocks.viewPos.On("MoveLineUp").Return(true)

	abstractWindowView.HandleAction(Action{ActionType: ActionPrevLine})

	mocks.viewPos.AssertCalled(t, "MoveLineUp")
	assertChildViewAndDisplayUpdated(t, mocks)
}

func TestActionNextLineIsHandledAndNoUpdatesResultWhenMoveLineDownReturnsFalse(t *testing.T) {
	abstractWindowView, mocks := setupAbstractWindowView()
	mocks.viewPos.On("MoveLineDown", uint(24)).Return(false)

	abstractWindowView.HandleAction(Action{ActionType: ActionNextLine})

	mocks.viewPos.AssertCalled(t, "MoveLineDown", uint(24))
	assertChildViewAndDisplayNotUpdated(t, mocks)
}

func TestActionNextLineIsHandledAndUpdatesResultWhenMoveLineDownReturnsTrue(t *testing.T) {
	abstractWindowView, mocks := setupAbstractWindowView()
	mocks.viewPos.On("MoveLineDown", uint(24)).Return(true)

	abstractWindowView.HandleAction(Action{ActionType: ActionNextLine})

	mocks.viewPos.AssertCalled(t, "MoveLineDown", uint(24))
	assertChildViewAndDisplayUpdated(t, mocks)
}

func TestActionPrevPageIsHandledAndNoUpdatesResultWhenMovePageUpReturnsFalse(t *testing.T) {
	abstractWindowView, mocks := setupAbstractWindowView()
	mocks.viewPos.On("MovePageUp", uint(22)).Return(false)

	abstractWindowView.HandleAction(Action{ActionType: ActionPrevPage})

	mocks.viewPos.AssertCalled(t, "MovePageUp", uint(22))
	assertChildViewAndDisplayNotUpdated(t, mocks)
}

func TestActionPrevPageIsHandledAndUpdatesResultWhenMovePageUpReturnsTrue(t *testing.T) {
	abstractWindowView, mocks := setupAbstractWindowView()
	mocks.viewPos.On("MovePageUp", uint(22)).Return(true)

	abstractWindowView.HandleAction(Action{ActionType: ActionPrevPage})

	mocks.viewPos.AssertCalled(t, "MovePageUp", uint(22))
	assertChildViewAndDisplayUpdated(t, mocks)
}

func TestActionNextPageIsHandledAndNoUpdatesResultWhenMovePageDownReturnsFalse(t *testing.T) {
	abstractWindowView, mocks := setupAbstractWindowView()
	mocks.viewPos.On("MovePageDown", uint(22), uint(24)).Return(false)

	abstractWindowView.HandleAction(Action{ActionType: ActionNextPage})

	mocks.viewPos.AssertCalled(t, "MovePageDown", uint(22), uint(24))
	assertChildViewAndDisplayNotUpdated(t, mocks)
}

func TestActionNextPageIsHandledAndUpdatesResultWhenMovePageDownReturnsTrue(t *testing.T) {
	abstractWindowView, mocks := setupAbstractWindowView()
	mocks.viewPos.On("MovePageDown", uint(22), uint(24)).Return(true)

	abstractWindowView.HandleAction(Action{ActionType: ActionNextPage})

	mocks.viewPos.AssertCalled(t, "MovePageDown", uint(22), uint(24))
	assertChildViewAndDisplayUpdated(t, mocks)
}

func TestActionPrevHalfPageIsHandledAndNoUpdatesResultWhenMovePageUpReturnsFalse(t *testing.T) {
	abstractWindowView, mocks := setupAbstractWindowView()
	mocks.viewPos.On("MovePageUp", uint(10)).Return(false)

	abstractWindowView.HandleAction(Action{ActionType: ActionPrevHalfPage})

	mocks.viewPos.AssertCalled(t, "MovePageUp", uint(10))
	assertChildViewAndDisplayNotUpdated(t, mocks)
}

func TestActionPrevHalfPageIsHandledAndUpdatesResultWhenMovePageUpReturnsTrue(t *testing.T) {
	abstractWindowView, mocks := setupAbstractWindowView()
	mocks.viewPos.On("MovePageUp", uint(10)).Return(true)

	abstractWindowView.HandleAction(Action{ActionType: ActionPrevHalfPage})

	mocks.viewPos.AssertCalled(t, "MovePageUp", uint(10))
	assertChildViewAndDisplayUpdated(t, mocks)
}

func TestActionNextHalfPageIsHandledAndNoUpdatesResultWhenMovePageDownReturnsFalse(t *testing.T) {
	abstractWindowView, mocks := setupAbstractWindowView()
	mocks.viewPos.On("MovePageDown", uint(10), uint(24)).Return(false)

	abstractWindowView.HandleAction(Action{ActionType: ActionNextHalfPage})

	mocks.viewPos.AssertCalled(t, "MovePageDown", uint(10), uint(24))
	assertChildViewAndDisplayNotUpdated(t, mocks)
}

func TestActionNextHalfPageIsHandledAndUpdatesResultWhenMovePageDownReturnsTrue(t *testing.T) {
	abstractWindowView, mocks := setupAbstractWindowView()
	mocks.viewPos.On("MovePageDown", uint(10), uint(24)).Return(true)

	abstractWindowView.HandleAction(Action{ActionType: ActionNextHalfPage})

	mocks.viewPos.AssertCalled(t, "MovePageDown", uint(10), uint(24))
	assertChildViewAndDisplayUpdated(t, mocks)
}

func TestActionScrollRightIsHandled(t *testing.T) {
	abstractWindowView, mocks := setupAbstractWindowView()
	mocks.viewPos.On("MovePageRight", uint(80)).Return()

	abstractWindowView.HandleAction(Action{ActionType: ActionScrollRight})

	mocks.viewPos.AssertCalled(t, "MovePageRight", uint(80))
	mocks.channels.AssertCalled(t, "UpdateDisplay")
}

func TestActionScrollLeftIsHandledAndNoUpdatesResultWhenMovePageLeftReturnsFalse(t *testing.T) {
	abstractWindowView, mocks := setupAbstractWindowView()
	mocks.viewPos.On("MovePageLeft", uint(80)).Return(false)

	abstractWindowView.HandleAction(Action{ActionType: ActionScrollLeft})

	mocks.viewPos.AssertCalled(t, "MovePageLeft", uint(80))
	mocks.channels.AssertNotCalled(t, "UpdateDisplay")
}

func TestActionScrollLeftIsHandledAndUpdatesResultWhenMovePageLeftReturnsTrue(t *testing.T) {
	abstractWindowView, mocks := setupAbstractWindowView()
	mocks.viewPos.On("MovePageLeft", uint(80)).Return(true)

	abstractWindowView.HandleAction(Action{ActionType: ActionScrollLeft})

	mocks.viewPos.AssertCalled(t, "MovePageLeft", uint(80))
	mocks.channels.AssertCalled(t, "UpdateDisplay")
}

func TestActionFirstLineIsHandledAndNoUpdatesResultWhenMoveToFirstLineReturnsFalse(t *testing.T) {
	abstractWindowView, mocks := setupAbstractWindowView()
	mocks.viewPos.On("MoveToFirstLine").Return(false)

	abstractWindowView.HandleAction(Action{ActionType: ActionFirstLine})

	mocks.viewPos.AssertCalled(t, "MoveToFirstLine")
	assertChildViewAndDisplayNotUpdated(t, mocks)
}

func TestActionFirstLineIsHandledAndUpdatesResultWhenMoveToFirstLineReturnsTrue(t *testing.T) {
	abstractWindowView, mocks := setupAbstractWindowView()
	mocks.viewPos.On("MoveToFirstLine").Return(true)

	abstractWindowView.HandleAction(Action{ActionType: ActionFirstLine})

	mocks.viewPos.AssertCalled(t, "MoveToFirstLine")
	assertChildViewAndDisplayUpdated(t, mocks)
}

func TestActionLastLineIsHandledAndNoUpdatesResultWhenMoveToLastLineReturnsFalse(t *testing.T) {
	abstractWindowView, mocks := setupAbstractWindowView()
	mocks.viewPos.On("MoveToLastLine", uint(24)).Return(false)

	abstractWindowView.HandleAction(Action{ActionType: ActionLastLine})

	mocks.viewPos.AssertCalled(t, "MoveToLastLine", uint(24))
	assertChildViewAndDisplayNotUpdated(t, mocks)
}

func TestActionLastLineIsHandledAndUpdatesResultWhenMoveToLastLineReturnsTrue(t *testing.T) {
	abstractWindowView, mocks := setupAbstractWindowView()
	mocks.viewPos.On("MoveToLastLine", uint(24)).Return(true)

	abstractWindowView.HandleAction(Action{ActionType: ActionLastLine})

	mocks.viewPos.AssertCalled(t, "MoveToLastLine", uint(24))
	assertChildViewAndDisplayUpdated(t, mocks)
}

func TestActionCenterViewIsHandledAndNoUpdatesResultWhenCenterActiveRowReturnsFalse(t *testing.T) {
	abstractWindowView, mocks := setupAbstractWindowView()
	mocks.viewPos.On("CenterActiveRow", uint(22)).Return(false)

	abstractWindowView.HandleAction(Action{ActionType: ActionCenterView})

	mocks.viewPos.AssertCalled(t, "CenterActiveRow", uint(22))
	mocks.channels.AssertNotCalled(t, "UpdateDisplay")
}

func TestActionCenterViewIsHandledAndUpdatesResultWhenCenterActiveRowReturnsTrue(t *testing.T) {
	abstractWindowView, mocks := setupAbstractWindowView()
	mocks.viewPos.On("CenterActiveRow", uint(22)).Return(true)

	abstractWindowView.HandleAction(Action{ActionType: ActionCenterView})

	mocks.viewPos.AssertCalled(t, "CenterActiveRow", uint(22))
	mocks.channels.AssertCalled(t, "UpdateDisplay")
}

func TestActionScrollCursorTopIsHandledAndNoUpdatesResultWhenScrollActiveRowTopReturnsFalse(t *testing.T) {
	abstractWindowView, mocks := setupAbstractWindowView()
	mocks.viewPos.On("ScrollActiveRowTop").Return(false)

	abstractWindowView.HandleAction(Action{ActionType: ActionScrollCursorTop})

	mocks.viewPos.AssertCalled(t, "ScrollActiveRowTop")
	mocks.channels.AssertNotCalled(t, "UpdateDisplay")
}

func TestActionScrollCursorTopIsHandledAndUpdatesResultWhenScrollActiveRowTopReturnsTrue(t *testing.T) {
	abstractWindowView, mocks := setupAbstractWindowView()
	mocks.viewPos.On("ScrollActiveRowTop").Return(true)

	abstractWindowView.HandleAction(Action{ActionType: ActionScrollCursorTop})

	mocks.viewPos.AssertCalled(t, "ScrollActiveRowTop")
	mocks.channels.AssertCalled(t, "UpdateDisplay")
}

func TestActionScrollCursorBottomIsHandledAndNoUpdatesResultWhenScrollActiveRowBottomReturnsFalse(t *testing.T) {
	abstractWindowView, mocks := setupAbstractWindowView()
	mocks.viewPos.On("ScrollActiveRowBottom", uint(22)).Return(false)

	abstractWindowView.HandleAction(Action{ActionType: ActionScrollCursorBottom})

	mocks.viewPos.AssertCalled(t, "ScrollActiveRowBottom", uint(22))
	mocks.channels.AssertNotCalled(t, "UpdateDisplay")
}

func TestActionScrollCursorBottomIsHandledAndUpdatesResultWhenScrollActiveRowBottomReturnsTrue(t *testing.T) {
	abstractWindowView, mocks := setupAbstractWindowView()
	mocks.viewPos.On("ScrollActiveRowBottom", uint(22)).Return(true)

	abstractWindowView.HandleAction(Action{ActionType: ActionScrollCursorBottom})

	mocks.viewPos.AssertCalled(t, "ScrollActiveRowBottom", uint(22))
	mocks.channels.AssertCalled(t, "UpdateDisplay")
}

func TestActionCursorTopViewIsHandledAndNoUpdatesResultWhenMoveCursorTopPageReturnsFalse(t *testing.T) {
	abstractWindowView, mocks := setupAbstractWindowView()
	mocks.viewPos.On("MoveCursorTopPage").Return(false)

	abstractWindowView.HandleAction(Action{ActionType: ActionCursorTopView})

	mocks.viewPos.AssertCalled(t, "MoveCursorTopPage")
	assertChildViewAndDisplayNotUpdated(t, mocks)
}

func TestActionCursorTopViewIsHandledAndUpdatesResultWhenMoveCursorTopPageReturnsTrue(t *testing.T) {
	abstractWindowView, mocks := setupAbstractWindowView()
	mocks.viewPos.On("MoveCursorTopPage").Return(true)

	abstractWindowView.HandleAction(Action{ActionType: ActionCursorTopView})

	mocks.viewPos.AssertCalled(t, "MoveCursorTopPage")
	assertChildViewAndDisplayUpdated(t, mocks)
}

func TestActionCursorMiddleViewIsHandledAndNoUpdatesResultWhenMoveCursorMiddlePageReturnsFalse(t *testing.T) {
	abstractWindowView, mocks := setupAbstractWindowView()
	mocks.viewPos.On("MoveCursorMiddlePage", uint(22), uint(24)).Return(false)

	abstractWindowView.HandleAction(Action{ActionType: ActionCursorMiddleView})

	mocks.viewPos.AssertCalled(t, "MoveCursorMiddlePage", uint(22), uint(24))
	assertChildViewAndDisplayNotUpdated(t, mocks)
}

func TestActionCursorMiddleViewIsHandledAndUpdatesResultWhenMoveCursorMiddlePageReturnsTrue(t *testing.T) {
	abstractWindowView, mocks := setupAbstractWindowView()
	mocks.viewPos.On("MoveCursorMiddlePage", uint(22), uint(24)).Return(true)

	abstractWindowView.HandleAction(Action{ActionType: ActionCursorMiddleView})

	mocks.viewPos.AssertCalled(t, "MoveCursorMiddlePage", uint(22), uint(24))
	assertChildViewAndDisplayUpdated(t, mocks)
}

func TestActionCursorBottomViewIsHandledAndNoUpdatesResultWhenMoveCursorBottomPageReturnsFalse(t *testing.T) {
	abstractWindowView, mocks := setupAbstractWindowView()
	mocks.viewPos.On("MoveCursorBottomPage", uint(22), uint(24)).Return(false)

	abstractWindowView.HandleAction(Action{ActionType: ActionCursorBottomView})

	mocks.viewPos.AssertCalled(t, "MoveCursorBottomPage", uint(22), uint(24))
	assertChildViewAndDisplayNotUpdated(t, mocks)
}

func TestActionCursorBottomViewIsHandledAndUpdatesResultWhenMoveCursorBottomPageReturnsTrue(t *testing.T) {
	abstractWindowView, mocks := setupAbstractWindowView()
	mocks.viewPos.On("MoveCursorBottomPage", uint(22), uint(24)).Return(true)

	abstractWindowView.HandleAction(Action{ActionType: ActionCursorBottomView})

	mocks.viewPos.AssertCalled(t, "MoveCursorBottomPage", uint(22), uint(24))
	assertChildViewAndDisplayUpdated(t, mocks)
}

func TestActionMouseSelectIsHandledAndResultsInAnErrorWhenActionIsInvalid(t *testing.T) {
	abstractWindowView, _ := setupAbstractWindowView()

	_, err := abstractWindowView.HandleAction(Action{ActionType: ActionMouseSelect})

	if err == nil {
		t.Errorf("Expected error but returned error was nil")
	}
}

func TestActionMouseSelectIsHandledAndNoUpdatesResultWhenClickIsOnBorders(t *testing.T) {
	abstractWindowView, mocks := setupAbstractWindowView()

	abstractWindowView.HandleAction(Action{
		ActionType: ActionMouseSelect,
		Args: []interface{}{MouseEvent{
			mouseEventType: MetLeftClick,
			row:            0,
			col:            0,
		}},
	})

	abstractWindowView.HandleAction(Action{
		ActionType: ActionMouseSelect,
		Args: []interface{}{MouseEvent{
			mouseEventType: MetLeftClick,
			row:            23,
			col:            0,
		}},
	})

	assertChildViewAndDisplayNotUpdated(t, mocks)
}

func TestActionMouseSelectIsHandledAndNoUpdatesResultWhenClickIsAfterRowsEnd(t *testing.T) {
	abstractWindowView, mocks := setupAbstractWindowView()
	mocks.viewPos.On("ViewStartRowIndex").Return(uint(0))

	abstractWindowView.HandleAction(Action{
		ActionType: ActionMouseSelect,
		Args: []interface{}{MouseEvent{
			mouseEventType: MetLeftClick,
			row:            128,
			col:            0,
		}},
	})

	assertChildViewAndDisplayNotUpdated(t, mocks)
}

func TestActionMouseSelectIsHandledAndUpdatesResultWhenClickIsOnRow(t *testing.T) {
	abstractWindowView, mocks := setupAbstractWindowView()
	mocks.viewPos.On("ViewStartRowIndex").Return(uint(0))
	mocks.viewPos.On("SetActiveRowIndex", uint(0)).Return()

	abstractWindowView.HandleAction(Action{
		ActionType: ActionMouseSelect,
		Args: []interface{}{MouseEvent{
			mouseEventType: MetLeftClick,
			row:            1,
			col:            0,
		}},
	})

	mocks.viewPos.AssertCalled(t, "SetActiveRowIndex", uint(0))
	assertChildViewAndDisplayUpdated(t, mocks)
}

func TestActionMouseScrollDownIsHandledAndNoUpdatesResultWhenScrollDownReturnsFalse(t *testing.T) {
	abstractWindowView, mocks := setupAbstractWindowView()
	mocks.viewPos.On("ScrollDown", uint(24), uint(22), uint(3)).Return(false)

	abstractWindowView.HandleAction(Action{ActionType: ActionMouseScrollDown})

	mocks.viewPos.AssertCalled(t, "ScrollDown", uint(24), uint(22), uint(3))
	assertChildViewAndDisplayNotUpdated(t, mocks)
}

func TestActionMouseScrollDownIsHandledAndUpdatesResultWhenScrollDownReturnsTrue(t *testing.T) {
	abstractWindowView, mocks := setupAbstractWindowView()
	mocks.viewPos.On("ScrollDown", uint(24), uint(22), uint(3)).Return(true)

	abstractWindowView.HandleAction(Action{ActionType: ActionMouseScrollDown})

	mocks.viewPos.AssertCalled(t, "ScrollDown", uint(24), uint(22), uint(3))
	assertChildViewAndDisplayUpdated(t, mocks)
}

func TestActionMouseScrollUpIsHandledAndNoUpdatesResultWhenScrollUpReturnsFalse(t *testing.T) {
	abstractWindowView, mocks := setupAbstractWindowView()
	mocks.viewPos.On("ScrollUp", uint(22), uint(3)).Return(false)

	abstractWindowView.HandleAction(Action{ActionType: ActionMouseScrollUp})

	mocks.viewPos.AssertCalled(t, "ScrollUp", uint(22), uint(3))
	assertChildViewAndDisplayNotUpdated(t, mocks)
}

func TestActionMouseScrollUpIsHandledAndUpdatesResultWhenScrollUpReturnsTrue(t *testing.T) {
	abstractWindowView, mocks := setupAbstractWindowView()
	mocks.viewPos.On("ScrollUp", uint(22), uint(3)).Return(true)

	abstractWindowView.HandleAction(Action{ActionType: ActionMouseScrollUp})

	mocks.viewPos.AssertCalled(t, "ScrollUp", uint(22), uint(3))
	assertChildViewAndDisplayUpdated(t, mocks)
}

func TestOnStateChangeSetsViewStateAndCallsSetVariablesWhenViewStateIsActive(t *testing.T) {
	abstractWindowView, mocks := setupAbstractWindowView()

	mocks.variables.On("SetViewVariable", VarLineNumer, "1", ViewStateActive).Return()
	mocks.variables.On("SetViewVariable", VarLineCount, "24", ViewStateActive).Return()
	mocks.variables.On("SetViewVariable", VarLineText, "", ViewStateActive).Return()

	abstractWindowView.OnStateChange(ViewStateActive)

	mocks.lock.AssertCalled(t, "Lock")
	mocks.lock.AssertCalled(t, "Unlock")
	assert.Equal(t, abstractWindowView.viewState, ViewStateActive, "viewState should be ViewStateActive")
}

func TestOnStateChangeSetsViewStateAndDoesNotCallSetVariablesWhenViewStateIsInactive(t *testing.T) {
	abstractWindowView, mocks := setupAbstractWindowView()

	abstractWindowView.OnStateChange(ViewStateInactiveAndVisible)

	mocks.lock.AssertCalled(t, "Lock")
	mocks.lock.AssertCalled(t, "Unlock")

	mocks.variables.AssertNotCalled(t, "SetViewVariable", VarLineNumer, "1", ViewStateActive)
	mocks.variables.AssertNotCalled(t, "SetViewVariable", VarLineCount, "24", ViewStateActive)
	mocks.variables.AssertNotCalled(t, "SetViewVariable", VarLineText, "", ViewStateActive)

	assert.Equal(t, abstractWindowView.viewState, ViewStateInactiveAndVisible, "viewState should be ViewStateInactiveAndVisible")
}

func TestOnStateChangeSetsViewStateAndDoesNotCallSetVariablesWhenViewStateIsInvisible(t *testing.T) {
	abstractWindowView, mocks := setupAbstractWindowView()

	abstractWindowView.OnStateChange(ViewStateInvisible)

	mocks.lock.AssertCalled(t, "Lock")
	mocks.lock.AssertCalled(t, "Unlock")

	mocks.variables.AssertNotCalled(t, "SetViewVariable", VarLineNumer, "1", ViewStateActive)
	mocks.variables.AssertNotCalled(t, "SetViewVariable", VarLineCount, "24", ViewStateActive)
	mocks.variables.AssertNotCalled(t, "SetViewVariable", VarLineText, "", ViewStateActive)

	assert.Equal(t, abstractWindowView.viewState, ViewStateInvisible, "viewState should be ViewStateInvisible")
}

func TestWhenNotifyChildRowSelectedIsCalledThenSetVariablesIsAsWell(t *testing.T) {
	abstractWindowView, mocks := setupAbstractWindowView()

	abstractWindowView.notifyChildRowSelected(0)

	mocks.variables.AssertCalled(t, "SetViewVariable", VarLineNumer, "1", ViewStateInvisible)
	mocks.variables.AssertCalled(t, "SetViewVariable", VarLineCount, "24", ViewStateInvisible)
	mocks.variables.AssertCalled(t, "SetViewVariable", VarLineText, "", ViewStateInvisible)
}

func TestLineNumberReturnsTheNumberOfRowsInTheView(t *testing.T) {
	abstractWindowView, mocks := setupAbstractWindowView()

	actualLineNumber := abstractWindowView.LineNumber()

	mocks.lock.AssertCalled(t, "Lock")
	mocks.lock.AssertCalled(t, "Unlock")

	assert.Equal(t, actualLineNumber, uint(24), "LineNumber should return 24")
}

func TestLineReturnsTheLineContentOfTheProvidedLine(t *testing.T) {
	abstractWindowView, mocks := setupAbstractWindowView()

	mocks.child.On("line", uint(10)).Return("line content")

	lineContent := abstractWindowView.Line(10)

	mocks.lock.AssertCalled(t, "Lock")
	mocks.lock.AssertCalled(t, "Unlock")

	assert.Equal(t, lineContent, "line content", "Line content does not match output from child")
}

func TestViewPosReturnsTheChildViewPos(t *testing.T) {
	abstractWindowView, mocks := setupAbstractWindowView()

	viewPos := abstractWindowView.ViewPos()

	mocks.lock.AssertCalled(t, "Lock")
	mocks.lock.AssertCalled(t, "Unlock")

	assert.Equal(t, viewPos, mocks.viewPos, "ViewPos should return child view pos")
}

func TestOnSearchMatchUpdatesTheViewPosAndNotifiesTheChild(t *testing.T) {
	abstractWindowView, mocks := setupAbstractWindowView()

	mocks.viewPos.On("SetActiveRowIndex", uint(10)).Return()
	mocks.child.On("onRowSelected", uint(10)).Return(nil)

	abstractWindowView.OnSearchMatch(mocks.viewPos, 10)

	mocks.lock.AssertCalled(t, "Lock")
	mocks.lock.AssertCalled(t, "Unlock")

	mocks.viewPos.AssertCalled(t, "SetActiveRowIndex", uint(10))
	mocks.child.AssertCalled(t, "onRowSelected", uint(10))
}

func TestOnSearchMatchDoesNotUpdateViewPosAndNotifyChildWhenViewPosHasChanged(t *testing.T) {
	abstractWindowView, mocks := setupAbstractWindowView()

	abstractWindowView.OnSearchMatch(NewViewPosition(), 10)

	mocks.lock.AssertCalled(t, "Lock")
	mocks.lock.AssertCalled(t, "Unlock")

	mocks.viewPos.AssertNotCalled(t, "SetActiveRowIndex", uint(10))
	mocks.child.AssertNotCalled(t, "onRowSelected", uint(10))
}

func TestOnSearchMatchDoesNotUpdateViewPosAndNotifyChildWhenMatchedLineIndexIsTooLarge(t *testing.T) {
	abstractWindowView, mocks := setupAbstractWindowView()

	abstractWindowView.OnSearchMatch(mocks.viewPos, 200)

	mocks.lock.AssertCalled(t, "Lock")
	mocks.lock.AssertCalled(t, "Unlock")

	mocks.viewPos.AssertNotCalled(t, "SetActiveRowIndex", uint(200))
	mocks.child.AssertNotCalled(t, "onRowSelected", uint(200))
}
