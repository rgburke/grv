package main

import (
	"strings"
)

// GenerateSetCommandHelpSections generates help documentation for the set command
func GenerateSetCommandHelpSections(config Config) (helpSections []*HelpSection) {
	description := []HelpSectionText{
		{text: "set", themeComponentID: CmpHelpViewSectionSubTitle},
		{},
		{text: "The set command allows configuration variables to be set. It has the form:"},
		{},
		{text: "set variable value", themeComponentID: CmpHelpViewSectionCodeBlock},
		{},
		{text: "For example, to set the tab width to tab width to 4, the currently active theme to \"mytheme\" and enable mouse support:"},
		{},
		{text: "set tabwidth 4", themeComponentID: CmpHelpViewSectionCodeBlock},
		{text: "set theme mytheme", themeComponentID: CmpHelpViewSectionCodeBlock},
		{text: "set mouse true", themeComponentID: CmpHelpViewSectionCodeBlock},
		{},
		{text: "GRV currently has the following themes available:"},
		{},
		{text: " - solarized"},
		{text: " - classic"},
		{},
		{text: "The solarized theme is the default theme for GRV and does not respect the terminals colour palette."},
		{text: "The classic theme respects the terminals colour palette."},
	}

	return []*HelpSection{
		{
			description: description,
		},
	}
}

// GenerateThemeCommandHelpSections generates help documentation for the theme command
func GenerateThemeCommandHelpSections(config Config) (helpSections []*HelpSection) {
	description := []HelpSectionText{
		{text: "theme", themeComponentID: CmpHelpViewSectionSubTitle},
		{},
		{text: "The theme command allows a custom theme to be defined."},
		{text: "This theme can then be activated using the theme config variable described above."},
		{text: "The form of the theme command is:"},
		{},
		{text: "theme --name [ThemeName] --component [ComponentId] --bgcolor [BackgroundColor] --fgcolor [ForegroundColor]", themeComponentID: CmpHelpViewSectionCodeBlock},
		{},
		{text: " - ThemeName: The name of the theme to be created/updated."},
		{text: " - ComponentId: The Id of the screen component (the part of the display to change)."},
		{text: " - BackgroundColor: The background color."},
		{text: " - ForegroundColor: The foreground color."},
		{},
		{text: "Using a sequence of theme commands it is possible to define a theme."},
		{text: "For example, to define a new theme \"mytheme\" and set it as the active theme:"},
		{},
		{text: "theme --name mytheme --component CommitView.Date      --bgcolor None --fgcolor Red", themeComponentID: CmpHelpViewSectionCodeBlock},
		{text: "theme --name mytheme --component RefView.Tag          --bgcolor Blue --fgcolor 36", themeComponentID: CmpHelpViewSectionCodeBlock},
		{text: "theme --name mytheme --component StatusBarView.Normal --bgcolor None --fgcolor f14a98", themeComponentID: CmpHelpViewSectionCodeBlock},
		{text: "set theme mytheme", themeComponentID: CmpHelpViewSectionCodeBlock},
		{},
		{text: "GRV supports 256 colors (when available). Provided colors will be mapped to the nearest available color."},
		{text: "The allowed color values are:"},
		{},
		{text: "System colors:"},
		{},
		{text: "None", themeComponentID: CmpHelpViewSectionCodeBlock},
		{text: "Black", themeComponentID: CmpHelpViewSectionCodeBlock},
		{text: "Red", themeComponentID: CmpHelpViewSectionCodeBlock},
		{text: "Green", themeComponentID: CmpHelpViewSectionCodeBlock},
		{text: "Yellow", themeComponentID: CmpHelpViewSectionCodeBlock},
		{text: "Blue", themeComponentID: CmpHelpViewSectionCodeBlock},
		{text: "Magenta", themeComponentID: CmpHelpViewSectionCodeBlock},
		{text: "Cyan", themeComponentID: CmpHelpViewSectionCodeBlock},
		{text: "White", themeComponentID: CmpHelpViewSectionCodeBlock},
		{},
		{text: "Terminal Color Numbers:"},
		{},
		{text: "0 - 255", themeComponentID: CmpHelpViewSectionCodeBlock},
		{},
		{text: "Hex Colors:"},
		{},
		{text: "000000 - ffffff", themeComponentID: CmpHelpViewSectionCodeBlock},
		{},
		{text: "The set of screen components that can be customised is:"},
		{},
	}

	var prevPrefix string
	for _, themeComponent := range ThemeComponentNames() {
		prefix := strings.Split(themeComponent, ".")[0]

		if prevPrefix != "" && prevPrefix != prefix {
			description = append(description, HelpSectionText{themeComponentID: CmpHelpViewSectionCodeBlock})
		}

		prevPrefix = prefix
		description = append(description, HelpSectionText{text: themeComponent, themeComponentID: CmpHelpViewSectionCodeBlock})
	}

	return []*HelpSection{
		{
			description: description,
		},
	}
}

