package main

// NewDefaultTheme creates the default theme of grv
func NewDefaultTheme() MutableTheme {
	return &ThemeComponents{
		components: map[ThemeComponentID]*ThemeComponent{
			CmpAllviewSearchMatch: {
				bgcolor: ColorYellow,
				fgcolor: ColorNone,
			},
			CmpAllviewActiveViewSelectedRow: {
				bgcolor: ColorWhite,
				fgcolor: ColorGreen,
			},
			CmpAllviewInactiveViewSelectedRow: {
				bgcolor: ColorNone,
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

// NewColdTheme creates the cold theme of grv
func NewColdTheme() MutableTheme {
	return &ThemeComponents{
		components: map[ThemeComponentID]*ThemeComponent{
			CmpAllviewSearchMatch: {
				bgcolor: ColorYellow,
				fgcolor: ColorNone,
			},
			CmpAllviewActiveViewSelectedRow: {
				bgcolor: ColorWhite,
				fgcolor: ColorBlue,
			},
			CmpAllviewInactiveViewSelectedRow: {
				bgcolor: ColorNone,
				fgcolor: ColorNone,
			},
			CmpCommitviewTitle: {
				bgcolor: ColorNone,
				fgcolor: ColorBlue,
			},
			CmpCommitviewFooter: {
				bgcolor: ColorNone,
				fgcolor: ColorBlue,
			},
			CmpCommitviewShortOid: {
				bgcolor: ColorNone,
				fgcolor: ColorMagenta,
			},
			CmpCommitviewDate: {
				bgcolor: ColorNone,
				fgcolor: ColorCyan,
			},
			CmpCommitviewAuthor: {
				bgcolor: ColorNone,
				fgcolor: ColorCyan,
			},
			CmpCommitviewSummary: {
				bgcolor: ColorNone,
				fgcolor: ColorNone,
			},
			CmpCommitviewTag: {
				bgcolor: ColorNone,
				fgcolor: ColorMagenta,
			},
			CmpCommitviewLocalBranch: {
				bgcolor: ColorNone,
				fgcolor: ColorMagenta,
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
				fgcolor: ColorCyan,
			},
			CmpDiffviewDifflineDiffCommitCommitter: {
				bgcolor: ColorNone,
				fgcolor: ColorCyan,
			},
			CmpDiffviewDifflineDiffCommitCommitterDate: {
				bgcolor: ColorNone,
				fgcolor: ColorCyan,
			},
			CmpDiffviewDifflineDiffCommitSummary: {
				bgcolor: ColorNone,
				fgcolor: ColorMagenta,
			},
			CmpDiffviewDifflineDiffStatsFile: {
				bgcolor: ColorNone,
				fgcolor: ColorNone,
			},
			CmpDiffviewDifflineGitDiffHeader: {
				bgcolor: ColorNone,
				fgcolor: ColorNone,
			},
			CmpDiffviewDifflineGitDiffExtendedHeader: {
				bgcolor: ColorNone,
				fgcolor: ColorNone,
			},
			CmpDiffviewDifflineUnifiedDiffHeader: {
				bgcolor: ColorNone,
				fgcolor: ColorNone,
			},
			CmpDiffviewDifflineHunkStart: {
				bgcolor: ColorNone,
				fgcolor: ColorMagenta,
			},
			CmpDiffviewDifflineHunkHeader: {
				bgcolor: ColorNone,
				fgcolor: ColorMagenta,
			},
			CmpDiffviewDifflineLineAdded: {
				bgcolor: ColorNone,
				fgcolor: ColorGreen,
			},
			CmpDiffviewDifflineLineRemoved: {
				bgcolor: ColorNone,
				fgcolor: ColorRed,
			},
			CmpRefviewTitle: {
				bgcolor: ColorNone,
				fgcolor: ColorBlue,
			},
			CmpRefviewFooter: {
				bgcolor: ColorNone,
				fgcolor: ColorBlue,
			},
			CmpRefviewLocalBranchesHeader: {
				bgcolor: ColorNone,
				fgcolor: ColorCyan,
			},
			CmpRefviewRemoteBranchesHeader: {
				bgcolor: ColorNone,
				fgcolor: ColorCyan,
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
				fgcolor: ColorCyan,
			},
			CmpRefviewTag: {
				bgcolor: ColorNone,
				fgcolor: ColorNone,
			},
			CmpStatusbarviewNormal: {
				bgcolor: ColorCyan,
				fgcolor: ColorWhite,
			},
			CmpHelpbarviewSpecial: {
				bgcolor: ColorNone,
				fgcolor: ColorCyan,
			},
			CmpHelpbarviewNormal: {
				bgcolor: ColorNone,
				fgcolor: ColorNone,
			},
			CmpErrorViewTitle: {
				bgcolor: ColorNone,
				fgcolor: ColorBlue,
			},
			CmpErrorViewFooter: {
				bgcolor: ColorNone,
				fgcolor: ColorBlue,
			},
			CmpErrorViewErrors: {
				bgcolor: ColorRed,
				fgcolor: ColorWhite,
			},
		},
	}
}
