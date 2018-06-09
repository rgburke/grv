package main

import (
	"fmt"
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
	ViewHistory
	ViewStatus
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
	ViewHelp
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
	OnActiveChange(bool)
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

type popupView struct {
	view          WindowView
	viewDimension ViewDimension
	win           *Window
}

// View is the top level view in grv
// All views in grv are children of this view
type View struct {
	views             []WindowViewCollection
	popupViews        []*popupView
	activeViewPos     uint
	grvStatusView     WindowViewCollection
	channels          Channels
	config            Config
	promptActive      bool
	errorView         *ErrorView
	errorViewWin      *Window
	activeViewWin     *Window
	errors            []error
	windowViewFactory *WindowViewFactory
	tabTitles         []string
	activeViewDim     ViewDimension
	lock              sync.Mutex
}

// NewView creates a new instance
func NewView(repoData RepoData, repoController RepoController, channels Channels, config ConfigSetter) (view *View) {
	view = &View{
		views: []WindowViewCollection{
			NewHistoryView(repoData, repoController, channels, config),
			NewStatusView(repoData, repoController, channels, config),
		},
		channels:          channels,
		config:            config,
		windowViewFactory: NewWindowViewFactory(repoData, repoController, channels, config),
	}

	view.grvStatusView = NewGRVStatusView(view, repoData, channels, config)
	view.errorView = NewErrorView()
	view.errorViewWin = NewWindow("errorView", config)
	view.activeViewWin = NewWindow("activeView", config)

	return
}

// Initialise sets up all child views
func (view *View) Initialise() (err error) {
	for _, childView := range view.views {
		if err = childView.Initialise(); err != nil {
			break
		}
	}

	view.OnActiveChange(true)

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
	childView := view.views[view.activeViewPos]
	view.activeViewDim = activeViewDim
	view.lock.Unlock()

	startRow := uint(0)
	if err = view.renderActiveView(activeViewDim.cols); err != nil {
		return
	}

	wins = append(wins, view.activeViewWin)
	startRow++

	activeViewWins, err := childView.Render(activeViewDim)
	if err != nil {
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
	tabTitles := make([]string, len(view.views))
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
		viewDimension := ViewDimension{
			rows: MinUint(popupView.viewDimension.rows, availableViewDimension.rows-2),
			cols: MinUint(popupView.viewDimension.cols, availableViewDimension.cols-2),
		}

		startRow := (availableViewDimension.rows - viewDimension.rows) / 2
		startCol := (availableViewDimension.cols - viewDimension.cols) / 2

		win := popupView.win
		win.Resize(viewDimension)
		win.SetPosition(startRow, startCol)
		win.Clear()

		if err = popupView.view.Render(win); err != nil {
			return
		}

		wins = append(wins, win)
	}

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

	err = view.ActiveView().RenderHelpBar(lineBuilder)

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

	if view.popupViewsActive() {
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
		view.lock.Lock()
		defer view.lock.Unlock()

		err = view.addView(action)
		return
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
	case ActionShowHelpView:
		view.lock.Lock()
		defer view.lock.Unlock()

		return view.showHelpView()
	}

	return view.ActiveView().HandleAction(action)
}

// OnActiveChange updates the active state of the currently active child view
func (view *View) OnActiveChange(active bool) {
	view.lock.Lock()
	defer view.lock.Unlock()

	log.Debugf("View active %v", active)
	view.onActiveChange(active)
}

func (view *View) onActiveChange(active bool) {
	view.activeView().OnActiveChange(active)
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

	return view.activeView()
}

func (view *View) activeView() BaseView {
	if view.popupViewsActive() {
		return view.popupViews[len(view.popupViews)-1].view
	} else if view.promptActive {
		return view.grvStatusView
	}

	return view.views[view.activeViewPos]
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
	view.views[view.activeViewPos].OnActiveChange(false)
	view.grvStatusView.OnActiveChange(true)
	view.promptActive = true
	view.lock.Unlock()

	err = view.grvStatusView.HandleAction(action)

	view.lock.Lock()
	view.promptActive = false
	view.grvStatusView.OnActiveChange(false)
	view.views[view.activeViewPos].OnActiveChange(true)
	view.lock.Unlock()

	view.channels.UpdateDisplay()

	return
}

func (view *View) nextTab() {
	view.activeViewPos++
	view.activeViewPos %= uint(len(view.views))
	view.onActiveChange(true)
	view.channels.UpdateDisplay()
}

func (view *View) prevTab() {
	if view.activeViewPos == 0 {
		view.activeViewPos = uint(len(view.views)) - 1
	} else {
		view.activeViewPos--
	}

	view.onActiveChange(true)
	view.channels.UpdateDisplay()
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
	view.views = append(view.views, containerView)
	view.activeViewPos = uint(len(view.views) - 1)
	view.channels.UpdateDisplay()

	return containerView
}

func (view *View) removeTab() {
	if len(view.views) <= 1 {
		log.Info("No more tabs left. Exiting GRV")
		view.channels.DoAction(Action{ActionType: ActionExit})
		return
	}

	index := view.activeViewPos
	view.views = append(view.views[:index], view.views[index+1:]...)

	if index >= uint(len(view.views)) {
		view.activeViewPos = uint(len(view.views) - 1)
	}

	view.onActiveChange(true)
	view.channels.UpdateDisplay()

	return
}

func (view *View) removeTabIfEmpty() {
	if containerView, isContainerView := view.activeView().(*ContainerView); isContainerView && containerView.IsEmpty() {
		view.removeTab()
	}
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

func (view *View) addView(action Action) (err error) {
	log.Debugf("Adding new view")
	args := action.Args

	if len(args) < 1 {
		return fmt.Errorf("Expected ActionAddViewArgs argument")
	}

	actionAddViewArgs, ok := args[0].(ActionAddViewArgs)
	if !ok {
		return fmt.Errorf("Expected first argument to have type ActionAddViewArgs but found %T", args[0])
	}

	newView, err := view.createView(actionAddViewArgs.CreateViewArgs)
	if err != nil {
		return
	}

	activeChildView := view.views[view.activeViewPos]
	containerView, ok := activeChildView.(*ContainerView)
	if !ok {
		return fmt.Errorf("This view can not be modified")
	}

	log.Infof("Adding view %T to child with index %v", newView, view.activeViewPos)

	containerView.AddChildViews(newView)
	view.onActiveChange(true)
	view.channels.UpdateDisplay()

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
			view.activeViewPos = uint(tabIndex)
			view.onActiveChange(true)
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

	view.addPopupView(&popupView{
		view:          NewContextMenuView(arg.config, view.channels, view.config),
		viewDimension: arg.viewDimension,
		win:           NewWindow(fmt.Sprintf("popupView-%v", len(view.popupViews)), view.config),
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

	commandOutputView := NewCommandOutputView(arg.command, view.channels, view.config)

	view.addPopupView(&popupView{
		view:          commandOutputView,
		viewDimension: arg.viewDimension,
		win:           NewWindow(fmt.Sprintf("popupView-%v", len(view.popupViews)), view.config),
	})

	arg.onCreation(commandOutputView)

	log.Debugf("Created command output view")

	view.channels.UpdateDisplay()

	return
}

func (view *View) addPopupView(popupView *popupView) {
	if !view.popupViewsActive() {
		view.onActiveChange(false)
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
		view.onActiveChange(true)
	}

	view.channels.UpdateDisplay()
}

func (view *View) popupViewsActive() bool {
	return len(view.popupViews) > 0
}

func (view *View) handlePopupViewAction(action Action) (err error) {
	if !view.popupViewsActive() {
		return fmt.Errorf("Expected at least one popup view to be active")
	}

	if action.ActionType == ActionRemoveView {
		view.removePopupView()
		return
	}

	return view.activeView().HandleAction(action)
}

func (view *View) showHelpView() (err error) {
	for childViewIndex, childView := range view.views {
		if childView.Title() == viewHelpViewTitle {
			view.activeViewPos = uint(childViewIndex)
			view.onActiveChange(true)
			view.channels.UpdateDisplay()
			return
		}
	}

	helpView := NewHelpView(view.channels, view.config)
	if err = helpView.Initialise(); err != nil {
		return
	}

	view.addTab(viewHelpViewTitle).AddChildViews(helpView)

	return
}
