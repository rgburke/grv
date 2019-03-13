package main

import (
	"fmt"
	"strings"
	"sync"

	log "github.com/Sirupsen/logrus"
)

const (
	viewMinActiveViewRows = 6
	viewHelpViewTitle     = "Help View"
)

// ViewID is an ID assigned to each view in grv
type ViewID int

// The set of view IDs
const (
	ViewAll ViewID = iota
	ViewMain
	ViewContainer
	ViewWindowContainer
	ViewHistory
	ViewStatus
	ViewSummary
	ViewGRVStatus
	ViewRef
	ViewCommit
	ViewDiff
	ViewStatusBar
	ViewHelpBar
	ViewError
	ViewGitStatus
	ViewContextMenu
	ViewCommandOutput
	ViewMessageBox
	ViewHelp
	ViewGRVVariable
	ViewRemote
	ViewGitSummary

	ViewCount // i.e. Number of views
)

// ViewState represents the current state of the view
type ViewState int

// The set of view states
const (
	ViewStateInvisible ViewState = iota
	ViewStateInactiveAndVisible
	ViewStateActive
)

// HelpRenderer renders help information
type HelpRenderer interface {
	RenderHelpBar(*LineBuilder) error
}

// BaseView exposes common functionality amongst all views
type BaseView interface {
	HelpRenderer
	EventListener
	Initialise() error
	Dispose()
	HandleAction(Action) error
	OnStateChange(ViewState)
	ViewID() ViewID
}

// WindowView is a single window view
type WindowView interface {
	BaseView
	Render(RenderWindow) error
}

// WindowViewCollection is a view that contains multiple child views
type WindowViewCollection interface {
	BaseView
	Render(ViewDimension) ([]*Window, error)
	ActiveView() BaseView
	Title() string
}

// ViewDimension describes the size of a view
type ViewDimension struct {
	rows uint
	cols uint
}

// String returns a string representation of the view dimensions
func (viewDimension ViewDimension) String() string {
	return fmt.Sprintf("rows:%v,cols:%v", viewDimension.rows, viewDimension.cols)
}

// RegisterViewListener is a function which registers an observer on a view
type RegisterViewListener func(observer interface{}) error

// ViewDimensionSupplier provides view dimensions
type ViewDimensionSupplier interface {
	ViewDimension() ViewDimension
}

type popupView interface {
	ViewDimensionSupplier
	windowView() WindowView
	window() *Window
}

type abstractPopupView struct {
	view WindowView
	win  *Window
}

func (abstractPopupView *abstractPopupView) windowView() WindowView {
	return abstractPopupView.view
}

func (abstractPopupView *abstractPopupView) window() *Window {
	return abstractPopupView.win
}

type fixedSizePopupView struct {
	*abstractPopupView
	viewDimension ViewDimension
}

func (fixedSizePopupView *fixedSizePopupView) ViewDimension() ViewDimension {
	return fixedSizePopupView.viewDimension
}

type dynamicSizePopupView struct {
	*abstractPopupView
	viewDimensionSupplier ViewDimensionSupplier
}

func (dynamicSizePopupView *dynamicSizePopupView) ViewDimension() ViewDimension {
	return dynamicSizePopupView.viewDimensionSupplier.ViewDimension()
}

type activeView struct {
	childView BaseView
}

func (activeView *activeView) isPresent() bool {
	return activeView.childView != nil
}

func (activeView *activeView) get() BaseView {
	return activeView.childView
}

func (activeView *activeView) ifPresent(onPresent func(childView BaseView)) *activeView {
	if activeView.isPresent() {
		onPresent(activeView.childView)
	}

	return activeView
}

func (activeView *activeView) orElse(defaultChildView BaseView) (childView BaseView) {
	if activeView.isPresent() {
		return activeView.childView
	}

	return defaultChildView
}

var popupViewPassThroughActions = map[ActionType]bool{
	ActionShowStatus:              true,
	ActionCreateContextMenu:       true,
	ActionCreateCommandOutputView: true,
	ActionCreateMessageBoxView:    true,
}

