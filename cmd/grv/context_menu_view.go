package main

import (
	"fmt"
	"strings"
	"sync"

	log "github.com/Sirupsen/logrus"
)

// OnContextMenuEntrySelected is called when an entry is selected
type OnContextMenuEntrySelected func(entry ContextMenuEntry, entryIndex uint)

// ContextMenuEntry stores the display name and value of
// a context menu item
type ContextMenuEntry struct {
	DisplayName string
	Value       interface{}
}

// ContextMenuConfig is the configuration for the ContextMenuView
// It contains all context menu contextMenuConfig and an observer
type ContextMenuConfig struct {
	Entity     string
	Entries    []ContextMenuEntry
	OnSelect   OnContextMenuEntrySelected
	ActionView ViewID
}

type contextMenuViewHandler func(*ContextMenuView, Action) error

// ContextMenuView is a view that displays a context menu
type ContextMenuView struct {
	*AbstractWindowView
	contextMenuConfig ContextMenuConfig
	activeViewPos     ViewPos
	lastViewDimension ViewDimension
	handlers          map[ActionType]contextMenuViewHandler
	lock              sync.Mutex
}

// NewContextMenuView creates a new instance
func NewContextMenuView(contextMenuConfig ContextMenuConfig, channels Channels, config Config, variables GRVVariableSetter) *ContextMenuView {
	contextMenuView := &ContextMenuView{
		contextMenuConfig: contextMenuConfig,
		activeViewPos:     NewViewPosition(),
		handlers: map[ActionType]contextMenuViewHandler{
			ActionSelect:   selectContextMenuEntry,
			ActionPrevLine: moveUpContextEntry,
			ActionNextLine: moveDownContextEntry,
		},
	}

	contextMenuView.AbstractWindowView = NewAbstractWindowView(contextMenuView, channels, config, variables, &contextMenuView.lock, "menu item")
	contextMenuView.processConfig()

	return contextMenuView
}

func (contextMenuView *ContextMenuView) processConfig() {
	contextMenuConfig := &contextMenuView.contextMenuConfig

	if contextMenuConfig.Entity == "" {
		contextMenuConfig.Entity = "Action"
	}

	if contextMenuConfig.ActionView != ViewAll {
		var keys []string
		maxKeyWidth := 0

		for i := uint(0); i < contextMenuView.rows(); i++ {
			entry := &contextMenuConfig.Entries[i]
			var key string

			if action, ok := entry.Value.(Action); ok {
				mappings := contextMenuView.config.KeyStrings(action.ActionType, ViewHierarchy{contextMenuConfig.ActionView})

				if len(mappings) > 0 {
					key = mappings[len(mappings)-1].keystring
				}
			}

			if key == "" {
				key = "None"
			}

			keys = append(keys, key)

			width := StringWidth(key)
			if width > maxKeyWidth {
				maxKeyWidth = width
			}
		}

		for keyIndex, key := range keys {
			if width := StringWidth(key); width < maxKeyWidth {
				key = fmt.Sprintf("%v%v", key, strings.Repeat(" ", maxKeyWidth-width))
			}

			entry := &contextMenuConfig.Entries[keyIndex]
			entry.DisplayName = fmt.Sprintf("%v  %v", key, entry.DisplayName)
		}
	}

	return
}

// ViewID returns the ViewID of the context menu view
func (contextMenuView *ContextMenuView) ViewID() ViewID {
	if contextMenuView.contextMenuConfig.ActionView != ViewAll {
		return contextMenuView.contextMenuConfig.ActionView
	}

	return ViewContextMenu
}

// Render generates the context menu view and writes it to the provided window
func (contextMenuView *ContextMenuView) Render(win RenderWindow) (err error) {
	contextMenuView.lock.Lock()
	defer contextMenuView.lock.Unlock()

	contextMenuView.lastViewDimension = win.ViewDimensions()

	winRows := win.Rows() - 2
	viewPos := contextMenuView.viewPos()

	viewRows := contextMenuView.rows()
	viewPos.DetermineViewStartRow(winRows, viewRows)

	viewRowIndex := viewPos.ViewStartRowIndex()
	startColumn := viewPos.ViewStartColumn()

	win.ApplyStyle(CmpContextMenuContent)

	var lineBuilder *LineBuilder
	for rowIndex := uint(0); rowIndex < winRows && viewRowIndex < viewRows; rowIndex++ {
		if lineBuilder, err = win.LineBuilder(rowIndex+1, startColumn); err != nil {
			return
		}

		entry := contextMenuView.contextMenuConfig.Entries[viewRowIndex]
		if contextMenuView.contextMenuConfig.ActionView != ViewAll {
			if displayParts := strings.SplitN(entry.DisplayName, " ", 2); len(displayParts) == 2 {
				lineBuilder.AppendWithStyle(CmpContextMenuKeyMapping, " %v ", displayParts[0]).
					AppendWithStyle(CmpContextMenuContent, "%v", displayParts[1])
			} else {
				return fmt.Errorf(`Expected entry of format "key  DisplayText" but found %v`, entry.DisplayName)
			}
		} else {
			lineBuilder.AppendWithStyle(CmpContextMenuContent, " %v", entry.DisplayName)
		}

		viewRowIndex++
	}

	if err = win.SetSelectedRow(viewPos.SelectedRowIndex()+1, ViewStateActive); err != nil {
		return
	}

	win.DrawBorderWithStyle(CmpContextMenuContent)
	entity := contextMenuView.contextMenuConfig.Entity

	if err = win.SetTitle(CmpContextMenuTitle, "Select %v", entity); err != nil {
		return
	}

	if err = win.SetFooter(CmpContextMenuTitle, "%v %v of %v", entity, viewPos.SelectedRowIndex()+1, viewRows); err != nil {
		return
	}

	return
}

