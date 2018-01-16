package main

// NewStatusView creates a new instance
func NewStatusView(repoData RepoData, channels *Channels, config Config) *ContainerView {
	gitStatusView := NewGitStatusView(repoData, channels)
	diffView := NewDiffView(repoData, channels)

	gitStatusView.RegisterGitStatusFileSelectedListener(diffView)

	statusView := NewContainerView(channels, config)
	statusView.SetTitle("Status View")
	statusView.SetOrientation(CoDynamic)
	statusView.SetViewID(ViewStatus)
	statusView.AddChildViews(gitStatusView, diffView)

	return statusView
}