// View is the top level view in grv
// All views in grv are children of this view
type View struct {
	views             []WindowViewCollection
	popupViews        []popupView
	activeViewPos     uint
	grvStatusView     WindowViewCollection
	channels          Channels
	config            ConfigSetter
	variables         GRVVariableSetter
	repoData          RepoData
	repoController    RepoController
	promptActive      bool
	errorView         *ErrorView
	errorViewWin      *Window
	activeViewWin     *Window
	emptyViewWin      *Window
	errors            []error
	windowViewFactory *WindowViewFactory
	tabTitles         []string
	activeViewDim     ViewDimension
	lock              sync.Mutex
}

// NewView creates a new instance
func NewView(repoData RepoData, repoController RepoController, channels Channels, config ConfigSetter, variables GRVVariableSetter) (view *View) {
	view = &View{
		channels:          channels,
		config:            config,
		variables:         variables,
		repoData:          repoData,
		repoController:    repoController,
		windowViewFactory: NewWindowViewFactory(repoData, repoController, channels, config, variables),
	}

	view.grvStatusView = NewGRVStatusView(view, repoData, channels, config)
	view.errorView = NewErrorView()
	view.errorViewWin = NewWindow("errorView", config)
	view.activeViewWin = NewWindow("activeView", config)
	view.emptyViewWin = NewWindow("emptyView", config)

	return
}

// Initialise sets up all child views
func (view *View) Initialise() (err error) {
	if defaultViewGenerator := view.config.GetString(CfDefaultView); defaultViewGenerator != "" {
		if errs := view.config.Evaluate(defaultViewGenerator); len(errs) > 0 {
			log.Errorf("Errors when executing default view command %v", defaultViewGenerator)
			view.channels.ReportErrors(errs)
		}
	} else {
		view.views = append(view.views,
			NewHistoryView(view.repoData, view.repoController, view.channels, view.config, view.variables),
			NewStatusView(view.repoData, view.repoController, view.channels, view.config, view.variables),
		)

		for _, childView := range view.views {
			if err = childView.Initialise(); err != nil {
				break
			}
		}
	}

	view.OnStateChange(ViewStateActive)

	return
}

// Dispose of any resources held by the view
func (view *View) Dispose() {
	view.lock.Lock()
	defer view.lock.Unlock()

	for _, view := range view.views {
		view.Dispose()
	}
}

// Render generates all windows to be drawn to the UI
func (view *View) Render(viewDimension ViewDimension) (wins []*Window, err error) {
	log.Debug("Rendering View")

	if viewDimension.rows < 4 {
		log.Errorf("Terminal is not large enough to render GRV")
		return
	}

	activeViewDim := viewDimension
	activeViewDim.rows -= 3

	statusViewDim := viewDimension
	statusViewDim.rows = 2

	errorViewDim := viewDimension
	errorViewDim.rows = 0

	if len(view.errors) > 0 {
		view.determineErrorViewDimensions(&errorViewDim, &activeViewDim)
	}

	view.lock.Lock()
	activeView := view.activeTabView()
	view.activeViewDim = activeViewDim
	view.lock.Unlock()

	startRow := uint(0)
	if err = view.renderActiveView(activeViewDim.cols); err != nil {
		return
	}

	wins = append(wins, view.activeViewWin)
	startRow++

	var activeViewWins []*Window
	if activeView.isPresent() {
		if childView, ok := activeView.get().(WindowViewCollection); ok {
			if activeViewWins, err = childView.Render(activeViewDim); err != nil {
				return
			}
		}
	} else if activeViewWins, err = view.renderEmptyView(activeViewDim); err != nil {
		return
	}

	for _, win := range activeViewWins {
		win.OffsetPosition(int(startRow), 0)
	}

	wins = append(wins, activeViewWins...)
	startRow += activeViewDim.rows

	if errorViewDim.rows > 0 {
		if wins, err = view.renderErrorView(wins, errorViewDim, activeViewDim); err != nil {
			return
		}

		view.errorViewWin.SetPosition(startRow, 0)
		startRow += errorViewDim.rows
	}

	statusViewWins, err := view.grvStatusView.Render(statusViewDim)
	if err != nil {
		return
	}

	for _, win := range statusViewWins {
		win.OffsetPosition(int(startRow), 0)
	}

	wins = append(wins, statusViewWins...)

	popupViewWins, err := view.renderPopupViews(viewDimension)
	if err != nil {
		return
	}

	wins = append(wins, popupViewWins...)

	return
}

