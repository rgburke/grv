package main

import (
	"strings"
)

// GenerateSetCommandHelpSections generates help documentation for the set command
func GenerateSetCommandHelpSections(config Config) (helpSections []*HelpSection) {
	description := []HelpSectionText{
		HelpSectionText{text: "set", themeComponentID: CmpHelpViewSectionSubTitle},
		HelpSectionText{},
		HelpSectionText{text: "The set command allows configuration variables to be set. It has the form:"},
		HelpSectionText{},
		HelpSectionText{text: "set variable value", themeComponentID: CmpHelpViewSectionCodeBlock},
		HelpSectionText{},
		HelpSectionText{text: "For example, to set the tab width to tab width to 4, the currently active theme to \"mytheme\" and enable mouse support:"},
		HelpSectionText{},
		HelpSectionText{text: "set tabwidth 4", themeComponentID: CmpHelpViewSectionCodeBlock},
		HelpSectionText{text: "set theme mytheme", themeComponentID: CmpHelpViewSectionCodeBlock},
		HelpSectionText{text: "set mouse true", themeComponentID: CmpHelpViewSectionCodeBlock},
		HelpSectionText{},
		HelpSectionText{text: "GRV currently has the following themes available:"},
		HelpSectionText{},
		HelpSectionText{text: " - solarized"},
		HelpSectionText{text: " - classic"},
		HelpSectionText{text: " - cold"},
		HelpSectionText{},
		HelpSectionText{text: "The solarized theme is the default theme for GRV and does not respect the terminals colour palette."},
		HelpSectionText{text: "The classic and cold themes do respect the terminals colour palette."},
	}

	return []*HelpSection{
		&HelpSection{
			description: description,
		},
	}
}

// GenerateThemeCommandHelpSections generates help documentation for the theme command
func GenerateThemeCommandHelpSections(config Config) (helpSections []*HelpSection) {
	description := []HelpSectionText{
		HelpSectionText{text: "theme", themeComponentID: CmpHelpViewSectionSubTitle},
		HelpSectionText{},
		HelpSectionText{text: "The theme command allows a custom theme to be defined."},
		HelpSectionText{text: "This theme can then be activated using the theme config variable described above."},
		HelpSectionText{text: "The form of the theme command is:"},
		HelpSectionText{},
		HelpSectionText{text: "theme --name [ThemeName] --component [ComponentId] --bgcolor [BackgroundColor] --fgcolor [ForegroundColor]", themeComponentID: CmpHelpViewSectionCodeBlock},
		HelpSectionText{},
		HelpSectionText{text: " - ThemeName: The name of the theme to be created/updated."},
		HelpSectionText{text: " - ComponentId: The Id of the screen component (the part of the display to change)."},
		HelpSectionText{text: " - BackgroundColor: The background color."},
		HelpSectionText{text: " - ForegroundColor: The foreground color."},
		HelpSectionText{},
		HelpSectionText{text: "Using a sequence of theme commands it is possible to define a theme."},
		HelpSectionText{text: "For example, to define a new theme \"mytheme\" and set it as the active theme:"},
		HelpSectionText{},
		HelpSectionText{text: "theme --name mytheme --component CommitView.Date      --bgcolor None --fgcolor Red", themeComponentID: CmpHelpViewSectionCodeBlock},
		HelpSectionText{text: "theme --name mytheme --component RefView.Tag          --bgcolor Blue --fgcolor 36", themeComponentID: CmpHelpViewSectionCodeBlock},
		HelpSectionText{text: "theme --name mytheme --component StatusBarView.Normal --bgcolor None --fgcolor f14a98", themeComponentID: CmpHelpViewSectionCodeBlock},
		HelpSectionText{text: "set theme mytheme", themeComponentID: CmpHelpViewSectionCodeBlock},
		HelpSectionText{},
		HelpSectionText{text: "GRV supports 256 colors (when available). Provided colors will be mapped to the nearest available color."},
		HelpSectionText{text: "The allowed color values are:"},
		HelpSectionText{},
		HelpSectionText{text: "System colors:"},
		HelpSectionText{},
		HelpSectionText{text: "None", themeComponentID: CmpHelpViewSectionCodeBlock},
		HelpSectionText{text: "Black", themeComponentID: CmpHelpViewSectionCodeBlock},
		HelpSectionText{text: "Red", themeComponentID: CmpHelpViewSectionCodeBlock},
		HelpSectionText{text: "Green", themeComponentID: CmpHelpViewSectionCodeBlock},
		HelpSectionText{text: "Yellow", themeComponentID: CmpHelpViewSectionCodeBlock},
		HelpSectionText{text: "Blue", themeComponentID: CmpHelpViewSectionCodeBlock},
		HelpSectionText{text: "Magenta", themeComponentID: CmpHelpViewSectionCodeBlock},
		HelpSectionText{text: "Cyan", themeComponentID: CmpHelpViewSectionCodeBlock},
		HelpSectionText{text: "White", themeComponentID: CmpHelpViewSectionCodeBlock},
		HelpSectionText{},
		HelpSectionText{text: "Terminal Color Numbers:"},
		HelpSectionText{},
		HelpSectionText{text: "0 - 255", themeComponentID: CmpHelpViewSectionCodeBlock},
		HelpSectionText{},
		HelpSectionText{text: "Hex Colors:"},
		HelpSectionText{},
		HelpSectionText{text: "000000 - ffffff", themeComponentID: CmpHelpViewSectionCodeBlock},
		HelpSectionText{},
		HelpSectionText{text: "The set of screen components that can be customised is:"},
		HelpSectionText{},
	}

	var prevPrefix string
	for _, themeComponent := range ThemeComponentNames() {
		prefix := strings.Split(themeComponent, ".")[0]

		if prevPrefix != "" && prevPrefix != prefix {
			description = append(description, HelpSectionText{})
		}

		prevPrefix = prefix
		description = append(description, HelpSectionText{text: themeComponent, themeComponentID: CmpHelpViewSectionCodeBlock})
	}

	return []*HelpSection{
		&HelpSection{
			description: description,
		},
	}
}

