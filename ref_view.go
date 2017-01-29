package main

type RenderRefList func(*RefView, *RefList, RenderWindow, uint) (uint, error)

type RefList struct {
	name        string
	expanded    bool
	activeIndex uint
	renderer    RenderRefList
}

type RefView struct {
	repoData           RepoData
	refLists           []RefList
	activeRefListIndex uint
	refListeners       []RefListener
}

type RefListener interface {
	OnRefSelect(*Oid) error
}

func NewRefView(repoData RepoData) *RefView {
	return &RefView{
		repoData: repoData,
		refLists: []RefList{
			RefList{
				name:     "Branches",
				renderer: DrawBranches,
				expanded: true,
			},
			RefList{
				name:     "Tags",
				renderer: DrawTags,
			},
		},
	}
}

func (refView *RefView) Initialise() (err error) {
	if err = refView.repoData.LoadHead(); err != nil {
		return
	}

	if err = refView.repoData.LoadLocalRefs(); err != nil {
		return
	}

	refView.notifyRefListeners(refView.repoData.Head())

	return
}

func (refView *RefView) RegisterRefListener(refListener RefListener) {
	refView.refListeners = append(refView.refListeners, refListener)
}

func (refView *RefView) notifyRefListeners(oid *Oid) (err error) {
	for _, refListener := range refView.refListeners {
		if err = refListener.OnRefSelect(oid); err != nil {
			break
		}
	}

	return
}

func (refView *RefView) Render(win RenderWindow) (err error) {
	rowIndex := uint(1)

	for _, refList := range refView.refLists {
		if rowIndex >= win.Rows() {
			break
		}

		expandChar := "+"
		if refList.expanded {
			expandChar = "-"
		}

		if err = win.SetRow(rowIndex, " %v%s:", expandChar, refList.name); err != nil {
			break
		}

		if refList.expanded {
			rowIndex++

			if rowIndex >= win.Rows() {
				break
			} else if rowIndex, err = refList.renderer(refView, &refList, win, rowIndex); err != nil {
				break
			}
		}

		rowIndex++
	}

	win.DrawBorder()

	return
}

func DrawBranches(refView *RefView, refList *RefList, win RenderWindow, rowIndex uint) (uint, error) {
	startRowIndex := rowIndex
	branches := refView.repoData.LocalBranches()

	for rowIndex < win.Rows() && rowIndex-startRowIndex < uint(len(branches)) {
		win.SetRow(rowIndex, "   %s", branches[rowIndex-startRowIndex].name)
		rowIndex++
	}

	return rowIndex, nil
}

func DrawTags(refView *RefView, refList *RefList, win RenderWindow, rowIndex uint) (uint, error) {
	startRowIndex := rowIndex
	tags := refView.repoData.LocalTags()

	for rowIndex < win.Rows() && rowIndex-startRowIndex < uint(len(tags)) {
		win.SetRow(rowIndex, "   %s", tags[rowIndex-startRowIndex].tag.Name())
		rowIndex++
	}

	return rowIndex, nil
}

func (refView *RefView) Handle(keyPressEvent KeyPressEvent) (err error) {
	return
}