func (view *View) determineErrorViewDimensions(errorViewDim, activeViewDim *ViewDimension) {
	view.errorView.SetErrors(view.errors)
	view.errors = nil

	errorRowsRequired := view.errorView.DisplayRowsRequired()

	if activeViewDim.rows > errorRowsRequired+viewMinActiveViewRows {
		errorViewDim.rows = errorRowsRequired
		activeViewDim.rows -= errorRowsRequired
	} else {
		log.Errorf("Unable to display all %v errors, not enough space", errorRowsRequired)

		if activeViewDim.rows > viewMinActiveViewRows {
			errorViewDim.rows = activeViewDim.rows - viewMinActiveViewRows
			activeViewDim.rows = viewMinActiveViewRows
		} else {
			log.Error("Unable to display any errors")
		}
	}
}

func (view *View) renderErrorView(wins []*Window, errorViewDim, activeViewDim ViewDimension) (allWins []*Window, err error) {
	view.errorViewWin.Resize(errorViewDim)
	view.errorViewWin.Clear()

	if err = view.errorView.Render(view.errorViewWin); err != nil {
		return
	}

	allWins = append(wins, view.errorViewWin)

	return
}

func (view *View) renderActiveView(availableCols uint) (err error) {
	tabTitles := make([]string, view.childViewNum())
	cols := uint(0)

	for index, childView := range view.views {
		tabTitles[index] = fmt.Sprintf(" %v ", childView.Title())
		cols += uint(len(tabTitles)) + 1
	}

	if cols > availableCols {
		maxColsPerView := availableCols / uint(len(tabTitles))

		for index, viewTitle := range tabTitles {
			if uint(len(viewTitle)) > maxColsPerView {
				tabTitles[index] = fmt.Sprintf("%*s ", maxColsPerView-1, tabTitles[index])
			}
		}
	}

	view.lock.Lock()
	view.tabTitles = tabTitles
	view.lock.Unlock()

	win := view.activeViewWin
	win.Resize(ViewDimension{rows: 1, cols: availableCols})
	win.Clear()
	win.SetPosition(0, 0)
	win.ApplyStyle(CmpMainviewNormalView)

	lineBuilder, err := view.activeViewWin.LineBuilder(0, 1)
	if err != nil {
		return
	}

	for index, viewTitle := range tabTitles {
		var themeComponentID ThemeComponentID

		if uint(index) == view.activeViewPos {
			themeComponentID = CmpMainviewActiveView
		} else {
			themeComponentID = CmpMainviewNormalView
		}

		lineBuilder.AppendWithStyle(themeComponentID, "%v", viewTitle)
	}

	return
}

func (view *View) renderPopupViews(availableViewDimension ViewDimension) (wins []*Window, err error) {
	for _, popupView := range view.popupViews {
		popupViewDimension := popupView.ViewDimension()

		viewDimension := ViewDimension{
			rows: MinUInt(popupViewDimension.rows, availableViewDimension.rows-2),
			cols: MinUInt(popupViewDimension.cols, availableViewDimension.cols-2),
		}

		startRow := (availableViewDimension.rows - viewDimension.rows) / 2
		startCol := (availableViewDimension.cols - viewDimension.cols) / 2

		win := popupView.window()
		win.Resize(viewDimension)
		win.SetPosition(startRow, startCol)
		win.Clear()

		if err = popupView.windowView().Render(win); err != nil {
			return
		}

		wins = append(wins, win)
	}

	return
}

