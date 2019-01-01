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
	client "github.com/maplain/control-tower/pkg/flyclient"
	"github.com/maplain/control-tower/pkg/io"
	"github.com/spf13/cobra"
)

// targetsCmd represents the targets command
var targetsCmd = &cobra.Command{
	Use:   "targets",
	Short: "list out target names",
	Long: `Examples:

ct fly targets`,
	Run: func(cmd *cobra.Command, args []string) {
		fclient := client.NewFlyClient()
		targets := fclient.Targets()
		var data [][]string
		for _, t := range targets {
			data = append(data, []string{string(t)})
		}
		io.WriteTable(data, []string{"targets"})
	},
}

func init() {
	flyCmd.AddCommand(targetsCmd)
}
