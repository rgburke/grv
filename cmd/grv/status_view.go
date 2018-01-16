package main

// StatusView manages the git status views
type StatusView struct {
	*ContainerView
}

// NewStatusView creates a new instance
func NewStatusView(repoData RepoData, channels *Channels, config Config) *StatusView {
	gitStatusView := NewGitStatusView(repoData, channels)
	diffView := NewDiffView(repoData, channels)

	gitStatusView.RegisterGitStatusFileSelectedListener(diffView)

	statusView := &StatusView{ContainerView: NewContainerView(channels, config)}
	statusView.SetTitle("Status View")
	statusView.SetOrientation(CoDynamic)
	statusView.AddChildViews(gitStatusView, diffView)

	return statusView
}
