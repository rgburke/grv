package main

type HelpBarView struct {
	rootView RootView
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

}

func (helpBarView *HelpBarView) ViewId() ViewId {
	return VIEW_HELP_BAR
}

func (helpBarView *HelpBarView) Render(RenderWindow) (err error) {
	return
}

func (helpBarView *HelpBarView) RenderStatusBar(RenderWindow) (err error) {
	return
}

func (helpBarView *HelpBarView) RenderHelpBar(RenderWindow) (err error) {
	return
}
