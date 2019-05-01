package main

import (
	"fmt"
	"strings"
	"sync"

	log "github.com/Sirupsen/logrus"
)

type refViewHandler func(*RefView, Action) error

// RenderedRefType is the type (branch, tag, etc...) of a rendered ref
type RenderedRefType int

// The set of RenderedRefTypes
const (
	RvLocalBranchGroup RenderedRefType = iota
	RvRemoteBranchGroup
	RvLocalBranch
	RvHead
	RvRemoteBranch
	RvTagGroup
	RvTag
	RvSpace
	RvLoading
)

var refToTheme = map[RenderedRefType]ThemeComponentID{
	RvLocalBranchGroup:  CmpRefviewLocalBranchesHeader,
	RvRemoteBranchGroup: CmpRefviewRemoteBranchesHeader,
	RvLocalBranch:       CmpRefviewLocalBranch,
	RvHead:              CmpRefviewHead,
	RvRemoteBranch:      CmpRefviewRemoteBranch,
	RvTagGroup:          CmpRefviewTagsHeader,
	RvTag:               CmpRefviewTag,
}

type renderedRefGenerator func(*RefView, *refList, renderedRefSet)

type refList struct {
	name            string
	expanded        bool
	renderer        renderedRefGenerator
	renderedRefType RenderedRefType
}

// RenderedRef represents a reference's string value and meta data
type RenderedRef struct {
	value           string
	ref             Ref
	renderedRefType RenderedRefType
	refList         *refList
	refNum          uint
}

func (renderedRef *RenderedRef) isSelectable() bool {
	return renderedRef.renderedRefType != RvSpace && renderedRef.renderedRefType != RvLoading
}

type renderedRefSet interface {
	Add(*RenderedRef)
	AddChild(renderedRefSet)
	RemoveChild() (removed bool)
	Child() renderedRefSet
	Clear()
	RenderedRefs() []*RenderedRef
	Children() uint
}

type renderedRefList struct {
	child        renderedRefSet
	renderedRefs []*RenderedRef
	refFilter    *RefFilter
}

func newRenderedRefList() *renderedRefList {
	return newFilteredRenderedRefList(nil)
}

func newFilteredRenderedRefList(refFilter *RefFilter) *renderedRefList {
	return &renderedRefList{
		refFilter: refFilter,
	}
}

// Add a ref to the list if it matches the filter (if set) and pass it down the child filter
func (renderedRefList *renderedRefList) Add(renderedRef *RenderedRef) {
	if renderedRefList.refFilter != nil && !renderedRefList.refFilter.MatchesFilter(renderedRef) {
		return
	}

	renderedRefList.renderedRefs = append(renderedRefList.renderedRefs, renderedRef)

	if renderedRefList.child != nil {
		renderedRefList.child.Add(renderedRef)
	}
}

// AddChild adds another ref set and initialises it with its parents references
func (renderedRefList *renderedRefList) AddChild(renderedRefs renderedRefSet) {
	if renderedRefList.child != nil {
		renderedRefList.child.AddChild(renderedRefs)
	} else {
		renderedRefList.child = renderedRefs

		for _, renderedRef := range renderedRefList.renderedRefs {
			renderedRefs.Add(renderedRef)
		}
	}
}

// Remove child removes the last child in the chain
func (renderedRefList *renderedRefList) RemoveChild() (removed bool) {
	switch {
	case renderedRefList.Child() == nil:
	case renderedRefList.Child().Child() == nil:
		renderedRefList.child = nil
		removed = true
	default:
		removed = renderedRefList.Child().RemoveChild()
	}

	return
}

// Child returns the child
func (renderedRefList *renderedRefList) Child() renderedRefSet {
	return renderedRefList.child
}

// Clear clears the list of rendered refs for this instance and all its children
func (renderedRefList *renderedRefList) Clear() {
	renderedRefList.renderedRefs = renderedRefList.renderedRefs[0:0]

	if renderedRefList.child != nil {
		renderedRefList.child.Clear()
	}
}

// RenderedRefs returns the leaf childs set of rendered refs
func (renderedRefList *renderedRefList) RenderedRefs() []*RenderedRef {
	if renderedRefList.child != nil {
		return renderedRefList.child.RenderedRefs()
	}

	return renderedRefList.renderedRefs
}

// Children returns a count of the number of children this instance has
func (renderedRefList *renderedRefList) Children() (children uint) {
	renderedRefs := renderedRefList.Child()

	for ; renderedRefs != nil; renderedRefs = renderedRefs.Child() {
		children++
	}

	return
}

// RefView manages the display of references
type RefView struct {
	*SelectableRowView
	channels          Channels
	repoData          RepoData
	repoController    RepoController
	config            Config
	refLists          []*refList
	refListeners      []RefListener
	renderedRefs      renderedRefSet
	activeViewPos     ViewPos
	lastViewDimension ViewDimension
	handlers          map[ActionType]refViewHandler
	variables         GRVVariableSetter
	lock              sync.Mutex
}

// RefListener is notified when a reference is selected
type RefListener interface {
	OnRefSelect(ref Ref) error
}

