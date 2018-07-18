package main

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
	VarRepoWorkTree

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
		name:           "line-numer",
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
		variable:    VarRepoWorkTree,
		name:        "repo-worktree",
		description: "Work tree file path",
	},
}

var activeViewOnlyVariables = map[GRVVariable]bool{}
var variableNameMap = map[string]*variableDescriptor{}

func init() {
	for _, variableDescriptor := range variableDescriptors {
		activeViewOnlyVariables[variableDescriptor.variable] = variableDescriptor.activeViewOnly
		variableNameMap[variableDescriptor.name] = &variableDescriptor
	}
}

// GRVVariableSetterClient sets the value of a GRV variable
type GRVVariableSetterClient interface {
	SetVariable(GRVVariable, string)
}

// GRVVariableSetter creates a view client
type GRVVariableSetter interface {
	GRVVariableSetterClient
	ViewClient(isActiveView bool) GRVVariableSetterClient
}

// GRVVariables stores the values of all variables
type GRVVariables struct {
	values map[GRVVariable]string
}

// NewGRVVariables creates a new instance
func NewGRVVariables() GRVVariableSetter {
	return &GRVVariables{
		values: make(map[GRVVariable]string),
	}
}

// SetVariable sets the value of a GRV variable
func (grvVariables *GRVVariables) SetVariable(variable GRVVariable, value string) {
	grvVariables.values[variable] = value
}

// ViewClient creates a view client
func (grvVariables *GRVVariables) ViewClient(isActive bool) GRVVariableSetterClient {
	return newGRVVariablesWindowClient(grvVariables, isActive)
}

type grvVariablesClient struct {
	variables *GRVVariables
	isActive  bool
}

func newGRVVariablesWindowClient(variables *GRVVariables, isActive bool) *grvVariablesClient {
	return &grvVariablesClient{
		variables: variables,
		isActive:  isActive,
	}
}

// SetVariable sets the value of a GRV variable
// Values may not be set for variables valid only for the active view
func (client *grvVariablesClient) SetVariable(variable GRVVariable, value string) {
	if activeViewOnlyVariables[variable] && !client.isActive {
		return
	}

	client.variables.SetVariable(variable, value)
}
