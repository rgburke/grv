package main

import (
	"sync"
)

// GRVVariable contains the state of a GRV property
type GRVVariable int

// The set of GRV variables
const (
	VarHead GRVVariable = iota
	VarBranch
	VarTag
	VarCommit
	VarFile
	VarDiffViewFile
	VarLineText
	VarLineNumer
	VarLineCount
	VarRepoPath
	VarRepoWorkDir

	VarCount
)

type variableDescriptor struct {
	variable       GRVVariable
	name           string
	description    string
	activeViewOnly bool
}

var variableDescriptors = []variableDescriptor{
	{
		variable:    VarHead,
		name:        "head",
		description: "Value of HEAD",
	},
	{
		variable:    VarBranch,
		name:        "branch",
		description: "Selected branch",
	},
	{
		variable:    VarTag,
		name:        "tag",
		description: "Selected tag",
	},
	{
		variable:    VarCommit,
		name:        "commit",
		description: "Selected commit",
	},
	{
		variable:    VarFile,
		name:        "file",
		description: "Selected file",
	},
	{
		variable:    VarDiffViewFile,
		name:        "diff-view-file",
		description: "Selected DiffView file",
	},
	{
		variable:       VarLineText,
		name:           "line-text",
		description:    "Selected lines content",
		activeViewOnly: true,
	},
	{
		variable:       VarLineNumer,
		name:           "line-number",
		description:    "Selected line number",
		activeViewOnly: true,
	},
	{
		variable:       VarLineCount,
		name:           "line-count",
		description:    "Number of lines in the active view",
		activeViewOnly: true,
	},
	{
		variable:    VarRepoPath,
		name:        "repo-path",
		description: "Repository file path",
	},
	{
		variable:    VarRepoWorkDir,
		name:        "repo-workdir",
		description: "Work directory path",
	},
}

var activeViewOnlyVariables = map[GRVVariable]bool{}
var variableNameDescriptorMap = map[string]*variableDescriptor{}
var variableNameMap = map[GRVVariable]*variableDescriptor{}

func init() {
	for index, variableDescriptor := range variableDescriptors {
		activeViewOnlyVariables[variableDescriptor.variable] = variableDescriptor.activeViewOnly
		variableNameDescriptorMap[variableDescriptor.name] = &variableDescriptors[index]
		variableNameMap[variableDescriptor.variable] = &variableDescriptors[index]
	}
}

// GRVVariableName returns the name of the provided variable
func GRVVariableName(variable GRVVariable) string {
	return variableNameMap[variable].name
}

// LookupGRVVariable returns the variable referenced by the provided name
func LookupGRVVariable(variableName string) (variable GRVVariable, exists bool) {
	variableDescriptor, exists := variableNameDescriptorMap[variableName]
	if exists {
		variable = variableDescriptor.variable
	}

	return
}

// GRVVariableGetter can read the values of GRV variables
type GRVVariableGetter interface {
	VariableValue(GRVVariable) (value string, isSet bool)
	VariableValues() map[GRVVariable]string
}

// GRVVariableSetter sets the value of a GRV variable
type GRVVariableSetter interface {
	GRVVariableGetter
	SetViewVariable(variable GRVVariable, value string, viewState ViewState)
	ClearViewVariable(variable GRVVariable, viewState ViewState)
}

// GRVVariables stores the values of all variables
type GRVVariables struct {
	values map[GRVVariable]string
	lock   sync.Mutex
}

// NewGRVVariables creates a new instance
func NewGRVVariables() *GRVVariables {
	return &GRVVariables{
		values: make(map[GRVVariable]string),
	}
}

// SetVariable sets the value of a GRV variable
func (grvVariables *GRVVariables) SetVariable(variable GRVVariable, value string) {
	grvVariables.lock.Lock()
	defer grvVariables.lock.Unlock()

	grvVariables.values[variable] = value
}

// SetViewVariable sets the value of a GRV variable for a view
func (grvVariables *GRVVariables) SetViewVariable(variable GRVVariable, value string, viewState ViewState) {
	if !activeViewOnlyVariables[variable] || viewState == ViewStateActive {
		grvVariables.SetVariable(variable, value)
	}
}

// ClearViewVariable clears the value of a GRV variable for a view
func (grvVariables *GRVVariables) ClearViewVariable(variable GRVVariable, viewState ViewState) {
	grvVariables.SetViewVariable(variable, "", viewState)
}

// VariableValues returns the current values of all variables
func (grvVariables *GRVVariables) VariableValues() map[GRVVariable]string {
	grvVariables.lock.Lock()
	defer grvVariables.lock.Unlock()

	values := map[GRVVariable]string{}

	for variable, value := range grvVariables.values {
		values[variable] = value
	}

	return values
}

// VariableValue returns the value of the provided variable and a boolean denoting if it's set
func (grvVariables *GRVVariables) VariableValue(variable GRVVariable) (value string, isSet bool) {
	grvVariables.lock.Lock()
	defer grvVariables.lock.Unlock()

	value, isSet = grvVariables.values[variable]
	return
}

// GenerateGRVVariablesHelpSection generates help information for GRV variables
func GenerateGRVVariablesHelpSection(config Config) *HelpSection {
	headers := []TableHeader{
		{text: "Variable", themeComponentID: CmpHelpViewSectionTableHeader},
		{text: "Description", themeComponentID: CmpHelpViewSectionTableHeader},
	}

	tableFormatter := NewTableFormatterWithHeaders(headers, config)
	tableFormatter.SetGridLines(true)

	tableFormatter.Resize(uint(len(variableDescriptors)))

	for rowIndex, variableDescriptor := range variableDescriptors {
		tableFormatter.SetCellWithStyle(uint(rowIndex), 0, CmpHelpViewSectionTableRow, "%v", variableDescriptor.name)
		tableFormatter.SetCellWithStyle(uint(rowIndex), 1, CmpHelpViewSectionTableRow, "%v", variableDescriptor.description)
	}

	return &HelpSection{
		tableFormatter: tableFormatter,
	}
}