// NewRefView creates a new instance
func NewRefView(repoData RepoData, repoController RepoController, channels Channels, config Config, variables GRVVariableSetter) *RefView {
	refView := &RefView{
		channels:       channels,
		repoData:       repoData,
		repoController: repoController,
		config:         config,
		variables:      variables,
		activeViewPos:  NewViewPosition(),
		renderedRefs:   newRenderedRefList(),
		refLists: []*refList{
			{
				name:            "Branches",
				renderer:        generateBranches,
				expanded:        true,
				renderedRefType: RvLocalBranchGroup,
			},
			{
				name:            "Remote Branches",
				renderer:        generateBranches,
				renderedRefType: RvRemoteBranchGroup,
			},
			{
				name:            "Tags",
				renderer:        generateTags,
				renderedRefType: RvTagGroup,
			},
		},
		handlers: map[ActionType]refViewHandler{
			ActionSelect:                  selectRef,
			ActionAddFilter:               addRefFilter,
			ActionRemoveFilter:            removeRefFilter,
			ActionMouseSelect:             mouseSelectRef,
			ActionCheckoutRef:             checkoutRef,
			ActionCheckoutPreviousRef:     checkoutPreviousRef,
			ActionCreateBranch:            createBranchFromRef,
			ActionCreateBranchAndCheckout: createBranchFromRefAndCheckout,
			ActionCreateTag:               createTagFromRef,
			ActionCreateAnnotatedTag:      createAnnotatedTagFromRef,
			ActionPushRef:                 pushRef,
			ActionDeleteRef:               deleteRef,
			ActionShowAvailableActions:    showActionsForRef,
			ActionMergeRef:                mergeRef,
			ActionRebase:                  rebase,
		},
	}

	refView.SelectableRowView = NewSelectableRowView(refView, channels, config, variables, &refView.lock, "ref")

	return refView
}

// Initialise loads the HEAD reference along with branches and tags
func (refView *RefView) Initialise() (err error) {
	log.Info("Initialising RefView")
	refView.lock.Lock()
	defer refView.lock.Unlock()

	if err = refView.repoData.LoadHead(); err != nil {
		return
	}

	refView.repoData.LoadRefs(func(refs []Ref) (err error) {
		log.Debug("Refs loaded")
		refView.lock.Lock()
		defer refView.lock.Unlock()

		refView.generateRenderedRefs()

		renderedRefs := refView.renderedRefs.RenderedRefs()
		var activeRowIndex uint

		for renderedRefIndex, renderedRef := range renderedRefs {
			if renderedRef.renderedRefType == RvHead {
				activeRowIndex = uint(renderedRefIndex)
				break
			}
		}

		refView.activeViewPos.SetActiveRowIndex(activeRowIndex)
		refView.setVariables()
		refView.channels.UpdateDisplay()

		refView.repoData.RegisterRefStateListener(refView)

		return
	})

	refView.generateRenderedRefs()
	head := refView.repoData.Head()

	err = refView.notifyRefListeners(head)

	return
}

// Dispose of any resources held by the view
func (refView *RefView) Dispose() {

}

// GetDetachedHeadDisplayValue generates a HEAD detached message
func GetDetachedHeadDisplayValue(oid *Oid) string {
	return fmt.Sprintf("HEAD detached at %s", oid.ShortID())
}

// RegisterRefListener adds a ref listener to be notified when a reference is selected
func (refView *RefView) RegisterRefListener(refListener RefListener) {
	if refListener == nil {
		return
	}

	log.Debugf("Registering RefListener %T", refListener)

	refView.lock.Lock()
	defer refView.lock.Unlock()

	refView.refListeners = append(refView.refListeners, refListener)
}

func (refView *RefView) notifyRefListeners(ref Ref) (err error) {
	refListeners := append([]RefListener(nil), refView.refListeners...)

	go func() {
		log.Debugf("Notifying RefListeners of selected ref %v", ref.Name())

		for _, refListener := range refListeners {
			if err = refListener.OnRefSelect(ref); err != nil {
				break
			}
		}
	}()

	return
}

// OnRefsChanged checks if refs have been added or removed and updates the ref view if so
func (refView *RefView) OnRefsChanged(addedRefs, removedRefs []Ref, updatedRefs []*UpdatedRef) {
	refView.lock.Lock()
	defer refView.lock.Unlock()

	updateDisplay := false

	if len(addedRefs) > 0 || len(removedRefs) > 0 {
		updateDisplay = true
	} else {
		for _, updatedRef := range updatedRefs {
			if updatedRef.NewRef.Name() == RdlHeadRef {
				updateDisplay = true
				break
			}
		}
	}

	if updateDisplay {
		refView.generateRenderedRefs()
		refView.channels.UpdateDisplay()
	}
}

// OnHeadChanged updates the ref view display when HEAD has changed
func (refView *RefView) OnHeadChanged(oldHead, newHead Ref) {
	refView.lock.Lock()
	defer refView.lock.Unlock()

	refView.generateRenderedRefs()
	refView.channels.UpdateDisplay()
}

// OnTrackingBranchesUpdated updates the ref view display when tracking branches have updated
func (refView *RefView) OnTrackingBranchesUpdated(trackingBranches []*LocalBranch) {
	refView.lock.Lock()
	defer refView.lock.Unlock()

	refView.generateRenderedRefs()
	refView.channels.UpdateDisplay()
}

