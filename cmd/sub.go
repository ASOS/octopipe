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
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

// inCmd represents the in command
var subCmd = &cobra.Command{
	Use:   "sub",
	Short: "Substitute Octopus variables in the format '#{variable}' in your deployment scripts",
	Long: `
Use sub to substitute Octopus variables in the format '#{variable}'
in your deployment scripts. This is for simplified local debugging
before committing changes. Pass the parent directory of your
scripts as the first argument followed by the environment to target

Examples:

octopipe sub scripts/ DevTest
octopipe sub -c scripts/ Production
octopipe sub -c -f deploystep1.ps1,deploystep2.ps1 scripts/ DevTest

`,
	Args: cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {

		co, _ := cmd.Flags().GetBool("check-only")
		fo, _ := cmd.Flags().GetString("filenames")
		sdir := args[0]
		scopes := args[1]

		var op octopipe
		op.importOctopipeFile()

		asc := strings.Split(scopes, ",")
		sc := make(map[string]string)
		for _, tsc := range asc {
			tssc := strings.Split(tsc, "=")
			sc[tssc[0]] = tssc[1]
		}

		vmap := make(map[string]string)
		for _, thisv := range op.Variables {
			if thisv.Value != "" {
				vmap[thisv.Name] = thisv.Value
			} else if thisv.ScopedValues != nil {
				for i, rsc := range sc {
					for _, scv := range thisv.ScopedValues {
						if len(scv) == 1 {
							if vmap[thisv.Name] == "" {
								vmap[thisv.Name] = scv["value"]
							}
						}
						regsc, _ := regexp.Match(rsc, []byte(scv[i]))
						if regsc {
							vmap[thisv.Name] = scv["value"]
						}
					}
				}
			}
		}

		var files []os.FileInfo
		if fo == "" {
			dirfiles, err := ioutil.ReadDir(sdir)
			if err != nil {
				logAndExitf("Could not read files in directory:\n%s\n", err.Error())
			}
			files = dirfiles
		} else {
			filenames := strings.Split(fo, ",")
			for _, filename := range filenames {
				fileinfo, err := os.Lstat(sdir + "/" + filename)
				if err != nil {
					logAndExitf("Could not lstat file %s\n:%s\n", filename, err.Error())
				}
				files = append(files, fileinfo)
			}
		}

		for _, file := range files {
			match, _ := regexp.Match("\\.octopipe", []byte(file.Name()))
			if !match {

				b, err := ioutil.ReadFile(sdir + "/" + file.Name())
				if err != nil {
					logAndExitf("Error opening file %s for reading:\n%s\n", file.Name(), err.Error())
				}

				if !co {
					exregex := regexp.MustCompile("(\\.[0-9a-zA-Z]+)$")
					nfile := exregex.ReplaceAll([]byte(file.Name()), []byte(".octopipe$1"))
					err = ioutil.WriteFile(sdir+string(nfile), b, 0644)
					if err != nil {
						logAndExitf("Error creating backup of file %s:\n%s\n", file.Name(), err.Error())
					} else {
						fmt.Printf("Created a backup of file '%s' at '%s'\n", file.Name(), sdir+string(nfile))
					}
				}

				regex := regexp.MustCompile("#{(.*?)}")

				notfounds := make(map[string]string)

				i := 0

				for i < 5 {
					matches := regex.FindAllSubmatch(b, -1)

					for _, match := range matches {
						regex := regexp.MustCompile(string(match[0]))
						if vmap[string(match[1])] != "" {
							if !co {
								b = regex.ReplaceAll(b, []byte(vmap[string(match[1])]))
							}
						} else {
							if notfounds[string(match[0])] == "" {
								notfounds[string(match[0])] = string(match[1])
							}
						}
					}
					i++
				}

				if !co {
					err = ioutil.WriteFile(sdir+"/"+file.Name(), b, 0644)
					if err != nil {
						logAndExitf("Failed to write modified file:\n%s", err.Error())
					}
				}

				if len(notfounds) > 0 {
					fmt.Printf("The following variables in file '%s' do not have a value defined in octopipe.yaml (ok if they are secret):\n", file.Name())
					for _, nf := range notfounds {
						fmt.Printf("%s\n", nf)
					}
				}
			}
		}

	},
}

func init() {
	rootCmd.AddCommand(subCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// inCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// inCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	subCmd.Flags().BoolP("check-only", "c", false, "Check only for variables not present in octopipe.yaml (do not sub)")
	subCmd.Flags().StringP("filenames", "f", "", "File names, separated by comma. If not specifed all files in directory are subbed")
}
