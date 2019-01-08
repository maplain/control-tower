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
	"fmt"
	"os"

	"github.com/concourse/fly/rc"
	"github.com/maplain/control-tower/pkg/concourseclient"
	"github.com/maplain/control-tower/pkg/config"
	cterror "github.com/maplain/control-tower/pkg/error"
	"github.com/spf13/cobra"
)

var (
	flyTriggerWatch bool
)

var flyTriggerCmd = &cobra.Command{
	Use:   "trigger",
	Short: "trigger a job build",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("ct fly trigger [job name]")
			os.Exit(1)
		}
		jobName := args[0]

		ctx, c, err := config.LoadInUseContext()
		cterror.Check(err)

		inusectx := ctx.Contexts[c]

		cli, err := concourseclient.NewConcourseClient(rc.TargetName(inusectx.Target))
		cterror.Check(err)

		jobs, err := cli.Team().ListJobs(inusectx.Pipeline)
		cterror.Check(err)

		for _, job := range jobs {
			if job.Name == jobName {
				err = concourseclient.TriggerJob(inusectx.Target, inusectx.Pipeline, jobName, flyTriggerWatch)
				cterror.Check(err)
				return
			}
		}
		fmt.Printf("%s does not exist in pipeline %s", jobName, inusectx.Pipeline)

		displayJobs(jobs)
	},
}

func init() {
	flyCmd.AddCommand(flyTriggerCmd)
	flyTriggerCmd.Flags().BoolVarP(&flyTriggerWatch, "watch", "w", false, "watch job build logs")
}