// RenderHelpBar renders a help message for the context menu
func (contextMenuView *ContextMenuView) RenderHelpBar(lineBuilder *LineBuilder) (err error) {
	var quitKeyText string

	quitKeys := contextMenuView.config.KeyStrings(ActionRemoveView, ViewHierarchy{ViewContextMenu, ViewAll})
	if len(quitKeys) > 0 {
		quitKeyText = fmt.Sprintf("(Press %v to close menu)", quitKeys[len(quitKeys)-1].keystring)
	}

	entity := strings.ToLower(contextMenuView.contextMenuConfig.Entity)
	lineBuilder.AppendWithStyle(CmpHelpbarviewSpecial, "Select %v %v", entity, quitKeyText)
	return
}

func (contextMenuView *ContextMenuView) viewPos() ViewPos {
	return contextMenuView.activeViewPos
}

func (contextMenuView *ContextMenuView) rows() uint {
	return uint(len(contextMenuView.contextMenuConfig.Entries))
}

func (contextMenuView *ContextMenuView) viewDimension() ViewDimension {
	return contextMenuView.lastViewDimension
}

func (contextMenuView *ContextMenuView) onRowSelected(rowIndex uint) (err error) {
	return
}

func (contextMenuView *ContextMenuView) line(lineIndex uint) (line string) {
	if lineIndex < contextMenuView.rows() {
		line = contextMenuView.contextMenuConfig.Entries[lineIndex].DisplayName
	}

	return
}

// HandleAction handles the action if supported
func (contextMenuView *ContextMenuView) HandleAction(action Action) (err error) {
	contextMenuView.lock.Lock()
	defer contextMenuView.lock.Unlock()

	var handled bool
	if handler, ok := contextMenuView.handlers[action.ActionType]; ok {
		log.Debugf("Action handled by ContextMenuView")
		err = handler(contextMenuView, action)
	} else if handled, err = contextMenuView.AbstractWindowView.HandleAction(action); handled {
		log.Debugf("Action handled by AbstractWindowView")
	} else if handled, err = contextMenuView.handleMenuAction(action); handled {
		log.Debugf("Menu action handled")
	} else {
		log.Debugf("Action not handled")
	}

	return
}

func (contextMenuView *ContextMenuView) handleMenuAction(action Action) (handled bool, err error) {
	if contextMenuView.contextMenuConfig.ActionView == ViewAll {
		return
	}

	for entryIndex, entry := range contextMenuView.contextMenuConfig.Entries {
		if entryAction, ok := entry.Value.(Action); ok && entryAction.ActionType == action.ActionType {
			contextMenuView.viewPos().SetActiveRowIndex(uint(entryIndex))
			err = selectContextMenuEntry(contextMenuView, Action{ActionType: ActionSelect})
			handled = true
			break
		}
	}

	return
}

func selectContextMenuEntry(contextMenuView *ContextMenuView, action Action) (err error) {
	viewPos := contextMenuView.viewPos()
	selectedIndex := viewPos.ActiveRowIndex()
	selectedEntry := contextMenuView.contextMenuConfig.Entries[selectedIndex]

	log.Debugf("Context menu entry selected: %v", selectedEntry)

	contextMenuView.channels.DoAction(Action{ActionType: ActionRemoveView})

	if contextMenuView.contextMenuConfig.OnSelect != nil {
		go contextMenuView.contextMenuConfig.OnSelect(selectedEntry, selectedIndex)
	}

	return
}

func moveUpContextEntry(contextMenuView *ContextMenuView, action Action) (err error) {
	viewPos := contextMenuView.viewPos()

	if viewPos.ActiveRowIndex() == 0 {
		_, err = contextMenuView.AbstractWindowView.HandleAction(Action{ActionType: ActionLastLine})
	} else {
		_, err = contextMenuView.AbstractWindowView.HandleAction(action)
	}

	return
}

func moveDownContextEntry(contextMenuView *ContextMenuView, action Action) (err error) {
	rows := contextMenuView.rows()
	viewPos := contextMenuView.viewPos()

	if viewPos.ActiveRowIndex() == rows-1 {
		_, err = contextMenuView.AbstractWindowView.HandleAction(Action{ActionType: ActionFirstLine})
	} else {
		_, err = contextMenuView.AbstractWindowView.HandleAction(action)
	}

	return
}
