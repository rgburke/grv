package main

import (
	"fmt"
	"sync"

	log "github.com/Sirupsen/logrus"
)

const (
	terminalAspectRatio = 80.0 / 24.0
)

// ContainerOrientation represents the orientation of the child views
type ContainerOrientation int

// Supported container orientations
const (
	CoVertical ContainerOrientation = iota
	CoHorizontal
	CoDynamic
)

// ChildViewPosition is the position and dimensions of a child view
type ChildViewPosition struct {
	viewDimension ViewDimension
	startRow      uint
	startCol      uint
}

// ViewLayoutData contains data which can be used to determine the child view layout
type ViewLayoutData struct {
	viewDimension   ViewDimension
	fullScreen      bool
	orientation     ContainerOrientation
	activeViewIndex uint
	childViewNum    uint
}

// ChildViewPositionCalculator calculates the child layout data for the view
type ChildViewPositionCalculator interface {
	CalculateChildViewPositions(*ViewLayoutData) []*ChildViewPosition
}

type containerViewHandler func(*ContainerView, Action) error

// ContainerView is a container with no visual presence that manages the
// layout of its child views
type ContainerView struct {
	channels                    Channels
	config                      Config
	childViews                  []BaseView
	title                       string
	viewWins                    map[WindowView]*Window
	emptyWin                    *Window
	activeViewIndex             uint
	handlers                    map[ActionType]containerViewHandler
	orientation                 ContainerOrientation
	childViewPositionCalculator ChildViewPositionCalculator
	viewID                      ViewID
	fullScreen                  bool
	childPositions              []*ChildViewPosition
	styleConfig                 WindowStyleConfig
	viewState                   ViewState
	lock                        sync.Mutex
}

// NewContainerView creates a new instance
func NewContainerView(channels Channels, config Config) *ContainerView {
	containerView := &ContainerView{
		config:      config,
		channels:    channels,
		orientation: CoVertical,
		viewID:      ViewContainer,
		viewWins:    make(map[WindowView]*Window),
		styleConfig: DefaultWindowStyleConfig(),
		handlers: map[ActionType]containerViewHandler{
			ActionNextView:         nextContainerChildView,
			ActionPrevView:         prevContainerChildView,
			ActionFullScreenView:   toggleFullScreenChildView,
			ActionToggleViewLayout: toggleViewOrientation,
			ActionAddView:          addView,
			ActionSplitView:        splitView,
			ActionRemoveView:       removeView,
			ActionMouseSelect:      childMouseClick,
		},
	}

	containerView.childViewPositionCalculator = containerView

	return containerView
}

// AddChildViews adds new child views to this container
func (containerView *ContainerView) AddChildViews(newViews ...BaseView) {
	containerView.lock.Lock()
	defer containerView.lock.Unlock()

	for _, newView := range newViews {
		containerView.addChildView(newView)
	}
}

func (containerView *ContainerView) addChildView(newView BaseView) {
	log.Debugf("Adding new view %T", newView)

	containerView.childViews = append(containerView.childViews, newView)

	if windowView, isWindowView := newView.(WindowView); isWindowView {
		log.Debugf("Creating window for new view %T", newView)
		viewIndex := len(containerView.childViews) - 1
		winID := fmt.Sprintf("%v-%T", viewIndex, windowView)
		win := NewWindowWithStyleConfig(winID, containerView.config, containerView.styleConfig)
		containerView.viewWins[windowView] = win
		containerView.onStateChange(containerView.viewState)
	}
}

// SetChildViewPositionCalculator sets the child layout calculator for this view
func (containerView *ContainerView) SetChildViewPositionCalculator(childViewPositionCalculator ChildViewPositionCalculator) {
	containerView.lock.Lock()
	defer containerView.lock.Unlock()

	containerView.childViewPositionCalculator = childViewPositionCalculator
}

// SetOrientation sets the orientation of the view
func (containerView *ContainerView) SetOrientation(orientation ContainerOrientation) {
	containerView.lock.Lock()
	defer containerView.lock.Unlock()

	containerView.orientation = orientation
}

// SetTitle sets the title of the view
func (containerView *ContainerView) SetTitle(title string) {
	containerView.lock.Lock()
	defer containerView.lock.Unlock()

	containerView.title = title
}

// SetViewID sets the ViewID of the view
func (containerView *ContainerView) SetViewID(viewID ViewID) {
	containerView.lock.Lock()
	defer containerView.lock.Unlock()

	containerView.viewID = viewID
}

