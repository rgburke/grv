package main

import (
	"fmt"
	"regexp"
	"strings"
	"sync"

	log "github.com/Sirupsen/logrus"
)

const (
	messageBoxRows             = 5
	messageBoxMaxCols          = 60
	messageBoxColPadding       = 4
	messageBoxMaxWritableWidth = messageBoxMaxCols - messageBoxColPadding
	messageBoxNonMessageRows   = 4
)

var whiteSpaceRegexp = regexp.MustCompile(`\s`)

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
	Title    string
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
func NewMessageBoxView(messageBoxConfig MessageBoxConfig, channels Channels, config Config, variables GRVVariableSetter) *MessageBoxView {
	messageBoxView := &MessageBoxView{
		messageBoxConfig: messageBoxConfig,
		activeViewPos:    NewViewPosition(),
		handlers: map[ActionType]messageBoxViewHandler{
			ActionSelect:      chooseMessageBoxButton,
			ActionNextButton:  selectNextMessageButton,
			ActionPrevButton:  selectPrevMessageButton,
			ActionMouseSelect: mouseSelectButton,
		},
	}

	messageBoxView.AbstractWindowView = NewAbstractWindowView(messageBoxView, channels, config, variables, &messageBoxView.lock, "message box row")

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
	buttonsWidth := messageBoxView.buttonsWidth()

	if win.Rows() < messageBoxRows || win.Cols() < buttonsWidth {
		log.Errorf("Unable to render MessageBoxView - too few rows and/or columns: %v", win.ViewDimensions())
		return
	}

	messageRows := win.Rows() - messageBoxNonMessageRows
	message := messageBoxView.messageBoxConfig.Message
	words := whiteSpaceRegexp.Split(message, -1)
	rowCols := win.Cols() - messageBoxColPadding

	rowIndex := uint(1)
	colIndex := uint(0)

	lineBuilder, err := win.LineBuilder(rowIndex, 1)
	if err != nil {
		return
	}

	lineBuilder.Append("  ")

OuterLoop:
	for _, word := range words {
		wordWidth := uint(StringWidth(word))

		if colIndex+wordWidth > rowCols {
			remainingWordWidth := wordWidth
			runeStartIndex := uint(0)
			runes := []rune(word)

			for {
				rowIndex++
				colIndex = 0

				if rowIndex-1 >= messageRows {
					lineBuilder.Append("...")
					break OuterLoop
				}

				lineBuilder, err = win.LineBuilder(rowIndex, 1)
				if err != nil {
					return
				}

				lineBuilder.Append("  ")
				endRuneIndex := runeStartIndex + MinUInt(rowCols, uint(len(runes))-runeStartIndex)
				subWord := string(runes[runeStartIndex:endRuneIndex])

				lineBuilder.Append("%v ", subWord)

				runeStartIndex = endRuneIndex
				subWordWidth := uint(StringWidth(subWord))
				remainingWordWidth -= subWordWidth
				colIndex += subWordWidth

				if remainingWordWidth == 0 || runeStartIndex >= uint(len(runes)) {
					break
				}
			}
		} else {
			lineBuilder.Append("%v ", word)
			colIndex += wordWidth + 1
		}
	}

	win.ApplyStyle(CmpMessageBoxContent)

	buttonsRowIndex := win.Rows() - 2

	lineBuilder, err = win.LineBuilder(buttonsRowIndex, 1)
	if err != nil {
		return
	}

	lineBuilder.AppendWithStyle(CmpMessageBoxContent, "  ")
	paddingSize := rowCols - buttonsWidth
	lineBuilder.AppendWithStyle(CmpMessageBoxContent, "%v", strings.Repeat(" ", int(paddingSize)))

	for buttonIndex, button := range messageBoxView.messageBoxConfig.Buttons {
		themeComponentID := CmpMessageBoxContent
		if uint(buttonIndex) == messageBoxView.selectedButtonIndex {
			themeComponentID = CmpMessageBoxSelectedButton
		}

		lineBuilder.AppendWithStyle(themeComponentID, "<%v>", button)
		lineBuilder.AppendWithStyle(CmpMessageBoxContent, " ")
	}

	win.DrawBorderWithStyle(CmpMessageBoxContent)

	if messageBoxView.messageBoxConfig.Title != "" {
		if err = win.SetTitle(CmpMessageBoxTitle, "%v", messageBoxView.messageBoxConfig.Title); err != nil {
			return
		}
	}

	return
}