// GenerateMapCommandHelpSections generates help documentation for the map command
func GenerateMapCommandHelpSections(config Config) (helpSections []*HelpSection) {
	description := []HelpSectionText{
		{text: "map", themeComponentID: CmpHelpViewSectionSubTitle},
		{},
		{text: "The map command allows a key sequence to be mapped to an action, another key sequence or a shell command for a specified view."},
		{text: "The form of the map command is:"},
		{},
		{text: "map view fromkeys tokeys", themeComponentID: CmpHelpViewSectionCodeBlock},
		{},
		{text: "For example, to map the key 'a' to the keys 'gg' in the Ref View:"},
		{},
		{text: "map RefView a gg", themeComponentID: CmpHelpViewSectionCodeBlock},
		{},
		{text: "When pressing 'a' in the Ref View, the first line would then become the selected line, as 'gg' moves the cursor to the first line."},
		{text: "All is a valid view argument when a binding should apply to all views."},
		{},
		{text: "GRV also has a text representation of actions that are independent of key bindings."},
		{text: "For example, the following commands can be used to make the <Up> key move a line down and the <Down> key move a line up:"},
		{},
		{text: "map All <Up> <grv-next-line>", themeComponentID: CmpHelpViewSectionCodeBlock},
		{text: "map All <Down> <grv-prev-line>", themeComponentID: CmpHelpViewSectionCodeBlock},
		{},
		{text: "The map command also allows a key sequence to be mapped directly to a shell command."},
		{text: "Prefix a shell command with the '!' character. For example, to map the key 'a' to the shell command 'ls -lh':"},
		{},
		{text: "map All a !ls -lh", themeComponentID: CmpHelpViewSectionCodeBlock},
		{},
		{text: "The set of actions available is described in the key binding tables above."},
	}

	return []*HelpSection{
		{
			description: description,
		},
	}
}

// GenerateUnmapCommandHelpSections generates help documentation for the unmap command
func GenerateUnmapCommandHelpSections(config Config) (helpSections []*HelpSection) {
	description := []HelpSectionText{
		{text: "unmap", themeComponentID: CmpHelpViewSectionSubTitle},
		{},
		{text: "The unmap command removes any defined key binding for a key sequence in the specified view."},
		{text: "The form of the unmap command is:"},
		{},
		{text: "unmap view fromkeys", themeComponentID: CmpHelpViewSectionCodeBlock},
		{},
		{text: "For example, to unmap the key 'c' in the Ref View:"},
		{},
		{text: "unmap RefView c", themeComponentID: CmpHelpViewSectionCodeBlock},
	}

	return []*HelpSection{
		{
			description: description,
		},
	}
}

// GenerateQuitCommandHelpSections generates help documentation for the quit command
func GenerateQuitCommandHelpSections(config Config) (helpSections []*HelpSection) {
	description := []HelpSectionText{
		{text: "q", themeComponentID: CmpHelpViewSectionSubTitle},
		{},
		{text: "The quit command is used to exit GRV and can be used with the following keys:"},
		{},
		{text: ":q<Enter>", themeComponentID: CmpHelpViewSectionCodeBlock},
	}

	return []*HelpSection{
		{
			description: description,
		},
	}
}