func (view *View) renderEmptyView(viewDimension ViewDimension) (wins []*Window, err error) {
	win := view.emptyViewWin
	win.SetPosition(0, 0)
	win.Resize(viewDimension)
	win.Clear()

	message := "No tabs defined to display"
	messageWidth := uint(StringWidth(message))

	messageStartRow := viewDimension.rows / 2
	messageStartCol := (viewDimension.cols - messageWidth) / 2

	if err = win.SetRow(messageStartRow, 1, CmpNone, "%v%v", strings.Repeat(" ", int(messageStartCol)), message); err != nil {
		return
	}

	wins = append(wins, win)

	return
}

// RenderHelpBar renders key binding help to the help bar for this view
func (view *View) RenderHelpBar(lineBuilder *LineBuilder) (err error) {
	view.lock.Lock()
	promptActive := view.promptActive
	view.lock.Unlock()

	if !promptActive && !view.popupViewsActive() {
		RenderKeyBindingHelp(view.ViewID(), lineBuilder, view.config, []ActionMessage{
			{action: ActionPrompt, message: "Cmd Prompt"},
			{action: ActionNextTab, message: "Next Tab"},
			{action: ActionPrevTab, message: "Prev Tab"},
		})
	}

	childView := view.ActiveView()
	if childView != view {
		err = childView.RenderHelpBar(lineBuilder)
	}

	return
}

// HandleEvent passes the event on to all child views
func (view *View) HandleEvent(event Event) (err error) {
	view.lock.Lock()
	defer view.lock.Unlock()

	for _, childView := range view.views {
		if err = childView.HandleEvent(event); err != nil {
			return
		}
	}

	switch event.EventType {
	case ViewRemovedEvent:
		for _, removedView := range event.Args {
			if baseView, ok := removedView.(BaseView); ok {
				baseView.Dispose()
			}
		}
	}

	return
}

// HandleAction checks if this view can handle the action
// If not the action is passed down to child views to handle
func (view *View) HandleAction(action Action) (err error) {
	log.Debugf("View handling action %v", action)

	if view.popupViewsActive() && !isPopupViewPassThroughAction(action) {
		return view.handlePopupViewAction(action)
	}

	if IsPromptAction(action.ActionType) {
		return view.prompt(action)
	}

	switch action.ActionType {
	case ActionShowStatus:
		view.lock.Lock()
		defer view.lock.Unlock()

		err = view.grvStatusView.HandleAction(action)
		return
	case ActionNextTab:
		view.lock.Lock()
		defer view.lock.Unlock()

		view.nextTab()
		return
	case ActionPrevTab:
		view.lock.Lock()
		defer view.lock.Unlock()

		view.prevTab()
		return
	case ActionSelectTabByName:
		view.lock.Lock()
		defer view.lock.Unlock()

		err = view.selectTabByName(action)
		return
	case ActionNewTab:
		view.lock.Lock()
		defer view.lock.Unlock()

		err = view.newTab(action)
		return
	case ActionRemoveTab:
		view.lock.Lock()
		defer view.lock.Unlock()

		view.removeTab()
		return
	case ActionAddView:
		if action, err = view.addView(action); err != nil {
			return
		}
	case ActionSplitView:
		if action, err = view.splitView(action); err != nil {
			return
		}
	case ActionRemoveView:
		if err = view.ActiveView().HandleAction(action); err != nil {
			return
		}

		view.lock.Lock()
		defer view.lock.Unlock()

		view.removeTabIfEmpty()

		return
	case ActionMouseSelect:
		view.lock.Lock()
		var handled bool
		action, handled, err = view.handleMouseClick(action)
		view.lock.Unlock()

		if handled || err != nil {
			return
		}
	case ActionCreateContextMenu:
		view.lock.Lock()
		defer view.lock.Unlock()

		return view.createContextMenuView(action)
	case ActionCreateCommandOutputView:
		view.lock.Lock()
		defer view.lock.Unlock()

		return view.createCommandOutputView(action)
	case ActionCreateMessageBoxView:
		view.lock.Lock()
		defer view.lock.Unlock()

		return view.createMessageBoxView(action)
	case ActionShowHelpView:
		view.lock.Lock()
		defer view.lock.Unlock()

		return view.showHelpView(action)
	}

	return view.ActiveView().HandleAction(action)
}