// GenerateMapCommandHelpSections generates help documentation for the map command
func GenerateMapCommandHelpSections(config Config) (helpSections []*HelpSection) {
	description := []HelpSectionText{
		HelpSectionText{text: "map", themeComponentID: CmpHelpViewSectionSubTitle},
		HelpSectionText{},
		HelpSectionText{text: "The map command allows a key sequence to be mapped to an action or another key sequence for a specified view."},
		HelpSectionText{text: "The form of the map command is:"},
		HelpSectionText{},
		HelpSectionText{text: "map view fromkeys tokeys", themeComponentID: CmpHelpViewSectionCodeBlock},
		HelpSectionText{},
		HelpSectionText{text: "For example, to map the key 'a' to the keys 'gg' in the Ref View:"},
		HelpSectionText{},
		HelpSectionText{text: "map RefView a gg", themeComponentID: CmpHelpViewSectionCodeBlock},
		HelpSectionText{},
		HelpSectionText{text: "When pressing 'a' in the Ref View, the first line would then become the selected line, as 'gg' moves the cursor to the first line."},
		HelpSectionText{text: "All is a valid view argument when a binding should apply to all views."},
		HelpSectionText{},
		HelpSectionText{text: "GRV also has a text representation of actions that are independent of key bindings."},
		HelpSectionText{text: "For example, the following commands can be used to make the <Up> key move a line down and the <Down> key move a line up:"},
		HelpSectionText{},
		HelpSectionText{text: "map All <Up> <grv-next-line>", themeComponentID: CmpHelpViewSectionCodeBlock},
		HelpSectionText{text: "map All <Down> <grv-prev-line>", themeComponentID: CmpHelpViewSectionCodeBlock},
		HelpSectionText{},
		HelpSectionText{text: "The set of actions available is described in the key binding tables above."},
	}

	return []*HelpSection{
		&HelpSection{
			description: description,
		},
	}
}

// GenerateUnmapCommandHelpSections generates help documentation for the unmap command
func GenerateUnmapCommandHelpSections(config Config) (helpSections []*HelpSection) {
	description := []HelpSectionText{
		HelpSectionText{text: "unmap", themeComponentID: CmpHelpViewSectionSubTitle},
		HelpSectionText{},
		HelpSectionText{text: "The unmap command removes any defined key binding for a key sequence in the specified view."},
		HelpSectionText{text: "The form of the unmap command is:"},
		HelpSectionText{},
		HelpSectionText{text: "unmap view fromkeys", themeComponentID: CmpHelpViewSectionCodeBlock},
		HelpSectionText{},
		HelpSectionText{text: "For example, to unmap the key 'c' in the Ref View:"},
		HelpSectionText{},
		HelpSectionText{text: "unmap RefView c", themeComponentID: CmpHelpViewSectionCodeBlock},
	}

	return []*HelpSection{
		&HelpSection{
			description: description,
		},
	}
}

