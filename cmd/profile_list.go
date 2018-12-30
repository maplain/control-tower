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
	"path"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/maplain/control-tower/pkg/config"
	cterror "github.com/maplain/control-tower/pkg/error"
	"github.com/maplain/control-tower/pkg/io"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "list all local profile names",
	Long: `Examples:

ct profile list.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Find home directory.
		home, err := homedir.Dir()
		cterror.Check(err)

		files, err := io.GetFilenames(path.Join(home, config.ProfileFolder))
		cterror.Check(err)

		var entries [][]string
		for _, f := range files {
			entries = append(entries, []string{path.Base(f)})
		}
		io.WriteTable(entries, []string{"Profile Name"})
	},
}

func init() {
	profileCmd.AddCommand(listCmd)
}
