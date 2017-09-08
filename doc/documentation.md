# GRV Documentation

## Introduction

GRV - Git Repository Viewer - is a TUI capable of displaying Git Repository
data. It provides a way to view refs, branches and diffs using vi like key
bindings.

GRV is comprised of three views:

 - **Ref View** - Lists branches and tags.
 - **Commit View** - Lists commits for the selected ref.
 - **Diff View** - Displays the diff for the selected commit.

## Key Bindings

The key bindings below are common to all views in GRV:

```
k       or <Up>         Move up one line
j       or <Down>       Move down one line
l       or <Right>      Scroll right
h       or <Left>       Scroll left
<C-b>   or <PageUp>     Move one page up
<C-f>   or <PageDown>   Move one page down
gg                      Move to first line
G                       Move to last line
/                       Search forwards
?                       Search backwards
n                       Move to next search match
N                       Move to last search match
:                       GRV Command prompt
<Tab>   or <C-w>w       Move to next view
<S-Tab> or <C-w>W       Move to previous view
f       or <C-w>o       Toggle current view full screen
<C-w>t                  Toggle views layout
<C-z>                   Suspend GRV
```

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

Below are the set of configuration commands supported:

### set

The set command allows configuration variables to be set. It has the form:

```
set variable value
```

Configuration variables available in GRV are:

```
 Variable | Type   | Description
 ---------+--------+----------------------------------------------
 tabwidth | int    | Tab character screen width (minimum value: 1)
 theme    | string | The currently active theme
```

For example, to set the tab width to tab width to 4 and the currently active
theme to "mytheme":

```
set tabwidth 4
set theme mytheme
```

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
theme --name mytheme --component RefView.Tag          --bgcolor Blue --fgcolor Yellow
theme --name mytheme --component StatusBarView.Normal --bgcolor None --fgcolor None
set theme mytheme
```

The set of possible colors is:

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

The set of screen components that can be customised is:

```
All.SearchMatch

CommitView.Author
CommitView.Date
CommitView.Footer
CommitView.LocalBranch
CommitView.RemoteBranch
CommitView.ShortOid
CommitView.Summary
CommitView.Tag
CommitView.Title

DiffView.AddedLine
DiffView.CommitAuthor
DiffView.CommitAuthorDate
DiffView.CommitCommitter
DiffView.CommitCommitterDate
DiffView.CommitSummary
DiffView.GitDiffExtendedHeader
DiffView.GitDiffHeader
DiffView.HunkHeader
DiffView.HunkStart
DiffView.Normal
DiffView.RemovedLine
DiffView.StatsFile
DiffView.UnifiedDiffHeader

ErrorView.Errors
ErrorView.Footer
ErrorView.Title

HelpBarView.Normal
HelpBarView.Special

RefView.Footer
RefView.LocalBranch
RefView.LocalBranchesHeader
RefView.RemoteBranch
RefView.RemoteBranchesHeader
RefView.Tag
RefView.TagsHeader
RefView.Title

StatusBarView.Normal
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

The set of views that can be customised is:

```
All
CommitView
DiffView
ErrorView
HelpBarView
HistoryView
RefView
StatusBarView
StatusView
```

GRV also has a text representation of actions that are independent of key
bindings. For example, the following commands can be used to make the `<Up>`
key move a line down and the `<Down>` key move a line up:

```
map All <Up>   <grv-next-line>
map All <Down> <grv-prev-line>
```

The set of actions available is:

```
<grv-clear-search>
<grv-exit>
<grv-suspend>
<grv-filter-prompt>
<grv-first-line>
<grv-full-screen-view>
<grv-last-line>
<grv-next-line>
<grv-next-page>
<grv-next-view>
<grv-nop>
<grv-prev-line>
<grv-prev-page>
<grv-prev-view>
<grv-prompt>
<grv-reverse-search-prompt>
<grv-scroll-left>
<grv-scroll-right>
<grv-search-find-next>
<grv-search-find-prev>
<grv-search-prompt>
<grv-select>
<grv-show-status>
<grv-toggle-view-layout>
```

### q

The quit command is used to exit GRV and can be used with the following
keys:

```
:q<Enter>
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

For more inforation about the supported GLOB syntax see 
[https://github.com/gobwas/glob](https://github.com/gobwas/glob)

For more information about the supported regex syntax see
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
```

The list of (case-insensitive) fields that can be used in the Ref View is:

```
 Field | Type
 ------+-------
 name  | string
```
