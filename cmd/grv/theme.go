package main

// ThemeComponentID represents different components of the display that makes up grv
// The style of each of these components can be customised using themes
type ThemeComponentID int16

// The set of theme components that make up the display of grv
const (
	CmpNone ThemeComponentID = iota

	CmpAllviewDefault
	CmpAllviewSearchMatch
	CmpAllviewActiveViewSelectedRow
	CmpAllviewInactiveViewSelectedRow

	CmpMainviewActiveView
	CmpMainviewNormalView

	CmpRefviewTitle
	CmpRefviewFooter
	CmpRefviewLocalBranchesHeader
	CmpRefviewRemoteBranchesHeader
	CmpRefviewLocalBranch
	CmpRefviewHead
	CmpRefviewRemoteBranch
	CmpRefviewTagsHeader
	CmpRefviewTag

	CmpCommitviewTitle
	CmpCommitviewFooter
	CmpCommitviewShortOid
	CmpCommitviewDate
	CmpCommitviewAuthor
	CmpCommitviewSummary
	CmpCommitviewTag
	CmpCommitviewLocalBranch
	CmpCommitviewRemoteBranch
	CmpCommitviewGraphCommit
	CmpCommitviewGraphMergeCommit
	CmpCommitviewGraphBranch1
	CmpCommitviewGraphBranch2
	CmpCommitviewGraphBranch3
	CmpCommitviewGraphBranch4
	CmpCommitviewGraphBranch5
	CmpCommitviewGraphBranch6
	CmpCommitviewGraphBranch7

	CmpDiffviewTitle
	CmpDiffviewFooter
	CmpDiffviewDifflineNormal
	CmpDiffviewDifflineDiffCommitAuthor
	CmpDiffviewDifflineDiffCommitAuthorDate
	CmpDiffviewDifflineDiffCommitCommitter
	CmpDiffviewDifflineDiffCommitCommitterDate
	CmpDiffviewDifflineDiffCommitMessage
	CmpDiffviewDifflineDiffStatsFile
	CmpDiffviewDifflineGitDiffHeader
	CmpDiffviewDifflineGitDiffExtendedHeader
	CmpDiffviewDifflineUnifiedDiffHeader
	CmpDiffviewDifflineHunkStart
	CmpDiffviewDifflineHunkHeader
	CmpDiffviewDifflineLineAdded
	CmpDiffviewDifflineLineRemoved
	CmpDiffviewFancyDiffLineSeparator
	CmpDiffviewFancyDiffLineFile
	CmpDiffviewFancyDifflineLineAdded
	CmpDiffviewFancyDifflineLineRemoved
	CmpDiffviewFancyDifflineLineAddedChange
	CmpDiffviewFancyDifflineLineRemovedChange
	CmpDiffviewFancyDifflineEmptyLineAdded
	CmpDiffviewFancyDifflineEmptyLineRemoved
	CmpDiffviewFancyDifflineTrailingWhitespace

	CmpGitStatusMessage
	CmpGitStatusStagedTitle
	CmpGitStatusUnstagedTitle
	CmpGitStatusUntrackedTitle
	CmpGitStatusConflictedTitle
	CmpGitStatusStagedFile
	CmpGitStatusUnstagedFile
	CmpGitStatusUntrackedFile
	CmpGitStatusConflictedFile

	CmpHelpViewTitle
	CmpHelpViewIndexTitle
	CmpHelpViewIndexSubTitle
	CmpHelpViewSectionTitle
	CmpHelpViewSectionSubTitle
	CmpHelpViewSectionDescription
	CmpHelpViewSectionCodeBlock
	CmpHelpViewSectionTableHeader
	CmpHelpViewSectionTableRow
	CmpHelpViewSectionTableRowHighlighted
	CmpHelpViewSectionTableCellSeparator
	CmpHelpViewFooter

	CmpStatusbarviewNormal

	CmpHelpbarviewSpecial
	CmpHelpbarviewNormal

	CmpErrorViewTitle
	CmpErrorViewFooter
	CmpErrorViewErrors

	CmpContextMenuTitle
	CmpContextMenuContent
	CmpContextMenuKeyMapping
	CmpContextMenuFooter

	CmpCommandOutputTitle
	CmpCommandOutputCommand
	CmpCommandOutputNormal
	CmpCommandOutputError
	CmpCommandOutputSuccess
	CmpCommandOutputFooter

	CmpMessageBoxTitle
	CmpMessageBoxContent
	CmpMessageBoxSelectedButton

	CmpGRVVariableViewTitle
	CmpGRVVariableViewVariable
	CmpGRVVariableViewValue
	CmpGRVVariableViewFooter

	CmpRemoteViewTitle
	CmpRemoteViewRemote
	CmpRemoteViewFooter

	CmpSummaryViewHeader
	CmpSummaryViewNormal
	CmpSummaryViewBranchAhead
	CmpSummaryViewBranchBehind
	CmpSummaryViewStagedFile
	CmpSummaryViewUnstagedFile
	CmpSummaryViewNoModifiedFiles

	CmpCount
)