// Render generates and writes the ref view to the provided window
func (refView *RefView) Render(win RenderWindow) (err error) {
	refView.lock.Lock()
	defer refView.lock.Unlock()

	refView.lastViewDimension = win.ViewDimensions()

	renderedRefs := refView.renderedRefs.RenderedRefs()
	renderedRefNum := uint(len(renderedRefs))
	rows := win.Rows() - 2
	viewPos := refView.activeViewPos
	viewPos.DetermineViewStartRow(rows, renderedRefNum)
	refIndex := viewPos.ViewStartRowIndex()
	startColumn := viewPos.ViewStartColumn()

	for winRowIndex := uint(0); winRowIndex < rows && refIndex < renderedRefNum; winRowIndex++ {
		renderedRef := renderedRefs[refIndex]

		themeComponentID, ok := refToTheme[renderedRef.renderedRefType]
		if !ok {
			themeComponentID = CmpNone
		}

		var lineBuilder *LineBuilder
		if lineBuilder, err = win.LineBuilder(winRowIndex+1, startColumn); err != nil {
			return
		}

		lineBuilder.AppendWithStyle(themeComponentID, "%v", renderedRef.value)

		if localBranch, isLocalBranch := renderedRef.ref.(*LocalBranch); isLocalBranch && localBranch.IsTrackingBranch() {
			lineBuilder.
				AppendWithStyle(themeComponentID, " (").
				AppendACSChar(AcsUarrow, themeComponentID).
				AppendWithStyle(themeComponentID, "%v ", localBranch.ahead).
				AppendACSChar(AcsDarrow, themeComponentID).
				AppendWithStyle(themeComponentID, "%v)", localBranch.behind)
		}

		refIndex++
	}

	if err = win.SetSelectedRow(viewPos.SelectedRowIndex()+1, refView.viewState); err != nil {
		return
	}

	win.DrawBorder()

	if err = win.SetTitle(CmpRefviewTitle, "Refs"); err != nil {
		return
	}

	selectedRenderedRef := renderedRefs[viewPos.ActiveRowIndex()]
	if err = refView.renderFooter(win, selectedRenderedRef); err != nil {
		return
	}

	if searchActive, searchPattern, lastSearchFoundMatch := refView.viewSearch.SearchActive(); searchActive && lastSearchFoundMatch {
		if err = win.Highlight(searchPattern, CmpAllviewSearchMatch); err != nil {
			return
		}
	}

	return
}

// RenderHelpBar generates key binding help info for the ref view
func (refView *RefView) RenderHelpBar(lineBuilder *LineBuilder) (err error) {
	RenderKeyBindingHelp(refView.ViewID(), lineBuilder, refView.config, []ActionMessage{
		{action: ActionSelect, message: "Select"},
		{action: ActionShowAvailableActions, message: "Show actions for ref"},
		{action: ActionFilterPrompt, message: "Add Filter"},
		{action: ActionRemoveFilter, message: "Remove Filter"},
	})

	return
}

func (refView *RefView) renderFooter(win RenderWindow, selectedRenderedRef *RenderedRef) (err error) {
	var footer string

	if filters := refView.renderedRefs.Children(); filters > 0 {
		plural := ""
		if filters > 1 {
			plural = "s"
		}

		footer = fmt.Sprintf("%v filter%v applied", filters, plural)
	} else {
		switch selectedRenderedRef.renderedRefType {
		case RvLocalBranchGroup:
			if localBranches, _, loading := refView.repoData.Branches(); loading && len(localBranches) == 0 {
				footer = "Branches: Loading..."
			} else {
				footer = fmt.Sprintf("Branches: %v", len(localBranches))
			}
		case RvRemoteBranchGroup:
			if _, remoteBranches, loading := refView.repoData.Branches(); loading && len(remoteBranches) == 0 {
				footer = "Remote Branches: Loading..."
			} else {
				footer = fmt.Sprintf("Remote Branches: %v", len(remoteBranches))
			}
		case RvLocalBranch, RvHead:
			localBranches, _, _ := refView.repoData.Branches()
			branchNum := len(localBranches)
			if _, isDetached := refView.repoData.Head().(*HEAD); isDetached {
				branchNum++
			}
			footer = fmt.Sprintf("Branch %v of %v", selectedRenderedRef.refNum, branchNum)
		case RvRemoteBranch:
			_, remoteBranches, _ := refView.repoData.Branches()
			footer = fmt.Sprintf("Remote Branch %v of %v", selectedRenderedRef.refNum, len(remoteBranches))
		case RvTagGroup:
			if tags, loading := refView.repoData.Tags(); loading && len(tags) == 0 {
				footer = "Tags: Loading"
			} else {
				footer = fmt.Sprintf("Tags: %v", len(tags))
			}
		case RvTag:
			tags, _ := refView.repoData.Tags()
			footer = fmt.Sprintf("Tag %v of %v", selectedRenderedRef.refNum, len(tags))
		}
	}

	if footer != "" {
		err = win.SetFooter(CmpRefviewFooter, "%v", footer)
	}

	return
}

func (refView *RefView) generateRenderedRefs() {
	log.Debug("Generating Rendered Refs")
	refView.renderedRefs.Clear()
	renderedRefs := refView.renderedRefs

	for refIndex, refList := range refView.refLists {
		expandChar := "+"
		if refList.expanded {
			expandChar = "-"
		}

		renderedRefs.Add(&RenderedRef{
			value:           fmt.Sprintf("  [%v] %v", expandChar, refList.name),
			refList:         refList,
			renderedRefType: refList.renderedRefType,
		})

		if refList.expanded {
			refList.renderer(refView, refList, renderedRefs)
		}

		if refIndex != len(refView.refLists)-1 {
			renderedRefs.Add(&RenderedRef{
				value:           "",
				renderedRefType: RvSpace,
			})
		}
	}

	viewPos := refView.activeViewPos
	renderedRefNum := uint(len(renderedRefs.RenderedRefs()))

	if viewPos.ActiveRowIndex() >= renderedRefNum {
		viewPos.SetActiveRowIndex(renderedRefNum - 1)
	} else {
		renderedRef := renderedRefs.RenderedRefs()[viewPos.ActiveRowIndex()]

		if renderedRef.renderedRefType == RvSpace {
			log.Debugf("Active row is empty. Moving to previous row")
			refView.SelectableRowView.HandleAction(Action{ActionType: ActionPrevLine})
		}
	}

	refView.channels.ReportError(refView.selectNearestSelectableRow())
	refView.setVariables()
}

