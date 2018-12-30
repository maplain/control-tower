// Copyright © 2018 NAME HERE <EMAIL ADDRESS>
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
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/maplain/control-tower/pkg/config"
	cterror "github.com/maplain/control-tower/pkg/error"
	"github.com/maplain/control-tower/pkg/io"
	"github.com/maplain/control-tower/pkg/secret"
	yaml "gopkg.in/yaml.v2"
)

var (
	configurations map[string]string
	profileName    string
	profileType    string
	overwrite      bool
	varFilePath    string
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		validate()
		var d []byte
		var err error

		if profileType == "" {
			if len(configurations) != 0 {
				d, err = yaml.Marshal(&configurations)
				cterror.Check(err)
			} else {
				d, err = io.ReadFromFile(varFilePath)
				cterror.Check(err)
			}
		} else {
			v := io.InteractivePopulateStringValues(profileRegistry[profileType])
			d, err = yaml.Marshal(&v)
			cterror.Check(err)
		}

		ed, err := secret.Encrypt(string(d[:]), encryptionKey)
		cterror.Check(err)

		filepath, err := config.GetProfilePath(profileName)
		cterror.Check(err)

		if !io.NotExist(filepath) {
			if !overwrite {
				fmt.Printf("profile with name %s already exists, set --overwrite if you want to overwrite it\n", profileName)
				return
			}
		}
		err = io.WriteToFile(ed, filepath)
		cterror.Check(err)
	},
}

func validate() {
	if profileType == "" {
		if len(configurations) == 0 && varFilePath == "" {
			cterror.Check(errors.New("at least one value should be provided for --vars and --var-file"))
		}
	} else {
		err := ValidateProfileTypes(profileType)
		cterror.Check(err)
	}
}

func init() {
	profileCmd.AddCommand(createCmd)
	createCmd.Flags().StringToStringVar(&configurations, "vars", make(map[string]string), "common separated key value pairs")
	createCmd.Flags().StringVar(&varFilePath, "var-file", "", "a file that has key-value pairs in yaml")
	createCmd.Flags().StringVarP(&profileName, "name", "n", "", "name of the profile")
	createCmd.Flags().StringVarP(&profileType, "type", "t", "", "type of the profile")
	createCmd.Flags().BoolVar(&overwrite, "overwrite", false, "whether to overwrite existing profile with the same name")
	createCmd.MarkFlagRequired("name")
}