// SetWindowStyleConfig sets the window style config for the child views
func (containerView *ContainerView) SetWindowStyleConfig(styleConfig WindowStyleConfig) {
	containerView.lock.Lock()
	defer containerView.lock.Unlock()

	containerView.styleConfig = styleConfig
}

// Initialise initialises this containers child views
func (containerView *ContainerView) Initialise() (err error) {
	containerView.lock.Lock()
	defer containerView.lock.Unlock()

	for _, childView := range containerView.childViews {
		if err = childView.Initialise(); err != nil {
			break
		}
	}

	return
}

// Dispose of any resources held by the view
func (containerView *ContainerView) Dispose() {
	containerView.lock.Lock()
	defer containerView.lock.Unlock()

	for !containerView.isEmpty() {
		removeView(containerView, Action{ActionType: ActionRemoveView})
	}
}

// HandleEvent passes the event on to all child views
func (containerView *ContainerView) HandleEvent(event Event) (err error) {
	containerView.lock.Lock()
	defer containerView.lock.Unlock()

	for _, childView := range containerView.childViews {
		if err = childView.HandleEvent(event); err != nil {
			return
		}
	}

	return
}

// HandleAction processes the action if supported or passes it on to the active child view
func (containerView *ContainerView) HandleAction(action Action) (err error) {
	containerView.lock.Lock()
	defer containerView.lock.Unlock()

	handler, handlerExists := containerView.handlers[action.ActionType]

	if handlerExists {
		err = handler(containerView, action)
	} else if !containerView.isEmpty() {
		err = containerView.activeChildView().HandleAction(action)
	}

	return
}

// OnStateChange updates the active state of this container and its child views
func (containerView *ContainerView) OnStateChange(viewState ViewState) {
	containerView.lock.Lock()
	defer containerView.lock.Unlock()

	containerView.onStateChange(viewState)
}

func (containerView *ContainerView) onStateChange(viewState ViewState) {
	containerView.viewState = viewState

	var inactiveChildViewState ViewState
	if viewState == ViewStateActive {
		inactiveChildViewState = ViewStateInactiveAndVisible
	} else {
		inactiveChildViewState = viewState
	}

	for index, childView := range containerView.childViews {
		if uint(index) == containerView.activeViewIndex {
			childView.OnStateChange(viewState)
		} else {
			childView.OnStateChange(inactiveChildViewState)
		}
	}
}

// ViewID returns container view id
func (containerView *ContainerView) ViewID() ViewID {
	containerView.lock.Lock()
	defer containerView.lock.Unlock()

	return containerView.viewID
}

// RenderHelpBar is proxied to the active child view
func (containerView *ContainerView) RenderHelpBar(lineBuilder *LineBuilder) (err error) {
	renderHelp := true

	if !containerView.isEmpty() {
		if _, isContainerView := containerView.activeChildView().(*ContainerView); isContainerView {
			renderHelp = false
		}
	}

	if renderHelp {
		RenderKeyBindingHelp(containerView.ViewID(), lineBuilder, containerView.config, []ActionMessage{
			{action: ActionNextView, message: "Next View"},
			{action: ActionPrevView, message: "Prev View"},
			{action: ActionFullScreenView, message: "Full Screen"},
			{action: ActionToggleViewLayout, message: "Layout"},
		})
	}

	if !containerView.isEmpty() {
		err = containerView.activeChildView().RenderHelpBar(lineBuilder)
	}

	return
}

// Render determines the layout of all child views, renders them and returns the resulting windows
func (containerView *ContainerView) Render(viewDimension ViewDimension) (wins []*Window, err error) {
	containerView.lock.Lock()
	defer containerView.lock.Unlock()

	if containerView.isEmpty() {
		wins = append(wins, containerView.renderEmptyView(viewDimension))
		return
	}

	if containerView.orientation == CoDynamic {
		containerView.determineOrientation(viewDimension)
	}

	viewLayoutData := ViewLayoutData{
		viewDimension:   viewDimension,
		fullScreen:      containerView.fullScreen,
		orientation:     containerView.orientation,
		activeViewIndex: containerView.activeViewIndex,
		childViewNum:    uint(len(containerView.childViews)),
	}

	containerView.childPositions = containerView.childViewPositionCalculator.CalculateChildViewPositions(&viewLayoutData)

	for childViewIndex, childView := range containerView.childViews {
		childPosition := containerView.childPositions[childViewIndex]
		if childPosition.viewDimension.cols == 0 || childPosition.viewDimension.rows == 0 {
			continue
		}

		switch view := childView.(type) {
		case WindowView:
			var win *Window
			win, err = containerView.renderWindowView(view, childPosition)
			if err != nil {
				return
			}

			wins = append(wins, win)
			win.SetPosition(childPosition.startRow, childPosition.startCol)
		case WindowViewCollection:
			var childWins []*Window
			childWins, err = view.Render(childPosition.viewDimension)
			if err != nil {
				return
			}

			for _, win := range childWins {
				win.OffsetPosition(int(childPosition.startRow), int(childPosition.startCol))
			}

			wins = append(wins, childWins...)
		default:
			log.Errorf("Unsupported view type: %T", view)
		}
	}

	return
}

