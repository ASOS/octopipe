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

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a skeleton octopipe.yaml",
	Long: `
Use create to create a skeleton octopipe.yaml file
in the current location

`,
	Run: func(cmd *cobra.Command, args []string) {
		info, _ := os.Lstat("octopipe.yaml")
		if info != nil {
			logAndExitf("octopipe.yaml already exists, will not overwrite")
		}

		values := make(map[string]string)
		values["Dev"] = "dev_value"
		values["Test"] = "test_value"

		variables := make([]variable, 0)
		variables = append(variables, variable{Name: "variable1", Value: "value1"})
		variables = append(variables, variable{Name: "variable2", Values: values})
		variables = append(variables, variable{Name: "variable3", Value: "octopusdeploy-account", Type: "AzureAccount", Description: "Account used for deployments"})

		steps := make([]step, 0)
		steps = append(steps, step{Name: "Init", Type: "PowerShell", File: "scripts/init.ps1"})
		steps = append(steps, step{Name: "Deploy VM", Type: "PowerShell", File: "scripts/deploy.ps1"})

		op := octopipe{
			Project: project{
				Name:         "OctopusProject",
				Description:  "My Octopus Project",
				ProjectGroup: "OctopusProjectGroup",
				Lifecycle:    "OctopusProject.Lifecycle",
			},
			Variables: variables,
			Process: process{
				Steps: steps,
			},
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
}
