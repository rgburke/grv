package main

// ErrorView contains state for the error view
type ErrorView struct {
	errors []error
}

// NewErrorView creates a new instance of the error view
func NewErrorView() *ErrorView {
	return &ErrorView{}
}

// Initialise does nothing
func (errorView *ErrorView) Initialise() (err error) {
	return
}

// SetErrors sets the errors for the error view to display
func (errorView *ErrorView) SetErrors(errors []error) {
	errorView.errors = errors
}

// DisplayRowsRequired calculates the number of rows on the display required to display the errors the error view currently has set
func (errorView *ErrorView) DisplayRowsRequired() uint {
	errorNum := len(errorView.errors)
	if errorNum == 0 {
		return 0
	}

	return uint(errorNum + 2)
}

// Render generates and writes the error view to the provided window
func (errorView *ErrorView) Render(win RenderWindow) (err error) {
	errorNum := uint(len(errorView.errors))

	for i := uint(1); i < win.Rows()-1 && i-1 < errorNum; i++ {
		var lineBuilder *LineBuilder
		if lineBuilder, err = win.LineBuilder(i, 1); err != nil {
			return err
		}

		err = errorView.errors[i-1]
		lineBuilder.AppendWithStyle(CmpErrorViewErrors, " %v", err)
	}

	win.DrawBorder()

	if err = win.SetTitle(CmpErrorViewTitle, "Errors"); err != nil {
		return
	}

	errorText := "Error"
	if errorNum > 1 {
		errorText += "s"
	}

	err = win.SetFooter(CmpErrorViewFooter, "%v %v", errorNum, errorText)

	return
}

// HandleKeyPress does nothing
func (errorView *ErrorView) HandleKeyPress(keystring string) (err error) {
	return
}

// HandleAction does nothing
func (errorView *ErrorView) HandleAction(Action) (err error) {
	return
}

// OnActiveChange does nothing
func (errorView *ErrorView) OnActiveChange(bool) {

}

// ViewID returns the view ID of the error view
func (errorView *ErrorView) ViewID() ViewID {
	return ViewError
}

// RenderStatusBar does nothing
func (errorView *ErrorView) RenderStatusBar(*LineBuilder) (err error) {
	return
}

// RenderHelpBar does nothing
func (errorView *ErrorView) RenderHelpBar(*LineBuilder) (err error) {
	return
}