// GenerateAddTabCommandHelpSections generates help documentation for the addtab command
func GenerateAddTabCommandHelpSections(config Config) (helpSections []*HelpSection) {
	description := []HelpSectionText{
		{text: "addtab", themeComponentID: CmpHelpViewSectionSubTitle},
		{},
		{text: "The addtab command creates a new named empty tab and switches to this new tab."},
		{text: "The format of the command is:"},
		{},
		{text: "addtab tabname", themeComponentID: CmpHelpViewSectionCodeBlock},
		{},
		{text: "For example, to add a new tab titled \"mycustomtab\" the following command can be used:"},
		{},
		{text: "addtab mycustomtab", themeComponentID: CmpHelpViewSectionCodeBlock},
	}

	return []*HelpSection{
		{
			description: description,
		},
	}
}

// GenerateRmTabCommandHelpSections generates help documentation for the rmtab command
func GenerateRmTabCommandHelpSections(config Config) (helpSections []*HelpSection) {
	description := []HelpSectionText{
		{text: "rmtab", themeComponentID: CmpHelpViewSectionSubTitle},
		{},
		{text: "The rmtab removes the currently active tab. If the tab removed is the last tab then GRV will exit."},
	}

	return []*HelpSection{
		{
			description: description,
		},
	}
}

// GenerateAddViewCommandHelpSections generates help documentation for the addview command
func GenerateAddViewCommandHelpSections(config Config) (helpSections []*HelpSection) {
	description := []HelpSectionText{
		{text: "addview", themeComponentID: CmpHelpViewSectionSubTitle},
		{},
		{text: "The addview command allows a view to be added to the currently active tab."},
		{text: "The form of the command is:"},
		{},
		{text: "addview view viewargs...", themeComponentID: CmpHelpViewSectionCodeBlock},
		{},
		{text: "Each view accepts a different set of arguments. This is described in the table below:"},
	}

	helpSections = append(helpSections, &HelpSection{
		description: description,
	})

	helpSections = append(helpSections, GenerateWindowViewFactoryHelpSection(config))

	description = []HelpSectionText{
		{text: "Examples usages for each view are given below:"},
		{},
		{text: "addview CommitView origin/master", themeComponentID: CmpHelpViewSectionCodeBlock},
		{text: "addview DiffView 4882ca9044661b49a26ae03ceb1be3a70d00c6a2", themeComponentID: CmpHelpViewSectionCodeBlock},
		{text: "addview GitStatusView", themeComponentID: CmpHelpViewSectionCodeBlock},
		{text: "addview RefView", themeComponentID: CmpHelpViewSectionCodeBlock},
	}

	helpSections = append(helpSections, &HelpSection{
		description: description,
	})

	return
}

// GenerateVSplitCommandHelpSections generates help documentation for the vsplit command
func GenerateVSplitCommandHelpSections(config Config) (helpSections []*HelpSection) {
	description := []HelpSectionText{
		{text: "vsplit", themeComponentID: CmpHelpViewSectionSubTitle},
		{},
		{text: "The vsplit command creates a vertical split between the currently selected view and the view specified in the command."},
		{text: "The form of the command is:"},
		{},
		{text: "vsplit view viewargs...", themeComponentID: CmpHelpViewSectionCodeBlock},
		{},
		{text: "For example, to create a vertical split between the currently selected view and a CommitView displaying commits for master:"},
		{},
		{text: "vsplit CommitView master", themeComponentID: CmpHelpViewSectionCodeBlock},
	}

	return []*HelpSection{
		{
			description: description,
		},
	}
}

// GenerateHSplitCommandHelpSections generates help documentation for the hsplit command
func GenerateHSplitCommandHelpSections(config Config) (helpSections []*HelpSection) {
	description := []HelpSectionText{
		{text: "hsplit", themeComponentID: CmpHelpViewSectionSubTitle},
		{},
		{text: "The hsplit command creates a horizontal split between the currently selected view and the view specified in the command."},
		{text: "The form of the command is:"},
		{},
		{text: "hsplit view viewargs...", themeComponentID: CmpHelpViewSectionCodeBlock},
		{},
		{text: "For example, to create a horizontal split between the currently selected view and a RefView:"},
		{},
		{text: "hsplit RefView", themeComponentID: CmpHelpViewSectionCodeBlock},
	}

	return []*HelpSection{
		{
			description: description,
		},
	}
}

