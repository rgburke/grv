package main

// NewClassicTheme creates the classic theme of grv
func NewClassicTheme() MutableTheme {
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
			CmpMainviewActiveView: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorNone),
			},
			CmpMainviewNormalView: {
				bgcolor: NewSystemColor(ColorBlue),
				fgcolor: NewSystemColor(ColorWhite),
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
			CmpCommitviewGraphCommit: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorBlue),
			},
			CmpCommitviewGraphMergeCommit: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorBlue),
			},
			CmpCommitviewGraphBranch1: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorYellow),
			},
			CmpCommitviewGraphBranch2: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorRed),
			},
			CmpCommitviewGraphBranch3: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorGreen),
			},
			CmpCommitviewGraphBranch4: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorMagenta),
			},
			CmpCommitviewGraphBranch5: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorCyan),
			},
			CmpCommitviewGraphBranch6: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorBlue),
			},
			CmpCommitviewGraphBranch7: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorWhite),
			},
			CmpDiffviewTitle: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorCyan),
			},
			CmpDiffviewFooter: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorCyan),
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
			CmpDiffviewDifflineDiffCommitMessage: {
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
			CmpRefviewHead: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorGreen),
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
			CmpGitStatusStagedTitle: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorMagenta),
			},
			CmpGitStatusUnstagedTitle: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorMagenta),
			},
			CmpGitStatusUntrackedTitle: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorMagenta),
			},
			CmpGitStatusConflictedTitle: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorMagenta),
			},
			CmpGitStatusStagedFile: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorGreen),
			},
			CmpGitStatusUnstagedFile: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorRed),
			},
			CmpGitStatusUntrackedFile: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorRed),
			},
			CmpGitStatusConflictedFile: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorRed),
			},
			CmpContextMenuTitle: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorCyan),
			},
			CmpContextMenuContent: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorNone),
			},
			CmpContextMenuFooter: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorCyan),
			},
			CmpCommandOutputTitle: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorCyan),
			},
			CmpCommandOutputNormal: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorNone),
			},
			CmpCommandOutputError: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorRed),
			},
			CmpCommandOutputSuccess: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorGreen),
			},
			CmpCommandOutputFooter: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorCyan),
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
			CmpMainviewActiveView: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorNone),
			},
			CmpMainviewNormalView: {
				bgcolor: NewSystemColor(ColorCyan),
				fgcolor: NewSystemColor(ColorWhite),
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
			CmpCommitviewGraphCommit: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorBlue),
			},
			CmpCommitviewGraphMergeCommit: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorBlue),
			},
			CmpCommitviewGraphBranch1: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorYellow),
			},
			CmpCommitviewGraphBranch2: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorRed),
			},
			CmpCommitviewGraphBranch3: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorGreen),
			},
			CmpCommitviewGraphBranch4: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorMagenta),
			},
			CmpCommitviewGraphBranch5: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorCyan),
			},
			CmpCommitviewGraphBranch6: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorBlue),
			},
			CmpCommitviewGraphBranch7: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorWhite),
			},
			CmpDiffviewTitle: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorBlue),
			},
			CmpDiffviewFooter: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorBlue),
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
			CmpDiffviewDifflineDiffCommitMessage: {
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
			CmpRefviewHead: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorGreen),
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
			CmpGitStatusStagedTitle: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorCyan),
			},
			CmpGitStatusUnstagedTitle: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorCyan),
			},
			CmpGitStatusUntrackedTitle: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorCyan),
			},
			CmpGitStatusConflictedTitle: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorCyan),
			},
			CmpGitStatusStagedFile: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorGreen),
			},
			CmpGitStatusUnstagedFile: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorRed),
			},
			CmpGitStatusUntrackedFile: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorRed),
			},
			CmpGitStatusConflictedFile: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorRed),
			},
			CmpContextMenuTitle: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorBlue),
			},
			CmpContextMenuContent: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorNone),
			},
			CmpContextMenuFooter: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorBlue),
			},
			CmpCommandOutputTitle: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorBlue),
			},
			CmpCommandOutputNormal: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorNone),
			},
			CmpCommandOutputError: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorRed),
			},
			CmpCommandOutputSuccess: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorGreen),
			},
			CmpCommandOutputFooter: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorBlue),
			},
		},
	}
}

