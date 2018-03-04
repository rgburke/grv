# GRV Documentation

## Table Of Contents

 - [Introduction](#introduction)
 - [Command Line Arguments](#command-line-arguments)
 - [Key Bindings](#key-bindings)
     * [Movement](#movement)
     * [Search](#search)
     * [View Navigation](#view-navigation)
     * [General](#general)
     * [View Specific Bindings](#view-specific-bindings)
 - [Configuration](#configuration)
     * [set](#set)
     * [theme](#theme)
     * [map](#map)
     * [q](#q)
     * [addtab](#addtab)
     * [rmtab](#rmtab)
     * [addview](#addview)
     * [vsplit](#vsplit)
     * [hsplit](#hsplit)
     * [split](#split)
 - [Filter Query Language](#filter-query-language)

## Introduction

GRV - Git Repository Viewer - is a TUI capable of displaying Git Repository
data. It provides a way to view refs, branches and diffs using vi like key
bindings.

GRV is comprised of two main tabs

 - **History View** - This tab is composed of:
     - **Ref View** - Lists branches and tags.
     - **Commit View** - Lists commits for the selected ref.
     - **Diff View** - Displays the diff for the selected commit.

 - **Status View** - This tab is composed of:
     - **Git Status View** - Displays the status of the repository
     - **Diff View** - Displays the diff of any selected modified files

## Command Line Arguments

GRV accepts the following command line arguments:

```
-logFile string
        Log file path (default "grv.log")
-logLevel string
        Logging level [NONE|PANIC|FATAL|ERROR|WARN|INFO|DEBUG] (default "NONE")
-repoFilePath string
        Repository file path (default ".")
-version
        Print version
-workTreeFilePath string
        Work tree file path
```

## Key Bindings

The key bindings below are common to all views in GRV:

### Movement

```
k       or <Up>         Move up one line
j       or <Down>       Move down one line
l       or <Right>      Scroll right
h       or <Left>       Scroll left
<C-b>   or <PageUp>     Move one page up
<C-f>   or <PageDown>   Move one page down
<C-u>                   Move half page up
<C-d>                   Move half page down
gg                      Move to first line
G                       Move to last line
zz                      Center view
zt                      Scroll the screen so cursor is at the top
zb                      Scroll the screen so cursor is at the bottom
H                       Move to the first line of the page
M                       Move to the middle line of the page
L                       Move to the last line of the page
```

### Search

```
/                       Search forwards
?                       Search backwards
n                       Move to next search match
N                       Move to last search match
```

### View Navigation

```
<Tab>   or <C-w>w       Move to next view
<S-Tab> or <C-w>W       Move to previous view
f       or <C-w>o       Toggle current view full screen
<C-w>t                  Toggle views layout
gt                      Move to next tab
gT                      Move to previous tab
q                       Close view (or close tab if empty)
```

### General

```
<Enter>                 Select item (opens listener view if none exists)
:                       GRV Command prompt
<C-z>                   Suspend GRV
```

### View Specific Bindings

Ref View specific key bindings:

```
<Enter>                 Select ref and load commits
<C-q>                   Add ref filter
<C-r>                   Remove ref filter
```

Commit View specific key bindings:

```
<C-q>                   Add commit filter
<C-r>                   Remove commit filter
```

## Configuration

The behaviour of GRV can be customised through the use of commands specified
in a configuration file. GRV will look for the following configuration files
on start up:

 - `$XDG_CONFIG_HOME/grv/grvrc`
 - `$HOME/.config/grv/grvrc`

GRV will attempt to process the first file which exists. Commands can also be
specified within GRV using the command prompt `:`

GRV supports configuration commands, some of which operate on views. When a
view argument is required it will be one of the following values:

```
CommitView
DiffView
GitStatusView
HistoryView
RefView
```

Below are the set of configuration commands supported:

### set

The set command allows configuration variables to be set. It has the form:

```
set variable value
```

Configuration variables available in GRV are:

```
 Variable          | Type   | Default Value | Description
 ------------------+--------+---------------+----------------------------------------------
 tabwidth          | int    | 8             | Tab character screen width (minimum value: 1)
 theme             | string | solarized     | The currently active theme
 mouse             | bool   | false         | Mouse support enabled
 mouse-scroll-rows | int    | 3             | Number of rows scrolled for each mouse event
```

For example, to set the tab width to tab width to 4, the currently active
theme to "mytheme" and enable mouse support:

```
set tabwidth 4
set theme mytheme
set mouse true
```

GRV currently has 3 built in themes available:
 - solarized
 - classic
 - cold

The solarized theme is the default theme for GRV and uses the colours
specified [here](http://ethanschoonover.com/solarized). This theme does not 
respect the terminals colour palette. The classic and cold themes do
respect the terminals colour palette.

### theme

The theme command allows a custom theme to be defined. This theme can then be
activated using the theme config variable described above. The form of the
theme command is:

```
theme --name [ThemeName] --component [ComponentId] --bgcolor [BackgroundColor] --fgcolor [ForegroundColor]
```

 - ThemeName: The name of the theme to be created/updated.
 - ComponentId: The Id of the screen component (the part of the display to change).
 - BackgroundColor: The background color.
 - ForegroundColor: The foreground color.

Using a sequence of theme commands it is possible to define a theme. For
example, to define a new theme "mytheme" and set it as the active theme:

```
theme --name mytheme --component CommitView.Date      --bgcolor None --fgcolor Red
theme --name mytheme --component RefView.Tag          --bgcolor Blue --fgcolor 36
theme --name mytheme --component StatusBarView.Normal --bgcolor None --fgcolor f14a98
set theme mytheme
```

GRV supports 256 colors (when available). Provided colors will be mapped to
the nearest available color. The allowed color values are:

**System Colors**
```
None
Black
Red
Green
Yellow
Blue
Magenta
Cyan
White
```

**Terminal Color Numbers**
```
0 - 255
```

**Hex Colors**
```
000000 - ffffff
```

The set of screen components that can be customised is:

```
All.Default
All.SearchMatch
All.ActiveViewSelectedRow
All.InactiveViewSelectedRow

MainView.ActiveView
MainView.NormalView

RefView.Title
RefView.Footer
RefView.LocalBranchesHeader
RefView.RemoteBranchesHeader
RefView.LocalBranch
RefView.Head
RefView.RemoteBranch
RefView.TagsHeader
RefView.Tag

CommitView.Title
CommitView.Footer
CommitView.ShortOid
CommitView.Date
CommitView.Author
CommitView.Summary
CommitView.Tag
CommitView.LocalBranch
CommitView.RemoteBranch

DiffView.Title
DiffView.Footer
DiffView.Normal
DiffView.CommitAuthor
DiffView.CommitAuthorDate
DiffView.CommitCommitter
DiffView.CommitCommitterDate
DiffView.CommitMessage
DiffView.StatsFile
DiffView.GitDiffHeader
DiffView.GitDiffExtendedHeader
DiffView.UnifiedDiffHeader
DiffView.HunkStart
DiffView.HunkHeader
DiffView.AddedLine
DiffView.RemovedLine

GitStatusView.StagedTitle
GitStatusView.UnstagedTitle
GitStatusView.UntrackedTitle
GitStatusView.ConflictedTitle
GitStatusView.StagedFile
GitStatusView.UnstagedFile
GitStatusView.UntrackedFile
GitStatusView.ConflictedFile

StatusBarView.Normal

HelpBarView.Special
HelpBarView.Normal

ErrorView.Title
ErrorView.Footer
ErrorView.Errors
```

### map

The map command allows a key sequence to be mapped to an action or another key
sequence for a specified view. The form of the map command is:

```
map view fromkeys tokeys
```

For example, to map the key 'a' to the keys 'gg' in the Ref View:

```
map RefView a gg
```

When pressing 'a' in the Ref View, the first line would then become the
selected line, as 'gg' moves the cursor to the first line.

All is a valid view argument when a binding should apply to all views.

GRV also has a text representation of actions that are independent of key
bindings. For example, the following commands can be used to make the `<Up>`
key move a line down and the `<Down>` key move a line up:

```
map All <Up>   <grv-next-line>
map All <Down> <grv-prev-line>
```

The set of actions available is:

```
<grv-nop>
<grv-exit>
<grv-suspend>
<grv-prompt>
<grv-search-prompt>
<grv-reverse-search-prompt>
<grv-filter-prompt>
<grv-search>
<grv-reverse-search>
<grv-search-find-next>
<grv-search-find-prev>
<grv-clear-search>
<grv-next-line>
<grv-prev-line>
<grv-next-page>
<grv-prev-page>
<grv-scroll-right>
<grv-scroll-left>
<grv-first-line>
<grv-last-line>
<grv-select>
<grv-next-view>
<grv-prev-view>
<grv-full-screen-view>
<grv-toggle-view-layout>
<grv-center-view>
<grv-next-tab>
<grv-prev-tab>
<grv-remove-tab>
<grv-remove-view>
```

### q

The quit command is used to exit GRV and can be used with the following
keys:

```
:q<Enter>
```

### addtab

The addtab command creates a new named empty tab and switches to this new tab.
The format of the command is:

```
addtab tabname
```

For example, to add a new tab titled "mycustomtab" the following command can
be used:

```
addtab mycustomtab
```

### rmtab

The rmtab removes the currently active tab. If the tab removed is the last tab
then GRV will exit.

### addview

The addview command allows a view to be added to the currently active tab.
The form of the command is:

```
addview view viewargs...
```

Each view accepts a different set of arguments. This is described in the
table below:

```
 View          | Args
 --------------+-----------
 CommitView    | ref or oid
 DiffView      | oid
 GitStatusView | none
 RefView       | none
```

Examples usages for each view are given below:

```
addview CommitView origin/master
addview DiffView 4882ca9044661b49a26ae03ceb1be3a70d00c6a2
addview GitStatusView
addview RefView
```

### vsplit

The vsplit command creates a vertical split between the currently selected
view and the view specified in the command. The form of the command is:

```
vsplit view viewargs...
```

For example, to create a vertical split between the currently selected view
and a CommitView displaying commits for master:

```
vsplit CommitView master
```

### hsplit

The hsplit command creates a horizontal split between the currently selected
view and the view specified in the command. The form of the command is:

```
hsplit view viewargs...
```

For example, to create a horizontal split between the currently selected view
and a RefView:

```
hsplit RefView
```

### split

The split command is similar to the vsplit and hsplit commands. It creates
either a new vsplit or hsplit determined by the current dimensions of the
active view. The form of the command is:

```
split view viewargs...
```

## Filter Query Language

GRV has a built in query language which can be used to filter the content of
the Ref and Commit views. All queries resolve to boolean values which
are tested against each item listed in the view. A query is composed of at
least one comparison:

```
field CMP value
```

CMP can be any of the following comparison operators, which are
case-insensitive:

```
=, !=, >, >=, <, <=, GLOB, REGEXP
```

Value is one of the following types:

```
string          (e.g. "test")
number          (e.g. 123 or 123.0)
date            (e.g. "2017-09-05 10:05:25" or "2017-09-05")
```

Field is specific to the view that is being filtered.  For example,
to filter commits to those whose commit messages start with
"Bug Fix:":

```
summary GLOB "Bug Fix:*"
```

Or equivalently:

```
summary REGEXP "^Bug Fix:.*"
```

For more inforation about the supported GLOB syntax see:
[https://github.com/gobwas/glob](https://github.com/gobwas/glob)

For more information about the supported regex syntax see:
[https://golang.org/s/re2syntax](https://golang.org/s/re2syntax)

Comparisons can be composed together using the following logical operators,
which are case-insensitive:

```
AND, OR, NOT
```

For example, to filter commits to those authored by John Smith or Jane Roe
in September 2017, ignoring merge commits:

```
authordate >= "2017-09-01" AND authordate < "2017-10-01" AND (authorname = "John Smith" OR authorname = "Jane Roe") AND parentcount < 2
```

As shown above, expressions can be grouped using parentheses.

The list of (case-insensitive) fields that can be used in the Commit View is:

```
 Field          | Type
 ---------------+-------
 authordate     | date
 authoremail    | string
 authorname     | string
 committerdate  | date
 committeremail | string
 committername  | string
 id             | string
 parentcount    | number
 summary        | string
 message        | string
```

The list of (case-insensitive) fields that can be used in the Ref View is:

```
 Field | Type
 ------+-------
 name  | string
```
