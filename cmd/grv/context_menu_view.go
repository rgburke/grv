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
	Entity   string
	Entries  []ContextMenuEntry
	OnSelect OnContextMenuEntrySelected
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
			ActionSelect: selectContextMenuEntry,
		},
	}

	contextMenuView.AbstractWindowView = NewAbstractWindowView(contextMenuView, channels, config, variables, &contextMenuView.lock, "menu item")

	if contextMenuConfig.Entity == "" {
		contextMenuView.contextMenuConfig.Entity = "Action"
	}

	return contextMenuView
}

// ViewID returns the ViewID of the context menu view
func (contextMenuView *ContextMenuView) ViewID() ViewID {
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

	for rowIndex := uint(0); rowIndex < winRows && viewRowIndex < viewRows; rowIndex++ {
		entry := contextMenuView.contextMenuConfig.Entries[viewRowIndex]

		if err = win.SetRow(rowIndex+1, startColumn, CmpNone, " %v", entry.DisplayName); err != nil {
			return
		}

		viewRowIndex++
	}

	win.ApplyStyle(CmpContextMenuContent)

	if err = win.SetSelectedRow(viewPos.SelectedRowIndex()+1, true); err != nil {
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
	} else {
		log.Debugf("Action not handled")
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