func (containerView *ContainerView) determineOrientation(viewDimension ViewDimension) {
	viewAspectRation := float64(viewDimension.cols) / float64(viewDimension.rows)
	log.Debugf("View aspect ration: %v", viewAspectRation)

	var orientation ContainerOrientation

	if viewAspectRation < terminalAspectRatio {
		orientation = CoHorizontal
	} else {
		orientation = CoVertical
	}

	containerView.orientation = orientation
}

// CalculateChildViewPositions calculates the child layout data for this view
func (containerView *ContainerView) CalculateChildViewPositions(viewLayoutData *ViewLayoutData) (childPositions []*ChildViewPosition) {
	switch {
	case viewLayoutData.fullScreen:
		for i := uint(0); i < viewLayoutData.childViewNum; i++ {
			childPositions = append(childPositions, &ChildViewPosition{
				viewDimension: ViewDimension{
					rows: 0,
					cols: 0,
				},
				startRow: 0,
				startCol: 0,
			})
		}

		childPositions[viewLayoutData.activeViewIndex].viewDimension = viewLayoutData.viewDimension
	case viewLayoutData.orientation == CoVertical:
		width := uint(viewLayoutData.viewDimension.cols / viewLayoutData.childViewNum)
		startCol := uint(0)

		for i := uint(0); i < viewLayoutData.childViewNum; i++ {
			childPositions = append(childPositions, &ChildViewPosition{
				viewDimension: ViewDimension{
					rows: viewLayoutData.viewDimension.rows,
					cols: width,
				},
				startRow: 0,
				startCol: startCol,
			})

			startCol += width
		}

		childPositions[len(childPositions)-1].viewDimension.cols += viewLayoutData.viewDimension.cols % viewLayoutData.childViewNum
	case viewLayoutData.orientation == CoHorizontal:
		height := uint(viewLayoutData.viewDimension.rows / viewLayoutData.childViewNum)
		startRow := uint(0)

		for i := uint(0); i < viewLayoutData.childViewNum; i++ {
			childPositions = append(childPositions, &ChildViewPosition{
				viewDimension: ViewDimension{
					rows: height,
					cols: viewLayoutData.viewDimension.cols,
				},
				startRow: startRow,
				startCol: 0,
			})

			startRow += height
		}

		childPositions[len(childPositions)-1].viewDimension.rows += viewLayoutData.viewDimension.rows % viewLayoutData.childViewNum
	}

	return
}

func (containerView *ContainerView) renderWindowView(childView WindowView, childPosition *ChildViewPosition) (*Window, error) {
	win := containerView.viewWins[childView]

	win.Resize(childPosition.viewDimension)
	win.SetPosition(childPosition.startRow, childPosition.startCol)
	win.Clear()

	if err := childView.Render(win); err != nil {
		return nil, err
	}

	return win, nil
}

func (containerView *ContainerView) renderEmptyView(viewDimension ViewDimension) *Window {
	if containerView.emptyWin == nil {
		containerView.emptyWin = NewWindowWithStyleConfig("empty", containerView.config, containerView.styleConfig)
	}

	win := containerView.emptyWin
	win.Resize(viewDimension)
	win.Clear()

	return win
}

// ActiveView returns the active child view
func (containerView *ContainerView) ActiveView() BaseView {
	containerView.lock.Lock()
	defer containerView.lock.Unlock()

	if containerView.isEmpty() {
		return containerView
	}

	return containerView.activeChildView()
}

// Title returns the title of the container view
func (containerView *ContainerView) Title() string {
	containerView.lock.Lock()
	defer containerView.lock.Unlock()

	return containerView.title
}

// IsEmpty returns true if this container view has no child views
func (containerView *ContainerView) IsEmpty() bool {
	containerView.lock.Lock()
	defer containerView.lock.Unlock()

	return containerView.isEmpty()
}

