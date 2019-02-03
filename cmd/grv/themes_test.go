package main

import (
	"testing"
)

func testThemeHasAllThemeComponentsSet(themeName string, theme Theme, t *testing.T) {
	themeComponents, ok := theme.(*ThemeComponents)
	if !ok {
		t.Errorf("Expected ThemeComponents instance")
	}

	for themeComponentID := CmpNone + 1; themeComponentID < CmpCount; themeComponentID++ {
		if _, ok = themeComponents.components[themeComponentID]; !ok {
			t.Errorf("Theme \"%v\" does not have entry for ThemeComponentID %v", themeName, themeComponentID)
		}
	}
}

func TestThemesHaveAllThemeComponentsSet(t *testing.T) {
	testThemeHasAllThemeComponentsSet("solarized", NewSolarizedTheme(), t)
	testThemeHasAllThemeComponentsSet("classic", NewClassicTheme(), t)
}
