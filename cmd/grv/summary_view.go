package main

const (
	// SummaryViewTitle is the title of the Summary View
	SummaryViewTitle = "Summary View"
)

// NewSummaryView creates a new instance
func NewSummaryView(repoData RepoData, repoController RepoController, channels Channels, config Config, variables GRVVariableSetter) *ContainerView {
	childViewContainer := NewWindowViewContainer(nil)
	gitSummaryView := NewGitSummaryView(repoData, repoController, channels, config, variables, childViewContainer)

	summaryView := NewContainerView(channels, config)
	summaryView.SetWindowStyleConfig(NewWindowStyleConfig(false, SrsUnderline))
	summaryView.SetTitle(SummaryViewTitle)
	summaryView.SetOrientation(CoDynamic)
	summaryView.SetViewID(ViewSummary)
	summaryView.AddChildViews(gitSummaryView, childViewContainer)

	return summaryView
}
