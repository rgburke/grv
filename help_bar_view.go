package main

import (
	log "github.com/Sirupsen/logrus"
)

type HelpBarView struct {
	rootView RootView
}

type ActionMessage struct {
	action  ActionType
	message string
}

func NewHelpBarView(rootView RootView) *HelpBarView {
	return &HelpBarView{
		rootView: rootView,
	}
}

func (helpBarView *HelpBarView) Initialise() (err error) {
	return
}

func (helpBarView *HelpBarView) HandleKeyPress(keystring string) (err error) {
	return
}

func (helpBarView *HelpBarView) HandleAction(Action) (err error) {
	return
}

func (helpBarView *HelpBarView) OnActiveChange(active bool) {
	return
}

func (helpBarView *HelpBarView) ViewId() ViewId {
	return VIEW_HELP_BAR
}

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

func (helpBarView *HelpBarView) RenderStatusBar(lineBuilder *LineBuilder) (err error) {
	return
}

func (helpBarView *HelpBarView) RenderHelpBar(lineBuilder *LineBuilder) (err error) {
	return
}

func RenderKeyBindingHelp(viewId ViewId, lineBuilder *LineBuilder, actionMessages []ActionMessage) {
	for _, actionMessage := range actionMessages {
		keys := DefaultKeyBindings(actionMessage.action, viewId)

		if len(keys) == 0 {
			log.Debugf("No keys mapped for action %v", actionMessage.action)
			continue
		}

		lineBuilder.
			AppendWithStyle(CMP_HELPBARVIEW_SPECIAL, "%v ", keys[0]).
			AppendWithStyle(CMP_HELPBARVIEW_NORMAL, "%v     ", actionMessage.message)
	}
}
