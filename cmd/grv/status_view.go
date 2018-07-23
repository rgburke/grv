package main

// NewStatusView creates a new instance
func NewStatusView(repoData RepoData, repoController RepoController, channels Channels, config Config, variables GRVVariableSetter) *ContainerView {
	gitStatusView := NewGitStatusView(repoData, repoController, channels, config, variables)
	diffView := NewDiffView(repoData, channels, config, variables)

	gitStatusView.RegisterGitStatusFileSelectedListener(diffView)

	statusView := NewContainerView(channels, config)
	statusView.SetTitle("Status View")
	statusView.SetOrientation(CoDynamic)
	statusView.SetViewID(ViewStatus)
	statusView.AddChildViews(gitStatusView, diffView)

	return statusView
}
