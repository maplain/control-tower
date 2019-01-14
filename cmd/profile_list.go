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
	"strconv"

	"github.com/maplain/control-tower/pkg/config"
	cterror "github.com/maplain/control-tower/pkg/error"
	"github.com/maplain/control-tower/pkg/io"
	"github.com/spf13/cobra"
)

var (
	profileListCmdProfileTag string
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "list all local profile names",
	Long: `Examples:

ct profile list.`,
	Run: func(cmd *cobra.Command, args []string) {
		profiles, err := config.LoadProfileControlInfo()
		cterror.Check(err)

		var entries [][]string
		var ps []config.Profile
		if profileListCmdProfileTag != "" {
			ps = profiles.GetProfilesByTag(profileListCmdProfileTag)
		} else {
			ps = profiles.GetProfileInfos()
		}
		for _, p := range ps {
			entries = append(entries, []string{p.Name, p.Tags.String(), strconv.FormatBool(p.IsTemplate())})
		}
		io.WriteTable(entries, []string{"Profile Name", "Tags", "IsTemplate"})
	},
}

func init() {
	profileCmd.AddCommand(listCmd)
	listCmd.Flags().StringVarP(&profileListCmdProfileTag, "tag", "t", "", "tag of profiles")
}