// GenerateQuitCommandHelpSections generates help documentation for the quit command
func GenerateQuitCommandHelpSections(config Config) (helpSections []*HelpSection) {
	description := []HelpSectionText{
		HelpSectionText{text: "q", themeComponentID: CmpHelpViewSectionSubTitle},
		HelpSectionText{},
		HelpSectionText{text: "The quit command is used to exit GRV and can be used with the following keys:"},
		HelpSectionText{},
		HelpSectionText{text: ":q<Enter>", themeComponentID: CmpHelpViewSectionCodeBlock},
	}

	return []*HelpSection{
		&HelpSection{
			description: description,
		},
	}
}

// GenerateAddTabCommandHelpSections generates help documentation for the addtab command
func GenerateAddTabCommandHelpSections(config Config) (helpSections []*HelpSection) {
	description := []HelpSectionText{
		HelpSectionText{text: "addtab", themeComponentID: CmpHelpViewSectionSubTitle},
		HelpSectionText{},
		HelpSectionText{text: "The addtab command creates a new named empty tab and switches to this new tab."},
		HelpSectionText{text: "The format of the command is:"},
		HelpSectionText{},
		HelpSectionText{text: "addtab tabname", themeComponentID: CmpHelpViewSectionCodeBlock},
		HelpSectionText{},
		HelpSectionText{text: "For example, to add a new tab titled \"mycustomtab\" the following command can be used:"},
		HelpSectionText{},
		HelpSectionText{text: "addtab mycustomtab", themeComponentID: CmpHelpViewSectionCodeBlock},
	}

	return []*HelpSection{
		&HelpSection{
			description: description,
		},
	}
}

// GenerateRmTabCommandHelpSections generates help documentation for the rmtab command
func GenerateRmTabCommandHelpSections(config Config) (helpSections []*HelpSection) {
	description := []HelpSectionText{
		HelpSectionText{text: "rmtab", themeComponentID: CmpHelpViewSectionSubTitle},
		HelpSectionText{},
		HelpSectionText{text: "The rmtab removes the currently active tab. If the tab removed is the last tab then GRV will exit."},
	}

	return []*HelpSection{
		&HelpSection{
			description: description,
		},
	}
}

// GenerateAddViewCommandHelpSections generates help documentation for the addview command
func GenerateAddViewCommandHelpSections(config Config) (helpSections []*HelpSection) {
	description := []HelpSectionText{
		HelpSectionText{text: "addview", themeComponentID: CmpHelpViewSectionSubTitle},
		HelpSectionText{},
		HelpSectionText{text: "The addview command allows a view to be added to the currently active tab."},
		HelpSectionText{text: "The form of the command is:"},
		HelpSectionText{},
		HelpSectionText{text: "addview view viewargs...", themeComponentID: CmpHelpViewSectionCodeBlock},
		HelpSectionText{},
		HelpSectionText{text: "Each view accepts a different set of arguments. This is described in the table below:"},
	}

	helpSections = append(helpSections, &HelpSection{
		description: description,
	})

	helpSections = append(helpSections, GenerateWindowViewFactoryHelpSection(config))

	description = []HelpSectionText{
		HelpSectionText{text: "Examples usages for each view are given below:"},
		HelpSectionText{},
		HelpSectionText{text: "addview CommitView origin/master", themeComponentID: CmpHelpViewSectionCodeBlock},
		HelpSectionText{text: "addview DiffView 4882ca9044661b49a26ae03ceb1be3a70d00c6a2", themeComponentID: CmpHelpViewSectionCodeBlock},
		HelpSectionText{text: "addview GitStatusView", themeComponentID: CmpHelpViewSectionCodeBlock},
		HelpSectionText{text: "addview RefView", themeComponentID: CmpHelpViewSectionCodeBlock},
	}

	helpSections = append(helpSections, &HelpSection{
		description: description,
	})

	return
}

// GenerateVSplitCommandHelpSections generates help documentation for the vsplit command
func GenerateVSplitCommandHelpSections(config Config) (helpSections []*HelpSection) {
	description := []HelpSectionText{
		HelpSectionText{text: "vsplit", themeComponentID: CmpHelpViewSectionSubTitle},
		HelpSectionText{},
		HelpSectionText{text: "The vsplit command creates a vertical split between the currently selected view and the view specified in the command."},
		HelpSectionText{text: "The form of the command is:"},
		HelpSectionText{},
		HelpSectionText{text: "vsplit view viewargs...", themeComponentID: CmpHelpViewSectionCodeBlock},
		HelpSectionText{},
		HelpSectionText{text: "For example, to create a vertical split between the currently selected view and a CommitView displaying commits for master:"},
		HelpSectionText{},
		HelpSectionText{text: "vsplit CommitView master", themeComponentID: CmpHelpViewSectionCodeBlock},
	}

	return []*HelpSection{
		&HelpSection{
			description: description,
		},
	}
}

