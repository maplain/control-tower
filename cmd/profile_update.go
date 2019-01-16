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
	"fmt"

	"github.com/maplain/control-tower/pkg/config"
	cterror "github.com/maplain/control-tower/pkg/error"
	"github.com/spf13/cobra"
	yaml "gopkg.in/yaml.v2"
)

var (
	profileUpdateCmdVars        map[string]string
	profileUpdateCmdProfileName string
	profileUpdateCmdDeleteKeys  []string
)

// profileUpdateCmd represents the profileUpdate command
var profileUpdateCmd = &cobra.Command{
	Use:     "update",
	Aliases: []string{"u", "ud"},
	Short:   "update content of a profile",
	Long: `only supports simple key-value now. examples:

ct profile profile -n deploy-kubo -k `,
	Run: func(cmd *cobra.Command, args []string) {
		profiles, err := config.LoadProfileControlInfo()
		cterror.Check(err)

		d, err := profiles.LoadProfileByName(profileUpdateCmdProfileName, encryptionKey)
		cterror.Check(err)

		res := make(map[string]interface{})
		err = yaml.Unmarshal([]byte(d), &res)
		cterror.Check(err)

		for k, v := range profileUpdateCmdVars {
			res[k] = v
		}

		for _, k := range profileUpdateCmdDeleteKeys {
			delete(res, k)
		}

		yamld, err := yaml.Marshal(&res)
		cterror.Check(err)

		err = profiles.UpdateProfileData(profileUpdateCmdProfileName, string(yamld[:]), encryptionKey)
		cterror.Check(err)

		fmt.Printf("profile %s updated:\n%s", profileUpdateCmdProfileName, string(yamld[:]))
	},
}

func init() {
	profileCmd.AddCommand(profileUpdateCmd)
	profileUpdateCmd.Flags().StringToStringVar(&profileUpdateCmdVars, "vars", make(map[string]string), "common separated key value pairs, eg: a=b,c=d")
	profileUpdateCmd.Flags().StringVarP(&profileUpdateCmdProfileName, "profile-name", "n", "", "name of the profile to be updated")
	profileUpdateCmd.Flags().StringSliceVarP(&profileUpdateCmdDeleteKeys, "delete", "d", []string{}, "name of the key in profile to be deleted")
	profileUpdateCmd.MarkFlagRequired("vars")
	profileUpdateCmd.MarkFlagRequired("profile-name")
}
