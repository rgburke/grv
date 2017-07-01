package main

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"strings"
	"sync"
)

type RefViewHandler func(*RefView) error

type RenderedRefType int

const (
	RV_LOCAL_BRANCH_GROUP RenderedRefType = iota
	RV_REMOTE_BRANCH_GROUP
	RV_LOCAL_BRANCH
	RV_REMOTE_BRANCH
	RV_TAG_GROUP
	RV_TAG
	RV_SPACE
	RV_LOADING
)

var refToTheme = map[RenderedRefType]ThemeComponentId{
	RV_LOCAL_BRANCH_GROUP:  CMP_REFVIEW_LOCAL_BRANCHES_HEADER,
	RV_REMOTE_BRANCH_GROUP: CMP_REFVIEW_REMOTE_BRANCHES_HEADER,
	RV_LOCAL_BRANCH:        CMP_REFVIEW_LOCAL_BRANCH,
	RV_REMOTE_BRANCH:       CMP_REFVIEW_REMOTE_BRANCH,
	RV_TAG_GROUP:           CMP_REFVIEW_TAGS_HEADER,
	RV_TAG:                 CMP_REFVIEW_TAG,
}

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
	refNum          uint
}

type RefView struct {
	channels      *Channels
	repoData      RepoData
	refLists      []*RefList
	refListeners  []RefListener
	active        bool
	renderedRefs  []RenderedRef
	viewPos       *ViewPos
	viewDimension ViewDimension
	handlers      map[ActionType]RefViewHandler
	viewSearch    *ViewSearch
	lock          sync.Mutex
}

type RefListener interface {
	OnRefSelect(refName string, oid *Oid) error
}

func NewRefView(repoData RepoData, channels *Channels) *RefView {
	refView := &RefView{
		channels: channels,
		repoData: repoData,
		viewPos:  NewViewPos(),
		refLists: []*RefList{
			&RefList{
				name:            "Branches",
				renderer:        GenerateBranches,
				expanded:        true,
				renderedRefType: RV_LOCAL_BRANCH_GROUP,
			},
			&RefList{
				name:            "Remote Branches",
				renderer:        GenerateBranches,
				renderedRefType: RV_REMOTE_BRANCH_GROUP,
			},
			&RefList{
				name:            "Tags",
				renderer:        GenerateTags,
				renderedRefType: RV_TAG_GROUP,
			},
		},
		handlers: map[ActionType]RefViewHandler{
			ACTION_PREV_LINE:    MoveUpRef,
			ACTION_NEXT_LINE:    MoveDownRef,
			ACTION_PREV_PAGE:    MoveUpRefPage,
			ACTION_NEXT_PAGE:    MoveDownRefPage,
			ACTION_SCROLL_RIGHT: ScrollRefViewRight,
			ACTION_SCROLL_LEFT:  ScrollRefViewLeft,
			ACTION_FIRST_LINE:   MoveToFirstRef,
			ACTION_LAST_LINE:    MoveToLastRef,
			ACTION_SELECT:       SelectRef,
		},
	}

	refView.viewSearch = NewViewSearch(refView, channels)

	return refView
}