// GenerateHSplitCommandHelpSections generates help documentation for the hsplit command
func GenerateHSplitCommandHelpSections(config Config) (helpSections []*HelpSection) {
	description := []HelpSectionText{
		HelpSectionText{text: "hsplit", themeComponentID: CmpHelpViewSectionSubTitle},
		HelpSectionText{},
		HelpSectionText{text: "The hsplit command creates a horizontal split between the currently selected view and the view specified in the command."},
		HelpSectionText{text: "The form of the command is:"},
		HelpSectionText{},
		HelpSectionText{text: "hsplit view viewargs...", themeComponentID: CmpHelpViewSectionCodeBlock},
		HelpSectionText{},
		HelpSectionText{text: "For example, to create a horizontal split between the currently selected view and a RefView:"},
		HelpSectionText{},
		HelpSectionText{text: "hsplit RefView", themeComponentID: CmpHelpViewSectionCodeBlock},
	}

	return []*HelpSection{
		&HelpSection{
			description: description,
		},
	}
}

// GenerateSplitCommandHelpSections generates help documentation for the split command
func GenerateSplitCommandHelpSections(config Config) (helpSections []*HelpSection) {
	description := []HelpSectionText{
		HelpSectionText{text: "split", themeComponentID: CmpHelpViewSectionSubTitle},
		HelpSectionText{},
		HelpSectionText{text: "The split command is similar to the vsplit and hsplit commands."},
		HelpSectionText{text: "It creates either a new vsplit or hsplit determined by the current dimensions of the active view."},
		HelpSectionText{text: "The form of the command is:"},
		HelpSectionText{},
		HelpSectionText{text: "split view viewargs...", themeComponentID: CmpHelpViewSectionCodeBlock},
	}

	return []*HelpSection{
		&HelpSection{
			description: description,
		},
	}
}

// GenerateGitCommandHelpSections generates help documentation for the git command
func GenerateGitCommandHelpSections(config Config) (helpSections []*HelpSection) {
	description := []HelpSectionText{
		HelpSectionText{text: "git", themeComponentID: CmpHelpViewSectionSubTitle},
		HelpSectionText{},
		HelpSectionText{text: "The git command is an alias to the git cli command."},
		HelpSectionText{text: "It allows a non-interactive git command to run without having to leave GRV."},
		HelpSectionText{text: "A pop-up window displays the output of the command."},
		HelpSectionText{text: "For example, to run 'git status' from within GRV use the following key sequence"},
		HelpSectionText{},
		HelpSectionText{text: ":git status", themeComponentID: CmpHelpViewSectionCodeBlock},
		HelpSectionText{},
		HelpSectionText{text: "Only non-interactive git commands (i.e. those that require no user input) can be run using the git command."},
		HelpSectionText{text: "For interactive git commands the giti command can be used."},
	}

	return []*HelpSection{
		&HelpSection{
			description: description,
		},
	}
}

// GenerateGitiCommandHelpSections generates help documentation for the giti command
func GenerateGitiCommandHelpSections(config Config) (helpSections []*HelpSection) {
	description := []HelpSectionText{
		HelpSectionText{text: "giti", themeComponentID: CmpHelpViewSectionSubTitle},
		HelpSectionText{},
		HelpSectionText{text: "The git command is an alias to the git cli command."},
		HelpSectionText{text: "It allows an interactive git command (i.e. those that require user input) to be run without having to leave GRV."},
		HelpSectionText{text: "The command is executed in the controlling terminal and GRV is resumed on command completion."},
		HelpSectionText{text: "For example, to run 'git rebase -i HEAD~2' use the following key sequence:"},
		HelpSectionText{},
		HelpSectionText{text: ":giti rebase -i HEAD~2", themeComponentID: CmpHelpViewSectionCodeBlock},
	}

	return []*HelpSection{
		&HelpSection{
			description: description,
		},
	}
}

// GenerateHelpCommandHelpSections generates help documentation for the help command
func GenerateHelpCommandHelpSections(config Config) (helpSections []*HelpSection) {
	description := []HelpSectionText{
		HelpSectionText{text: "help", themeComponentID: CmpHelpViewSectionSubTitle},
		HelpSectionText{},
		HelpSectionText{text: "The help command opens a tab containing documentation for GRV."},
	}

	return []*HelpSection{
		&HelpSection{
			description: description,
		},
	}
}
