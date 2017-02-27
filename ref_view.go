package main

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	gc "github.com/rthornton128/goncurses"
	"sync"
)

type RefViewHandler func(*RefView, HandlerChannels) error

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
	repoData       RepoData
	refLists       []*RefList
	refListeners   []RefListener
	active         bool
	renderedRefs   []RenderedRef
	activeIndex    uint
	viewStartIndex uint
	handlers       map[gc.Key]RefViewHandler
	lock           sync.Mutex
}

type RefListener interface {
	OnRefSelect(*Oid, HandlerChannels) error
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
		handlers: map[gc.Key]RefViewHandler{
			gc.KEY_UP:   MoveUpRef,
			gc.KEY_DOWN: MoveDownRef,
			'\n':        SelectRef,
		},
	}
}

func (refView *RefView) Initialise(channels HandlerChannels) (err error) {
	log.Info("Initialising RefView")

	if err = refView.repoData.LoadHead(); err != nil {
		return
	}

	if err = refView.repoData.LoadLocalBranches(func(branches []*Branch) {
		log.Debug("Local branches loaded")
		refView.lock.Lock()
		defer refView.lock.Unlock()

		refView.GenerateRenderedRefs()

		_, headBranch := refView.repoData.Head()
		refView.activeIndex = 1

		if headBranch != nil {
			refView.activeIndex = 1

			for _, branch := range branches {
				if branch.name == headBranch.name {
					break
				}

				refView.activeIndex++
			}
		}

		channels.displayCh <- true
	}); err != nil {
		return
	}

	if err = refView.repoData.LoadLocalTags(func(tags []*Tag) {
		log.Debug("Local tags loaded")
		refView.lock.Lock()
		defer refView.lock.Unlock()

		refView.GenerateRenderedRefs()
		channels.displayCh <- true
	}); err != nil {
		return
	}

	refView.GenerateRenderedRefs()
	head, _ := refView.repoData.Head()
	refView.notifyRefListeners(head, channels)

	return
}

func (refView *RefView) RegisterRefListener(refListener RefListener) {
	refView.refListeners = append(refView.refListeners, refListener)
}

func (refView *RefView) notifyRefListeners(oid *Oid, channels HandlerChannels) (err error) {
	log.Debugf("Notifying RefListeners of selected oid %v", oid)

	for _, refListener := range refView.refListeners {
		if err = refListener.OnRefSelect(oid, channels); err != nil {
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
			value:           fmt.Sprintf("   HEAD detached at %s", head.oid.String()[0:7]),
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
	log.Debugf("RefView active %v", active)
	refView.lock.Lock()
	defer refView.lock.Unlock()

	refView.active = active
}

func (refView *RefView) Handle(keyPressEvent KeyPressEvent, channels HandlerChannels) (err error) {
	log.Debugf("RefView handling key %v", keyPressEvent)
	refView.lock.Lock()
	defer refView.lock.Unlock()

	if handler, ok := refView.handlers[keyPressEvent.key]; ok {
		err = handler(refView, channels)
	}

	return
}

func MoveUpRef(refView *RefView, channels HandlerChannels) (err error) {
	if refView.activeIndex == 0 {
		return
	}

	log.Debug("Moving up one ref")

	startIndex := refView.activeIndex
	refView.activeIndex--

	for refView.activeIndex > 0 {
		renderedRef := refView.renderedRefs[refView.activeIndex]

		if renderedRef.renderedRefType != RV_SPACE && renderedRef.renderedRefType != RV_LOADING {
			break
		}

		refView.activeIndex--
	}

	renderedRef := refView.renderedRefs[refView.activeIndex]
	if renderedRef.renderedRefType == RV_SPACE || renderedRef.renderedRefType == RV_LOADING {
		refView.activeIndex = startIndex
		log.Debug("No valid ref entry to move to")
	} else {
		channels.displayCh <- true
	}

	return
}

func MoveDownRef(refView *RefView, channels HandlerChannels) (err error) {
	indexLimit := uint(len(refView.renderedRefs)) - 1

	if refView.activeIndex >= indexLimit {
		return
	}

	log.Debug("Moving down one ref")

	startIndex := refView.activeIndex
	refView.activeIndex++

	for refView.activeIndex < indexLimit {
		renderedRef := refView.renderedRefs[refView.activeIndex]

		if renderedRef.renderedRefType != RV_SPACE && renderedRef.renderedRefType != RV_LOADING {
			break
		}

		refView.activeIndex++
	}

	renderedRef := refView.renderedRefs[refView.activeIndex]
	if renderedRef.renderedRefType == RV_SPACE || renderedRef.renderedRefType == RV_LOADING {
		refView.activeIndex = startIndex
		log.Debug("No valid ref entry to move to")
	} else {
		channels.displayCh <- true
	}

	return
}

func SelectRef(refView *RefView, channels HandlerChannels) (err error) {
	renderedRef := refView.renderedRefs[refView.activeIndex]

	switch renderedRef.renderedRefType {
	case RV_BRANCH_GROUP, RV_TAG_GROUP:
		renderedRef.refList.expanded = !renderedRef.refList.expanded
		log.Debugf("Setting ref group %v to expanded %v", renderedRef.refList.name, renderedRef.refList.expanded)
		refView.GenerateRenderedRefs()
		channels.displayCh <- true
	case RV_BRANCH, RV_TAG:
		log.Debugf("Selecting ref %v:%v", renderedRef.value, renderedRef.oid)
		if err = refView.notifyRefListeners(renderedRef.oid, channels); err != nil {
			return
		}
		channels.displayCh <- true
	default:
		log.Warn("Unexpected ref type %v", renderedRef.renderedRefType)
	}

	return
}