// OnStateChange updates the active state of the currently active child view
func (view *View) OnStateChange(viewState ViewState) {
	view.lock.Lock()
	defer view.lock.Unlock()

	log.Debugf("View state %v", viewState)
	view.onStateChange(viewState)
}

func (view *View) onStateChange(viewState ViewState) {
	view.activeView().ifPresent(func(childView BaseView) {
		childView.OnStateChange(viewState)
	})
}

// ViewID returns the view ID of this view
func (view *View) ViewID() ViewID {
	return ViewMain
}

// ActiveViewHierarchy generates the currently active view hierarchy and returns the views that define it
func (view *View) ActiveViewHierarchy() []BaseView {
	viewHierarchy := []BaseView{view}
	var parentView WindowViewCollection = view
	var ok bool

	for {
		childView := parentView.ActiveView()
		if childView == parentView {
			break
		}

		viewHierarchy = append(viewHierarchy, childView)

		if parentView, ok = childView.(WindowViewCollection); !ok {
			break
		}
	}

	return viewHierarchy
}

// ActiveViewIDHierarchy generates the currently active view hierarchy and returns the view ID's that define it
func (view *View) ActiveViewIDHierarchy() (viewIds []ViewID) {
	viewHierarchy := view.ActiveViewHierarchy()

	for _, activeView := range viewHierarchy {
		viewIds = append(viewIds, activeView.ViewID())
	}

	return
}

// ActiveView returns the currently active child view
func (view *View) ActiveView() BaseView {
	view.lock.Lock()
	defer view.lock.Unlock()

	return view.activeView().orElse(view)
}

func (view *View) activeView() *activeView {
	var childView BaseView

	if view.popupViewsActive() {
		childView = view.popupViews[len(view.popupViews)-1].windowView()
	} else if view.promptActive {
		childView = view.grvStatusView
	} else if view.childViewNum() > 0 {
		childView = view.views[view.activeViewPos]
	}

	return &activeView{childView: childView}
}

func (view *View) activeTabView() *activeView {
	var childView BaseView

	if view.childViewNum() > 0 {
		childView = view.views[view.activeViewPos]
	}

	return &activeView{childView: childView}
}

func (view *View) childViewNum() uint {
	return uint(len(view.views))
}

// SetErrors sets errors to be displayed in the error view
func (view *View) SetErrors(errors []error) {
	view.lock.Lock()
	defer view.lock.Unlock()

	view.errors = errors
}

// Title returns the title of this view
func (view *View) Title() string {
	return "Main View"
}

// ReportStatus reports the provided status in the status bar
func (view *View) ReportStatus(status string) error {
	return view.grvStatusView.HandleAction(Action{
		ActionType: ActionShowStatus,
		Args:       []interface{}{status},
	})
}

func (view *View) prompt(action Action) (err error) {
	view.lock.Lock()
	view.activeView().ifPresent(func(childView BaseView) {
		childView.OnStateChange(ViewStateInactiveAndVisible)
	})
	view.grvStatusView.OnStateChange(ViewStateActive)
	view.promptActive = true
	view.lock.Unlock()

	err = view.grvStatusView.HandleAction(action)

	view.lock.Lock()
	view.promptActive = false
	view.grvStatusView.OnStateChange(ViewStateInactiveAndVisible)
	view.activeView().ifPresent(func(childView BaseView) {
		childView.OnStateChange(ViewStateActive)
	})
	view.lock.Unlock()

	view.channels.UpdateDisplay()

	return
}

func (view *View) nextTab() {
	if view.childViewNum() < 1 {
		return
	}

	view.onStateChange(ViewStateInvisible)
	view.activeViewPos++
	view.activeViewPos %= view.childViewNum()
	view.onStateChange(ViewStateActive)
	view.channels.UpdateDisplay()
}

