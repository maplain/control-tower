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
	"os"

	"github.com/concourse/fly/rc"
	"github.com/maplain/control-tower/pkg/concourseclient"
	cterror "github.com/maplain/control-tower/pkg/error"
	"github.com/spf13/cobra"
)

var (
	logCmdBuildID string
)

// logCmd represents the log command
var logCmd = &cobra.Command{
	Use:   "logs",
	Short: "fetch the logs of a job build",
	Run: func(cmd *cobra.Command, args []string) {
		err := logCmdValidate()
		cterror.Check(err)

		cli, err := concourseclient.NewConcourseClient(rc.TargetName(flyTarget))
		cterror.Check(err)

		err = cli.ReadBuildLog(logCmdBuildID, os.Stdout)
		cterror.Check(err)
	},
}

func logCmdValidate() error {
	return nil
}

func init() {
	flyCmd.AddCommand(logCmd)

	logCmd.Flags().StringVarP(&logCmdBuildID, "id", "i", "", "build id")
	logCmd.MarkFlagRequired("id")
}
