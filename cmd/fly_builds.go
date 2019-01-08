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
	"github.com/concourse/fly/rc"
	"github.com/maplain/control-tower/pkg/concourseclient"
	"github.com/maplain/control-tower/pkg/config"
	cterror "github.com/maplain/control-tower/pkg/error"
	"github.com/spf13/cobra"
)

var flyBuildsCmd = &cobra.Command{
	Use: "jobs",
	Run: func(cmd *cobra.Command, args []string) {
		ctx, c, err := config.LoadInUseContext()
		cterror.Check(err)

		inusectx := ctx.Contexts[c]

		_, err = concourseclient.NewConcourseClient(rc.TargetName(inusectx.Target))
		//cli.Team().JobBuilds(pipelineName string, jobName string, page Page) ([]atc.Build, Pagination, bool, error)
		cterror.Check(err)
	},
}

func init() {
	flyCmd.AddCommand(flyBuildsCmd)
}
