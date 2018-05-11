package main

import (
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
func NewContextMenuView(contextMenuConfig ContextMenuConfig, channels Channels, config Config) *ContextMenuView {
	contextMenuView := &ContextMenuView{
		contextMenuConfig: contextMenuConfig,
		activeViewPos:     NewViewPosition(),
		handlers: map[ActionType]contextMenuViewHandler{
			ActionSelect: selectContextMenuEntry,
		},
	}

	contextMenuView.AbstractWindowView = NewAbstractWindowView(contextMenuView, channels, config, "menu item")

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

	if err = win.SetSelectedRow(viewPos.SelectedRowIndex()+1, true); err != nil {
		return
	}

	win.DrawBorder()

	if err = win.SetTitle(CmpNone, "Select action"); err != nil {
		return
	}

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
