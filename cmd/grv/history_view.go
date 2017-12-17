package main

const (
	hvMaxRefViewWidth = uint(35)
)

// HistoryView manages the history view and it's child views
type HistoryView struct {
	*ContainerView
}

// NewHistoryView creates a new instance of the history view
func NewHistoryView(repoData RepoData, channels *Channels, config Config) *HistoryView {
	refView := NewRefView(repoData, channels)
	commitView := NewCommitView(repoData, channels)
	diffView := NewDiffView(repoData, channels)

	refView.RegisterRefListener(commitView)
	commitView.RegisterCommitViewListener(diffView)

	subContainer := NewContainerView(channels, config, CoHorizontal, []AbstractView{commitView, diffView})

	historyView := &HistoryView{
		ContainerView: NewContainerView(channels, config, CoVertical, []AbstractView{refView, subContainer}),
	}

	historyView.SetChildViewPositionCalculator(historyView)

	return historyView
}

// Title returns the title for the history view
func (historyView *HistoryView) Title() string {
	return "History View"
}

// ViewID returns container view id
func (historyView *HistoryView) ViewID() ViewID {
	return ViewHistory
}

// CalculateChildViewPositions calculates the child layout data for the history view
func (historyView *HistoryView) CalculateChildViewPositions(viewLayoutData *ViewLayoutData) (childPositions []*ChildViewPosition) {
	childPositions = historyView.ContainerView.CalculateChildViewPositions(viewLayoutData)

	if !viewLayoutData.fullScreen && viewLayoutData.orientation == CoVertical && len(childPositions) > 0 {
		refViewPosition := childPositions[0]

		if refViewPosition.viewDimension.cols > hvMaxRefViewWidth {
			cols := refViewPosition.viewDimension.cols - hvMaxRefViewWidth
			extraColsPerView := cols / uint(len(childPositions)-1)
			startCol := hvMaxRefViewWidth

			for i := 1; i < len(childPositions); i++ {
				childPosition := childPositions[i]
				childPosition.startCol = startCol
				childPosition.viewDimension.cols += extraColsPerView

				startCol += childPosition.viewDimension.cols
			}

			childPositions[len(childPositions)-1].viewDimension.cols += cols % uint(len(childPositions)-1)
			refViewPosition.viewDimension.cols = hvMaxRefViewWidth
		}
	}

	return
}
