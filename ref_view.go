package main

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	gc "github.com/rthornton128/goncurses"
	"strings"
	"sync"
)

type RefViewHandler func(*RefView) error

type RenderedRefType int

const (
	RV_BRANCH_GROUP RenderedRefType = iota
	RV_BRANCH
	RV_TAG_GROUP
	RV_TAG
	RV_SPACE
	RV_LOADING
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
	refList         *RefList
}

type RefView struct {
	channels          *Channels
	repoData          RepoData
	refLists          []*RefList
	refListeners      []RefListener
	active            bool
	renderedRefs      []RenderedRef
	activeRowIndex    uint
	viewStartRowIndex uint
	handlers          map[gc.Key]RefViewHandler
	lock              sync.Mutex
}

type RefListener interface {
	OnRefSelect(refName string, oid *Oid) error
}

func NewRefView(repoData RepoData, channels *Channels) *RefView {
	return &RefView{
		channels: channels,
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
		handlers: map[gc.Key]RefViewHandler{
			gc.KEY_UP:   MoveUpRef,
			gc.KEY_DOWN: MoveDownRef,
			'\n':        SelectRef,
		},
	}
}

func (refView *RefView) Initialise() (err error) {
	log.Info("Initialising RefView")

	if err = refView.repoData.LoadHead(); err != nil {
		return
	}

	if err = refView.repoData.LoadLocalBranches(func(branches []*Branch) error {
		log.Debug("Local branches loaded")
		refView.lock.Lock()
		defer refView.lock.Unlock()

		refView.GenerateRenderedRefs()

		_, headBranch := refView.repoData.Head()
		refView.activeRowIndex = 1

		if headBranch != nil {
			refView.activeRowIndex = 1

			for _, branch := range branches {
				if branch.name == headBranch.name {
					break
				}

				refView.activeRowIndex++
			}
		}

		refView.channels.UpdateDisplay()

		return nil
	}); err != nil {
		return
	}

	if err = refView.repoData.LoadLocalTags(func(tags []*Tag) error {
		log.Debug("Local tags loaded")
		refView.lock.Lock()
		defer refView.lock.Unlock()

		refView.GenerateRenderedRefs()
		refView.channels.UpdateDisplay()

		return nil
	}); err != nil {
		return
	}

	refView.GenerateRenderedRefs()
	head, branch := refView.repoData.Head()

	var branchName string
	if branch == nil {
		branchName = getDetachedHeadDisplayValue(head)
	} else {
		branchName = branch.name
	}

	refView.notifyRefListeners(branchName, head)

	return
}

func getDetachedHeadDisplayValue(oid *Oid) string {
	return fmt.Sprintf("HEAD detached at %s", oid.String()[0:7])
}

func (refView *RefView) RegisterRefListener(refListener RefListener) {
	refView.refListeners = append(refView.refListeners, refListener)
}

func (refView *RefView) notifyRefListeners(refName string, oid *Oid) (err error) {
	log.Debugf("Notifying RefListeners of selected oid %v", oid)

	for _, refListener := range refView.refListeners {
		if err = refListener.OnRefSelect(refName, oid); err != nil {
			break
		}
	}

	return
}