func generateBranches(refView *RefView, refList *refList, renderedRefs renderedRefSet) {
	localBranches, remoteBranches, loading := refView.repoData.Branches()

	if loading && len(localBranches) == 0 && len(remoteBranches) == 0 {
		renderedRefs.Add(&RenderedRef{
			value:           "   Loading...",
			renderedRefType: RvLoading,
		})

		return
	}

	branchNum := uint(1)
	var branches []Branch
	var branchRenderedRefType RenderedRefType

	if refList.renderedRefType == RvLocalBranchGroup {
		branchRenderedRefType = RvLocalBranch
		branches = localBranches
		head := refView.repoData.Head()

		if _, isDetached := head.(*HEAD); isDetached {
			renderedRefs.Add(&RenderedRef{
				value:           fmt.Sprintf("   %s", GetDetachedHeadDisplayValue(head.Oid())),
				renderedRefType: branchRenderedRefType,
				refNum:          branchNum,
				ref:             head,
			})

			branchNum++
		}
	} else {
		branchRenderedRefType = RvRemoteBranch
		branches = remoteBranches
	}

	for _, branch := range branches {
		renderedRefs.Add(&RenderedRef{
			value:           fmt.Sprintf("   %s", branch.Shorthand()),
			ref:             branch,
			renderedRefType: branchRenderedRefType,
			refNum:          branchNum,
		})

		branchNum++
	}

	if refList.renderedRefType == RvLocalBranchGroup {
		head := refView.repoData.Head()

		for _, renderedRef := range renderedRefs.RenderedRefs() {
			if head.Equal(renderedRef.ref) {
				renderedRef.value = fmt.Sprintf(" * %v", strings.TrimLeft(renderedRef.value, " "))
				renderedRef.renderedRefType = RvHead
				break
			}
		}
	}
}

func generateTags(refView *RefView, refList *refList, renderedRefs renderedRefSet) {
	tags, loading := refView.repoData.Tags()

	if loading && len(tags) == 0 {
		renderedRefs.Add(&RenderedRef{
			value:           "   Loading...",
			renderedRefType: RvLoading,
		})

		return
	}

	for tagIndex, tag := range tags {
		renderedRefs.Add(&RenderedRef{
			value:           fmt.Sprintf("   %s", tag.Shorthand()),
			ref:             tag,
			renderedRefType: RvTag,
			refNum:          uint(tagIndex + 1),
		})
	}
}

func (refView *RefView) createRefListenerView(ref Ref) {
	createViewArgs := CreateViewArgs{
		viewID:   ViewCommit,
		viewArgs: []interface{}{ref.Name()},
		registerViewListener: func(observer interface{}) (err error) {
			if observer == nil {
				return fmt.Errorf("Invalid RefListener: %v", observer)
			}

			if refListener, ok := observer.(RefListener); ok {
				refView.RegisterRefListener(refListener)
			} else {
				err = fmt.Errorf("Observer is not a RefListener but has type %T", observer)
			}

			return
		},
	}

	refView.channels.DoAction(Action{
		ActionType: ActionSplitView,
		Args: []interface{}{
			ActionSplitViewArgs{
				CreateViewArgs: createViewArgs,
				orientation:    CoDynamic,
			},
		},
	})
}

// ViewID returns the view ID of the ref view
func (refView *RefView) ViewID() ViewID {
	return ViewRef
}

func (refView *RefView) viewPos() ViewPos {
	return refView.activeViewPos
}

func (refView *RefView) line(lineIndex uint) (line string) {
	renderedRefs := refView.renderedRefs.RenderedRefs()
	renderedRefNum := uint(len(renderedRefs))

	if lineIndex >= renderedRefNum {
		log.Errorf("Invalid lineIndex: %v", lineIndex)
		return
	}

	renderedRef := renderedRefs[lineIndex]
	line = renderedRef.value

	return
}

func (refView *RefView) rows() uint {
	renderedRefs := refView.renderedRefs.RenderedRefs()
	return uint(len(renderedRefs))
}

func (refView *RefView) viewDimension() ViewDimension {
	return refView.lastViewDimension
}

func (refView *RefView) onRowSelected(rowIndex uint) (err error) {
	refView.setVariables()
	return
}

func (refView *RefView) isSelectableRow(rowIndex uint) (isSelectable bool) {
	if rowIndex >= refView.rows() {
		return
	}

	renderedRefs := refView.renderedRefs.RenderedRefs()
	renderedRef := renderedRefs[rowIndex]

	return renderedRef.isSelectable()
}

