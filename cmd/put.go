// Copyright © 2019 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/spf13/cobra"
)

var apiKey = os.Getenv("OCTOPUS_API_KEY")
var uri = os.Getenv("OCTOPUS_URI")
var client = http.Client{}
var validVariableTypes = []string{"AzureAccount", "AWSAccount", "Certificate", "String"}
var validScriptSyntaxTypes = []string{"PowerShell", "Bash", "CSharp", "FSharp"}

// putCmd represents the put command
var putCmd = &cobra.Command{
	Use:   "put",
	Short: "Write the local configuration data to Octopus",
	Long: `
Use the put command to write the local
project configuration data into Octopus

Usage:

octopipe put

`,
	Run: func(cmd *cobra.Command, args []string) {

		// Start
		start := time.Now()

		if apiKey == "" || uri == "" {
			logAndExitf("Octopus Api Key and Octopus Uri must be specified in environment variables with names OCTOPUS_API_KEY and OCTOPUS_URI")
		}

		var op octopipe
		op.importOctopipeFile()

		//Project
		apg := make([]octopusProjectGroup, 0)
		al := make([]octopusLifecycle, 0)

		pgresp, status := doOctopusRequest(nil, uri+"/api/projectgroups/all", "GET")
		if status != 200 {
			logAndExitf("Failed to fetch Octopus Project Groups:\n%s", string(pgresp))
		}
		json.Unmarshal(pgresp, &apg)

		lresp, status := doOctopusRequest(nil, uri+"/api/lifecycles/all", "GET")
		if status != 200 {
			logAndExitf("Failed to fetch Octopus Lifecycles:\n%s", string(lresp))
		}
		json.Unmarshal(lresp, &al)

		lifecycle, err := getLifecycle(al, op.Project.Lifecycle)
		if err != nil {
			logAndExitf(err.Error())
		}

		projectGroup, err := getProjectGroup(apg, op.Project.ProjectGroup)
		if err != nil {
			logAndExitf(err.Error())
		}

		p := &octopusProject{}
		slug := getProjectSlug(op.Project.Name)
		presp, status := doOctopusRequest(nil, uri+"/api/projects/"+slug, "GET")
		json.Unmarshal(presp, &p)

		if status == 404 {

			newp := &octopusProject{
				Name:           op.Project.Name,
				Description:    op.Project.Description,
				LifecycleID:    lifecycle.ID,
				ProjectGroupID: projectGroup.ID,
			}

			postOctopusData(newp, uri+"/api/projects")
			p = newp

		} else {

			p.Name = op.Project.Name
			p.LifecycleID = lifecycle.ID
			p.ProjectGroupID = projectGroup.ID
			p.Description = op.Project.Description

			putOctopusData(p, uri+"/api/projects/"+p.ID)
		}

		// Variables
		v := &octopusVariableSet{}
		getOctopusData(v, uri+"/api/variables/"+p.VariableSetID)

		newv := []octopusVariable{}

		for _, sv := range op.Variables {
			if len(sv.Values) != 0 {
				for i, svv := range sv.Values {
					thistype, err := verifyVariableType(sv)
					if err != nil {
						logAndExitf(err.Error())
					}
					if i == "local" {
						// do nothing
					} else if i == "default" {
						tv := octopusVariable{
							Name:        sv.Name,
							Value:       svv,
							Type:        thistype,
							Description: sv.Description,
						}
						newv = append(newv, tv)
					} else {
						envID, err := v.ScopeValues.getEnvironmentID(i)
						if err != nil {
							logAndExitf("Variable %s:\n%s", sv.Name, err.Error())
						} else {
							tv := octopusVariable{
								Name:        sv.Name,
								Value:       svv,
								Scope:       envID,
								Type:        thistype,
								Description: sv.Description,
							}
							newv = append(newv, tv)
						}
					}
				}
			}
			thistype, err := verifyVariableType(sv)
			if err != nil {
				logAndExitf(err.Error())
			}
			if sv.Value != "" {
				tv := octopusVariable{
					Name:        sv.Name,
					Value:       sv.Value,
					Type:        thistype,
					Description: sv.Description,
				}
				newv = append(newv, tv)
			}
		}

		v.Variables = newv
		putOctopusData(v, uri+"/api/variables/"+p.VariableSetID)

		// Deployment process
		d := &octopusDeploymentProcess{}
		getOctopusData(d, uri+"/api/deploymentprocesses/"+p.DeploymentProcessID)

		news := make([]octopusDeploymentStep, 0)

		for _, s := range op.Process.Steps {
			newa := make([]octopusDeploymentAction, 0)

			taa, err := ioutil.ReadFile(s.File)
			if err != nil {
				logAndExitf("Error opening %s:\n%s", s.File, err.Error())
			}

			tap := make(map[string]string)

			thistype, err := verifySyntaxType(s)
			if err != nil {
				logAndExitf(err.Error())
			}

			tap["Octopus.Action.Script.Syntax"] = thistype
			tap["Octopus.Action.Script.ScriptSource"] = "Inline"
			tap["Octopus.Action.Script.ScriptBody"] = string(taa)
			tap["Octopus.Action.RunOnServer"] = "True"

			ta := octopusDeploymentAction{
				Name:         s.Name,
				ActionType:   "Octopus.Script",
				WorkerPoolID: "WorkerPools-1",
				Properties:   tap,
			}

			newa = append(newa, ta)

			ts := octopusDeploymentStep{
				Name:    s.Name,
				Actions: newa,
			}

			news = append(news, ts)
		}

		d.Steps = news
		putOctopusData(d, uri+"/api/deploymentprocesses/"+p.DeploymentProcessID)

		// End
		finish := time.Now()
		elapsed := finish.Sub(start)
		fmt.Printf("Put completed successfully in %gs\n", elapsed.Seconds())
	},
}

func init() {
	rootCmd.AddCommand(putCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// putCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// putCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
