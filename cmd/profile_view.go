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
	"github.com/maplain/control-tower/pkg/io"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var (
	viewName string
)

// viewCmd represents the view command
var viewCmd = &cobra.Command{
	Use:     "view",
	Aliases: []string{"v"},
	Short:   "view contents of a profile",
	Long: `Examples:

ct profile view -n deploy-kubo --key=1234567891123456`,
	Run: func(cmd *cobra.Command, args []string) {
		profiles, err := config.LoadProfileControlInfo()
		cterror.Check(err)

		profileNames := []string{}
		if viewName != "" {
			profileNames = append(profileNames, viewName)
		}
		// profile cmd global parameter
		for _, tag := range profileTags {
			ps := profiles.GetProfilesByTag(tag)
			for _, p := range ps {
				profileNames = append(profileNames, p.Name)
			}
		}

		if len(profileNames) == 0 {
			cterror.Check(errors.Wrap(EmptyParameterError, "no profile match provided --name or --tags"))
		}

		for _, name := range profileNames {
			d, err := profiles.LoadProfileByName(name, encryptionKey)
			cterror.Check(err)

			yamld, err := io.DumpYaml(d)
			cterror.Check(err)

			fmt.Printf("%s", yamld)
		}
	},
}

func init() {
	profileCmd.AddCommand(viewCmd)
	viewCmd.Flags().StringVarP(&viewName, "name", "n", "", "name of the profile")
}
