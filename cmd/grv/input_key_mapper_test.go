package main

import (
	"fmt"
	"io/ioutil"
	"reflect"
	"testing"

	log "github.com/Sirupsen/logrus"
	gc "github.com/rgburke/goncurses"
	"github.com/stretchr/testify/mock"
)

func init() {
	log.SetOutput(ioutil.Discard)
}

type MockInputUI struct {
	mock.Mock
}

func (inputUI *MockInputUI) GetInput(force bool) (Key, error) {
	args := inputUI.Called(force)
	return args.Get(0).(Key), args.Error(1)
}

func (inputUI *MockInputUI) CancelGetInput() error {
	args := inputUI.Called()
	return args.Error(0)
}

func (inputUI *MockInputUI) GetMouseEvent() (mouseEvent MouseEvent, exists bool) {
	return
}

func checkOutput(actualKey string, actualError error, expectedKey string, expectedError error, t *testing.T) {
	if actualError != expectedError {
		t.Errorf("Error did not match expected value. Expected: %v. Actual: %v", expectedError, actualError)
	} else if actualKey != expectedKey {
		t.Errorf("Key did not match expected value. Expected: %v. Actual: %v", expectedKey, actualKey)
	}
}

func TestInputKeyMapperCanProcessASCIIChar(t *testing.T) {
	inputUI := &MockInputUI{}
	inputKeyMapper := NewInputKeyMapper(inputUI)

	inputUI.On("GetInput", false).Return(Key('a'), nil)

	key, err := inputKeyMapper.GetKeyInput()
	checkOutput(key, err, "a", nil, t)
}

func TestInputKeyMapperCanProcessMultiByteUTF8Char(t *testing.T) {
	inputUI := &MockInputUI{}
	inputKeyMapper := NewInputKeyMapper(inputUI)

	inputUI.On("GetInput", false).Return(Key(0xE4), nil).Once().
		On("GetInput", false).Return(Key(0xB8), nil).Once().
		On("GetInput", false).Return(Key(0x96), nil)

	key, err := inputKeyMapper.GetKeyInput()
	checkOutput(key, err, "世", nil, t)
}

func TestInputKeyMapperDiscardsInvalidUTF8Bytes(t *testing.T) {
	inputUI := &MockInputUI{}
	inputKeyMapper := NewInputKeyMapper(inputUI)

	inputUI.On("GetInput", false).Return(Key(0xE4), nil).Once().
		On("GetInput", false).Return(Key(0xE4), nil).Once().
		On("GetInput", false).Return(Key('a'), nil).Once()

	key, err := inputKeyMapper.GetKeyInput()
	checkOutput(key, err, "a", nil, t)
}

func TestInputKeyMapperMapsNCursesKeysToStringKeys(t *testing.T) {
	inputUI := &MockInputUI{}
	inputKeyMapper := NewInputKeyMapper(inputUI)

	inputUI.On("GetInput", false).Return(Key(gc.KEY_TAB), nil).Once().
		On("GetInput", false).Return(Key(gc.KEY_RETURN), nil)

	key, err := inputKeyMapper.GetKeyInput()
	checkOutput(key, err, "<Tab>", nil, t)

	key, err = inputKeyMapper.GetKeyInput()
	checkOutput(key, err, "<Enter>", nil, t)
}

func TestInputKeyMapperMapsControlKeyComboToString(t *testing.T) {
	inputUI := &MockInputUI{}
	inputKeyMapper := NewInputKeyMapper(inputUI)

	inputUI.On("GetInput", false).Return(Key(1), nil)

	key, err := inputKeyMapper.GetKeyInput()
	checkOutput(key, err, "<C-a>", nil, t)
}

func TestInputKeyMapperMapsMetaKeyComboToString(t *testing.T) {
	inputUI := &MockInputUI{}
	inputKeyMapper := NewInputKeyMapper(inputUI)

	inputUI.On("GetInput", false).Return(Key(0x1B), nil).Once().
		On("GetInput", true).Return(Key('c'), nil)

	key, err := inputKeyMapper.GetKeyInput()
	checkOutput(key, err, "<M-c>", nil, t)
}

func TestInputKeyMapperReturnsEscapeIfNoMoreInputAvailbleAfterEscapeReturned(t *testing.T) {
	inputUI := &MockInputUI{}
	inputKeyMapper := NewInputKeyMapper(inputUI)

	inputUI.On("GetInput", false).Return(Key(0x1B), nil).Once().
		On("GetInput", true).Return(Key(0), nil)

	key, err := inputKeyMapper.GetKeyInput()
	checkOutput(key, err, "<Escape>", nil, t)
}

func TestEmptyStringIsReturnedIfNoInputAvailable(t *testing.T) {
	inputUI := &MockInputUI{}
	inputKeyMapper := NewInputKeyMapper(inputUI)

	inputUI.On("GetInput", false).Return(Key(UINoKey), nil)

	key, err := inputKeyMapper.GetKeyInput()
	checkOutput(key, err, "", nil, t)
}

func TestEmptyStringIsReturnedIfErrorReturnedFromInputUI(t *testing.T) {
	inputUI := &MockInputUI{}
	inputKeyMapper := NewInputKeyMapper(inputUI)
	expectedError := fmt.Errorf("Error")

	inputUI.On("GetInput", false).Return(Key(UINoKey), expectedError)

	key, err := inputKeyMapper.GetKeyInput()
	checkOutput(key, err, "", expectedError, t)
}

func TestTokeniseKeysBreaksDownKeys(t *testing.T) {
	keysString := "abc<grv-select>123<C-a>!<grv-nop><C-c><<C-b>><Enter>世<M-a><Tab>"
	expected := []string{
		"a", "b", "c", "<grv-select>", "1", "2", "3", "<C-a>", "!", "<", "g", "r", "v", "-", "n", "o", "p", ">", "<C-c>", "<", "<C-b>", ">", "<Enter>", "世", "<M-a>", "<Tab>",
	}

	actual := TokeniseKeys(keysString)

	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("TokeniseKeys did not produce expected output. Expected: %v, Actual: %v", expected, actual)
	}
}
