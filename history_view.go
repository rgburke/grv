package main

const (
	BRANCH_VIEW_WIDTH = 75
)

type HistoryView struct {
	refView         WindowView
	commitView      WindowView
	refViewWin      *Window
	commitViewWin   *Window
	views           []WindowView
	activeViewIndex uint
}

func NewHistoryView(repoData RepoData) *HistoryView {
	refView := NewRefView(repoData)
	commitView := NewCommitView(repoData)
	refView.RegisterRefListener(commitView)

	return &HistoryView{
		refView:       refView,
		commitView:    commitView,
		refViewWin:    NewWindow(),
		commitViewWin: NewWindow(),
		views:         []WindowView{refView, commitView},
	}
}

func (historyView *HistoryView) Initialise() (err error) {
	for _, childView := range historyView.views {
		if err = childView.Initialise(); err != nil {
			break
		}
	}

	return
}

func (historyView *HistoryView) Render(viewDimension ViewDimension) (wins []*Window, err error) {
	refViewDim := viewDimension
	refViewDim.cols = Min(BRANCH_VIEW_WIDTH, refViewDim.cols/2)

	commitViewDim := viewDimension
	commitViewDim.cols = viewDimension.cols - refViewDim.cols

	historyView.refViewWin.Resize(refViewDim)
	historyView.commitViewWin.Resize(commitViewDim)

	historyView.refViewWin.Clear()
	historyView.commitViewWin.Clear()

	if err = historyView.refView.Render(historyView.refViewWin); err != nil {
		return
	}

	if err = historyView.commitView.Render(historyView.commitViewWin); err != nil {
		return
	}

	historyView.refViewWin.SetPosition(0, 0)
	historyView.commitViewWin.SetPosition(0, refViewDim.cols)

	wins = []*Window{historyView.refViewWin, historyView.commitViewWin}
	return
}

func (historyView *HistoryView) Handle(keyPressEvent KeyPressEvent, channels HandlerChannels) error {
	view := historyView.views[historyView.activeViewIndex]
	return view.Handle(keyPressEvent, channels)
}
