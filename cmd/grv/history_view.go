package main

const (
	hvMaxRefViewWidth = uint(35)
	hvMenu            = "HistoryViewMenu"
	hvRemoteViewRows  = uint(4)
)

// NewHistoryView creates a new instance of the history view
func NewHistoryView(repoData RepoData, repoController RepoController, channels Channels, config Config, variables GRVVariableSetter) *ContainerView {
	refView := NewRefView(repoData, repoController, channels, config, variables)
	remoteView := NewRemoteView(repoData, repoController, channels, config, variables)
	commitView := NewCommitView(repoData, repoController, channels, config, variables)
	diffView := NewDiffView(repoData, channels, config, variables)

	refView.RegisterRefListener(commitView)
	commitView.RegisterCommitViewListener(diffView)

	leftContainer := NewContainerView(channels, config)
	leftContainer.SetTitle(hvMenu)
	leftContainer.SetOrientation(CoHorizontal)
	leftContainer.AddChildViews(refView, remoteView)
	leftContainer.SetChildViewPositionCalculator(&historyViewMenuPositionCalculator{menuView: leftContainer})

	rightContainer := NewContainerView(channels, config)
	rightContainer.SetOrientation(CoVertical)
	rightContainer.AddChildViews(commitView, diffView)

	historyView := NewContainerView(channels, config)
	historyView.SetTitle("History View")
	historyView.SetOrientation(CoVertical)
	historyView.SetViewID(ViewHistory)
	historyView.SetChildViewPositionCalculator(&historyViewPositionCalculator{historyView: historyView})
	historyView.AddChildViews(leftContainer, rightContainer)

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
		if containerView, isContainerView := calculator.historyView.childViews[0].(*ContainerView); isContainerView && containerView.Title() == hvMenu {
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

type historyViewMenuPositionCalculator struct {
	menuView *ContainerView
}

// CalculateChildViewPositions calculates the child layout data for the history view menu
func (calculator *historyViewMenuPositionCalculator) CalculateChildViewPositions(viewLayoutData *ViewLayoutData) (childPositions []*ChildViewPosition) {
	childPositions = calculator.menuView.CalculateChildViewPositions(viewLayoutData)
	childPositionNum := uint(len(childPositions))

	if !viewLayoutData.fullScreen && viewLayoutData.orientation == CoHorizontal && childPositionNum == 2 {
		_, isRefView := calculator.menuView.childViews[0].(*RefView)
		_, isRemoteView := calculator.menuView.childViews[1].(*RemoteView)

		if isRefView && isRemoteView {
			refViewPosition := childPositions[0]
			remoteViewPosition := childPositions[1]

			offset := remoteViewPosition.viewDimension.rows - hvRemoteViewRows
			remoteViewPosition.viewDimension.rows = hvRemoteViewRows
			remoteViewPosition.startRow += offset
			refViewPosition.viewDimension.rows += offset
		}
	}

	return
}