func (containerView *ContainerView) isEmpty() bool {
	return len(containerView.childViews) == 0
}

func (containerView *ContainerView) activeChildView() BaseView {
	return containerView.childViews[containerView.activeViewIndex]
}

func (containerView *ContainerView) removeActiveChildView() {
	if containerView.isEmpty() {
		return
	}

	index := containerView.activeViewIndex
	childView := containerView.activeChildView()
	log.Debugf("Removing child view %T at index %v", childView, index)

	containerView.childViews = append(containerView.childViews[:index], containerView.childViews[index+1:]...)
	childViewNum := uint(len(containerView.childViews))

	if index > 0 && index >= childViewNum {
		if childViewNum > 0 {
			containerView.activeViewIndex = childViewNum - 1
		} else {
			containerView.activeViewIndex = 0
		}
	}
}

// NextView changes the active view to the next child view
// Return value is true if the active child view wrapped back to the first
func (containerView *ContainerView) NextView() (wrapped bool) {
	containerView.lock.Lock()
	defer containerView.lock.Unlock()

	return containerView.nextView()
}

func (containerView *ContainerView) nextView() (wrapped bool) {
	if containerView.isEmpty() {
		return
	}

	switch childView := containerView.activeChildView().(type) {
	case WindowView:
		if len(containerView.childViews) > 1 {
			if containerView.activeViewIndex == uint(len(containerView.childViews)-1) {
				wrapped = true
			} else {
				containerView.setActiveViewAndActivateFirstChild(containerView.activeViewIndex + 1)
			}
		} else {
			wrapped = true
		}
	case *ContainerView:
		if childView.NextView() {
			wrapped = containerView.activeViewIndex == uint(len(containerView.childViews)-1)

			if !wrapped {
				containerView.setActiveViewAndActivateFirstChild(containerView.activeViewIndex + 1)
			}
		}
	}

	return
}

// PrevView changes the active child view to the previous child view
// Return value is true if the active child view wrapped back to the last
func (containerView *ContainerView) PrevView() (wrapped bool) {
	containerView.lock.Lock()
	defer containerView.lock.Unlock()

	return containerView.prevView()
}

func (containerView *ContainerView) prevView() (wrapped bool) {
	if containerView.isEmpty() {
		return
	}

	switch childView := containerView.activeChildView().(type) {
	case WindowView:
		if len(containerView.childViews) > 1 {
			if containerView.activeViewIndex == 0 {
				wrapped = true
			} else {
				containerView.setActiveViewAndActivateLastChild(containerView.activeViewIndex - 1)
			}
		} else {
			wrapped = true
		}
	case *ContainerView:
		if childView.PrevView() {
			wrapped = containerView.activeViewIndex == 0

			if !wrapped {
				containerView.setActiveViewAndActivateLastChild(containerView.activeViewIndex - 1)
			}
		}
	}

	return
}

func (containerView *ContainerView) setActiveViewAndActivateFirstChild(activeViewIndex uint) {
	containerView.activeViewIndex = activeViewIndex

	if newChildView, isContainerView := containerView.activeChildView().(*ContainerView); isContainerView {
		newChildView.setActiveViewAndActivateFirstChild(0)
	}
}

func (containerView *ContainerView) setActiveViewAndActivateLastChild(activeViewIndex uint) {
	containerView.activeViewIndex = activeViewIndex

	if newChildView, isContainerView := containerView.activeChildView().(*ContainerView); isContainerView {
		newChildView.setActiveViewAndActivateLastChild(uint(len(newChildView.childViews) - 1))
	}
}

func nextContainerChildView(containerView *ContainerView, action Action) (err error) {
	if containerView.nextView() {
		if containerView.activeViewIndex == uint(len(containerView.childViews)-1) {
			containerView.setActiveViewAndActivateFirstChild(0)
		} else {
			containerView.setActiveViewAndActivateFirstChild(containerView.activeViewIndex + 1)
		}
	}

	containerView.onStateChange(containerView.viewState)
	containerView.channels.UpdateDisplay()

	return
}

func prevContainerChildView(containerView *ContainerView, action Action) (err error) {
	if containerView.prevView() {
		if containerView.activeViewIndex == 0 {
			containerView.setActiveViewAndActivateLastChild(uint(len(containerView.childViews) - 1))
		} else {
			containerView.setActiveViewAndActivateLastChild(containerView.activeViewIndex - 1)
		}
	}

	containerView.onStateChange(containerView.viewState)
	containerView.channels.UpdateDisplay()

	return
}