func (view *View) prevTab() {
	if view.childViewNum() < 1 {
		return
	}

	view.onStateChange(ViewStateInvisible)

	if view.activeViewPos == 0 {
		view.activeViewPos = view.childViewNum() - 1
	} else {
		view.activeViewPos--
	}

	view.onStateChange(ViewStateActive)
	view.channels.UpdateDisplay()
}

func (view *View) selectTabByName(action Action) (err error) {
	if len(action.Args) == 0 {
		return fmt.Errorf("No tab name provided")
	}

	tabName, ok := action.Args[0].(string)
	if !ok {
		return fmt.Errorf("Expected tab name argument to be of type string, but got %T", action.Args[0])
	}

	for childIndex, child := range view.views {
		if child.Title() == tabName {
			view.onStateChange(ViewStateInvisible)
			view.activeViewPos = uint(childIndex)
			view.onStateChange(ViewStateActive)
			view.channels.UpdateDisplay()
			return
		}
	}

	return fmt.Errorf("No tab with name %v", tabName)
}

func (view *View) newTab(action Action) (err error) {
	if len(action.Args) == 0 {
		err = fmt.Errorf("No tab name provided")
	} else if tabName, ok := action.Args[0].(string); !ok {
		err = fmt.Errorf("Expected tab name argument to be of type string, but got %T", action.Args[0])
	} else {
		view.addTab(tabName)
	}

	return
}

func (view *View) addTab(tabName string) *ContainerView {
	containerView := NewContainerView(view.channels, view.config)
	containerView.SetTitle(tabName)
	view.onStateChange(ViewStateInvisible)

	if view.childViewNum() > 0 {
		view.activeViewPos++
		view.views = append(view.views, nil)
		copy(view.views[view.activeViewPos+1:], view.views[view.activeViewPos:])
		view.views[view.activeViewPos] = containerView
	} else {
		view.views = append(view.views, containerView)
		view.activeViewPos = 0
	}

	view.onStateChange(ViewStateActive)
	view.channels.UpdateDisplay()

	return containerView
}

func (view *View) removeTab() {
	if view.childViewNum() <= 1 {
		log.Info("No more tabs left. Exiting GRV")
		view.channels.DoAction(Action{ActionType: ActionExit})
		return
	}

	view.onStateChange(ViewStateInvisible)
	view.activeView().ifPresent(func(childView BaseView) {
		childView.Dispose()
	})

	index := view.activeViewPos
	view.views = append(view.views[:index], view.views[index+1:]...)

	if index >= view.childViewNum() {
		view.activeViewPos = view.childViewNum() - 1
	} else if index > 0 {
		view.activeViewPos--
	}

	view.onStateChange(ViewStateActive)
	view.channels.UpdateDisplay()

	return
}

func (view *View) removeTabIfEmpty() {
	view.activeView().ifPresent(func(childView BaseView) {
		if containerView, isContainerView := childView.(*ContainerView); isContainerView && containerView.IsEmpty() {
			view.removeTab()
		}
	})
}

func (view *View) createView(createViewArgs CreateViewArgs) (windowView WindowView, err error) {
	if windowView, err = view.windowViewFactory.CreateWindowViewWithArgs(createViewArgs.viewID, createViewArgs.viewArgs); err != nil {
		err = fmt.Errorf("Failed to create new view: %v", err)
		return
	}

	if err = windowView.Initialise(); err != nil {
		err = fmt.Errorf("Failed to initialise new view: %v", err)
		return
	}

	if createViewArgs.registerViewListener != nil {
		err = createViewArgs.registerViewListener(windowView)
	}

	return
}

