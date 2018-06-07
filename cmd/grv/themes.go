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
			CmpCommandOutputCommand: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorYellow),
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
			CmpHelpViewTitle: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorCyan),
			},
			CmpHelpViewSectionTitle: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorCyan),
				style:   ThemeStyle{styleTypes: TstBold | TstUnderline},
			},
			CmpHelpViewSectionSubTitle: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorMagenta),
				style:   ThemeStyle{styleTypes: TstUnderline},
			},
			CmpHelpViewSectionDescription: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorNone),
			},
			CmpHelpViewSectionTableHeader: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorBlue),
			},
			CmpHelpViewSectionTableRow: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorNone),
			},
			CmpHelpViewSectionTableRowHighlighted: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorYellow),
				style:   ThemeStyle{styleTypes: TstBold},
			},
			CmpHelpViewSectionTableCellSeparator: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorCyan),
			},
			CmpHelpViewFooter: {
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
			CmpCommandOutputCommand: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorYellow),
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
			CmpHelpViewTitle: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorCyan),
			},
			CmpHelpViewSectionTitle: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorCyan),
				style:   ThemeStyle{styleTypes: TstBold | TstUnderline},
			},
			CmpHelpViewSectionSubTitle: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorMagenta),
				style:   ThemeStyle{styleTypes: TstUnderline},
			},
			CmpHelpViewSectionDescription: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorNone),
			},
			CmpHelpViewSectionTableHeader: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorBlue),
			},
			CmpHelpViewSectionTableRow: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorNone),
			},
			CmpHelpViewSectionTableRowHighlighted: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorYellow),
				style:   ThemeStyle{styleTypes: TstBold},
			},
			CmpHelpViewSectionTableCellSeparator: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorCyan),
			},
			CmpHelpViewFooter: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorCyan),
			},
		},
	}
}

const (
	solarizedBrightBlack   = 234
	solarizedBlack         = 235
	solarizedBrightGreen   = 240
	solarizedBrightYellow  = 241
	solarizedBrightBlue    = 244
	solarizedBrightCyan    = 245
	solarizedWhite         = 254
	solarizedBrightWhite   = 230
	solarizedYellow        = 136
	solarizedBrightRed     = 166
	solarizedRed           = 160
	solarizedMagenta       = 125
	solarizedBrightMagenta = 61
	solarizedBlue          = 33
	solarizedCyan          = 37
	solarizedGreen         = 64
)

