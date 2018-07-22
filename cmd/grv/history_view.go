package main

const (
	hvMaxRefViewWidth = uint(35)
)

// NewHistoryView creates a new instance of the history view
func NewHistoryView(repoData RepoData, repoController RepoController, channels Channels, config Config, variables GRVVariableSetter) *ContainerView {
	refView := NewRefView(repoData, repoController, channels, config, variables)
	commitView := NewCommitView(repoData, repoController, channels, config, variables)
	diffView := NewDiffView(repoData, channels, config)

	refView.RegisterRefListener(commitView)
	commitView.RegisterCommitViewListener(diffView)

	subContainer := NewContainerView(channels, config)
	subContainer.SetOrientation(CoDynamic)
	subContainer.AddChildViews(commitView, diffView)

	historyView := NewContainerView(channels, config)
	historyView.SetTitle("History View")
	historyView.SetOrientation(CoVertical)
	historyView.SetViewID(ViewHistory)
	historyView.SetChildViewPositionCalculator(&historyViewPositionCalculator{historyView: historyView})
	historyView.AddChildViews(refView, subContainer)

	return historyView
}

type historyViewPositionCalculator struct {
	historyView *ContainerView
}

// CalculateChildViewPositions calculates the child layout data for the history view
func (calculator *historyViewPositionCalculator) CalculateChildViewPositions(viewLayoutData *ViewLayoutData) (childPositions []*ChildViewPosition) {
	childPositions = calculator.historyView.CalculateChildViewPositions(viewLayoutData)
	childPositionNum := uint(len(childPositions))

	if !viewLayoutData.fullScreen && viewLayoutData.orientation == CoVertical && childPositionNum > 0 {
		if _, isRefView := calculator.historyView.childViews[0].(*RefView); isRefView {
			refViewPosition := childPositions[0]

			if refViewPosition.viewDimension.cols > hvMaxRefViewWidth {
				if childPositionNum > 1 {
					cols := refViewPosition.viewDimension.cols - hvMaxRefViewWidth
					extraColsPerView := cols / (childPositionNum - 1)
					startCol := hvMaxRefViewWidth

					for i := 1; i < len(childPositions); i++ {
						childPosition := childPositions[i]
						childPosition.startCol = startCol
						childPosition.viewDimension.cols += extraColsPerView

						startCol += childPosition.viewDimension.cols
					}

					childPositions[childPositionNum-1].viewDimension.cols += cols % (childPositionNum - 1)
				}

				refViewPosition.viewDimension.cols = hvMaxRefViewWidth
			}
		}
	}

	return
}