func (view *View) addView(action Action) (newAction Action, err error) {
	log.Debugf("Adding new view")
	args := action.Args

	if len(args) < 1 {
		err = fmt.Errorf("Expected ActionAddViewArgs argument")
		return
	}

	actionAddViewArgs, ok := args[0].(ActionAddViewArgs)
	if !ok {
		err = fmt.Errorf("Expected first argument to have type ActionAddViewArgs but found %T", args[0])
		return
	}

	newView, err := view.createView(actionAddViewArgs.CreateViewArgs)
	if err != nil {
		return
	}

	newAction = Action{
		ActionType: ActionAddView,
		Args:       []interface{}{newView},
	}

	return
}

func (view *View) splitView(action Action) (newAction Action, err error) {
	log.Debug("Splitting view")
	args := action.Args

	if len(args) < 1 {
		err = fmt.Errorf("Expected ActionSplitViewArgs argument")
		return
	}

	actionSplitViewArgs, ok := args[0].(ActionSplitViewArgs)
	if !ok {
		err = fmt.Errorf("Expected first argument to have type ActionSplitViewArgs but found %T", args[0])
		return
	}

	newView, err := view.createView(actionSplitViewArgs.CreateViewArgs)
	if err != nil {
		return
	}

	newAction = Action{
		ActionType: ActionSplitView,
		Args:       []interface{}{actionSplitViewArgs.orientation, newView},
	}

	return
}

func (view *View) handleMouseClick(action Action) (processedAction Action, handled bool, err error) {
	mouseEvent, err := GetMouseEventFromAction(action)
	if err != nil {
		return
	}

	if mouseEvent.row == 0 {
		view.handleTabClick(mouseEvent.col)
		handled = true
	} else if uint(mouseEvent.row) <= view.activeViewDim.rows {
		mouseEvent.row--
		processedAction = action
		processedAction.Args[0] = mouseEvent
	} else {
		handled = true
	}

	return
}

func (view *View) handleTabClick(col uint) {
	cols := uint(0)

	for tabIndex, tabTitle := range view.tabTitles {
		width := uint(StringWidth(tabTitle))

		if col >= cols && col < cols+width {
			log.Debugf("Tab at index %v selected", tabIndex)
			view.onStateChange(ViewStateInvisible)
			view.activeViewPos = uint(tabIndex)
			view.onStateChange(ViewStateActive)
			view.channels.UpdateDisplay()
			return
		}

		cols += width
	}
}

func (view *View) createContextMenuView(action Action) (err error) {
	if len(action.Args) == 0 {
		return fmt.Errorf("Expected ActionCreateContextMenuArgs argument")
	}

	arg, ok := action.Args[0].(ActionCreateContextMenuArgs)
	if !ok {
		return fmt.Errorf("Expected ActionCreateContextMenuArgs argument but got %T", action.Args[0])
	}

	view.addPopupView(&fixedSizePopupView{
		abstractPopupView: &abstractPopupView{
			view: NewContextMenuView(arg.config, view.channels, view.config, view.variables),
			win:  NewWindow(fmt.Sprintf("popupView-%v", len(view.popupViews)), view.config),
		},
		viewDimension: arg.viewDimension,
	})

	log.Debugf("Created context menu")

	return
}

func (view *View) createCommandOutputView(action Action) (err error) {
	if len(action.Args) == 0 {
		return fmt.Errorf("Expected ActionCreateCommandOutputViewArgs argument")
	}

	arg, ok := action.Args[0].(ActionCreateCommandOutputViewArgs)
	if !ok {
		return fmt.Errorf("Expected ActionCreateCommandOutputViewArgs argument but got %T", action.Args[0])
	}

	commandOutputView := NewCommandOutputView(arg.command, view.channels, view.config, view.variables)

	view.addPopupView(&fixedSizePopupView{
		abstractPopupView: &abstractPopupView{
			view: commandOutputView,
			win:  NewWindow(fmt.Sprintf("popupView-%v", len(view.popupViews)), view.config),
		},
		viewDimension: arg.viewDimension,
	})

	arg.onCreation(commandOutputView)

	log.Debugf("Created command output view")

	view.channels.UpdateDisplay()

	return
}

