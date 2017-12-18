package main

import (
	"fmt"
	"sync"

	log "github.com/Sirupsen/logrus"
)

// ContainerOrientation represents the orientation of the child views
type ContainerOrientation int

// Supported container orientations
const (
	CoVertical ContainerOrientation = iota
	CoHorizontal
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
	channels                    *Channels
	config                      Config
	childViews                  []AbstractView
	viewWins                    map[WindowView]*Window
	activeViewIndex             uint
	handlers                    map[ActionType]containerViewHandler
	orientation                 ContainerOrientation
	childViewPositionCalculator ChildViewPositionCalculator
	fullScreen                  bool
	lock                        sync.Mutex
}

// NewContainerView creates a new instance
func NewContainerView(channels *Channels, config Config, orientation ContainerOrientation, childViews []AbstractView) *ContainerView {
	containerView := &ContainerView{
		config:      config,
		channels:    channels,
		orientation: orientation,
		viewWins:    make(map[WindowView]*Window),
		handlers: map[ActionType]containerViewHandler{
			ActionNextView:         nextContainerChildView,
			ActionPrevView:         prevContainerChildView,
			ActionFullScreenView:   toggleFullScreenChildView,
			ActionToggleViewLayout: toggleViewOrientation,
		},
	}

	containerView.childViewPositionCalculator = containerView

	for _, childView := range childViews {
		containerView.AddChildView(childView)
	}

	return containerView
}

// AddChildView adds a new child view to this container
func (containerView *ContainerView) AddChildView(childView AbstractView) {
	containerView.lock.Lock()
	defer containerView.lock.Unlock()

	containerView.childViews = append(containerView.childViews, childView)

	if windowView, isWindowView := childView.(WindowView); isWindowView {
		viewIndex := len(containerView.childViews) - 1
		winID := fmt.Sprintf("%v-%T", viewIndex, windowView)
		win := NewWindow(winID, containerView.config)
		containerView.viewWins[windowView] = win
	}
}

// SetChildViewPositionCalculator sets the child layout calculator for this view
func (containerView *ContainerView) SetChildViewPositionCalculator(childViewPositionCalculator ChildViewPositionCalculator) {
	containerView.lock.Lock()
	defer containerView.lock.Unlock()

	containerView.childViewPositionCalculator = childViewPositionCalculator
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

// HandleKeyPress passes the keystring to the active child view to process
func (containerView *ContainerView) HandleKeyPress(keystring string) (err error) {
	containerView.lock.Lock()
	defer containerView.lock.Unlock()

	if containerView.isEmpty() {
		return
	}

	return containerView.activeChildView().HandleKeyPress(keystring)
}

// HandleAction processes the action if supported or passes it on to the active child view
func (containerView *ContainerView) HandleAction(action Action) (err error) {
	containerView.lock.Lock()
	defer containerView.lock.Unlock()

	if containerView.isEmpty() {
		return
	}

	handler, handlerExists := containerView.handlers[action.ActionType]

	if handlerExists {
		err = handler(containerView, action)
	} else {
		err = containerView.activeChildView().HandleAction(action)
	}

	return
}

// OnActiveChange updates the active state of this container and its child views
func (containerView *ContainerView) OnActiveChange(active bool) {
	containerView.lock.Lock()
	defer containerView.lock.Unlock()

	containerView.onActiveChange(active)
}

func (containerView *ContainerView) onActiveChange(active bool) {
	for index, childView := range containerView.childViews {
		if uint(index) == containerView.activeViewIndex {
			childView.OnActiveChange(active)
		} else {
			childView.OnActiveChange(false)
		}
	}
}

// ViewID returns container view id
func (containerView *ContainerView) ViewID() ViewID {
	return ViewContainer
}

// RenderHelpBar is proxied to the active child view
func (containerView *ContainerView) RenderHelpBar(lineBuilder *LineBuilder) (err error) {
	RenderKeyBindingHelp(containerView.ViewID(), lineBuilder, []ActionMessage{
		{action: ActionNextView, message: "Next View"},
		{action: ActionPrevView, message: "Prev View"},
		{action: ActionFullScreenView, message: "Full Screen"},
		{action: ActionToggleViewLayout, message: "Layout"},
	})

	return
}

// Render determines the layout of all child views, renders them and returns the resulting windows
func (containerView *ContainerView) Render(viewDimension ViewDimension) (wins []*Window, err error) {
	containerView.lock.Lock()
	defer containerView.lock.Unlock()

	if containerView.isEmpty() {
		return
	}

	viewLayoutData := ViewLayoutData{
		viewDimension:   viewDimension,
		fullScreen:      containerView.fullScreen,
		orientation:     containerView.orientation,
		activeViewIndex: containerView.activeViewIndex,
		childViewNum:    uint(len(containerView.childViews)),
	}

	childPositions := containerView.childViewPositionCalculator.CalculateChildViewPositions(&viewLayoutData)

	for childViewIndex, childView := range containerView.childViews {
		childPosition := childPositions[childViewIndex]
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

// ActiveView returns the active child view
func (containerView *ContainerView) ActiveView() AbstractView {
	containerView.lock.Lock()
	defer containerView.lock.Unlock()

	if containerView.isEmpty() {
		return containerView
	}

	return containerView.activeChildView()
}

// Title returns the title of the container view
func (containerView *ContainerView) Title() string {
	return "Container View"
}

func (containerView *ContainerView) isEmpty() bool {
	return len(containerView.childViews) == 0
}

func (containerView *ContainerView) activeChildView() AbstractView {
	return containerView.childViews[containerView.activeViewIndex]
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
		newChildView.setActiveViewAndActivateLastChild(uint(len(containerView.childViews) - 1))
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

	containerView.onActiveChange(true)
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

	containerView.onActiveChange(true)
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
