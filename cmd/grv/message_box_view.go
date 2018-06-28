package main

import (
	"strings"
	"sync"

	log "github.com/Sirupsen/logrus"
)

const (
	messageBoxRows = 7
)

// OnMessageBoxButtonSelected is called when a button is selected
type OnMessageBoxButtonSelected func(MessageBoxButton)

// MessageBoxButton represents a button
type MessageBoxButton string

// The set of buttons available
const (
	ButtonCancel MessageBoxButton = "Cancel"
	ButtonOK     MessageBoxButton = "OK"
	ButtonYes    MessageBoxButton = "Yes"
	ButtonNo     MessageBoxButton = "No"
)

// MessageBoxConfig is the configuration for the MessageBoxView
type MessageBoxConfig struct {
	Message  string
	Buttons  []MessageBoxButton
	OnSelect OnMessageBoxButtonSelected
}

type messageBoxViewHandler func(*MessageBoxView, Action) error

// MessageBoxView is a view that displays message box
type MessageBoxView struct {
	*AbstractWindowView
	messageBoxConfig    MessageBoxConfig
	activeViewPos       ViewPos
	lastViewDimension   ViewDimension
	selectedButtonIndex uint
	handlers            map[ActionType]messageBoxViewHandler
	lock                sync.Mutex
}

// NewMessageBoxView creates a new instance
func NewMessageBoxView(messageBoxConfig MessageBoxConfig, channels Channels, config Config) *MessageBoxView {
	messageBoxView := &MessageBoxView{
		messageBoxConfig: messageBoxConfig,
		activeViewPos:    NewViewPosition(),
		handlers: map[ActionType]messageBoxViewHandler{
			ActionSelect:      selectMessageBoxButton,
			ActionScrollRight: nextMessageButton,
			ActionScrollLeft:  prevMessageButton,
		},
	}

	messageBoxView.AbstractWindowView = NewAbstractWindowView(messageBoxView, channels, config, "message box row")

	return messageBoxView
}

// ViewID returns the ViewID of the message box view
func (messageBoxView *MessageBoxView) ViewID() ViewID {
	return ViewMessageBox
}

// Render generates the message box view and writes it to the provided window
func (messageBoxView *MessageBoxView) Render(win RenderWindow) (err error) {
	messageBoxView.lock.Lock()
	defer messageBoxView.lock.Unlock()

	messageBoxView.lastViewDimension = win.ViewDimensions()

	if win.Rows() < messageBoxRows {
		log.Errorf("Unable to render MessageBoxView - too few rows: %v", win.Rows())
		return
	}

	win.SetRow(2, 1, CmpNone, "  %v", messageBoxView.messageBoxConfig.Message)

	lineBuilder, err := win.LineBuilder(4, 1)
	if err != nil {
		return
	}

	lineBuilder.Append("  ")

	buttons := messageBoxView.messageBoxConfig.Buttons
	currentIndex := uint(0)

	for buttonIndex, colIndex := range messageBoxView.calculateButtonColumns(buttons) {
		button := buttons[buttonIndex]

		lineBuilder.Append(strings.Repeat(" ", int(colIndex-currentIndex)))
		lineBuilder.Append(string(button))

		currentIndex = colIndex + uint(StringWidth(string(button)))
	}

	win.ApplyStyle(CmpContextMenuContent)
	win.DrawBorderWithStyle(CmpContextMenuContent)

	return
}

func (messageBoxView *MessageBoxView) calculateButtonColumns(buttons []MessageBoxButton) (colIndexes []uint) {
	viewDimension := messageBoxView.lastViewDimension
	buttonOffset := (viewDimension.cols - 4) / (uint(len(buttons)) + 1)
	columnIndex := uint(buttonOffset)

	for _, button := range buttons {
		buttonWidth := uint(StringWidth(string(button)))

		var startIndex uint
		if buttonWidth%2 == 0 {
			startIndex = columnIndex - (buttonWidth-1)/2
		} else {
			startIndex = columnIndex - (buttonWidth / 2)
		}

		colIndexes = append(colIndexes, startIndex)
		columnIndex += buttonOffset
	}

	return
}

func (messageBoxView *MessageBoxView) viewPos() ViewPos {
	return messageBoxView.activeViewPos
}

func (messageBoxView *MessageBoxView) rows() uint {
	return messageBoxRows - 2
}

func (messageBoxView *MessageBoxView) viewDimension() ViewDimension {
	return messageBoxView.lastViewDimension
}

func (messageBoxView *MessageBoxView) onRowSelected(rowIndex uint) (err error) {
	return
}

func (messageBoxView *MessageBoxView) buttons() uint {
	return uint(len(messageBoxView.messageBoxConfig.Buttons))
}

// HandleAction handles the action if supported
func (messageBoxView *MessageBoxView) HandleAction(action Action) (err error) {
	messageBoxView.lock.Lock()
	defer messageBoxView.lock.Unlock()

	if handler, ok := messageBoxView.handlers[action.ActionType]; ok {
		log.Debugf("Action handled by MessageBoxView")
		err = handler(messageBoxView, action)
	} else {
		log.Debugf("Action not handled")
	}

	return
}

func selectMessageBoxButton(messageBoxView *MessageBoxView, action Action) (err error) {
	button := messageBoxView.messageBoxConfig.Buttons[messageBoxView.selectedButtonIndex]
	log.Debugf("Selected button: %v", button)

	messageBoxView.channels.DoAction(Action{ActionType: ActionRemoveView})

	if messageBoxView.messageBoxConfig.OnSelect != nil {
		go messageBoxView.messageBoxConfig.OnSelect(button)
	}

	return
}

func nextMessageButton(messageBoxView *MessageBoxView, action Action) (err error) {
	messageBoxView.selectedButtonIndex++

	if messageBoxView.selectedButtonIndex == messageBoxView.buttons() {
		messageBoxView.selectedButtonIndex = 0
	}

	messageBoxView.channels.UpdateDisplay()

	return
}

func prevMessageButton(messageBoxView *MessageBoxView, action Action) (err error) {
	if messageBoxView.selectedButtonIndex == 0 {
		messageBoxView.selectedButtonIndex = messageBoxView.buttons() - 1
	} else {
		messageBoxView.selectedButtonIndex--
	}

	messageBoxView.channels.UpdateDisplay()

	return
}