func (view *View) createMessageBoxView(action Action) (err error) {
	if len(action.Args) == 0 {
		return fmt.Errorf("Expected ActionCreateMessageBoxViewArgs argument")
	}

	arg, ok := action.Args[0].(ActionCreateMessageBoxViewArgs)
	if !ok {
		return fmt.Errorf("Expected ActionCreateMessageBoxViewArgs argument but got %T", action.Args[0])
	}

	messageBoxView := NewMessageBoxView(arg.config, view.channels, view.config, view.variables)

	view.addPopupView(&dynamicSizePopupView{
		abstractPopupView: &abstractPopupView{
			view: messageBoxView,
			win:  NewWindow(fmt.Sprintf("popupView-%v", len(view.popupViews)), view.config),
		},
		viewDimensionSupplier: messageBoxView,
	})

	log.Debugf("Created context menu")

	return
}

func (view *View) addPopupView(popupView popupView) {
	if !view.popupViewsActive() {
		view.onStateChange(ViewStateInactiveAndVisible)
	}

	view.popupViews = append(view.popupViews, popupView)
	log.Debugf("Added popupView. %v popup view(s) active", len(view.popupViews))

	view.channels.UpdateDisplay()
}

func (view *View) removePopupView() {
	removedPopupView := view.popupViews[len(view.popupViews)-1]
	view.popupViews = view.popupViews[:len(view.popupViews)-1]
	log.Debugf("Removed popupView. %v popup view(s) active", len(view.popupViews))

	view.channels.ReportEvent(Event{
		EventType: ViewRemovedEvent,
		Args:      []interface{}{removedPopupView},
	})

	if !view.popupViewsActive() {
		view.onStateChange(ViewStateActive)
	}

	view.channels.UpdateDisplay()
}

func (view *View) popupViewsActive() bool {
	return len(view.popupViews) > 0
}

func isPopupViewPassThroughAction(action Action) (passThrough bool) {
	_, passThrough = popupViewPassThroughActions[action.ActionType]
	return
}

func (view *View) handlePopupViewAction(action Action) (err error) {
	if !view.popupViewsActive() {
		return fmt.Errorf("Expected at least one popup view to be active")
	}

	if action.ActionType == ActionRemoveView {
		view.removePopupView()
		return
	} else if action.ActionType == ActionMouseSelect {
		var handled bool
		if handled, err = view.processMouseEventForPopupView(action); handled || err != nil {
			return
		}
	}

	view.activeView().ifPresent(func(childView BaseView) {
		err = childView.HandleAction(action)
	})

	return
}

func (view *View) processMouseEventForPopupView(action Action) (handled bool, err error) {
	mouseEvent, err := GetMouseEventFromAction(action)
	if err != nil {
		return
	}

	popupView := view.popupViews[len(view.popupViews)-1]
	win := popupView.window()
	startRow, startCol := win.Position()

	if startRow > mouseEvent.row ||
		startCol > mouseEvent.col ||
		mouseEvent.row >= startRow+win.Rows() ||
		mouseEvent.col >= startCol+win.Cols() {
		handled = true
		return
	}

	mouseEvent.row -= startRow
	mouseEvent.col -= startCol
	action.Args[0] = mouseEvent

	return
}

func (view *View) showHelpView(action Action) (err error) {
	var helpView *HelpView

	for childViewIndex, childView := range view.views {
		if childView.Title() == viewHelpViewTitle {
			helpView, _ = childView.ActiveView().(*HelpView)
			view.onStateChange(ViewStateInvisible)
			view.activeViewPos = uint(childViewIndex)
			view.onStateChange(ViewStateActive)
			break
		}
	}

	if helpView == nil {
		helpView = NewHelpView(view.channels, view.config, view.variables)
		if err = helpView.Initialise(); err != nil {
			return
		}

		view.addTab(viewHelpViewTitle).AddChildViews(helpView)
	}

	if len(action.Args) > 0 {
		if searchTerm, ok := action.Args[0].(string); ok {
			helpView.SearchHelp(searchTerm)
		}
	}

	view.channels.UpdateDisplay()

	return
}