func (refView *RefView) setVariables() {
	refView.SelectableRowView.setVariables()

	selectedRefIndex := refView.viewPos().ActiveRowIndex()
	var branch, tag string

	if selectedRefIndex < refView.rows() {
		renderedRefs := refView.renderedRefs.RenderedRefs()
		renderedRef := renderedRefs[selectedRefIndex]

		if renderedRef.renderedRefType == RvLocalBranch || renderedRef.renderedRefType == RvRemoteBranch {
			branch = renderedRef.ref.Name()
		} else if renderedRef.renderedRefType == RvTag {
			tag = renderedRef.ref.Name()
		} else if renderedRef.renderedRefType == RvHead {
			if _, isDetached := renderedRef.ref.(*HEAD); !isDetached {
				branch = renderedRef.ref.Name()
			}
		}
	}

	if branch != "" || tag != "" {
		refView.variables.SetViewVariable(VarBranch, branch, refView.viewState)
		refView.variables.SetViewVariable(VarTag, tag, refView.viewState)
	}
}

// HandleEvent reacts to an event
func (refView *RefView) HandleEvent(event Event) (err error) {
	refView.lock.Lock()
	defer refView.lock.Unlock()

	switch event.EventType {
	case ViewRemovedEvent:
		refView.removeRefListeners(event.Args)
	}

	return
}

func (refView *RefView) removeRefListeners(views []interface{}) {
	for _, view := range views {
		if refListener, ok := view.(RefListener); ok {
			refView.removeRefListener(refListener)
		}
	}
}

func (refView *RefView) removeRefListener(refListener RefListener) {
	for index, listener := range refView.refListeners {
		if refListener == listener {
			log.Debugf("Removing RefListener %T", refListener)
			refView.refListeners = append(refView.refListeners[:index], refView.refListeners[index+1:]...)
			break
		}
	}
}

func (refView *RefView) selectedRef() (renderedRef *RenderedRef) {
	selectedIndex := refView.activeViewPos.ActiveRowIndex()

	if refView.rows() == 0 || selectedIndex >= refView.rows() {
		return
	}

	renderedRefs := refView.renderedRefs.RenderedRefs()
	renderedRef = renderedRefs[selectedIndex]

	return
}

// HandleAction checks if the rev view supports an action and executes it if so
func (refView *RefView) HandleAction(action Action) (err error) {
	log.Debugf("RefView handling action %v", action)
	refView.lock.Lock()
	defer refView.lock.Unlock()

	var handled bool
	if handler, ok := refView.handlers[action.ActionType]; ok {
		log.Debugf("Action handled by RefView")
		err = handler(refView, action)
	} else if handled, err = refView.SelectableRowView.HandleAction(action); handled {
		log.Debugf("Action handled by SelectableRowView")
	} else {
		log.Debugf("Action not handled")
	}

	return
}

func selectRef(refView *RefView, action Action) (err error) {
	renderedRefs := refView.renderedRefs.RenderedRefs()
	renderedRef := renderedRefs[refView.activeViewPos.ActiveRowIndex()]

	switch renderedRef.renderedRefType {
	case RvLocalBranchGroup, RvRemoteBranchGroup, RvTagGroup:
		renderedRef.refList.expanded = !renderedRef.refList.expanded
		log.Debugf("Setting ref group %v to expanded %v", renderedRef.refList.name, renderedRef.refList.expanded)
		refView.generateRenderedRefs()
		refView.channels.UpdateDisplay()
	case RvLocalBranch, RvHead, RvRemoteBranch, RvTag:
		log.Debugf("Selecting ref %v:%v", renderedRef.ref.Name(), renderedRef.ref.Oid())

		if len(refView.refListeners) == 0 {
			refView.createRefListenerView(renderedRef.ref)
		} else {
			if err = refView.notifyRefListeners(renderedRef.ref); err != nil {
				return
			}
		}
		refView.channels.UpdateDisplay()
	default:
		log.Warnf("Unexpected ref type %v", renderedRef.renderedRefType)
	}

	return
}

func addRefFilter(refView *RefView, action Action) (err error) {
	if !(len(action.Args) > 0) {
		return fmt.Errorf("Expected filter query argument")
	}

	query, ok := action.Args[0].(string)
	if !ok {
		return fmt.Errorf("Expected filter query argument to have type string")
	}

	refFilter, errors := CreateRefFilter(query)
	if len(errors) > 0 {
		refView.channels.ReportErrors(errors)
		return
	} else if refFilter == nil {
		log.Debugf("Query string does not define ref filter: \"%v\"", query)
		return
	}

	beforeRenderedRefNum := len(refView.renderedRefs.RenderedRefs())
	refView.renderedRefs.AddChild(newFilteredRenderedRefList(refFilter))
	afterRenderedRefNum := len(refView.renderedRefs.RenderedRefs())

	if afterRenderedRefNum < beforeRenderedRefNum {
		refView.channels.ReportStatus("Filter applied")
	} else {
		refView.channels.ReportStatus("Filter had no effect")
	}

	return
}

func removeRefFilter(refView *RefView, action Action) (err error) {
	if refView.renderedRefs.RemoveChild() {
		refView.channels.ReportStatus("Removed ref filter")
	} else {
		refView.channels.ReportStatus("No ref filter applied to remove")
	}

	return
}

func mouseSelectRef(refView *RefView, action Action) (err error) {
	mouseEvent, err := GetMouseEventFromAction(action)
	if err != nil {
		return
	}

	if mouseEvent.row == 0 || mouseEvent.row == refView.lastViewDimension.rows-1 {
		return
	}

	viewPos := refView.activeViewPos
	selectedIndex := viewPos.ViewStartRowIndex() + mouseEvent.row - 1

	renderedRefs := refView.renderedRefs.RenderedRefs()
	renderedRefNum := uint(len(renderedRefs))

	if selectedIndex >= renderedRefNum {
		return
	}

	renderedRef := renderedRefs[selectedIndex]

	if !renderedRef.isSelectable() {
		return
	}

	if viewPos.ActiveRowIndex() == selectedIndex {
		err = selectRef(refView, action)
	} else {
		viewPos.SetActiveRowIndex(selectedIndex)
		refView.channels.UpdateDisplay()
	}

	return
}