// ThemeColor is a color that can be specified for a theme
type ThemeColor interface {
	themeColor()
}

// SystemColorValue represents one of the 8 basic system colors
type SystemColorValue int

// The set of SystemColorValues
const (
	ColorNone SystemColorValue = iota
	ColorBlack
	ColorRed
	ColorGreen
	ColorYellow
	ColorBlue
	ColorMagenta
	ColorCyan
	ColorWhite
)

// SystemColor stores the value of a system color
type SystemColor struct {
	systemColorValue SystemColorValue
}

// NewSystemColor creates a new instance
func NewSystemColor(systemColorValue SystemColorValue) ThemeColor {
	return &SystemColor{
		systemColorValue: systemColorValue,
	}
}

func (systemColor *SystemColor) themeColor() {}

// ColorNumber stores the terminal color number (0 - 255) for a color
type ColorNumber struct {
	number int16
}

// NewColorNumber creates a new instance
func NewColorNumber(number int16) ThemeColor {
	return &ColorNumber{
		number: number,
	}
}

func (colorNumber *ColorNumber) themeColor() {}

// RGBColor stores the red, green and blue components of a color
type RGBColor struct {
	red   byte
	green byte
	blue  byte
}

// NewRGBColor creates a new instance
func NewRGBColor(red, green, blue byte) ThemeColor {
	return &RGBColor{
		red:   red,
		green: green,
		blue:  blue,
	}
}

func (rgbColor *RGBColor) themeColor() {}

// ThemeStyleType represents a type of styling applied to text
type ThemeStyleType int

// The set of supported text styles
const (
	TstNormal   ThemeStyleType = 0
	TstStandout ThemeStyleType = 1 << (iota - 1)
	TstUnderline
	TstReverse
	TstBlink
	TstDim
	TstBold
	TstProtect
	TstInvis
	TstAltcharset
	TstChartext
)

// ThemeStyle contains styles that should be applied to text
type ThemeStyle struct {
	styleTypes ThemeStyleType
}

// ThemeComponent stores the color information for a theme component
type ThemeComponent struct {
	bgcolor ThemeColor
	fgcolor ThemeColor
	style   ThemeStyle
}

// Theme provides read only access to the style information of a theme
type Theme interface {
	GetComponent(ThemeComponentID) ThemeComponent
	GetAllComponents() map[ThemeComponentID]ThemeComponent
}

// MutableTheme allows defaults to be set on a theme that has not defined color information for all theme components
type MutableTheme interface {
	Theme
	CreateOrGetComponent(ThemeComponentID) *ThemeComponent
}

// ThemeComponents stores all of the style information for a theme
type ThemeComponents struct {
	components map[ThemeComponentID]*ThemeComponent
}

// GetComponent returns the configured color information for the specified theme component ID
func (themeComponents *ThemeComponents) GetComponent(themeComponentID ThemeComponentID) ThemeComponent {
	if themeComponent, ok := themeComponents.components[themeComponentID]; ok {
		return *themeComponent
	}

	return getDefaultThemeComponent()
}

// GetAllComponents returns all configured color information the theme contains
func (themeComponents *ThemeComponents) GetAllComponents() map[ThemeComponentID]ThemeComponent {
	components := make(map[ThemeComponentID]ThemeComponent, CmpCount)

	for themeComponentID := ThemeComponentID(1); themeComponentID < CmpCount; themeComponentID++ {
		themeComponent := themeComponents.GetComponent(themeComponentID)
		components[themeComponentID] = themeComponent
	}

	return components
}

// CreateOrGetComponent returns the configured info if the component has been defined on this theme
// Otherwise a component is created and set on the theme using default values. This default is then returned
func (themeComponents *ThemeComponents) CreateOrGetComponent(themeComponentID ThemeComponentID) *ThemeComponent {
	themeComponent, ok := themeComponents.components[themeComponentID]

	if !ok {
		defultThemeComponent := getDefaultThemeComponent()
		themeComponent = &defultThemeComponent
		themeComponents.components[themeComponentID] = themeComponent
	}

	return themeComponent
}

func getDefaultThemeComponent() ThemeComponent {
	return ThemeComponent{
		bgcolor: NewSystemColor(ColorNone),
		fgcolor: NewSystemColor(ColorNone),
	}
}

// NewTheme creates a new empty theme
func NewTheme() MutableTheme {
	return &ThemeComponents{
		components: make(map[ThemeComponentID]*ThemeComponent),
	}
}
