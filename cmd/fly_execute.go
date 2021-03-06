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
	"github.com/spf13/cobra"
)

// logCmd represents the log command
var flyExecuteCmd = &cobra.Command{
	Use:   "execute",
	Short: "execute a fly task",
	Run: func(cmd *cobra.Command, args []string) {
	},
}

func init() {
	flyCmd.AddCommand(flyExecuteCmd)
}