func checkoutRef(refView *RefView, action Action) (err error) {
	renderedRefs := refView.renderedRefs.RenderedRefs()
	renderedRef := renderedRefs[refView.activeViewPos.ActiveRowIndex()]

	if renderedRef.ref == nil {
		return
	}

	if refView.config.GetBool(CfConfirmCheckout) {
		question := fmt.Sprintf("Are you sure you want to checkout ref %v?", renderedRef.ref.Shorthand())

		refView.channels.DoAction(YesNoQuestion(question, func(response QuestionResponse) {
			if response == ResponseYes {
				refView.performCheckoutRef(renderedRef)
			}
		}))
	} else {
		refView.performCheckoutRef(renderedRef)
	}

	return
}

func checkoutPreviousRef(refView *RefView, action Action) (err error) {
	if refView.config.GetBool(CfConfirmCheckout) {
		question := "Are you sure you want to checkout the previous ref?"
		refView.channels.DoAction(YesNoQuestion(question, func(response QuestionResponse) {
			if response == ResponseYes {
				refView.repoController.CheckoutPreviousRef(func(ref Ref, err error) {
					refView.onRefCheckoutOut(ref, err)
				})
			}
		}))
	} else {
		refView.repoController.CheckoutPreviousRef(func(ref Ref, err error) {
			refView.onRefCheckoutOut(ref, err)
		})
	}

	return
}

func (refView *RefView) performCheckoutRef(renderedRef *RenderedRef) {
	refView.repoController.CheckoutRef(renderedRef.ref, func(ref Ref, err error) {
		refView.onRefCheckoutOut(ref, err)
	})
}

func (refView *RefView) onRefCheckoutOut(ref Ref, err error) {
	refView.lock.Lock()
	defer refView.lock.Unlock()

	if err != nil {
		refView.channels.ReportError(err)
		return
	}

	refView.generateRenderedRefs()

	if err = refView.setSelectedRowToRef(ref); err != nil {
		refView.channels.ReportError(err)
		return
	}

	refView.channels.ReportStatus("Checked out %v", ref.Shorthand())
}

func (refView *RefView) setSelectedRowToRef(ref Ref) (err error) {
	for renderedRefIndex, renderedRef := range refView.renderedRefs.RenderedRefs() {
		if renderedRef.ref != nil && renderedRef.ref.Equal(ref) {
			refView.activeViewPos.SetActiveRowIndex(uint(renderedRefIndex))
			return selectRef(refView, Action{})
		}
	}

	log.Errorf("Unable to find ref %v to select", ref.Name())
	return
}

func (refView *RefView) processRefNameAction(action Action, promptAction, nextAction ActionType) (refName string, ref Ref, err error) {
	if len(action.Args) == 0 {
		refView.channels.DoAction(Action{
			ActionType: promptAction,
			Args:       []interface{}{nextAction},
		})

		return
	}

	refName, isString := action.Args[0].(string)
	if !isString {
		err = fmt.Errorf("Expected first argument to be ref name but found %T", action.Args[0])
		return
	}

	renderedRefs := refView.renderedRefs.RenderedRefs()
	renderedRef := renderedRefs[refView.activeViewPos.ActiveRowIndex()]
	ref = renderedRef.ref

	return
}

func createBranchFromRef(refView *RefView, action Action) (err error) {
	branchName, ref, err := refView.processRefNameAction(action, ActionBranchNamePrompt, ActionCreateBranch)
	if err != nil || ref == nil || branchName == "" {
		return
	}

	if err = refView.repoController.CreateBranch(branchName, ref.Oid()); err != nil {
		return
	}

	refView.channels.ReportStatus("Created branch %v at %v", branchName, ref.Oid().ShortID())

	return
}

func createBranchFromRefAndCheckout(refView *RefView, action Action) (err error) {
	branchName, ref, err := refView.processRefNameAction(action, ActionBranchNamePrompt, ActionCreateBranchAndCheckout)
	if err != nil || ref == nil || branchName == "" {
		return
	}

	refView.repoController.CreateBranchAndCheckout(branchName, ref.Oid(), func(ref Ref, err error) {
		refView.lock.Lock()
		defer refView.lock.Unlock()

		if err != nil {
			refView.channels.ReportError(fmt.Errorf("Failed to create and checkout branch %v", branchName))
			return
		}

		refView.generateRenderedRefs()

		if err = refView.setSelectedRowToRef(ref); err != nil {
			refView.channels.ReportError(err)
		}

		refView.channels.ReportStatus("Created and checked out branch %v", branchName)
	})

	return
}

func createTagFromRef(refView *RefView, action Action) (err error) {
	tagName, ref, err := refView.processRefNameAction(action, ActionTagNamePrompt, ActionCreateTag)
	if err != nil || ref == nil || tagName == "" {
		return
	}

	if err = refView.repoController.CreateTag(tagName, ref.Oid()); err != nil {
		return
	}

	refView.channels.ReportStatus("Created tag %v at %v", tagName, ref.Oid().ShortID())

	return
}

