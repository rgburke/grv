package main

import (
	"testing"
)

func checkThemeComponent(expectedThemeComponent, actualThemeComponent ThemeComponent, t *testing.T) {
	if expectedThemeComponent != actualThemeComponent {
		t.Errorf("ThemeComponent does not match expected value. Expected: %v, Actual: %v", expectedThemeComponent, actualThemeComponent)
	}
}

func TestDefaultThemeHasComponentsSet(t *testing.T) {
	theme := NewDefaultTheme()

	tests := map[ThemeComponentID]ThemeComponent{
		CmpAllviewSearchMatch: {
			bgcolor: ColorYellow,
			fgcolor: ColorNone,
		},
		CmpErrorViewErrors: {
			bgcolor: ColorRed,
			fgcolor: ColorWhite,
		},
	}

	for themeComponentID, expectedThemeComponent := range tests {
		actualThemeComponent := theme.GetComponent(themeComponentID)
		checkThemeComponent(expectedThemeComponent, actualThemeComponent, t)
	}
}

func TestDefaultThemeComponentIsReturnedIfNotConfiguredForProvidedId(t *testing.T) {
	theme := NewTheme()
	expectedThemeComponent := getDefaultThemeComponent()

	actualThemeComponent := theme.GetComponent(CmpAllviewSearchMatch)

	checkThemeComponent(expectedThemeComponent, actualThemeComponent, t)
}

func TestThemeComponentCanBeSet(t *testing.T) {
	theme := NewTheme()
	expectedThemeComponent := ThemeComponent{
		bgcolor: ColorBlack,
		fgcolor: ColorMagenta,
	}

	themeComponent := theme.CreateOrGetComponent(CmpAllviewSearchMatch)
	*themeComponent = expectedThemeComponent

	actualThemeComponent := theme.GetComponent(CmpAllviewSearchMatch)

	checkThemeComponent(expectedThemeComponent, actualThemeComponent, t)
}

func TestGetAllComponentsDefaultsComponentsNotSet(t *testing.T) {
	theme := NewTheme()
	tests := map[ThemeComponentID]ThemeComponent{
		CmpAllviewSearchMatch: {
			bgcolor: ColorYellow,
			fgcolor: ColorBlue,
		},
		CmpErrorViewErrors: getDefaultThemeComponent(),
	}

	searchMatchComponent := theme.CreateOrGetComponent(CmpAllviewSearchMatch)
	*searchMatchComponent = tests[CmpAllviewSearchMatch]

	allComponents := theme.GetAllComponents()

	if len(allComponents) != int(CmpCount-1) {
		t.Errorf("Size of GetAllComponents does not match expected value. Expected: %v, Actual: %v", CmpCount-1, len(allComponents))
	}

	for themeComponentID, expectedThemeComponent := range tests {
		actualThemeComponent := allComponents[themeComponentID]
		checkThemeComponent(expectedThemeComponent, actualThemeComponent, t)
	}
}
