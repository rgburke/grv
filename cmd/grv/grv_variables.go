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

// GRVVariableSetter sets the value of a GRV variable
type GRVVariableSetter interface {
	SetViewVariable(variable GRVVariable, value string, isActiveView bool)
	VariableValues() map[GRVVariable]string
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
func (grvVariables *GRVVariables) SetViewVariable(variable GRVVariable, value string, isActiveView bool) {
	if !activeViewOnlyVariables[variable] || isActiveView {
		grvVariables.SetVariable(variable, value)
	}
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
