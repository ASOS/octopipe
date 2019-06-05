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
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"gopkg.in/yaml.v2"

	"github.com/spf13/cobra"
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create an octopipe.yaml file",
	Long: `
Use create to create a skeleton octopipe.yaml file or
import an existing Octopus project.  If using -i, specify
the project name you wish to create from

Examples

octopipe create
octopipe create -i My.Octopus.Project

`,
	Run: func(cmd *cobra.Command, args []string) {

		pn, _ := cmd.Flags().GetString("import")

		if pn != "" {

			if apiKey == "" || uri == "" {
				logAndExitf("Octopus Api Key and Octopus Uri must be specified in environment variables with names OCTOPUS_API_KEY and OCTOPUS_URI")
			}

			// Project
			slug := getProjectSlug(pn)

			p := octopusProject{}
			v := octopusVariableSet{}
			d := octopusDeploymentProcess{}
			l := octopusLifecycles{}
			g := octopusProjectGroups{}

			getOctopusData(&p, uri+"/api/projects/"+slug)
			getOctopusData(&v, uri+"/api/variables/"+p.VariableSetID)
			getOctopusData(&d, uri+"/api/deploymentprocesses/"+p.DeploymentProcessID)
			getOctopusData(&l, uri+"/api/lifecycles/all")
			getOctopusData(&g, uri+"/api/projectgroups/all")

			tl, _ := getLifecycle(l, "", p.LifecycleID)
			tg, _ := getProjectGroup(g, "", p.ProjectGroupID)

			ds := v.ScopeValues.makeScopeDataSet()

			// Variables
			vss := []variable{}

			for _, vs := range v.Variables {
				exist := false
				for i, sv := range vss {
					if sv.Name == vs.Name {
						exist = true
						tv := variable{}
						tv.Name = vs.Name
						tv.Description = vs.Description
						if vs.Type != "String" {
							tv.Type = vs.Type
						}

						values := sv.ScopedValues

						if vs.Scope != nil {
							scvalue := make(map[string]string)
							for sc, es := range vs.Scope {
								_, scopeNames, _ := v.ScopeValues.getScope(ds, "", es, sc)
								j := ""
								for _, sc := range scopeNames {
									j = j + sc + ","
								}
								j = strings.TrimRight(j, ",")
								scvalue[sc] = j
							}
							scvalue["value"] = vs.Value
							values = append(values, scvalue)
						} else {
							scvalue := make(map[string]string)
							scvalue["value"] = vs.Value
							values = append(values, scvalue)
						}

						tv.ScopedValues = values

						vss[i] = tv

						break
					}
				}
				if exist {
					continue
				}

				tv := variable{}
				tv.Name = vs.Name
				tv.Description = vs.Description
				if vs.Type != "String" {
					tv.Type = vs.Type
				}

				if len(vs.Scope) != 0 {
					values := make([]map[string]string, 0)
					scvalue := make(map[string]string)
					for sc, es := range vs.Scope {
						_, scopeNames, _ := v.ScopeValues.getScope(ds, "", es, sc)
						j := ""
						for _, sc := range scopeNames {
							j = j + sc + ","
						}
						j = strings.TrimRight(j, ",")
						scvalue[sc] = j
					}
					scvalue["value"] = vs.Value
					values = append(values, scvalue)
					tv.ScopedValues = values
				} else {
					tv.Value = vs.Value
				}

				vss = append(vss, tv)
			}

			// Deployment process
			dss := process{}
			dsa := []step{}

			_, err := os.Stat("scripts")
			if os.IsNotExist(err) {
				os.Mkdir("scripts", 0644)
			}

			ext := make(map[string]string)
			ext["PowerShell"] = "ps1"
			ext["Bash"] = "sh"
			ext["FSharp"] = "f"
			ext["Csharp"] = "c"

			for _, s := range d.Steps {
				for _, a := range s.Actions {
					if a.ActionType == "Octopus.Script" {
						script := a.Properties["Octopus.Action.Script.ScriptBody"]
						filename := getProjectSlug(a.Name)
						ffilename := "scripts/" + filename + "." + ext[a.Properties["Octopus.Action.Script.Syntax"]]

						err = ioutil.WriteFile(ffilename, []byte(script), 0644)
						if err != nil {
							logAndExitf("Failed to write deployment script to disk:\n%s\n", err.Error())
						}
						ts := step{}
						ts.Name = a.Name
						ts.Type = a.Properties["Octopus.Action.Script.Syntax"]
						ts.File = ffilename

						dsa = append(dsa, ts)
					}
				}
			}

			dss.Steps = dsa

			op := octopipe{}

			op.Project.Name = p.Name
			op.Project.Description = p.Description
			op.Project.Tenanted = p.TenantedDeploymentMode
			op.Project.Lifecycle = tl.Name
			op.Project.ProjectGroup = tg.Name

			op.Variables = vss

			op.Process = dss

			info, _ := os.Lstat("octopipe.yaml")
			if info != nil {
				logAndExitf("octopipe.yaml already exists, will not overwrite")
			}

			contents, err := yaml.Marshal(op)
			if err != nil {
				logAndExitf("Failed to serialize yaml data:\n%s\n", err.Error())
			}

			err = ioutil.WriteFile("octopipe.yaml", contents, 0644)
			if err != nil {
				logAndExitf("Failed to write file to disk:\n%s\n", err.Error())
			}

			os.Exit(0)
		}

		info, _ := os.Lstat("octopipe.yaml")
		if info != nil {
			logAndExitf("octopipe.yaml already exists, will not overwrite")
		}

		values := make([]map[string]string, 0)
		values1 := make(map[string]string)
		values2 := make(map[string]string)
		values1["Environment"] = "Dev,Test"
		values1["value"] = "aks-devtest-rg"
		values2["Environment"] = "Production"
		values2["Machine"] = "deploynode01"
		values2["value"] = "aks-pd-rg"
		values = append(values, values1, values2)

		variables := make([]variable, 0)
		variables = append(variables, variable{Name: "AksName", Value: "aks-01"})
		variables = append(variables, variable{Name: "AksResourceGroupName", ScopedValues: values})
		variables = append(variables, variable{Name: "AzureUsername", Value: "user@github.com"})
		variables = append(variables, variable{Name: "AzureTenantId", Value: "1234-5678-abcd-efgh"})
		variables = append(variables, variable{Name: "variable3", Value: "octopusdeploy-account", Type: "AzureAccount", Description: "Account used for deployments"})

		steps := make([]step, 0)
		steps = append(steps, step{Name: "Init", Type: "PowerShell", File: "scripts/init.ps1"})
		steps = append(steps, step{Name: "Deploy VM", Type: "PowerShell", File: "scripts/deploy.ps1"})

		op := octopipe{
			Project: project{
				Name:         "OctopusProject",
				Description:  "My Octopus Project",
				ProjectGroup: "Octopus Project Group",
				Lifecycle:    "OctopusProject.Lifecycle",
				Tenanted:     "TenantedOrUntenanted",
			},
			Variables: variables,
			Process: process{
				Steps: steps,
			},
		}

		inits := `$credential = New-Object PSCredential -ArgumentList @("#{AzureUsername}", ("#{AzurePassword}" | ConvertTo-SecureString -Force -AsPlainText))
			Login-AzAccount -ServicePrincipal -Tenant "#{AzureTenantId}" -Credential $credential
			`
		deps := `az aks create -g "#{AksResourceGroupName}" -n "#{AksName}" --node-count 5
			$cluster = az aks show -n "#{AksName}" -g "#{AksResourceGroupName}" | ConvertFrom-Json
			Write-Host $cluster.Status`

		_, err := os.Stat("scripts")
		if os.IsNotExist(err) {
			os.Mkdir("scripts", 0644)
		}

		err = ioutil.WriteFile("scripts/init.ps1", []byte(inits), 0644)
		if err != nil {
			logAndExitf("Failed to write example script to disk:\n%s\n", err.Error())
		}

		err = ioutil.WriteFile("scripts/deploy.ps1", []byte(deps), 0644)
		if err != nil {
			logAndExitf("Failed to write example script to disk:\n%s\n", err.Error())
		}

		contents, err := yaml.Marshal(op)
		if err != nil {
			logAndExitf("Failed to serialize yaml data:\n%s\n", err.Error())
		}

		err = ioutil.WriteFile("octopipe.yaml", contents, 0644)
		if err != nil {
			logAndExitf("Failed to write file to disk:\n%s\n", err.Error())
		}

		fmt.Println("octopipe.yaml created")

	},
}

func init() {
	rootCmd.AddCommand(createCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// createCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// createCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	createCmd.Flags().StringP("import", "i", "", "Create an octopipe.yaml file from an existing project, supply the project name")
}
