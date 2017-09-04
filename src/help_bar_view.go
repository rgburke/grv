package main

import (
	log "github.com/Sirupsen/logrus"
)

// HelpBarView manages displaying help information in the help bar
type HelpBarView struct {
	rootView RootView
}

// ActionMessage groups a description to an action
type ActionMessage struct {
	action  ActionType
	message string
}

// NewHelpBarView creates a new instance of the help bar view
func NewHelpBarView(rootView RootView) *HelpBarView {
	return &HelpBarView{
		rootView: rootView,
	}
}

// Initialise does nothing
func (helpBarView *HelpBarView) Initialise() (err error) {
	return
}

// HandleKeyPress does nothing
func (helpBarView *HelpBarView) HandleKeyPress(keystring string) (err error) {
	return
}

// HandleAction does nothing
func (helpBarView *HelpBarView) HandleAction(Action) (err error) {
	return
}

// OnActiveChange does nothing
func (helpBarView *HelpBarView) OnActiveChange(active bool) {

}

// ViewID returns the help bar view ID
func (helpBarView *HelpBarView) ViewID() ViewID {
	return ViewHelpBar
}

// Render generates and writes the help view to the provided window
func (helpBarView *HelpBarView) Render(win RenderWindow) (err error) {
	log.Debug("Rendering HelpBarView")

	lineBuilder, err := win.LineBuilder(0, 1)
	if err != nil {
		return
	}

	lineBuilder.Append(" ")

	viewHierarchy := helpBarView.rootView.ActiveViewHierarchy()

	for _, view := range viewHierarchy {
		if err = view.RenderHelpBar(lineBuilder); err != nil {
			return
		}
	}

	return
}

// RenderStatusBar does nothing
func (helpBarView *HelpBarView) RenderStatusBar(lineBuilder *LineBuilder) (err error) {
	return
}

// RenderHelpBar does nothing
func (helpBarView *HelpBarView) RenderHelpBar(lineBuilder *LineBuilder) (err error) {
	return
}

// RenderKeyBindingHelp is a helper method for views to generate key binding help
func RenderKeyBindingHelp(viewID ViewID, lineBuilder *LineBuilder, actionMessages []ActionMessage) {
	for _, actionMessage := range actionMessages {
		keys := DefaultKeyBindings(actionMessage.action, viewID)

		if len(keys) == 0 {
			log.Debugf("No keys mapped for action %v", actionMessage.action)
			continue
		}

		lineBuilder.
			AppendWithStyle(CmpHelpbarviewSpecial, "%v ", keys[0]).
			AppendWithStyle(CmpHelpbarviewNormal, "%v     ", actionMessage.message)
	}
}
