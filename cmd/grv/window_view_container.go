package main

// ChildViewContainer is a container for a window view
type ChildViewContainer interface {
	SetChild(WindowView)
}

// WindowViewContainer is a decorator for the WindowView interface
// and can be used as a place holder for a WindowView
type WindowViewContainer struct {
	child WindowView
}

// NewWindowViewContainer creates a new instance
func NewWindowViewContainer(child WindowView) *WindowViewContainer {
	return &WindowViewContainer{
		child: child,
	}
}

// SetChild sets the underlying child view that is decorated
func (container *WindowViewContainer) SetChild(child WindowView) {
	container.child = child
}

// Initialise is forwarded onto the child view
func (container *WindowViewContainer) Initialise() (err error) {
	if container.child != nil {
		return container.child.Initialise()
	}

	return
}

// Dispose is forwarded onto the child view
func (container *WindowViewContainer) Dispose() {
	if container.child != nil {
		container.child.Dispose()
	}
}

// RenderHelpBar is forwarded onto the child view
func (container *WindowViewContainer) RenderHelpBar(lineBuilder *LineBuilder) (err error) {
	if container.child != nil {
		return container.child.RenderHelpBar(lineBuilder)
	}

	return
}

// HandleEvent is forwarded onto the child view
func (container *WindowViewContainer) HandleEvent(event Event) (err error) {
	if container.child != nil {
		return container.child.HandleEvent(event)
	}

	return
}

// HandleAction is forwarded onto the child view
func (container *WindowViewContainer) HandleAction(action Action) (err error) {
	if container.child != nil {
		return container.child.HandleAction(action)
	}

	return
}

// OnStateChange is forwarded onto the child view
func (container *WindowViewContainer) OnStateChange(viewState ViewState) {
	if container.child != nil {
		container.child.OnStateChange(viewState)
	}
}

// ViewID is forwarded onto the child view
func (container *WindowViewContainer) ViewID() ViewID {
	if container.child != nil {
		return container.child.ViewID()
	}

	return ViewWindowContainer
}

// Render is forwarded onto the child view
func (container *WindowViewContainer) Render(win RenderWindow) (err error) {
	if container.child != nil {
		return container.child.Render(win)
	}

	return
}