func toggleFullScreenChildView(containerView *ContainerView, action Action) (err error) {
	containerView.fullScreen = !containerView.fullScreen

	for _, childView := range containerView.childViews {
		if _, isContainerView := childView.(WindowViewCollection); isContainerView {
			if err = childView.HandleAction(action); err != nil {
				break
			}
		}
	}

	containerView.channels.UpdateDisplay()

	return
}

func toggleViewOrientation(containerView *ContainerView, action Action) (err error) {
	if containerView.isEmpty() {
		return
	}

	switch childView := containerView.activeChildView().(type) {
	case WindowView:
		if containerView.orientation == CoVertical {
			containerView.orientation = CoHorizontal
		} else {
			containerView.orientation = CoVertical
		}
	case WindowViewCollection:
		err = childView.HandleAction(action)
	}

	containerView.channels.UpdateDisplay()

	return
}

func addView(containerView *ContainerView, action Action) (err error) {
	if !containerView.isEmpty() {
		if child, isContainerView := containerView.activeChildView().(*ContainerView); isContainerView {
			return child.HandleAction(action)
		}
	}

	args := action.Args

	if len(args) < 1 {
		return fmt.Errorf("Execpted view but received: %v", args)
	}

	newView, ok := args[0].(BaseView)
	if !ok {
		return fmt.Errorf("Execpted second argument to be BaseView but got %T", args[0])
	}

	containerView.addChildView(newView)

	return
}

func splitView(containerView *ContainerView, action Action) (err error) {
	args := action.Args

	if len(args) < 2 {
		return fmt.Errorf("Execpted orientation and view but received: %v", args)
	}

	orientation, ok := args[0].(ContainerOrientation)
	if !ok {
		return fmt.Errorf("Execpted first argument to be orientation but got %T", args[0])
	}

	newView, ok := args[1].(WindowView)
	if !ok {
		return fmt.Errorf("Execpted second argument to be WindowView but got %T", args[1])
	}

	if containerView.isEmpty() {
		containerView.addChildView(newView)
		return
	}

	switch childView := containerView.activeChildView().(type) {
	case WindowView:
		if len(containerView.childViews) < 2 {
			containerView.addChildView(newView)
			containerView.orientation = orientation
		} else {
			newContainer := NewContainerView(containerView.channels, containerView.config)
			newContainer.SetWindowStyleConfig(containerView.styleConfig)
			newContainer.SetOrientation(orientation)
			newContainer.AddChildViews(childView, newView)
			containerView.childViews[containerView.activeViewIndex] = newContainer
			newContainer.OnStateChange(containerView.viewState)
		}

		containerView.channels.UpdateDisplay()
		log.Infof("Created orientation %v split between views %T and %T", orientation, childView, newView)
	case WindowViewCollection:
		err = childView.HandleAction(action)
	}

	return
}

func removeView(containerView *ContainerView, action Action) (err error) {
	if containerView.isEmpty() {
		return
	}

	if childView, isContainerView := containerView.activeChildView().(*ContainerView); isContainerView {
		err = childView.HandleAction(action)

		if childView.isEmpty() {
			containerView.removeActiveChildView()
		}
	} else {
		childView := containerView.activeChildView()
		containerView.removeActiveChildView()
		containerView.channels.ReportEvent(Event{
			EventType: ViewRemovedEvent,
			Args:      []interface{}{childView},
		})
	}

	containerView.onStateChange(containerView.viewState)
	containerView.channels.UpdateDisplay()

	return
}

func childMouseClick(containerView *ContainerView, action Action) (err error) {
	if containerView.isEmpty() {
		return
	}

	mouseEvent, err := GetMouseEventFromAction(action)
	if err != nil {
		return
	}

	for childIndex, childPosition := range containerView.childPositions {
		if mouseEvent.row >= childPosition.startRow && mouseEvent.row < childPosition.startRow+childPosition.viewDimension.rows &&
			mouseEvent.col >= childPosition.startCol && mouseEvent.col < childPosition.startCol+childPosition.viewDimension.cols {
			containerView.activeViewIndex = uint(childIndex)
			mouseEvent.col -= childPosition.startCol
			mouseEvent.row -= childPosition.startRow
			action.Args[0] = mouseEvent
			err = containerView.activeChildView().HandleAction(action)
			break
		}
	}

	containerView.onStateChange(containerView.viewState)
	containerView.channels.UpdateDisplay()

	return
}