// GenerateSplitCommandHelpSections generates help documentation for the split command
func GenerateSplitCommandHelpSections(config Config) (helpSections []*HelpSection) {
	description := []HelpSectionText{
		{text: "split", themeComponentID: CmpHelpViewSectionSubTitle},
		{},
		{text: "The split command is similar to the vsplit and hsplit commands."},
		{text: "It creates either a new vsplit or hsplit determined by the current dimensions of the active view."},
		{text: "The form of the command is:"},
		{},
		{text: "split view viewargs...", themeComponentID: CmpHelpViewSectionCodeBlock},
	}

	return []*HelpSection{
		{
			description: description,
		},
	}
}

// GenerateGitCommandHelpSections generates help documentation for the git command
func GenerateGitCommandHelpSections(config Config) (helpSections []*HelpSection) {
	description := []HelpSectionText{
		{text: "git", themeComponentID: CmpHelpViewSectionSubTitle},
		{},
		{text: "The git command is an alias to the git cli command."},
		{text: "It allows a non-interactive git command to run without having to leave GRV."},
		{text: "A pop-up window displays the output of the command."},
		{text: "For example, to run 'git status' from within GRV use the following key sequence"},
		{},
		{text: ":git status", themeComponentID: CmpHelpViewSectionCodeBlock},
		{},
		{text: "Only non-interactive git commands (i.e. those that require no user input) can be run using the git command."},
		{text: "For interactive git commands the giti command can be used."},
	}

	return []*HelpSection{
		{
			description: description,
		},
	}
}

// GenerateGitiCommandHelpSections generates help documentation for the giti command
func GenerateGitiCommandHelpSections(config Config) (helpSections []*HelpSection) {
	description := []HelpSectionText{
		{text: "giti", themeComponentID: CmpHelpViewSectionSubTitle},
		{},
		{text: "The git command is an alias to the git cli command."},
		{text: "It allows an interactive git command (i.e. those that require user input) to be run without having to leave GRV."},
		{text: "The command is executed in the controlling terminal and GRV is resumed on command completion."},
		{text: "For example, to run 'git rebase -i HEAD~2' use the following key sequence:"},
		{},
		{text: ":giti rebase -i HEAD~2", themeComponentID: CmpHelpViewSectionCodeBlock},
	}

	return []*HelpSection{
		{
			description: description,
		},
	}
}

// GenerateHelpCommandHelpSections generates help documentation for the help command
func GenerateHelpCommandHelpSections(config Config) (helpSections []*HelpSection) {
	description := []HelpSectionText{
		{text: "help", themeComponentID: CmpHelpViewSectionSubTitle},
		{},
		{text: "The help command opens a tab containing documentation for GRV."},
		{text: "A search term can be provided as an argument. For example:"},
		{},
		{text: "help vsplit", themeComponentID: CmpHelpViewSectionCodeBlock},
		{},
		{text: "will display the section for the command vsplit in the help tab"},
	}

	return []*HelpSection{
		{
			description: description,
		},
	}
}

