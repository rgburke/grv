package main

type ThemeComponentId int16

const (
	CMP_NONE ThemeComponentId = iota

	CMP_ALLVIEW_SEARCH_MATCH

	CMP_REFVIEW_TITLE
	CMP_REFVIEW_FOOTER
	CMP_REFVIEW_BRANCHES_HEADER
	CMP_REFVIEW_BRANCH
	CMP_REFVIEW_TAGS_HEADER
	CMP_REFVIEW_TAG

	CMP_COMMITVIEW_TITLE
	CMP_COMMITVIEW_FOOTER
	CMP_COMMITVIEW_DATE
	CMP_COMMITVIEW_AUTHOR
	CMP_COMMITVIEW_SUMMARY

	CMP_DIFFVIEW_DIFFLINE_NORMAL
	CMP_DIFFVIEW_DIFFLINE_DIFF_COMMIT_AUTHOR
	CMP_DIFFVIEW_DIFFLINE_DIFF_COMMIT_AUTHOR_DATE
	CMP_DIFFVIEW_DIFFLINE_DIFF_COMMIT_COMMITTER
	CMP_DIFFVIEW_DIFFLINE_DIFF_COMMIT_COMMITTER_DATE
	CMP_DIFFVIEW_DIFFLINE_DIFF_COMMIT_SUMMARY
	CMP_DIFFVIEW_DIFFLINE_DIFF_STATS_FILE
	CMP_DIFFVIEW_DIFFLINE_GIT_DIFF_HEADER
	CMP_DIFFVIEW_DIFFLINE_GIT_DIFF_EXTENDED_HEADER
	CMP_DIFFVIEW_DIFFLINE_UNIFIED_DIFF_HEADER
	CMP_DIFFVIEW_DIFFLINE_HUNK_START
	CMP_DIFFVIEW_DIFFLINE_HUNK_HEADER
	CMP_DIFFVIEW_DIFFLINE_LINE_ADDED
	CMP_DIFFVIEW_DIFFLINE_LINE_REMOVED

	CMP_STATUSBARVIEW_NORMAL

	CMP_HELPBARVIEW_SPECIAL
	CMP_HELPBARVIEW_NORMAL

	CMP_ERROR_VIEW_TITLE
	CMP_ERROR_VIEW_FOOTER
	CMP_ERROR_VIEW_ERRORS

	CMP_COUNT
)

type ThemeColor int

const (
	COLOR_NONE ThemeColor = iota
	COLOR_BLACK
	COLOR_RED
	COLOR_GREEN
	COLOR_YELLOW
	COLOR_BLUE
	COLOR_MAGENTA
	COLOR_CYAN
	COLOR_WHITE
)

type ThemeComponent struct {
	bgcolor ThemeColor
	fgcolor ThemeColor
}

type Theme interface {
	GetComponent(ThemeComponentId) ThemeComponent
	GetAllComponents() map[ThemeComponentId]ThemeComponent
}

type MutableTheme interface {
	Theme
	CreateOrGetComponent(ThemeComponentId) *ThemeComponent
}

type ThemeComponents struct {
	components map[ThemeComponentId]*ThemeComponent
}

func (themeComponents *ThemeComponents) GetComponent(themeComponentId ThemeComponentId) ThemeComponent {
	if themeComponent, ok := themeComponents.components[themeComponentId]; ok {
		return *themeComponent
	}

	return getDefaultThemeComponent()
}

func (themeComponents *ThemeComponents) GetAllComponents() map[ThemeComponentId]ThemeComponent {
	components := make(map[ThemeComponentId]ThemeComponent, CMP_COUNT)

	for themeComponentId := ThemeComponentId(1); themeComponentId < CMP_COUNT; themeComponentId++ {
		themeComponent := themeComponents.GetComponent(themeComponentId)
		components[themeComponentId] = themeComponent
	}

	return components
}

func (themeComponents *ThemeComponents) CreateOrGetComponent(themeComponentId ThemeComponentId) *ThemeComponent {
	themeComponent, ok := themeComponents.components[themeComponentId]

	if !ok {
		defultThemeComponent := getDefaultThemeComponent()
		themeComponent = &defultThemeComponent
		themeComponents.components[themeComponentId] = themeComponent
	}

	return themeComponent
}

func getDefaultThemeComponent() ThemeComponent {
	return ThemeComponent{
		bgcolor: COLOR_NONE,
		fgcolor: COLOR_NONE,
	}
}

func NewTheme() MutableTheme {
	return &ThemeComponents{
		components: make(map[ThemeComponentId]*ThemeComponent),
	}
}

