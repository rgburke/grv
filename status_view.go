package main

type StatusView struct {
}

func NewStatusView() *StatusView {
	return &StatusView{}
}

func (statusView *StatusView) Initialise() (err error) {
	return
}

func (statusView *StatusView) HandleKeyPress(keystring string) (err error) {
	return
}

func (statusView *StatusView) HandleAction(Action) (err error) {
	return
}

func (statusView *StatusView) OnActiveChange(active bool) {

}

func (statusView *StatusView) ViewId() ViewId {
	return VIEW_STATUS
}

func (statusView *StatusView) Render(ViewDimension) (wins []*Window, err error) {
	return
}

func (statusView *StatusView) ActiveViewHierarchy() (viewIds []ViewId) {
	return
}