// NewSolarizedTheme creates the solarized theme of grv
// Solarized color codes Copyright (c) 2011 Ethan Schoonover
func NewSolarizedTheme() MutableTheme {
	return &ThemeComponents{
		components: map[ThemeComponentID]*ThemeComponent{
			CmpAllviewDefault: {
				bgcolor: NewColorNumber(solarizedBrightBlack),
				fgcolor: NewColorNumber(solarizedBrightBlue),
			},
			CmpAllviewSearchMatch: {
				bgcolor: NewColorNumber(solarizedYellow),
				fgcolor: NewColorNumber(solarizedWhite),
			},
			CmpAllviewActiveViewSelectedRow: {
				bgcolor: NewColorNumber(solarizedBrightBlack),
				fgcolor: NewColorNumber(solarizedWhite),
			},
			CmpAllviewInactiveViewSelectedRow: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewColorNumber(solarizedBrightCyan),
			},
			CmpMainviewActiveView: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewColorNumber(solarizedWhite),
			},
			CmpMainviewNormalView: {
				bgcolor: NewColorNumber(solarizedBlack),
				fgcolor: NewSystemColor(ColorNone),
			},
			CmpCommitviewTitle: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewColorNumber(solarizedCyan),
			},
			CmpCommitviewFooter: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewColorNumber(solarizedCyan),
			},
			CmpCommitviewShortOid: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewColorNumber(solarizedYellow),
			},
			CmpCommitviewDate: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewColorNumber(solarizedBlue),
			},
			CmpCommitviewAuthor: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewColorNumber(solarizedGreen),
			},
			CmpCommitviewSummary: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorNone),
			},
			CmpCommitviewTag: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewColorNumber(solarizedRed),
			},
			CmpCommitviewLocalBranch: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewColorNumber(solarizedCyan),
			},
			CmpCommitviewRemoteBranch: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewColorNumber(solarizedMagenta),
			},
			CmpCommitviewGraphCommit: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewColorNumber(solarizedBlue),
			},
			CmpCommitviewGraphMergeCommit: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewColorNumber(solarizedBlue),
			},
			CmpCommitviewGraphBranch1: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewColorNumber(solarizedYellow),
			},
			CmpCommitviewGraphBranch2: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewColorNumber(solarizedRed),
			},
			CmpCommitviewGraphBranch3: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewColorNumber(solarizedGreen),
			},
			CmpCommitviewGraphBranch4: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewColorNumber(solarizedMagenta),
			},
			CmpCommitviewGraphBranch5: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewColorNumber(solarizedCyan),
			},
			CmpCommitviewGraphBranch6: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewColorNumber(solarizedBlue),
			},
			CmpCommitviewGraphBranch7: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewColorNumber(solarizedWhite),
			},
			CmpDiffviewTitle: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewColorNumber(solarizedCyan),
			},
			CmpDiffviewFooter: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewColorNumber(solarizedCyan),
			},
			CmpDiffviewDifflineNormal: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorNone),
			},
			CmpDiffviewDifflineDiffCommitAuthor: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewColorNumber(solarizedBrightMagenta),
			},
			CmpDiffviewDifflineDiffCommitAuthorDate: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewColorNumber(solarizedYellow),
			},
			CmpDiffviewDifflineDiffCommitCommitter: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewColorNumber(solarizedMagenta),
			},
			CmpDiffviewDifflineDiffCommitCommitterDate: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewColorNumber(solarizedYellow),
			},
			CmpDiffviewDifflineDiffCommitMessage: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorNone),
			},
			CmpDiffviewDifflineDiffStatsFile: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewColorNumber(solarizedBlue),
			},
			CmpDiffviewDifflineGitDiffHeader: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewColorNumber(solarizedYellow),
			},
			CmpDiffviewDifflineGitDiffExtendedHeader: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewColorNumber(solarizedBlue),
			},
			CmpDiffviewDifflineUnifiedDiffHeader: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewColorNumber(solarizedYellow),
			},
			CmpDiffviewDifflineHunkStart: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewColorNumber(solarizedCyan),
			},
			CmpDiffviewDifflineHunkHeader: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewColorNumber(solarizedBlue),
			},
			CmpDiffviewDifflineLineAdded: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewColorNumber(solarizedGreen),
			},
			CmpDiffviewDifflineLineRemoved: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewColorNumber(solarizedRed),
			},
			CmpRefviewTitle: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewColorNumber(solarizedCyan),
			},
			CmpRefviewFooter: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewColorNumber(solarizedCyan),
			},
			CmpRefviewLocalBranchesHeader: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewColorNumber(solarizedMagenta),
			},
			CmpRefviewRemoteBranchesHeader: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewColorNumber(solarizedMagenta),
			},
			CmpRefviewLocalBranch: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorNone),
			},
			CmpRefviewHead: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewColorNumber(solarizedGreen),
			},
			CmpRefviewRemoteBranch: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorNone),
			},
			CmpRefviewTagsHeader: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewColorNumber(solarizedMagenta),
			},
			CmpRefviewTag: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorNone),
			},
			CmpStatusbarviewNormal: {
				bgcolor: NewColorNumber(solarizedBlack),
				fgcolor: NewColorNumber(solarizedYellow),
			},
			CmpHelpbarviewSpecial: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewColorNumber(solarizedMagenta),
			},
			CmpHelpbarviewNormal: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorNone),
			},
			CmpErrorViewTitle: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewColorNumber(solarizedCyan),
			},
			CmpErrorViewFooter: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewColorNumber(solarizedCyan),
			},
			CmpErrorViewErrors: {
				bgcolor: NewColorNumber(solarizedRed),
				fgcolor: NewColorNumber(solarizedBrightCyan),
			},
			CmpGitStatusStagedTitle: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewColorNumber(solarizedBrightMagenta),
			},
			CmpGitStatusUnstagedTitle: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewColorNumber(solarizedBrightMagenta),
			},
			CmpGitStatusUntrackedTitle: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewColorNumber(solarizedBrightMagenta),
			},
			CmpGitStatusConflictedTitle: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewColorNumber(solarizedBrightMagenta),
			},
			CmpGitStatusStagedFile: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewColorNumber(solarizedGreen),
			},
			CmpGitStatusUnstagedFile: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewColorNumber(solarizedRed),
			},
			CmpGitStatusUntrackedFile: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewColorNumber(solarizedRed),
			},
			CmpGitStatusConflictedFile: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewColorNumber(solarizedRed),
			},
			CmpContextMenuTitle: {
				bgcolor: NewColorNumber(solarizedBlack),
				fgcolor: NewColorNumber(solarizedCyan),
			},
			CmpContextMenuContent: {
				bgcolor: NewColorNumber(solarizedBlack),
				fgcolor: NewColorNumber(solarizedWhite),
			},
			CmpContextMenuFooter: {
				bgcolor: NewColorNumber(solarizedBlack),
				fgcolor: NewColorNumber(solarizedCyan),
			},
			CmpCommandOutputTitle: {
				bgcolor: NewColorNumber(solarizedBlack),
				fgcolor: NewColorNumber(solarizedCyan),
			},
			CmpCommandOutputCommand: {
				bgcolor: NewColorNumber(solarizedBlack),
				fgcolor: NewColorNumber(solarizedYellow),
			},
			CmpCommandOutputNormal: {
				bgcolor: NewColorNumber(solarizedBlack),
				fgcolor: NewColorNumber(solarizedWhite),
			},
			CmpCommandOutputError: {
				bgcolor: NewColorNumber(solarizedBlack),
				fgcolor: NewColorNumber(solarizedRed),
			},
			CmpCommandOutputSuccess: {
				bgcolor: NewColorNumber(solarizedBlack),
				fgcolor: NewColorNumber(solarizedGreen),
			},
			CmpCommandOutputFooter: {
				bgcolor: NewColorNumber(solarizedBlack),
				fgcolor: NewColorNumber(solarizedCyan),
			},
			CmpHelpViewTitle: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewColorNumber(solarizedCyan),
			},
			CmpHelpViewSectionTitle: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewColorNumber(solarizedBrightMagenta),
				style:   ThemeStyle{styleTypes: TstBold | TstUnderline},
			},
			CmpHelpViewSectionSubTitle: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewColorNumber(solarizedMagenta),
				style:   ThemeStyle{styleTypes: TstUnderline},
			},
			CmpHelpViewSectionDescription: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewSystemColor(ColorNone),
			},
			CmpHelpViewSectionTableHeader: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewColorNumber(solarizedBlue),
			},
			CmpHelpViewSectionTableRow: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewColorNumber(solarizedWhite),
			},
			CmpHelpViewSectionTableRowHighlighted: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewColorNumber(solarizedYellow),
				style:   ThemeStyle{styleTypes: TstBold},
			},
			CmpHelpViewSectionTableCellSeparator: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewColorNumber(solarizedCyan),
			},
			CmpHelpViewFooter: {
				bgcolor: NewSystemColor(ColorNone),
				fgcolor: NewColorNumber(solarizedCyan),
			},
		},
	}
}
