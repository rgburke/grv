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

	childViews := []AbstractView{
		gitStatusView,
		diffView,
	}

	return &StatusView{
		ContainerView: NewContainerView(channels, config, CoVertical, childViews),
	}
}

// Title returns the title of the status view
func (statusView *StatusView) Title() string {
	return "Status View"
}