func createAnnotatedTagFromRef(refView *RefView, action Action) (err error) {
	tagName, ref, err := refView.processRefNameAction(action, ActionTagNamePrompt, ActionCreateAnnotatedTag)
	if err != nil || ref == nil || tagName == "" {
		return
	}

	refView.repoController.CreateAnnotatedTag(tagName, ref.Oid(), func(ref Ref, err error) {
		refView.lock.Lock()
		defer refView.lock.Unlock()

		if err != nil {
			refView.channels.ReportError(fmt.Errorf("Failed to create annotated tag %v", tagName))
			return
		}

		refView.generateRenderedRefs()

		refView.channels.ReportStatus("Created annotated tag %v at %v", tagName, ref.Oid().ShortID())
	})

	return
}

func pushRef(refView *RefView, action Action) (err error) {
	renderedRef := refView.selectedRef()
	if renderedRef == nil || renderedRef.ref == nil {
		return
	}

	ref := renderedRef.ref

	var track bool

	switch rawRef := ref.(type) {
	case *LocalBranch:
		track = !rawRef.IsTrackingBranch()
	case *RemoteBranch:
		return
	case *HEAD:
		return
	}

	remotes := refView.repoData.Remotes()
	var remote string

	if len(remotes) == 0 {
		return fmt.Errorf("Cannot push ref: No remotes configured")
	} else if len(remotes) > 1 {
		if len(action.Args) == 0 {
			refView.showRemotesMenu(func(selectedValue interface{}) {
				refView.channels.DoAction(Action{
					ActionType: action.ActionType,
					Args:       []interface{}{selectedValue},
				})
			})
			return
		} else if remoteName, ok := action.Args[0].(string); ok {
			remote = remoteName
		} else {
			return fmt.Errorf("Expected to find remote argument")
		}
	} else {
		remote = remotes[0]
	}

	refView.runReportingTask("Running git push", func(quit chan bool) {
		refView.repoController.Push(remote, ref, track, func(err error) {
			if err != nil {
				refView.channels.ReportError(err)
				refView.channels.ReportStatus("git push failed")
			} else {
				refView.channels.ReportStatus("git push for remote %v and ref %v complete", remote, ref.Shorthand())
			}

			close(quit)
		})
	})

	return
}

func deleteRef(refView *RefView, action Action) (err error) {
	renderedRef := refView.selectedRef()
	if renderedRef == nil || renderedRef.ref == nil {
		return
	} else if _, isDetached := renderedRef.ref.(*HEAD); isDetached {
		return
	}

	ref := renderedRef.ref
	remote := false

	question := fmt.Sprintf("Are you sure you want to delete %v?", ref.Shorthand())

	refView.channels.DoAction(YesNoQuestion(question, func(deleteResponse QuestionResponse) {
		if deleteResponse == ResponseNo {
			return
		}

		switch rawRef := ref.(type) {
		case *LocalBranch:
			if refView.repoData.Head().Equal(ref) {
				refView.channels.ReportError(fmt.Errorf("Cannot delete currently checked out branch"))
				return
			}

			if err = refView.repoController.DeleteLocalRef(ref); err != nil {
				refView.channels.ReportError(err)
				return
			}

			remote = rawRef.IsTrackingBranch()
		case *Tag:
			if err = refView.repoController.DeleteLocalRef(ref); err != nil {
				refView.channels.ReportError(err)
				return
			}

			remote = true
		case *RemoteBranch:
			refView.deleteRemoteRef(rawRef.remoteName, ref)
			return
		default:
			return
		}

		refView.channels.ReportStatus("Deleted ref %v", ref.Shorthand())

		remotes := refView.repoData.Remotes()
		if remote && len(remotes) > 0 {
			question := "Do you want to delete the corresponding remote ref as well?"

			refView.channels.DoAction(YesNoQuestion(question, func(deleteRemoteResponse QuestionResponse) {
				if deleteRemoteResponse == ResponseYes {
					if len(remotes) > 1 {
						refView.showRemotesMenu(func(selectedValue interface{}) {
							if remote, ok := selectedValue.(string); ok {
								refView.deleteRemoteRef(remote, ref)
							} else {
								log.Debugf("Expected string value for remote but found: %T", selectedValue)
							}
						})
					} else {
						refView.deleteRemoteRef(remotes[0], ref)
					}
				}
			}))
		}
	}))

	return
}

func (refView *RefView) deleteRemoteRef(remote string, ref Ref) {
	refView.runReportingTask(fmt.Sprintf("Deleting ref %v on remote %v", ref.Shorthand(), remote), func(quit chan bool) {
		refView.repoController.DeleteRemoteRef(remote, ref, func(err error) {
			if err != nil {
				refView.channels.ReportError(err)
				refView.channels.ReportStatus("Deleting ref %v on remote %v failed", ref.Shorthand(), remote)
			} else {
				refView.channels.ReportStatus("Deleted ref %v on remote %v", ref.Shorthand(), remote)
			}

			close(quit)
		})
	})
}

func (refView *RefView) showRemotesMenu(consumer Consumer) {
	remotes := refView.repoData.Remotes()
	contextMenuEntries := []ContextMenuEntry{}

	for _, remote := range remotes {
		contextMenuEntries = append(contextMenuEntries, ContextMenuEntry{
			DisplayName: remote,
			Value:       remote,
		})
	}

	refView.channels.DoAction(Action{
		ActionType: ActionCreateContextMenu,
		Args: []interface{}{
			ActionCreateContextMenuArgs{
				viewDimension: ViewDimension{
					rows: 5,
					cols: 40,
				},
				config: ContextMenuConfig{
					Entity:  "Remote",
					Entries: contextMenuEntries,
					OnSelect: func(entry ContextMenuEntry, entryIndex uint) {
						consumer(entry.Value)
					},
				},
			},
		},
	})
}

