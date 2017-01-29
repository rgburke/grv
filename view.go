package main

type WindowView interface {
	Initialise() error
	Render(RenderWindow) error
	Handle(KeyPressEvent) error
}

type WindowViewCollection interface {
	Initialise() error
	Render(ViewDimension) ([]*Window, error)
	Handle(KeyPressEvent) error
}

type ViewDimension struct {
	rows uint
	cols uint
}

type View struct {
	views           []WindowViewCollection
	activeViewIndex uint
}

func NewView(repoData RepoData) (view *View) {
	view = &View{}
	view.views = []WindowViewCollection{
		NewHistoryView(repoData),
	}

	return
}

func (view *View) Initialise() (err error) {
	for _, childView := range view.views {
		if err = childView.Initialise(); err != nil {
			break
		}
	}

	return
}

func (view *View) Render(viewDimension ViewDimension) ([]*Window, error) {
	return view.views[view.activeViewIndex].Render(viewDimension)
}

func (view *View) Handle(keyPressEvent KeyPressEvent) error {
	return view.views[view.activeViewIndex].Handle(keyPressEvent)
}
