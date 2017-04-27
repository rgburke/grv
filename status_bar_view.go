package main

type StatusBarView struct {
}

func (statusBarView *StatusBarView) Initialise() (err error) {
	return
}

func (statusBarView *StatusBarView) HandleKeyPress(keystring string) (err error) {
	return
}

func (statusBarView *StatusBarView) HandleAction(Action) (err error) {
	return
}

func (statusBarView *StatusBarView) OnActiveChange(active bool) {
	return
}

func (statusBarView *StatusBarView) ViewId() ViewId {
	return VIEW_STATUS_BAR
}

func (statusBarView *StatusBarView) Render(RenderWindow) (err error) {
	return
}
