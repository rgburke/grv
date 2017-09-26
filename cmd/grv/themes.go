package main

// NewDefaultTheme creates the default theme of grv
func NewDefaultTheme() MutableTheme {
	return &ThemeComponents{
		components: map[ThemeComponentID]*ThemeComponent{
			CmpAllviewDefault: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorNone),
			},
			CmpAllviewSearchMatch: {
				bgcolor: NewSystemColor(ColorYellow),
				fgcolor: NewSystemColor(ColorNone),
			},
			CmpAllviewActiveViewSelectedRow: {
				bgcolor: NewSystemColor(ColorWhite),
				fgcolor: NewSystemColor(ColorGreen),
			},
			CmpAllviewInactiveViewSelectedRow: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorNone),
			},
			CmpCommitviewTitle: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorCyan),
			},
			CmpCommitviewFooter: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorCyan),
			},
			CmpCommitviewShortOid: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorYellow),
			},
			CmpCommitviewDate: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorBlue),
			},
			CmpCommitviewAuthor: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorGreen),
			},
			CmpCommitviewSummary: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorNone),
			},
			CmpCommitviewTag: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorRed),
			},
			CmpCommitviewLocalBranch: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorCyan),
			},
			CmpCommitviewRemoteBranch: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorMagenta),
			},
			CmpDiffviewDifflineNormal: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorNone),
			},
			CmpDiffviewDifflineDiffCommitAuthor: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorCyan),
			},
			CmpDiffviewDifflineDiffCommitAuthorDate: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorYellow),
			},
			CmpDiffviewDifflineDiffCommitCommitter: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorMagenta),
			},
			CmpDiffviewDifflineDiffCommitCommitterDate: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorYellow),
			},
			CmpDiffviewDifflineDiffCommitSummary: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorNone),
			},
			CmpDiffviewDifflineDiffStatsFile: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorBlue),
			},
			CmpDiffviewDifflineGitDiffHeader: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorYellow),
			},
			CmpDiffviewDifflineGitDiffExtendedHeader: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorBlue),
			},
			CmpDiffviewDifflineUnifiedDiffHeader: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorYellow),
			},
			CmpDiffviewDifflineHunkStart: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorCyan),
			},
			CmpDiffviewDifflineHunkHeader: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorBlue),
			},
			CmpDiffviewDifflineLineAdded: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorGreen),
			},
			CmpDiffviewDifflineLineRemoved: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorRed),
			},
			CmpRefviewTitle: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorCyan),
			},
			CmpRefviewFooter: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorCyan),
			},
			CmpRefviewLocalBranchesHeader: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorMagenta),
			},
			CmpRefviewRemoteBranchesHeader: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorMagenta),
			},
			CmpRefviewLocalBranch: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorNone),
			},
			CmpRefviewRemoteBranch: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorNone),
			},
			CmpRefviewTagsHeader: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorMagenta),
			},
			CmpRefviewTag: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorNone),
			},
			CmpStatusbarviewNormal: {
				bgcolor: NewSystemColor(ColorBlue),
				fgcolor: NewSystemColor(ColorYellow),
			},
			CmpHelpbarviewSpecial: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorMagenta),
			},
			CmpHelpbarviewNormal: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorNone),
			},
			CmpErrorViewTitle: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorCyan),
			},
			CmpErrorViewFooter: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorCyan),
			},
			CmpErrorViewErrors: {
				bgcolor: NewSystemColor(ColorRed),
				fgcolor: NewSystemColor(ColorWhite),
			},
		},
	}
}

// NewColdTheme creates the cold theme of grv
func NewColdTheme() MutableTheme {
	return &ThemeComponents{
		components: map[ThemeComponentID]*ThemeComponent{
			CmpAllviewDefault: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorNone),
			},
			CmpAllviewSearchMatch: {
				bgcolor: NewSystemColor(ColorYellow),
				fgcolor: NewSystemColor(ColorNone),
			},
			CmpAllviewActiveViewSelectedRow: {
				bgcolor: NewSystemColor(ColorWhite),
				fgcolor: NewSystemColor(ColorBlue),
			},
			CmpAllviewInactiveViewSelectedRow: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorNone),
			},
			CmpCommitviewTitle: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorBlue),
			},
			CmpCommitviewFooter: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorBlue),
			},
			CmpCommitviewShortOid: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorMagenta),
			},
			CmpCommitviewDate: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorCyan),
			},
			CmpCommitviewAuthor: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorCyan),
			},
			CmpCommitviewSummary: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorNone),
			},
			CmpCommitviewTag: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorMagenta),
			},
			CmpCommitviewLocalBranch: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorMagenta),
			},
			CmpCommitviewRemoteBranch: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorMagenta),
			},
			CmpDiffviewDifflineNormal: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorNone),
			},
			CmpDiffviewDifflineDiffCommitAuthor: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorCyan),
			},
			CmpDiffviewDifflineDiffCommitAuthorDate: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorCyan),
			},
			CmpDiffviewDifflineDiffCommitCommitter: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorCyan),
			},
			CmpDiffviewDifflineDiffCommitCommitterDate: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorCyan),
			},
			CmpDiffviewDifflineDiffCommitSummary: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorMagenta),
			},
			CmpDiffviewDifflineDiffStatsFile: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorNone),
			},
			CmpDiffviewDifflineGitDiffHeader: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorNone),
			},
			CmpDiffviewDifflineGitDiffExtendedHeader: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorNone),
			},
			CmpDiffviewDifflineUnifiedDiffHeader: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorNone),
			},
			CmpDiffviewDifflineHunkStart: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorMagenta),
			},
			CmpDiffviewDifflineHunkHeader: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorMagenta),
			},
			CmpDiffviewDifflineLineAdded: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorGreen),
			},
			CmpDiffviewDifflineLineRemoved: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorRed),
			},
			CmpRefviewTitle: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorBlue),
			},
			CmpRefviewFooter: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorBlue),
			},
			CmpRefviewLocalBranchesHeader: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorCyan),
			},
			CmpRefviewRemoteBranchesHeader: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorCyan),
			},
			CmpRefviewLocalBranch: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorNone),
			},
			CmpRefviewRemoteBranch: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorNone),
			},
			CmpRefviewTagsHeader: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorCyan),
			},
			CmpRefviewTag: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorNone),
			},
			CmpStatusbarviewNormal: {
				bgcolor: NewSystemColor(ColorCyan),
				fgcolor: NewSystemColor(ColorWhite),
			},
			CmpHelpbarviewSpecial: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorCyan),
			},
			CmpHelpbarviewNormal: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorNone),
			},
			CmpErrorViewTitle: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorBlue),
			},
			CmpErrorViewFooter: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorBlue),
			},
			CmpErrorViewErrors: {
				bgcolor: NewSystemColor(ColorRed),
				fgcolor: NewSystemColor(ColorWhite),
			},
		},
	}
}
