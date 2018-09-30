package main

const (
	// SummaryViewTitle is the title of the Summary View
	SummaryViewTitle            = "Summary View"
	svVerticalSplitOffsetFactor = 0.15
)

// NewSummaryView creates a new instance
func NewSummaryView(repoData RepoData, repoController RepoController, channels Channels, config Config, variables GRVVariableSetter) *ContainerView {
	childViewContainer := NewWindowViewContainer(nil)
	gitSummaryView := NewGitSummaryView(repoData, repoController, channels, config, variables, childViewContainer)

	summaryView := NewContainerView(channels, config)
	summaryView.SetWindowStyleConfig(NewWindowStyleConfig(false, SrsUnderline))
	summaryView.SetChildViewPositionCalculator(&summaryViewPositionCalculator{summaryView})
	summaryView.SetTitle(SummaryViewTitle)
	summaryView.SetOrientation(CoDynamic)
	summaryView.SetViewID(ViewSummary)
	summaryView.AddChildViews(gitSummaryView, childViewContainer)

	return summaryView
}

type summaryViewPositionCalculator struct {
	summaryView *ContainerView
}

// CalculateChildViewPositions calculates the child layout data for the summary view
func (calculator *summaryViewPositionCalculator) CalculateChildViewPositions(viewLayoutData *ViewLayoutData) (childPositions []*ChildViewPosition) {
	childPositions = calculator.summaryView.CalculateChildViewPositions(viewLayoutData)
	childPositionNum := uint(len(childPositions))

	if !viewLayoutData.fullScreen && viewLayoutData.orientation == CoVertical && childPositionNum == 2 {
		offset := uint(float64(childPositions[0].viewDimension.cols) * svVerticalSplitOffsetFactor)

		childPositions[0].viewDimension.cols -= offset
		childPositions[1].viewDimension.cols += offset
		childPositions[1].startCol -= offset
	}

	return
}
