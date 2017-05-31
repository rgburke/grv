package main

type ErrorView struct {
	errors []error
}

func NewErrorView() *ErrorView {
	return &ErrorView{}
}

func (errorView *ErrorView) Initialise() (err error) {
	return
}

func (errorView *ErrorView) SetErrors(errors []error) {
	errorView.errors = errors
}

func (errorView *ErrorView) DisplayRowsRequired() uint {
	errorNum := len(errorView.errors)
	if errorNum == 0 {
		return 0
	}

	return uint(errorNum + 2)
}

func (errorView *ErrorView) Render(win RenderWindow) (err error) {
	errorNum := uint(len(errorView.errors))

	for i := uint(1); i < win.Rows()-1 && i-1 < errorNum; i++ {
		lineBuilder, err := win.LineBuilder(i, 1)
		if err != nil {
			return err
		}

		err = errorView.errors[i-1]
		lineBuilder.AppendWithStyle(CMP_ERROR_VIEW_ERRORS, " %v", err)
	}

	win.DrawBorder()

	win.SetTitle(CMP_ERROR_VIEW_TITLE, "Errors")

	errorText := "Error"
	if errorNum > 1 {
		errorText += "s"
	}

	win.SetFooter(CMP_ERROR_VIEW_FOOTER, "%v %v", errorNum, errorText)

	return
}

func (errorView *ErrorView) HandleKeyPress(keystring string) (err error) {
	return
}

func (errorView *ErrorView) HandleAction(Action) (err error) {
	return
}

func (errorView *ErrorView) OnActiveChange(bool) {
	return
}

func (errorView *ErrorView) ViewId() ViewId {
	return VIEW_ERROR
}

func (errorView *ErrorView) RenderStatusBar(*LineBuilder) (err error) {
	return
}

func (errorView *ErrorView) RenderHelpBar(*LineBuilder) (err error) {
	return
}
