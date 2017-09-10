package main

// ThemeComponentID represents different components of the display that makes up grv
// The style of each of these components can be customised using themes
type ThemeComponentID int16

// The set of theme components that make up the display of grv
const (
	CmpNone ThemeComponentID = iota

	CmpAllviewSearchMatch

	CmpRefviewTitle
	CmpRefviewFooter
	CmpRefviewLocalBranchesHeader
	CmpRefviewRemoteBranchesHeader
	CmpRefviewLocalBranch
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

	CmpDiffviewDifflineNormal
	CmpDiffviewDifflineDiffCommitAuthor
	CmpDiffviewDifflineDiffCommitAuthorDate
	CmpDiffviewDifflineDiffCommitCommitter
	CmpDiffviewDifflineDiffCommitCommitterDate
	CmpDiffviewDifflineDiffCommitSummary
	CmpDiffviewDifflineDiffStatsFile
	CmpDiffviewDifflineGitDiffHeader
	CmpDiffviewDifflineGitDiffExtendedHeader
	CmpDiffviewDifflineUnifiedDiffHeader
	CmpDiffviewDifflineHunkStart
	CmpDiffviewDifflineHunkHeader
	CmpDiffviewDifflineLineAdded
	CmpDiffviewDifflineLineRemoved

	CmpStatusbarviewNormal

	CmpHelpbarviewSpecial
	CmpHelpbarviewNormal

	CmpErrorViewTitle
	CmpErrorViewFooter
	CmpErrorViewErrors

	CmpCount
)

// ThemeColor is a display color that grv supports
type ThemeColor int

// The set of display colors grv supports
const (
	ColorNone ThemeColor = iota
	ColorBlack
	ColorRed
	ColorGreen
	ColorYellow
	ColorBlue
	ColorMagenta
	ColorCyan
	ColorWhite
)

// ThemeComponent stores the color information for a theme component
type ThemeComponent struct {
	bgcolor ThemeColor
	fgcolor ThemeColor
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
		bgcolor: ColorNone,
		fgcolor: ColorNone,
	}
}

// NewTheme creates a new empty theme
func NewTheme() MutableTheme {
	return &ThemeComponents{
		components: make(map[ThemeComponentID]*ThemeComponent),
	}
}