func mergeRef(refView *RefView, action Action) (err error) {
	renderedRef := refView.selectedRef()
	if renderedRef == nil || renderedRef.ref == nil {
		return
	}

	ref := renderedRef.ref
	head := refView.repoData.Head()

	if refView.repoController.MergeRef(ref) == nil {
		refView.channels.ReportStatus("Merged %v into %v", ref.Shorthand(), head.Shorthand())
	} else {
		refView.channels.ReportStatus("Merged failed")
		refView.channels.DoAction(Action{
			ActionType: ActionSelectTabByName,
			Args:       []interface{}{StatusViewTitle},
		})
	}

	return
}

func rebase(refView *RefView, action Action) (err error) {
	renderedRef := refView.selectedRef()
	head := refView.repoData.Head()

	if renderedRef == nil || renderedRef.ref == nil {
		return
	} else if _, isLocalBranch := renderedRef.ref.(*LocalBranch); !isLocalBranch {
		return fmt.Errorf("Selected ref is not a local branch")
	} else if _, isHeadLocalBranch := head.(*LocalBranch); !isHeadLocalBranch {
		return fmt.Errorf("HEAD is not a local branch")
	}

	ref := renderedRef.ref

	if refView.repoController.Rebase(ref) == nil {
		refView.channels.ReportStatus("Rebased %v onto %v", head.Shorthand(), ref.Shorthand())
	} else {
		refView.channels.ReportStatus("Rebase failed")
		refView.channels.DoAction(Action{
			ActionType: ActionSelectTabByName,
			Args:       []interface{}{StatusViewTitle},
		})
	}

	return
}

func showActionsForRef(refView *RefView, action Action) (err error) {
	if refView.rows() == 0 {
		return
	}

	renderedRefs := refView.renderedRefs.RenderedRefs()
	renderedRef := renderedRefs[refView.activeViewPos.ActiveRowIndex()]

	if renderedRef.ref == nil {
		return
	}

	refName := renderedRef.ref.Shorthand()
	if StringWidth(refName) > 15 {
		refName = refName[:15] + "..."
	}

	var contextMenuEntries []ContextMenuEntry

	_, isHead := renderedRef.ref.(*HEAD)
	if !isHead {
		contextMenuEntries = append(contextMenuEntries, ContextMenuEntry{
			DisplayName: fmt.Sprintf("Checkout %v", refName),
			Value:       Action{ActionType: ActionCheckoutRef},
		})
	}

	contextMenuEntries = append(contextMenuEntries,
		ContextMenuEntry{
			DisplayName: "Checkout previous ref",
			Value:       Action{ActionType: ActionCheckoutPreviousRef},
		},
		ContextMenuEntry{
			DisplayName: fmt.Sprintf("Create branch from %v", refName),
			Value:       Action{ActionType: ActionCreateBranch},
		},
		ContextMenuEntry{
			DisplayName: fmt.Sprintf("Create branch from %v and checkout", refName),
			Value:       Action{ActionType: ActionCreateBranchAndCheckout},
		},
		ContextMenuEntry{
			DisplayName: fmt.Sprintf("Create tag at %v", refName),
			Value:       Action{ActionType: ActionCreateTag},
		},
		ContextMenuEntry{
			DisplayName: fmt.Sprintf("Create annotated tag at %v", refName),
			Value:       Action{ActionType: ActionCreateAnnotatedTag},
		},
	)

	_, isLocalBranch := renderedRef.ref.(*LocalBranch)
	_, isTag := renderedRef.ref.(*Tag)

	if isLocalBranch || isTag {
		contextMenuEntries = append(contextMenuEntries, ContextMenuEntry{
			DisplayName: fmt.Sprintf("Push %v to remote", refName),
			Value:       Action{ActionType: ActionPushRef},
		})
	}

	head := refView.repoData.Head()
	headName := head.Shorthand()
	if StringWidth(headName) > 12 {
		headName = headName[:12] + "..."
	}

	if !isHead {
		contextMenuEntries = append(contextMenuEntries, ContextMenuEntry{
			DisplayName: fmt.Sprintf("Delete %v", refName),
			Value:       Action{ActionType: ActionDeleteRef},
		})

		if !head.Equal(renderedRef.ref) {
			contextMenuEntries = append(contextMenuEntries, ContextMenuEntry{
				DisplayName: fmt.Sprintf("Merge %v into %v", refName, headName),
				Value:       Action{ActionType: ActionMergeRef},
			})
			contextMenuEntries = append(contextMenuEntries, ContextMenuEntry{
				DisplayName: fmt.Sprintf("Rebase %v onto %v", headName, refName),
				Value:       Action{ActionType: ActionRebase},
			})
		}
	}

	refView.channels.DoAction(Action{
		ActionType: ActionCreateContextMenu,
		Args: []interface{}{
			ActionCreateContextMenuArgs{
				viewDimension: ViewDimension{
					rows: uint(MinInt(len(contextMenuEntries)+2, 15)),
					cols: 60,
				},
				config: ContextMenuConfig{
					ActionView: ViewRef,
					Entries:    contextMenuEntries,
					OnSelect: func(entry ContextMenuEntry, entryIndex uint) {
						if selectedAction, ok := entry.Value.(Action); ok {
							refView.channels.DoAction(selectedAction)
						} else {
							log.Errorf("Expected Action instance but found: %v", entry.Value)
						}
					},
				},
			},
		},
	})

	return
}