func (refView *RefView) Render(win RenderWindow) (err error) {
	log.Debug("Rendering RefView")
	refView.lock.Lock()
	defer refView.lock.Unlock()

	rows := win.Rows() - 2

	if refView.viewStartRowIndex > refView.activeRowIndex {
		refView.viewStartRowIndex = refView.activeRowIndex
	} else if rowDiff := refView.activeRowIndex - refView.viewStartRowIndex; rowDiff >= rows {
		refView.viewStartRowIndex += (rowDiff - rows) + 1
	}

	refIndex := refView.viewStartRowIndex

	for winRowIndex := uint(0); winRowIndex < rows && refIndex < uint(len(refView.renderedRefs)); winRowIndex++ {
		if err = win.SetRow(winRowIndex+1, "%v", refView.renderedRefs[refIndex].value); err != nil {
			return
		}

		refIndex++
	}

	if err = win.SetSelectedRow((refView.activeRowIndex-refView.viewStartRowIndex)+1, refView.active); err != nil {
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
			refList:         refList,
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
	branches, loading := refView.repoData.LocalBranches()

	if loading {
		*renderedRefs = append(*renderedRefs, RenderedRef{
			value:           "   Loading...",
			renderedRefType: RV_LOADING,
		})

		return
	}

	if head, headBranch := refView.repoData.Head(); headBranch == nil {
		*renderedRefs = append(*renderedRefs, RenderedRef{
			value:           fmt.Sprintf("   %s", getDetachedHeadDisplayValue(head)),
			oid:             head,
			renderedRefType: RV_BRANCH,
		})
	}

	for _, branch := range branches {
		*renderedRefs = append(*renderedRefs, RenderedRef{
			value:           fmt.Sprintf("   %s", branch.name),
			oid:             branch.oid,
			renderedRefType: RV_BRANCH,
		})
	}
}

func GenerateTags(refView *RefView, refList *RefList, renderedRefs *[]RenderedRef) {
	tags, loading := refView.repoData.LocalTags()

	if loading {
		*renderedRefs = append(*renderedRefs, RenderedRef{
			value:           "   Loading...",
			renderedRefType: RV_LOADING,
		})

		return
	}

	for _, tag := range tags {
		*renderedRefs = append(*renderedRefs, RenderedRef{
			value:           fmt.Sprintf("   %s", tag.name),
			oid:             tag.oid,
			renderedRefType: RV_TAG,
		})
	}
}

func (refView *RefView) OnActiveChange(active bool) {
	log.Debugf("RefView active: %v", active)
	refView.lock.Lock()
	defer refView.lock.Unlock()

	refView.active = active
}

func (refView *RefView) Handle(keyPressEvent KeyPressEvent) (err error) {
	log.Debugf("RefView handling key %v", keyPressEvent)
	refView.lock.Lock()
	defer refView.lock.Unlock()

	if handler, ok := refView.handlers[keyPressEvent.key]; ok {
		err = handler(refView)
	}

	return
}

func MoveUpRef(refView *RefView) (err error) {
	if refView.activeRowIndex == 0 {
		return
	}

	log.Debug("Moving up one ref")

	startIndex := refView.activeRowIndex
	refView.activeRowIndex--

	for refView.activeRowIndex > 0 {
		renderedRef := refView.renderedRefs[refView.activeRowIndex]

		if renderedRef.renderedRefType != RV_SPACE && renderedRef.renderedRefType != RV_LOADING {
			break
		}

		refView.activeRowIndex--
	}

	renderedRef := refView.renderedRefs[refView.activeRowIndex]
	if renderedRef.renderedRefType == RV_SPACE || renderedRef.renderedRefType == RV_LOADING {
		refView.activeRowIndex = startIndex
		log.Debug("No valid ref entry to move to")
	} else {
		refView.channels.UpdateDisplay()
	}

	return
}

func MoveDownRef(refView *RefView) (err error) {
	indexLimit := uint(len(refView.renderedRefs)) - 1

	if refView.activeRowIndex >= indexLimit {
		return
	}

	log.Debug("Moving down one ref")

	startIndex := refView.activeRowIndex
	refView.activeRowIndex++

	for refView.activeRowIndex < indexLimit {
		renderedRef := refView.renderedRefs[refView.activeRowIndex]

		if renderedRef.renderedRefType != RV_SPACE && renderedRef.renderedRefType != RV_LOADING {
			break
		}

		refView.activeRowIndex++
	}

	renderedRef := refView.renderedRefs[refView.activeRowIndex]
	if renderedRef.renderedRefType == RV_SPACE || renderedRef.renderedRefType == RV_LOADING {
		refView.activeRowIndex = startIndex
		log.Debug("No valid ref entry to move to")
	} else {
		refView.channels.UpdateDisplay()
	}

	return
}

func SelectRef(refView *RefView) (err error) {
	renderedRef := refView.renderedRefs[refView.activeRowIndex]

	switch renderedRef.renderedRefType {
	case RV_BRANCH_GROUP, RV_TAG_GROUP:
		renderedRef.refList.expanded = !renderedRef.refList.expanded
		log.Debugf("Setting ref group %v to expanded %v", renderedRef.refList.name, renderedRef.refList.expanded)
		refView.GenerateRenderedRefs()
		refView.channels.UpdateDisplay()
	case RV_BRANCH, RV_TAG:
		log.Debugf("Selecting ref %v:%v", renderedRef.value, renderedRef.oid)
		if err = refView.notifyRefListeners(strings.TrimLeft(renderedRef.value, " "), renderedRef.oid); err != nil {
			return
		}
		refView.channels.UpdateDisplay()
	default:
		log.Warn("Unexpected ref type %v", renderedRef.renderedRefType)
	}

	return
}
