package cmd

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"strings"

	"gopkg.in/yaml.v2"
)

type project struct {
	Name         string `yaml:"name"`
	Description  string `yaml:"description"`
	ProjectGroup string `yaml:"group"`
	Lifecycle    string `yaml:"lifecycle"`
}

type variable struct {
	Name        string            `yaml:"name"`
	Value       string            `yaml:"value,omitempty"`
	Values      map[string]string `yaml:"values,omitempty"`
	Type        string            `yaml:"type,omitempty"`
	Description string            `yaml:"description,omitempty"`
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

type octopusProjectGroup struct {
	ID   string `json:"Id"`
	Name string `json:"Name"`
}

type octopusProject struct {
	ID                  string            `json:"Id"`
	Name                string            `json:"Name"`
	Description         string            `json:"Description"`
	VariableSetID       string            `json:"VariableSetId"`
	LifecycleID         string            `json:"LifecycleId"`
	ProjectGroupID      string            `json:"ProjectGroupId"`
	DeploymentProcessID string            `json:"DeploymentProcessId"`
	Links               map[string]string `json:"Links"`
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

func (v *octopusVariableSetScopeValues) getEnvironmentID(names string) (environmentID map[string][]string, err error) {
	tnames := strings.Split(names, ",")
	emap := make(map[string][]string)
	environments := make([]string, 0)

	for _, name := range tnames {
		found := false
		for _, env := range v.Environments {
			if env.Name == name {
				environments = append(environments, env.ID)
				found = true
			}
		}

		if !found {
			return nil, errors.New("Environment with name " + name + " not found")
		}
	}

	emap["Environment"] = environments
	return emap, nil
}