func NewDefaultTheme() MutableTheme {
	return &ThemeComponents{
		components: map[ThemeComponentId]*ThemeComponent{
			CMP_ALLVIEW_SEARCH_MATCH: &ThemeComponent{
				bgcolor: COLOR_YELLOW,
				fgcolor: COLOR_NONE,
			},
			CMP_COMMITVIEW_TITLE: &ThemeComponent{
				bgcolor: COLOR_NONE,
				fgcolor: COLOR_CYAN,
			},
			CMP_COMMITVIEW_FOOTER: &ThemeComponent{
				bgcolor: COLOR_NONE,
				fgcolor: COLOR_CYAN,
			},
			CMP_COMMITVIEW_DATE: &ThemeComponent{
				bgcolor: COLOR_NONE,
				fgcolor: COLOR_BLUE,
			},
			CMP_COMMITVIEW_AUTHOR: &ThemeComponent{
				bgcolor: COLOR_NONE,
				fgcolor: COLOR_GREEN,
			},
			CMP_COMMITVIEW_SUMMARY: &ThemeComponent{
				bgcolor: COLOR_NONE,
				fgcolor: COLOR_YELLOW,
			},
			CMP_DIFFVIEW_DIFFLINE_NORMAL: &ThemeComponent{
				bgcolor: COLOR_NONE,
				fgcolor: COLOR_NONE,
			},
			CMP_DIFFVIEW_DIFFLINE_DIFF_COMMIT_AUTHOR: &ThemeComponent{
				bgcolor: COLOR_NONE,
				fgcolor: COLOR_CYAN,
			},
			CMP_DIFFVIEW_DIFFLINE_DIFF_COMMIT_AUTHOR_DATE: &ThemeComponent{
				bgcolor: COLOR_NONE,
				fgcolor: COLOR_YELLOW,
			},
			CMP_DIFFVIEW_DIFFLINE_DIFF_COMMIT_COMMITTER: &ThemeComponent{
				bgcolor: COLOR_NONE,
				fgcolor: COLOR_MAGENTA,
			},
			CMP_DIFFVIEW_DIFFLINE_DIFF_COMMIT_COMMITTER_DATE: &ThemeComponent{
				bgcolor: COLOR_NONE,
				fgcolor: COLOR_YELLOW,
			},
			CMP_DIFFVIEW_DIFFLINE_DIFF_COMMIT_SUMMARY: &ThemeComponent{
				bgcolor: COLOR_NONE,
				fgcolor: COLOR_NONE,
			},
			CMP_DIFFVIEW_DIFFLINE_DIFF_STATS_FILE: &ThemeComponent{
				bgcolor: COLOR_NONE,
				fgcolor: COLOR_BLUE,
			},
			CMP_DIFFVIEW_DIFFLINE_GIT_DIFF_HEADER: &ThemeComponent{
				bgcolor: COLOR_NONE,
				fgcolor: COLOR_YELLOW,
			},
			CMP_DIFFVIEW_DIFFLINE_GIT_DIFF_EXTENDED_HEADER: &ThemeComponent{
				bgcolor: COLOR_NONE,
				fgcolor: COLOR_BLUE,
			},
			CMP_DIFFVIEW_DIFFLINE_UNIFIED_DIFF_HEADER: &ThemeComponent{
				bgcolor: COLOR_NONE,
				fgcolor: COLOR_YELLOW,
			},
			CMP_DIFFVIEW_DIFFLINE_HUNK_START: &ThemeComponent{
				bgcolor: COLOR_NONE,
				fgcolor: COLOR_CYAN,
			},
			CMP_DIFFVIEW_DIFFLINE_HUNK_HEADER: &ThemeComponent{
				bgcolor: COLOR_NONE,
				fgcolor: COLOR_BLUE,
			},
			CMP_DIFFVIEW_DIFFLINE_LINE_ADDED: &ThemeComponent{
				bgcolor: COLOR_NONE,
				fgcolor: COLOR_GREEN,
			},
			CMP_DIFFVIEW_DIFFLINE_LINE_REMOVED: &ThemeComponent{
				bgcolor: COLOR_NONE,
				fgcolor: COLOR_RED,
			},
			CMP_REFVIEW_TITLE: &ThemeComponent{
				bgcolor: COLOR_NONE,
				fgcolor: COLOR_CYAN,
			},
			CMP_REFVIEW_FOOTER: &ThemeComponent{
				bgcolor: COLOR_NONE,
				fgcolor: COLOR_CYAN,
			},
			CMP_REFVIEW_BRANCHES_HEADER: &ThemeComponent{
				bgcolor: COLOR_NONE,
				fgcolor: COLOR_MAGENTA,
			},
			CMP_REFVIEW_BRANCH: &ThemeComponent{
				bgcolor: COLOR_NONE,
				fgcolor: COLOR_NONE,
			},
			CMP_REFVIEW_TAGS_HEADER: &ThemeComponent{
				bgcolor: COLOR_NONE,
				fgcolor: COLOR_MAGENTA,
			},
			CMP_REFVIEW_TAG: &ThemeComponent{
				bgcolor: COLOR_NONE,
				fgcolor: COLOR_NONE,
			},
			CMP_STATUSBARVIEW_NORMAL: &ThemeComponent{
				bgcolor: COLOR_BLUE,
				fgcolor: COLOR_YELLOW,
			},
			CMP_HELPBARVIEW_SPECIAL: &ThemeComponent{
				bgcolor: COLOR_NONE,
				fgcolor: COLOR_MAGENTA,
			},
			CMP_HELPBARVIEW_NORMAL: &ThemeComponent{
				bgcolor: COLOR_NONE,
				fgcolor: COLOR_NONE,
			},
			CMP_ERROR_VIEW_TITLE: &ThemeComponent{
				bgcolor: COLOR_NONE,
				fgcolor: COLOR_CYAN,
			},
			CMP_ERROR_VIEW_FOOTER: &ThemeComponent{
				bgcolor: COLOR_NONE,
				fgcolor: COLOR_CYAN,
			},
			CMP_ERROR_VIEW_ERRORS: &ThemeComponent{
				bgcolor: COLOR_RED,
				fgcolor: COLOR_WHITE,
			},
		},
	}
}