// GenerateDefCommandHelpSections generates help documentation for the def command
func GenerateDefCommandHelpSections(config Config) (helpSections []*HelpSection) {
	description := []HelpSectionText{
		{text: "def", themeComponentID: CmpHelpViewSectionSubTitle},
		{},
		{text: "The def command allows a custom GRV command to be defined. It has the form:"},
		{},
		{text: "def NAME {", themeComponentID: CmpHelpViewSectionCodeBlock},
		{text: "\tBODY", themeComponentID: CmpHelpViewSectionCodeBlock},
		{text: "}", themeComponentID: CmpHelpViewSectionCodeBlock},
		{},
		{text: "where NAME is the name of the new command and BODY is a sequence of commands to execute."},
		{text: "For example, to define a command \"maintab\" to open a new tab containing the CommitView for master:"},
		{},
		{text: "def maintab {", themeComponentID: CmpHelpViewSectionCodeBlock},
		{text: "\taddtab Main", themeComponentID: CmpHelpViewSectionCodeBlock},
		{text: "\taddview CommitView master", themeComponentID: CmpHelpViewSectionCodeBlock},
		{text: "}", themeComponentID: CmpHelpViewSectionCodeBlock},
		{},
		{text: "This command can be invoked at the command prompt with:"},
		{},
		{text: ":maintab", themeComponentID: CmpHelpViewSectionCodeBlock},
		{},
		{text: "The command body can contain argument placeholders that will be substituted on invocation."},
		{text: "Argument placeholders have the form $n or ${n} where n is an integer greater than zero corresponding to the argument to substitute."},
		{text: "For example, the \"maintab\" command defined earlier can be altered to accept the branch in as an argument:"},
		{},
		{text: "def maintab {", themeComponentID: CmpHelpViewSectionCodeBlock},
		{text: "\taddtab Main", themeComponentID: CmpHelpViewSectionCodeBlock},
		{text: "\taddview CommitView $1", themeComponentID: CmpHelpViewSectionCodeBlock},
		{text: "}", themeComponentID: CmpHelpViewSectionCodeBlock},
		{},
		{text: "To invoke this command for the branch \"feature-branch\" and open a new tab containing the commit view for this branch:"},
		{},
		{text: ":maintab feature-branch", themeComponentID: CmpHelpViewSectionCodeBlock},
		{},
		{text: "All arguments can be substituted using the placeholder $@ or ${@}"},
		{text: "For example, the following command acts as an alias for the vsplit command:"},
		{},
		{text: "def vs { vsplit $@ }", themeComponentID: CmpHelpViewSectionCodeBlock},
		{},
		{text: "and can be invoked with:"},
		{},
		{text: ":vs CommitView master", themeComponentID: CmpHelpViewSectionCodeBlock},
		{},
		{text: "Argument placeholders can be escaped by prepending a dollar sign."},
		{text: "For example, to specify the literal string $1 in a command body specify $$1."},
	}

	return []*HelpSection{
		{
			description: description,
		},
	}
}

// GenerateUndefCommandHelpSections generates help documentation for the addtab command
func GenerateUndefCommandHelpSections(config Config) (helpSections []*HelpSection) {
	description := []HelpSectionText{
		{text: "undef", themeComponentID: CmpHelpViewSectionSubTitle},
		{},
		{text: "The undef command removes a user defined command."},
		{text: "The format of the command is:"},
		{},
		{text: "undef commandname", themeComponentID: CmpHelpViewSectionCodeBlock},
		{},
		{text: "For example, to undefine a comamnd \"mycustomcommand\" the following can be used:"},
		{},
		{text: "undef mycustomcommand", themeComponentID: CmpHelpViewSectionCodeBlock},
	}

	return []*HelpSection{
		{
			description: description,
		},
	}
}

// GenerateEvalKeysCommandHelpSections generates help documentation for the addtab command
func GenerateEvalKeysCommandHelpSections(config Config) (helpSections []*HelpSection) {
	description := []HelpSectionText{
		{text: "evalkeys", themeComponentID: CmpHelpViewSectionSubTitle},
		{},
		{text: "The evalkeys command executes the provided key string sequence."},
		{text: "The format of the command is:"},
		{},
		{text: "evalkeys keys", themeComponentID: CmpHelpViewSectionCodeBlock},
		{},
		{text: "For example, running the following will switch to the next tab:"},
		{},
		{text: "evalkeys <grv-next-tab>", themeComponentID: CmpHelpViewSectionCodeBlock},
	}

	return []*HelpSection{
		{
			description: description,
		},
	}
}

// GenerateSleepCommandHelpSections generates help documentation for the addtab command
func GenerateSleepCommandHelpSections(config Config) (helpSections []*HelpSection) {
	description := []HelpSectionText{
		{text: "sleep", themeComponentID: CmpHelpViewSectionSubTitle},
		{},
		{text: "The sleep command causes grv to pause execution for the provided number of seconds."},
		{text: "The format of the command is:"},
		{},
		{text: "sleep seconds", themeComponentID: CmpHelpViewSectionCodeBlock},
		{},
		{text: "For example, running the following will pause execution for 0.5 seconds:"},
		{},
		{text: "sleep 0.5", themeComponentID: CmpHelpViewSectionCodeBlock},
	}

	return []*HelpSection{
		{
			description: description,
		},
	}
}
