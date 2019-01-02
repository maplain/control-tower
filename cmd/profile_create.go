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

	"github.com/maplain/control-tower/pkg/config"
	cterror "github.com/maplain/control-tower/pkg/error"
	"github.com/maplain/control-tower/pkg/io"
	"github.com/maplain/control-tower/pkg/secret"
	"github.com/spf13/cobra"
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
	Long: `a profile can be created from key-value pairs, eg:
ct profile create  --vars="a=b,c=d" --name=deploy-kubo

or from a variable file, eg:
ct profile create --var-file ../secrets/common-secrets.yml  --name=common-secrets

or interactively for a supported built-in type, eg:
ct profile create --type deploy-kubo --name test

You can find out all supported types by:
ct profile types`,
	Run: func(cmd *cobra.Command, args []string) {
		err := profileCreateValidate()
		cterror.Check(err)

		var d []byte

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
		err = io.WriteToFile(ed, filepath)
		cterror.Check(err)
	},
}

func profileCreateValidate() error {
	if profileType == "" {
		if len(configurations) == 0 && varFilePath == "" {
			return errors.New("at least one value should be provided for --vars and --var-file")
		}
	} else {
		err := ValidateProfileTypes(profileType)
		if err != nil {
			return err
		}
	}
	filepath, err := config.GetProfilePath(profileName)
	if err != nil {
		return err
	}

	if !io.NotExist(filepath) {
		if !overwrite {
			return errors.New(fmt.Sprintf("profile with name %s already exists, set --overwrite if you want to overwrite it\n", profileName))
		}
	}
	return nil
}

func init() {
	profileCmd.AddCommand(createCmd)
	createCmd.Flags().StringToStringVar(&configurations, "vars", make(map[string]string), "common separated key value pairs, eg: a=b,c=d")
	createCmd.Flags().StringVar(&varFilePath, "var-file", "", "a yaml file that contains key-value pairs")
	createCmd.Flags().StringVarP(&profileName, "name", "n", "", "name of the profile")
	createCmd.Flags().StringVarP(&profileType, "type", "t", "", "type of the profile")
	createCmd.Flags().BoolVar(&overwrite, "overwrite", false, "whether to overwrite existing profile with the same name")
	createCmd.MarkFlagRequired("name")
}
