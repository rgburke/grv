package main

// ThemeComponentID represents different components of the display that makes up grv
// The style of each of these components can be customised using themes
type ThemeComponentID int16

// The set of theme components that make up the display of grv
const (
	CmpNone ThemeComponentID = iota

	CmpAllviewSearchMatch

	CmpMainviewActiveView
	CmpMainviewNormalView

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

// NewDefaultTheme creates the default theme of grv
func NewDefaultTheme() MutableTheme {
	return &ThemeComponents{
		components: map[ThemeComponentID]*ThemeComponent{
			CmpAllviewSearchMatch: {
				bgcolor: ColorYellow,
				fgcolor: ColorNone,
			},
			CmpCommitviewTitle: {
				bgcolor: ColorNone,
				fgcolor: ColorCyan,
			},
			CmpCommitviewFooter: {
				bgcolor: ColorNone,
				fgcolor: ColorCyan,
			},
			CmpCommitviewShortOid: {
				bgcolor: ColorNone,
				fgcolor: ColorYellow,
			},
			CmpCommitviewDate: {
				bgcolor: ColorNone,
				fgcolor: ColorBlue,
			},
			CmpCommitviewAuthor: {
				bgcolor: ColorNone,
				fgcolor: ColorGreen,
			},
			CmpCommitviewSummary: {
				bgcolor: ColorNone,
				fgcolor: ColorNone,
			},
			CmpCommitviewTag: {
				bgcolor: ColorNone,
				fgcolor: ColorRed,
			},
			CmpCommitviewLocalBranch: {
				bgcolor: ColorNone,
				fgcolor: ColorCyan,
			},
			CmpCommitviewRemoteBranch: {
				bgcolor: ColorNone,
				fgcolor: ColorMagenta,
			},
			CmpDiffviewDifflineNormal: {
				bgcolor: ColorNone,
				fgcolor: ColorNone,
			},
			CmpDiffviewDifflineDiffCommitAuthor: {
				bgcolor: ColorNone,
				fgcolor: ColorCyan,
			},
			CmpDiffviewDifflineDiffCommitAuthorDate: {
				bgcolor: ColorNone,
				fgcolor: ColorYellow,
			},
			CmpDiffviewDifflineDiffCommitCommitter: {
				bgcolor: ColorNone,
				fgcolor: ColorMagenta,
			},
			CmpDiffviewDifflineDiffCommitCommitterDate: {
				bgcolor: ColorNone,
				fgcolor: ColorYellow,
			},
			CmpDiffviewDifflineDiffCommitSummary: {
				bgcolor: ColorNone,
				fgcolor: ColorNone,
			},
			CmpDiffviewDifflineDiffStatsFile: {
				bgcolor: ColorNone,
				fgcolor: ColorBlue,
			},
			CmpDiffviewDifflineGitDiffHeader: {
				bgcolor: ColorNone,
				fgcolor: ColorYellow,
			},
			CmpDiffviewDifflineGitDiffExtendedHeader: {
				bgcolor: ColorNone,
				fgcolor: ColorBlue,
			},
			CmpDiffviewDifflineUnifiedDiffHeader: {
				bgcolor: ColorNone,
				fgcolor: ColorYellow,
			},
			CmpDiffviewDifflineHunkStart: {
				bgcolor: ColorNone,
				fgcolor: ColorCyan,
			},
			CmpDiffviewDifflineHunkHeader: {
				bgcolor: ColorNone,
				fgcolor: ColorBlue,
			},
			CmpDiffviewDifflineLineAdded: {
				bgcolor: ColorNone,
				fgcolor: ColorGreen,
			},
			CmpDiffviewDifflineLineRemoved: {
				bgcolor: ColorNone,
				fgcolor: ColorRed,
			},
			CmpMainviewActiveView: {
				bgcolor: ColorWhite,
				fgcolor: ColorBlue,
			},
			CmpMainviewNormalView: {
				bgcolor: ColorBlue,
				fgcolor: ColorWhite,
			},
			CmpRefviewTitle: {
				bgcolor: ColorNone,
				fgcolor: ColorCyan,
			},
			CmpRefviewFooter: {
				bgcolor: ColorNone,
				fgcolor: ColorCyan,
			},
			CmpRefviewLocalBranchesHeader: {
				bgcolor: ColorNone,
				fgcolor: ColorMagenta,
			},
			CmpRefviewRemoteBranchesHeader: {
				bgcolor: ColorNone,
				fgcolor: ColorMagenta,
			},
			CmpRefviewLocalBranch: {
				bgcolor: ColorNone,
				fgcolor: ColorNone,
			},
			CmpRefviewRemoteBranch: {
				bgcolor: ColorNone,
				fgcolor: ColorNone,
			},
			CmpRefviewTagsHeader: {
				bgcolor: ColorNone,
				fgcolor: ColorMagenta,
			},
			CmpRefviewTag: {
				bgcolor: ColorNone,
				fgcolor: ColorNone,
			},
			CmpStatusbarviewNormal: {
				bgcolor: ColorBlue,
				fgcolor: ColorYellow,
			},
			CmpHelpbarviewSpecial: {
				bgcolor: ColorNone,
				fgcolor: ColorMagenta,
			},
			CmpHelpbarviewNormal: {
				bgcolor: ColorNone,
				fgcolor: ColorNone,
			},
			CmpErrorViewTitle: {
				bgcolor: ColorNone,
				fgcolor: ColorCyan,
			},
			CmpErrorViewFooter: {
				bgcolor: ColorNone,
				fgcolor: ColorCyan,
			},
			CmpErrorViewErrors: {
				bgcolor: ColorRed,
				fgcolor: ColorWhite,
			},
		},
	}
}
