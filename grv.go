package main

type GRV struct {
	repoData *RepositoryData
	view     *View
	ui       UI
}

func NewGRV() *GRV {
	repoDataLoader := NewRepoDataLoader()
	repoData := NewRepositoryData(repoDataLoader)

	return &GRV{
		repoData: repoData,
		view:     NewView(repoData),
		ui:       NewNcursesDisplay(),
	}
}

func (grv *GRV) Initialise(repoPath string) (err error) {
	if err = grv.repoData.Initialise(repoPath); err != nil {
		return
	}

	if err = grv.ui.Initialise(); err != nil {
		return
	}

	if err = grv.view.Initialise(); err != nil {
		return
	}

	return
}

func (grv *GRV) Free() {
	grv.repoData.Free()
	grv.ui.End()
}
