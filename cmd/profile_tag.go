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
)

var (
	profileTagCmdProfileName string
	profileTagCmdTags        []string
	profileTagCmdDeleteTags  []string
)

// profileTagCmd represents the view command
var profileTagCmd = &cobra.Command{
	Use:     "tag",
	Aliases: []string{"t"},
	Short:   "tag a profile",
	Long: `Examples:
ct profile tag -n deploy-kubo -t=kubo`,
	Run: func(cmd *cobra.Command, args []string) {
		profiles, err := config.LoadProfileControlInfo()
		cterror.Check(err)

		for _, tag := range profileTagCmdTags {
			err = profiles.AddTagForProfile(profileTagCmdProfileName, tag)
			cterror.Check(err)
		}
		for _, tag := range profileTagCmdDeleteTags {
			err = profiles.RemoveTagForProfile(profileTagCmdProfileName, tag)
			cterror.Check(err)
		}

		err = profiles.Save()
		cterror.Check(err)

		p, err := profiles.GetProfileInfoByName(profileTagCmdProfileName)
		cterror.Check(err)

		tagString := p.Tags.String()
		if tagString != "" {
			fmt.Printf("tags of profile %s are updated to %s\n", profileTagCmdProfileName, p.Tags.String())
		} else {
			fmt.Printf("tags of profile %s are all removed\n", profileTagCmdProfileName)
		}
	},
}

func init() {
	profileCmd.AddCommand(profileTagCmd)
	profileTagCmd.Flags().StringVarP(&profileTagCmdProfileName, "name", "n", "", "name of the profile")
	profileTagCmd.Flags().StringSliceVarP(&profileTagCmdTags, "tag", "t", profileTagCmdTags, "tag of the profile, can be used multiple times to specify different tags")
	profileTagCmd.Flags().StringSliceVarP(&profileTagCmdDeleteTags, "delete", "d", profileTagCmdDeleteTags, "tag of the profile to be deleted")
	profileTagCmd.MarkFlagRequired("name")
}
