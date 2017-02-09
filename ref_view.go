package main

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	gc "github.com/rthornton128/goncurses"
)

type RenderedRefType int

const (
	RV_BRANCH_GROUP RenderedRefType = iota
	RV_BRANCH
	RV_TAG_GROUP
	RV_TAG
	RV_SPACE
)

type RenderedRefGenerator func(*RefView, *RefList, *[]RenderedRef)

type RefList struct {
	name            string
	expanded        bool
	renderer        RenderedRefGenerator
	renderedRefType RenderedRefType
}

type RenderedRef struct {
	value           string
	oid             *Oid
	renderedRefType RenderedRefType
}

type RefView struct {
	repoData       RepoData
	refLists       []*RefList
	refListeners   []RefListener
	active         bool
	renderedRefs   []RenderedRef
	activeIndex    uint
	viewStartIndex uint
}

type RefListener interface {
	OnRefSelect(*Oid) error
}

func NewRefView(repoData RepoData) *RefView {
	return &RefView{
		repoData: repoData,
		refLists: []*RefList{
			&RefList{
				name:            "Branches",
				renderer:        GenerateBranches,
				expanded:        true,
				renderedRefType: RV_BRANCH_GROUP,
			},
			&RefList{
				name:            "Tags",
				renderer:        GenerateTags,
				renderedRefType: RV_TAG_GROUP,
			},
		},
	}
}

func (refView *RefView) Initialise() (err error) {
	log.Info("Initialising RefView")

	if err = refView.repoData.LoadHead(); err != nil {
		return
	}

	if err = refView.repoData.LoadLocalRefs(); err != nil {
		return
	}

	refView.GenerateRenderedRefs()

	head := refView.repoData.Head()
	branches := refView.repoData.LocalBranches()
	refView.activeIndex = 1

	for _, branch := range branches {
		if branch.oid.oid.Equal(head.oid) {
			break
		}

		refView.activeIndex++
	}

	refView.notifyRefListeners(refView.repoData.Head())

	return
}

func (refView *RefView) RegisterRefListener(refListener RefListener) {
	refView.refListeners = append(refView.refListeners, refListener)
}

func (refView *RefView) notifyRefListeners(oid *Oid) (err error) {
	log.Debugf("Notifying RefListeners of selected oid %v", oid)

	for _, refListener := range refView.refListeners {
		if err = refListener.OnRefSelect(oid); err != nil {
			break
		}
	}

	return
}

func (refView *RefView) Render(win RenderWindow) (err error) {
	log.Debug("Rendering RefView")

	rows := win.Rows() - 2

	if refView.viewStartIndex > refView.activeIndex {
		refView.viewStartIndex = refView.activeIndex
	} else if rowDiff := refView.activeIndex - refView.viewStartIndex; rowDiff >= rows {
		refView.viewStartIndex += (rowDiff - rows) + 1
	}

	refIndex := refView.viewStartIndex

	for winRowIndex := uint(0); winRowIndex < rows && refIndex < uint(len(refView.renderedRefs)); winRowIndex++ {
		if err = win.SetRow(winRowIndex+1, "%v", refView.renderedRefs[refIndex].value); err != nil {
			return
		}

		refIndex++
	}

	if err = win.SetSelectedRow((refView.activeIndex-refView.viewStartIndex)+1, refView.active); err != nil {
		return
	}

	win.DrawBorder()

	return
}

func (refView *RefView) GenerateRenderedRefs() {
	log.Debug("Generating Rendered Refs")
	var renderedRefs []RenderedRef

	for refIndex, refList := range refView.refLists {
		expandChar := "+"
		if refList.expanded {
			expandChar = "-"
		}

		renderedRefs = append(renderedRefs, RenderedRef{
			value:           fmt.Sprintf("  %v%v", expandChar, refList.name),
			renderedRefType: refList.renderedRefType,
		})

		if refList.expanded {
			refList.renderer(refView, refList, &renderedRefs)
		}

		if refIndex != len(refView.refLists)-1 {
			renderedRefs = append(renderedRefs, RenderedRef{
				value:           "",
				renderedRefType: RV_SPACE,
			})
		}
	}

	refView.renderedRefs = renderedRefs
}

func GenerateBranches(refView *RefView, refList *RefList, renderedRefs *[]RenderedRef) {
	branches := refView.repoData.LocalBranches()

	for _, branch := range branches {
		*renderedRefs = append(*renderedRefs, RenderedRef{
			value:           fmt.Sprintf("   %s", branch.name),
			oid:             branch.oid,
			renderedRefType: RV_BRANCH,
		})
	}
}

func GenerateTags(refView *RefView, refList *RefList, renderedRefs *[]RenderedRef) {
	tags := refView.repoData.LocalTags()

	for _, tag := range tags {
		*renderedRefs = append(*renderedRefs, RenderedRef{
			value:           fmt.Sprintf("   %s", tag.tag.Name()),
			oid:             tag.oid,
			renderedRefType: RV_TAG,
		})
	}
}

func (refView *RefView) Handle(keyPressEvent KeyPressEvent, channels HandlerChannels) (err error) {
	log.Debugf("RefView handling key %v", keyPressEvent)

	switch keyPressEvent.key {
	case gc.KEY_UP:
		if refView.activeIndex == 0 {
			return
		}

		refView.activeIndex--

		for refView.renderedRefs[refView.activeIndex].renderedRefType == RV_SPACE && refView.activeIndex > 0 {
			refView.activeIndex--
		}

		channels.displayCh <- true
	case gc.KEY_DOWN:
		indexLimit := uint(len(refView.renderedRefs)) - 1

		if refView.activeIndex >= indexLimit {
			return
		}

		refView.activeIndex++

		for refView.renderedRefs[refView.activeIndex].renderedRefType == RV_SPACE && refView.activeIndex < indexLimit {
			refView.activeIndex++
		}

		channels.displayCh <- true
	}

	return
}

func (refView *RefView) OnActiveChange(active bool) {
	log.Debugf("RefView active %v", active)
	refView.active = active
}
