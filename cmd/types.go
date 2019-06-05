package cmd

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"gopkg.in/yaml.v2"
)

var apiKey = os.Getenv("OCTOPUS_API_KEY")
var uri = os.Getenv("OCTOPUS_URI")
var client = http.Client{}
var validVariableTypes = []string{"AzureAccount", "AWSAccount", "Certificate", "String"}
var validScriptSyntaxTypes = []string{"PowerShell", "Bash", "CSharp", "FSharp"}
var validTenancyTypes = []string{"Tenanted", "Untenanted", "TenantedOrUntenanted"}
var validScopeTypes = []string{"TenantTag", "Environment", "Machine", "Channel", "Action", "Role"}

type project struct {
	Name         string `yaml:"name"`
	Description  string `yaml:"description"`
	ProjectGroup string `yaml:"group"`
	Lifecycle    string `yaml:"lifecycle"`
	Tenanted     string `yaml:"tenanted"`
}

type variable struct {
	Name         string              `yaml:"name"`
	Value        string              `yaml:"value,omitempty"`
	ScopedValues []map[string]string `yaml:"scopedValues,omitempty"`
	Type         string              `yaml:"type,omitempty"`
	Description  string              `yaml:"description,omitempty"`
}

type step struct {
	Name string `yaml:"name"`
	Type string `yaml:"type"`
	File string `yaml:"file"`
}

type process struct {
	Steps []step `yaml:"steps"`
}

type octopipe struct {
	Variables []variable `yaml:"variables"`
	Project   project    `yaml:"project"`
	Process   process    `yaml:"process"`
}

type octopusResource interface {
}

type octopusLifecycle struct {
	ID   string `json:"Id"`
	Name string `json:"Name"`
}

type octopusLifecycles []octopusLifecycle

type octopusProjectGroup struct {
	ID   string `json:"Id"`
	Name string `json:"Name"`
}

type octopusProjectGroups []octopusProjectGroup

type octopusProject struct {
	ID                     string            `json:"Id"`
	Name                   string            `json:"Name"`
	Description            string            `json:"Description"`
	VariableSetID          string            `json:"VariableSetId"`
	LifecycleID            string            `json:"LifecycleId"`
	ProjectGroupID         string            `json:"ProjectGroupId"`
	DeploymentProcessID    string            `json:"DeploymentProcessId"`
	TenantedDeploymentMode string            `json:"TenantedDeploymentMode"`
	Links                  map[string]string `json:"Links"`
}

type octopusVariable struct {
	Name        string              `json:"Name"`
	Value       string              `json:"Value"`
	Description string              `json:"Description"`
	IsSensitive bool                `json:"IsSensitive"`
	Scope       map[string][]string `json:"Scope,omitempty"`
	Type        string              `json:"Type"`
}

type octopusVariableSetScopeValue struct {
	ID   string `json:"Id"`
	Name string `json:"Name"`
}

type octopusVariableSetScopeValues struct {
	Environments []octopusVariableSetScopeValue `json:"Environments"`
	Machines     []octopusVariableSetScopeValue `json:"Machines"`
	Actions      []octopusVariableSetScopeValue `json:"Actions"`
	Roles        []octopusVariableSetScopeValue `json:"Roles"`
	Channels     []octopusVariableSetScopeValue `json:"Channels"`
	TenantTags   []octopusVariableSetScopeValue `json:"TenantTags"`
}

type octopusVariableSet struct {
	ID          string                        `json:"Id"`
	Variables   []octopusVariable             `json:"Variables"`
	Version     int                           `json:"Version"`
	Links       map[string]string             `json:"Links"`
	OwnerID     string                        `json:"OwnerId"`
	ScopeValues octopusVariableSetScopeValues `json:"ScopeValues"`
}

type octopusDeploymentAction struct {
	Name         string            `json:"Name"`
	ActionType   string            `json:"ActionType"`
	WorkerPoolID string            `json:"WorkerPoolId"`
	Properties   map[string]string `json:"Properties"`
}

type octopusDeploymentStep struct {
	Name    string                    `json:"Name"`
	Actions []octopusDeploymentAction `json:"Actions"`
}

type octopusDeploymentProcess struct {
	ID        string                  `json:"Id"`
	ProjectID string                  `json:"ProjectId"`
	Version   int                     `json:"Version"`
	Steps     []octopusDeploymentStep `json:"Steps"`
}

func getOctopusData(o octopusResource, uri string) {
	responsebody, status := doOctopusRequest(nil, uri, "GET")
	json.Unmarshal(responsebody, o)

	if status != 200 {
		logAndExitf("Failed to get Octopus resource:\n%s", string(responsebody))
	}
}

func putOctopusData(o octopusResource, uri string) {
	body, err := json.Marshal(o)
	if err != nil {
		logAndExitf("Failed to serialize Octopus resource before put:\n%s", err.Error())
	}
	responsebody, status := doOctopusRequest(body, uri, "PUT")

	if status != 200 {
		logAndExitf("Failed to put Octopus resource:\n%s", string(responsebody))
	}

	json.Unmarshal(responsebody, o)
}

func postOctopusData(o octopusResource, uri string) {
	body, err := json.Marshal(o)
	if err != nil {
		logAndExitf("Failed to serialize Octopus resource before post:\n%s", err.Error())
	}
	responsebody, status := doOctopusRequest(body, uri, "POST")

	if status != 201 {
		logAndExitf("Failed to post Octopus resource:\n%s", string(responsebody))
	}

	json.Unmarshal(responsebody, o)
}

func (op *octopipe) importOctopipeFile() {

	ofile, err := ioutil.ReadFile("octopipe.yaml")
	if err != nil {
		logAndExitf("Error opening octopipe.yaml:\n%s", err.Error())
	}

	err = yaml.Unmarshal(ofile, op)
	if err != nil {
		logAndExitf("Error importing octopipe.yaml:\n%s", err.Error())
	}
}

func (v *octopusVariableSetScopeValues) makeScopeDataSet() (dataset map[string][]octopusVariableSetScopeValue) {
	scmap := make(map[string][]octopusVariableSetScopeValue)
	scmap["Environment"] = v.Environments
	scmap["Machine"] = v.Machines
	scmap["Role"] = v.Roles
	scmap["TenantTag"] = v.TenantTags
	scmap["Action"] = v.Actions
	scmap["Channel"] = v.Channels

	return scmap
}

func (v *octopusVariableSetScopeValues) getScope(dataset map[string][]octopusVariableSetScopeValue, names string, IDs []string, scopeType string) (scopeIDs []string, scopeNames []string, err error) {
	if names != "" {
		tnames := strings.Split(names, ",")
		scopes := make([]string, 0)

		for _, name := range tnames {
			found := false
			for _, sc := range dataset[scopeType] {
				if scopeType == "TenantTag" {
					if sc.ID == name {
						scopes = append(scopes, sc.ID)
						found = true
					}
				} else {
					if sc.Name == name {
						scopes = append(scopes, sc.ID)
						found = true
					}
				}
			}
			if !found {
				return nil, nil, errors.New("Scope value with name '" + name + "' not found")
			}
		}
		return scopes, nil, nil
	} else if IDs != nil {
		names := make([]string, 0)
		for _, ID := range IDs {
			for _, sc := range dataset[scopeType] {
				if sc.ID == ID {
					if scopeType == "TenantTag" {
						names = append(names, sc.ID)
					} else {
						names = append(names, sc.Name)
					}
				}
			}
		}
		return nil, names, nil
	}

	return nil, nil, errors.New("No names or Ids to process")
}
