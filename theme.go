package main

type ThemeComponentId int16

const (
	CMP_NONE ThemeComponentId = iota

	CMP_COMMITVIEW_DATE
	CMP_COMMITVIEW_AUTHOR
	CMP_COMMITVIEW_SUMMARY

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

type ThemeComponents struct {
	components map[ThemeComponentId]ThemeComponent
}

func (themeComponents *ThemeComponents) GetComponent(themeComponentId ThemeComponentId) ThemeComponent {
	if themeComponent, ok := themeComponents.components[themeComponentId]; ok {
		return themeComponent
	}

	return ThemeComponent{
		bgcolor: COLOR_NONE,
		fgcolor: COLOR_NONE,
	}
}

func (themeComponents *ThemeComponents) GetAllComponents() map[ThemeComponentId]ThemeComponent {
	components := make(map[ThemeComponentId]ThemeComponent, CMP_COUNT)

	for themeComponentId := ThemeComponentId(1); themeComponentId < CMP_COUNT; themeComponentId++ {
		components[themeComponentId] = themeComponents.GetComponent(themeComponentId)
	}

	return components
}

func NewDefaultTheme() Theme {
	return &ThemeComponents{
		components: map[ThemeComponentId]ThemeComponent{
			CMP_COMMITVIEW_DATE: ThemeComponent{
				bgcolor: COLOR_NONE,
				fgcolor: COLOR_BLUE,
			},
			CMP_COMMITVIEW_AUTHOR: ThemeComponent{
				bgcolor: COLOR_NONE,
				fgcolor: COLOR_GREEN,
			},
			CMP_COMMITVIEW_SUMMARY: ThemeComponent{
				bgcolor: COLOR_NONE,
				fgcolor: COLOR_YELLOW,
			},
		},
	}
}
