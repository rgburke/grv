package main

const (
	evMaxErrorDisplayNum = uint(5)
)

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
	errorNum := uint(len(errorView.errors))

	var rowsRequired uint
	switch {
	case errorNum == 0:
		return 0
	case errorNum > evMaxErrorDisplayNum:
		rowsRequired = evMaxErrorDisplayNum + 1
	default:
		rowsRequired = errorNum
	}

	return rowsRequired + 2
}

// Render generates and writes the error view to the provided window
func (errorView *ErrorView) Render(win RenderWindow) (err error) {
	errorNum := uint(len(errorView.errors))
	errorDisplayNum := MinUInt(errorNum, evMaxErrorDisplayNum)

	var lineBuilder *LineBuilder
	for i := uint(1); i < win.Rows()-1 && i-1 < errorDisplayNum; i++ {
		if lineBuilder, err = win.LineBuilder(i, 1); err != nil {
			return err
		}

		err = errorView.errors[i-1]
		lineBuilder.AppendWithStyle(CmpErrorViewErrors, " %v", err)
	}

	if errorDisplayNum < errorNum && win.Rows() >= evMaxErrorDisplayNum+3 {
		if lineBuilder, err = win.LineBuilder(win.Rows()-2, 1); err != nil {
			return
		}

		errorsNotDisplayed := errorNum - errorDisplayNum
		suffix := ""
		if errorsNotDisplayed > 1 {
			suffix += "s"
		}

		lineBuilder.AppendWithStyle(CmpNone, " + %v more error%v", errorsNotDisplayed, suffix)
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

// HandleEvent does nothing
func (errorView *ErrorView) HandleEvent(event Event) (err error) {
	return
}

// HandleAction does nothing
func (errorView *ErrorView) HandleAction(Action) (err error) {
	return
}

// OnStateChange does nothing
func (errorView *ErrorView) OnStateChange(ViewState) {

}

// ViewID returns the view ID of the error view
func (errorView *ErrorView) ViewID() ViewID {
	return ViewError
}

// RenderHelpBar does nothing
func (errorView *ErrorView) RenderHelpBar(*LineBuilder) (err error) {
	return
}