// NewSolarizedTheme creates the solarized theme of grv
// Solarized color codes Copyright (c) 2011 Ethan Schoonover
func NewSolarizedTheme() MutableTheme {
	return &ThemeComponents{
		components: map[ThemeComponentID]*ThemeComponent{
			CmpAllviewDefault: {
				bgcolor: NewColorNumber(234),
				fgcolor: NewColorNumber(244),
			},
			CmpAllviewSearchMatch: {
				bgcolor: NewColorNumber(136),
				fgcolor: NewColorNumber(254),
			},
			CmpAllviewActiveViewSelectedRow: {
				bgcolor: NewColorNumber(234),
				fgcolor: NewColorNumber(254),
			},
			CmpAllviewInactiveViewSelectedRow: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewColorNumber(245),
			},
			CmpMainviewActiveView: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewColorNumber(254),
			},
			CmpMainviewNormalView: {
				bgcolor: NewColorNumber(235),
				fgcolor: NewSystemColor(ColorNone),
			},
			CmpCommitviewTitle: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewColorNumber(37),
			},
			CmpCommitviewFooter: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewColorNumber(37),
			},
			CmpCommitviewShortOid: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewColorNumber(136),
			},
			CmpCommitviewDate: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewColorNumber(33),
			},
			CmpCommitviewAuthor: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewColorNumber(64),
			},
			CmpCommitviewSummary: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorNone),
			},
			CmpCommitviewTag: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewColorNumber(160),
			},
			CmpCommitviewLocalBranch: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewColorNumber(37),
			},
			CmpCommitviewRemoteBranch: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewColorNumber(125),
			},
			CmpCommitviewGraphCommit: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewColorNumber(33),
			},
			CmpCommitviewGraphMergeCommit: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewColorNumber(33),
			},
			CmpCommitviewGraphBranch1: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewColorNumber(136),
			},
			CmpCommitviewGraphBranch2: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewColorNumber(160),
			},
			CmpCommitviewGraphBranch3: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewColorNumber(64),
			},
			CmpCommitviewGraphBranch4: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewColorNumber(125),
			},
			CmpCommitviewGraphBranch5: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewColorNumber(37),
			},
			CmpCommitviewGraphBranch6: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewColorNumber(33),
			},
			CmpCommitviewGraphBranch7: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewColorNumber(254),
			},
			CmpDiffviewTitle: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewColorNumber(37),
			},
			CmpDiffviewFooter: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewColorNumber(37),
			},
			CmpDiffviewDifflineNormal: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorNone),
			},
			CmpDiffviewDifflineDiffCommitAuthor: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewColorNumber(61),
			},
			CmpDiffviewDifflineDiffCommitAuthorDate: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewColorNumber(136),
			},
			CmpDiffviewDifflineDiffCommitCommitter: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewColorNumber(125),
			},
			CmpDiffviewDifflineDiffCommitCommitterDate: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewColorNumber(136),
			},
			CmpDiffviewDifflineDiffCommitMessage: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorNone),
			},
			CmpDiffviewDifflineDiffStatsFile: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewColorNumber(33),
			},
			CmpDiffviewDifflineGitDiffHeader: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewColorNumber(136),
			},
			CmpDiffviewDifflineGitDiffExtendedHeader: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewColorNumber(33),
			},
			CmpDiffviewDifflineUnifiedDiffHeader: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewColorNumber(136),
			},
			CmpDiffviewDifflineHunkStart: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewColorNumber(37),
			},
			CmpDiffviewDifflineHunkHeader: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewColorNumber(33),
			},
			CmpDiffviewDifflineLineAdded: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewColorNumber(64),
			},
			CmpDiffviewDifflineLineRemoved: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewColorNumber(160),
			},
			CmpRefviewTitle: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewColorNumber(37),
			},
			CmpRefviewFooter: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewColorNumber(37),
			},
			CmpRefviewLocalBranchesHeader: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewColorNumber(125),
			},
			CmpRefviewRemoteBranchesHeader: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewColorNumber(125),
			},
			CmpRefviewLocalBranch: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorNone),
			},
			CmpRefviewHead: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewColorNumber(64),
			},
			CmpRefviewRemoteBranch: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorNone),
			},
			CmpRefviewTagsHeader: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewColorNumber(125),
			},
			CmpRefviewTag: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorNone),
			},
			CmpStatusbarviewNormal: {
				bgcolor: NewColorNumber(235),
				fgcolor: NewColorNumber(136),
			},
			CmpHelpbarviewSpecial: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewColorNumber(125),
			},
			CmpHelpbarviewNormal: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorNone),
			},
			CmpErrorViewTitle: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewColorNumber(37),
			},
			CmpErrorViewFooter: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewColorNumber(37),
			},
			CmpErrorViewErrors: {
				bgcolor: NewColorNumber(160),
				fgcolor: NewColorNumber(245),
			},
			CmpGitStatusStagedTitle: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewColorNumber(61),
			},
			CmpGitStatusUnstagedTitle: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewColorNumber(61),
			},
			CmpGitStatusUntrackedTitle: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewColorNumber(61),
			},
			CmpGitStatusConflictedTitle: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewColorNumber(61),
			},
			CmpGitStatusStagedFile: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewColorNumber(64),
			},
			CmpGitStatusUnstagedFile: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewColorNumber(160),
			},
			CmpGitStatusUntrackedFile: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewColorNumber(160),
			},
			CmpGitStatusConflictedFile: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewColorNumber(160),
			},
			CmpContextMenuTitle: {
				bgcolor: NewColorNumber(235),
				fgcolor: NewColorNumber(37),
			},
			CmpContextMenuContent: {
				bgcolor: NewColorNumber(235),
				fgcolor: NewColorNumber(254),
			},
			CmpContextMenuFooter: {
				bgcolor: NewColorNumber(235),
				fgcolor: NewColorNumber(37),
			},
			CmpCommandOutputTitle: {
				bgcolor: NewColorNumber(235),
				fgcolor: NewColorNumber(37),
			},
			CmpCommandOutputNormal: {
				bgcolor: NewColorNumber(235),
				fgcolor: NewColorNumber(254),
			},
			CmpCommandOutputError: {
				bgcolor: NewColorNumber(235),
				fgcolor: NewColorNumber(160),
			},
			CmpCommandOutputSuccess: {
				bgcolor: NewColorNumber(235),
				fgcolor: NewColorNumber(64),
			},
			CmpCommandOutputFooter: {
				bgcolor: NewColorNumber(235),
				fgcolor: NewColorNumber(37),
			},
		},
	}
}