// RenderHelpBar renders a help message for the message box view
func (messageBoxView *MessageBoxView) RenderHelpBar(lineBuilder *LineBuilder) (err error) {
	quitKeys := messageBoxView.config.KeyStrings(ActionRemoveView, ViewHierarchy{ViewMessageBox, ViewAll})

	if len(quitKeys) > 0 {
		quitKeyText := fmt.Sprintf("Press %v to close the message box", quitKeys[len(quitKeys)-1].keystring)
		lineBuilder.AppendWithStyle(CmpHelpbarviewSpecial, " %v", quitKeyText)
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

func (messageBoxView *MessageBoxView) line(lineIndex uint) (line string) {
	return
}

// ViewDimension calculates the view dimensions required to display
// the message box correctly
func (messageBoxView *MessageBoxView) ViewDimension() ViewDimension {
	messageBoxView.lock.Lock()
	defer messageBoxView.lock.Unlock()

	message := messageBoxView.messageBoxConfig.Message
	words := whiteSpaceRegexp.Split(message, -1)

	requiredRows := uint(1)
	requiredCols := uint(0)

	for _, word := range words {
		wordWidth := uint(StringWidth(word)) + 1

		if requiredCols+wordWidth > messageBoxMaxWritableWidth {
			if wordWidth > messageBoxMaxWritableWidth {
				requiredRows += (wordWidth / messageBoxMaxWritableWidth) + 1
				requiredCols = wordWidth % messageBoxMaxWritableWidth
			} else {
				requiredRows++
				requiredCols = wordWidth
			}
		} else {
			requiredCols += wordWidth
		}
	}

	if requiredRows > 1 {
		requiredCols = messageBoxMaxCols
	} else {
		requiredCols += messageBoxColPadding
		requiredCols--
	}

	buttonsRequiredCols := messageBoxView.buttonsWidth() + messageBoxColPadding
	requiredCols = MaxUInt(requiredCols, buttonsRequiredCols)

	requiredRows += messageBoxNonMessageRows

	return ViewDimension{
		rows: requiredRows,
		cols: requiredCols,
	}
}

func (messageBoxView *MessageBoxView) buttonsWidth() (buttonsWidth uint) {
	for _, button := range messageBoxView.messageBoxConfig.Buttons {
		buttonsWidth += uint(StringWidth(string(button))) + 3
	}

	if buttonsWidth > 0 {
		buttonsWidth--
	}

	return
}

func (messageBoxView *MessageBoxView) selectButton(button MessageBoxButton) {
	log.Debugf("Selected button: %v", button)
	messageBoxView.channels.DoAction(Action{ActionType: ActionRemoveView})

	if messageBoxView.messageBoxConfig.OnSelect != nil {
		go messageBoxView.messageBoxConfig.OnSelect(button)
	}
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

func chooseMessageBoxButton(messageBoxView *MessageBoxView, action Action) (err error) {
	button := messageBoxView.messageBoxConfig.Buttons[messageBoxView.selectedButtonIndex]
	messageBoxView.selectButton(button)
	return
}

func selectNextMessageButton(messageBoxView *MessageBoxView, action Action) (err error) {
	messageBoxView.selectedButtonIndex++

	if messageBoxView.selectedButtonIndex == messageBoxView.buttons() {
		messageBoxView.selectedButtonIndex = 0
	}

	messageBoxView.channels.UpdateDisplay()

	return
}

func selectPrevMessageButton(messageBoxView *MessageBoxView, action Action) (err error) {
	if messageBoxView.selectedButtonIndex == 0 {
		messageBoxView.selectedButtonIndex = messageBoxView.buttons() - 1
	} else {
		messageBoxView.selectedButtonIndex--
	}

	messageBoxView.channels.UpdateDisplay()

	return
}

func mouseSelectButton(messageBoxView *MessageBoxView, action Action) (err error) {
	mouseEvent, err := GetMouseEventFromAction(action)
	if err != nil {
		return
	}

	if mouseEvent.row != messageBoxView.lastViewDimension.rows-2 {
		return
	}

	buttonsWidth := messageBoxView.buttonsWidth()
	startColumn := messageBoxView.lastViewDimension.cols - (buttonsWidth + 2)

	if mouseEvent.col < startColumn {
		return
	}

	currentColumn := startColumn

	for _, button := range messageBoxView.messageBoxConfig.Buttons {
		buttonLength := uint(len(button)) + 2
		if mouseEvent.col >= currentColumn && mouseEvent.col < currentColumn+buttonLength {
			log.Debugf("Mouse click on button: %v", button)
			messageBoxView.selectButton(button)
			break
		}

		currentColumn += buttonLength + 1
	}

	return
}