func (refView *RefView) Initialise() (err error) {
	log.Info("Initialising RefView")

	if err = refView.repoData.LoadHead(); err != nil {
		return
	}

	if err = refView.repoData.LoadBranches(func(localBranches, remoteBranches []*Branch) error {
		log.Debug("Branches loaded")
		refView.lock.Lock()
		defer refView.lock.Unlock()

		refView.GenerateRenderedRefs()

		_, headBranch := refView.repoData.Head()
		viewPos := refView.viewPos
		viewPos.activeRowIndex = 1

		if headBranch != nil {
			viewPos.activeRowIndex = 1

			for _, branch := range localBranches {
				if branch.name == headBranch.name {
					log.Debugf("Setting branch %v as selected branch", branch.name)
					break
				}

				viewPos.activeRowIndex++
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

	err = refView.notifyRefListeners(branchName, head)

	return
}

func getDetachedHeadDisplayValue(oid *Oid) string {
	return fmt.Sprintf("HEAD detached at %s", oid.String()[0:7])
}

func isSelectableRenderedRef(renderedRefType RenderedRefType) bool {
	return renderedRefType != RV_SPACE && renderedRefType != RV_LOADING
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

	refView.viewDimension = win.ViewDimensions()

	renderedRefNum := uint(len(refView.renderedRefs))
	rows := win.Rows() - 2
	viewPos := refView.viewPos
	viewPos.DetermineViewStartRow(rows, renderedRefNum)
	refIndex := viewPos.viewStartRowIndex
	startColumn := viewPos.viewStartColumn

	for winRowIndex := uint(0); winRowIndex < rows && refIndex < renderedRefNum; winRowIndex++ {
		renderedRef := refView.renderedRefs[refIndex]

		themeComponentId, ok := refToTheme[renderedRef.renderedRefType]
		if !ok {
			themeComponentId = CMP_NONE
		}

		if err = win.SetRow(winRowIndex+1, startColumn, themeComponentId, "%v", renderedRef.value); err != nil {
			return
		}

		refIndex++
	}

	if err = win.SetSelectedRow((viewPos.activeRowIndex-viewPos.viewStartRowIndex)+1, refView.active); err != nil {
		return
	}

	win.DrawBorder()

	if err = win.SetTitle(CMP_REFVIEW_TITLE, "Refs"); err != nil {
		return
	}

	selectedRenderedRef := refView.renderedRefs[viewPos.activeRowIndex]
	if err = refView.renderFooter(win, selectedRenderedRef); err != nil {
		return
	}

	if searchActive, searchPattern := refView.viewSearch.SearchActive(); searchActive {
		if err = win.Highlight(searchPattern, CMP_ALLVIEW_SEARCH_MATCH); err != nil {
			return
		}
	}

	return
}

func (refView *RefView) RenderStatusBar(lineBuilder *LineBuilder) (err error) {
	return
}

func (refView *RefView) RenderHelpBar(lineBuilder *LineBuilder) (err error) {
	RenderKeyBindingHelp(refView.ViewId(), lineBuilder, []ActionMessage{
		ActionMessage{action: ACTION_SELECT, message: "Select"},
	})

	return
}

func (refView *RefView) renderFooter(win RenderWindow, selectedRenderedRef RenderedRef) (err error) {
	var footer string

	switch selectedRenderedRef.renderedRefType {
	case RV_LOCAL_BRANCH_GROUP:
		if localBranches, _, loading := refView.repoData.Branches(); loading {
			footer = "Branches: Loading..."
		} else {
			footer = fmt.Sprintf("Branches: %v", len(localBranches))
		}
	case RV_REMOTE_BRANCH_GROUP:
		if _, remoteBranches, loading := refView.repoData.Branches(); loading {
			footer = "Remote Branches: Loading..."
		} else {
			footer = fmt.Sprintf("Remote Branches: %v", len(remoteBranches))
		}
	case RV_LOCAL_BRANCH:
		localBranches, _, _ := refView.repoData.Branches()
		footer = fmt.Sprintf("Branch %v of %v", selectedRenderedRef.refNum, len(localBranches))
	case RV_REMOTE_BRANCH:
		_, remoteBranches, _ := refView.repoData.Branches()
		footer = fmt.Sprintf("Remote Branch %v of %v", selectedRenderedRef.refNum, len(remoteBranches))
	case RV_TAG_GROUP:
		if tags, loading := refView.repoData.LocalTags(); loading {
			footer = "Tags: Loading"
		} else {
			footer = fmt.Sprintf("Tags: %v", len(tags))
		}
	case RV_TAG:
		tags, _ := refView.repoData.LocalTags()
		footer = fmt.Sprintf("Tag %v of %v", selectedRenderedRef.refNum, len(tags))
	}

	if footer != "" {
		err = win.SetFooter(CMP_REFVIEW_FOOTER, "%v", footer)
	}

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
			value:           fmt.Sprintf("  [%v] %v", expandChar, refList.name),
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
	localBranches, remoteBranches, loading := refView.repoData.Branches()

	if loading {
		*renderedRefs = append(*renderedRefs, RenderedRef{
			value:           "   Loading...",
			renderedRefType: RV_LOADING,
		})

		return
	}

	branchNum := uint(1)
	var branches []*Branch
	var branchRenderedRefType RenderedRefType

	if refList.renderedRefType == RV_LOCAL_BRANCH_GROUP {
		branchRenderedRefType = RV_LOCAL_BRANCH
		branches = localBranches

		if head, headBranch := refView.repoData.Head(); headBranch == nil {
			*renderedRefs = append(*renderedRefs, RenderedRef{
				value:           fmt.Sprintf("   %s", getDetachedHeadDisplayValue(head)),
				oid:             head,
				renderedRefType: branchRenderedRefType,
				refNum:          branchNum,
			})

			branchNum++
		}
	} else {
		branchRenderedRefType = RV_REMOTE_BRANCH
		branches = remoteBranches
	}

	for _, branch := range branches {
		*renderedRefs = append(*renderedRefs, RenderedRef{
			value:           fmt.Sprintf("   %s", branch.name),
			oid:             branch.oid,
			renderedRefType: branchRenderedRefType,
			refNum:          branchNum,
		})

		branchNum++
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

	for tagIndex, tag := range tags {
		*renderedRefs = append(*renderedRefs, RenderedRef{
			value:           fmt.Sprintf("   %s", tag.name),
			oid:             tag.oid,
			renderedRefType: RV_TAG,
			refNum:          uint(tagIndex + 1),
		})
	}
}

func (refView *RefView) OnActiveChange(active bool) {
	log.Debugf("RefView active: %v", active)
	refView.lock.Lock()
	defer refView.lock.Unlock()

	refView.active = active
}

func (refView *RefView) ViewId() ViewId {
	return VIEW_REF
}

func (refView *RefView) ViewPos() *ViewPos {
	return refView.viewPos
}

func (refView *RefView) OnSearchMatch(startPos *ViewPos, matchLineIndex uint) {
	refView.lock.Lock()
	defer refView.lock.Unlock()

	renderedRef := refView.renderedRefs[matchLineIndex]

	if isSelectableRenderedRef(renderedRef.renderedRefType) {
		refView.viewPos.activeRowIndex = matchLineIndex
	} else {
		log.Debugf("Unable to select search match at index %v as it is not a selectable type", matchLineIndex)
	}
}

func (refView *RefView) Line(lineIndex uint) (line string, lineExists bool) {
	refView.lock.Lock()
	defer refView.lock.Unlock()

	renderedRefNum := uint(len(refView.renderedRefs))

	if lineIndex < renderedRefNum {
		renderedRef := refView.renderedRefs[lineIndex]
		line = renderedRef.value
		lineExists = true
	}

	return
}

func (refView *RefView) LineNumber() (lineNumber uint) {
	refView.lock.Lock()
	defer refView.lock.Unlock()

	renderedRefNum := uint(len(refView.renderedRefs))
	return renderedRefNum
}

func (refView *RefView) HandleKeyPress(keystring string) (err error) {
	log.Debugf("RefView handling key %v - NOP", keystring)
	return
}

func (refView *RefView) HandleAction(action Action) (err error) {
	log.Debugf("RefView handling action %v", action)
	refView.lock.Lock()
	defer refView.lock.Unlock()

	if handler, ok := refView.handlers[action.ActionType]; ok {
		err = handler(refView)
	} else {
		_, err = refView.viewSearch.HandleAction(action)
	}

	return
}

func MoveUpRef(refView *RefView) (err error) {
	viewPos := refView.viewPos

	if viewPos.activeRowIndex == 0 {
		return
	}

	log.Debug("Moving up one ref")

	startIndex := viewPos.activeRowIndex
	viewPos.activeRowIndex--

	for viewPos.activeRowIndex > 0 {
		renderedRef := refView.renderedRefs[viewPos.activeRowIndex]

		if isSelectableRenderedRef(renderedRef.renderedRefType) {
			break
		}

		viewPos.activeRowIndex--
	}

	renderedRef := refView.renderedRefs[viewPos.activeRowIndex]
	if isSelectableRenderedRef(renderedRef.renderedRefType) {
		refView.channels.UpdateDisplay()
	} else {
		viewPos.activeRowIndex = startIndex
		log.Debug("No valid ref entry to move to")
	}

	return
}

func MoveDownRef(refView *RefView) (err error) {
	renderedRefNum := uint(len(refView.renderedRefs))
	viewPos := refView.viewPos

	if renderedRefNum == 0 || !(viewPos.activeRowIndex < renderedRefNum-1) {
		return
	}

	log.Debug("Moving down one ref")

	startIndex := viewPos.activeRowIndex
	viewPos.activeRowIndex++

	for viewPos.activeRowIndex < renderedRefNum-1 {
		renderedRef := refView.renderedRefs[viewPos.activeRowIndex]

		if isSelectableRenderedRef(renderedRef.renderedRefType) {
			break
		}

		viewPos.activeRowIndex++
	}

	renderedRef := refView.renderedRefs[viewPos.activeRowIndex]
	if isSelectableRenderedRef(renderedRef.renderedRefType) {
		refView.channels.UpdateDisplay()
	} else {
		viewPos.activeRowIndex = startIndex
		log.Debug("No valid ref entry to move to")
	}

	return
}

func MoveUpRefPage(refView *RefView) (err error) {
	pageSize := refView.viewDimension.rows - 2
	viewPos := refView.viewPos

	for viewPos.activeRowIndex > 0 && pageSize > 0 {
		if err = MoveUpRef(refView); err != nil {
			break
		} else {
			pageSize--
		}
	}

	return
}

func MoveDownRefPage(refView *RefView) (err error) {
	renderedRefNum := uint(len(refView.renderedRefs))
	pageSize := refView.viewDimension.rows - 2
	viewPos := refView.viewPos

	for viewPos.activeRowIndex+1 < renderedRefNum && pageSize > 0 {
		if err = MoveDownRef(refView); err != nil {
			break
		} else {
			pageSize--
		}
	}

	return
}

func ScrollRefViewRight(refView *RefView) (err error) {
	viewPos := refView.viewPos
	viewPos.MovePageRight(refView.viewDimension.cols)
	log.Debugf("Scrolling right. View starts at column %v", viewPos.viewStartColumn)
	refView.channels.UpdateDisplay()

	return
}

func ScrollRefViewLeft(refView *RefView) (err error) {
	viewPos := refView.viewPos

	if viewPos.MovePageLeft(refView.viewDimension.cols) {
		log.Debugf("Scrolling left. View starts at column %v", viewPos.viewStartColumn)
		refView.channels.UpdateDisplay()
	}

	return
}

func MoveToFirstRef(refView *RefView) (err error) {
	viewPos := refView.viewPos

	if viewPos.MoveToFirstLine() {
		log.Debugf("Moving to first ref")
		refView.channels.UpdateDisplay()
	}

	return
}

func MoveToLastRef(refView *RefView) (err error) {
	viewPos := refView.viewPos
	renderedRefNum := uint(len(refView.renderedRefs))

	if viewPos.MoveToLastLine(renderedRefNum) {
		log.Debugf("Moving to last ref")
		refView.channels.UpdateDisplay()
	}

	return
}

func SelectRef(refView *RefView) (err error) {
	renderedRef := refView.renderedRefs[refView.viewPos.activeRowIndex]

	switch renderedRef.renderedRefType {
	case RV_LOCAL_BRANCH_GROUP, RV_REMOTE_BRANCH_GROUP, RV_TAG_GROUP:
		renderedRef.refList.expanded = !renderedRef.refList.expanded
		log.Debugf("Setting ref group %v to expanded %v", renderedRef.refList.name, renderedRef.refList.expanded)
		refView.GenerateRenderedRefs()
		refView.channels.UpdateDisplay()
	case RV_LOCAL_BRANCH, RV_REMOTE_BRANCH, RV_TAG:
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
