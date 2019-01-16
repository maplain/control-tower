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
	"sort"
	"strconv"

	"github.com/concourse/fly/rc"
	"github.com/maplain/control-tower/pkg/concourseclient"
	"github.com/maplain/control-tower/pkg/config"
	cterror "github.com/maplain/control-tower/pkg/error"
	"github.com/maplain/control-tower/pkg/io"
	"github.com/spf13/cobra"
)

var (
	flyBuildsCmdJobName string
	flyBuildsCmdJobNum  int
)

var flyBuildsCmd = &cobra.Command{
	Use:     "build",
	Aliases: []string{"b"},
	Short:   "builds of a job",
	Run: func(cmd *cobra.Command, args []string) {
		ctx, c, err := config.LoadInUseContext()
		cterror.Check(err)

		inusectx := ctx.Contexts[c]

		cli, err := concourseclient.NewConcourseClient(rc.TargetName(inusectx.Target))
		cterror.Check(err)

		if flyBuildsCmdJobNum > concourseclient.QueryBuildLimit {
			flyBuildsCmdJobNum = concourseclient.QueryBuildLimit
		}
		builds, err := cli.LatestJobBuilds(inusectx.Team, inusectx.Pipeline, flyBuildsCmdJobName, flyBuildsCmdJobNum)
		cterror.Check(err)

		var data [][]string
		sort.Slice(builds, func(i, j int) bool { return builds[i].ID < builds[j].ID })
		for _, b := range builds {
			data = append(data, []string{strconv.Itoa(b.ID), b.TeamName, b.Name, b.Status, b.JobName})
		}
		header := []string{"ID", "Team", "Name", "Status", "Job"}

		p, err := io.NewPrinter(outputFormat)
		p.Display(!outputNoHeader, data, header)
	},
}

func init() {
	flyCmd.AddCommand(flyBuildsCmd)
	flyBuildsCmd.Flags().StringVarP(&flyBuildsCmdJobName, "job", "j", "", "job name")
	flyBuildsCmd.Flags().IntVarP(&flyBuildsCmdJobNum, "num", "n", 5, "number of builds to display(maximum: 20)")
	flyBuildsCmd.MarkFlagRequired("job")
}
