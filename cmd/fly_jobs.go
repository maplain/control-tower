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
	"strconv"

	"github.com/concourse/atc"
	"github.com/concourse/fly/rc"
	"github.com/maplain/control-tower/pkg/concourseclient"
	"github.com/maplain/control-tower/pkg/config"
	cterror "github.com/maplain/control-tower/pkg/error"
	"github.com/maplain/control-tower/pkg/io"
	"github.com/spf13/cobra"
)

var flyJobsCmd = &cobra.Command{
	Use:     "jobs",
	Short:   "pipeline job information",
	Aliases: []string{"j"},
	Run: func(cmd *cobra.Command, args []string) {
		ctx, c, err := config.LoadInUseContext()
		cterror.Check(err)

		inusectx := ctx.Contexts[c]

		cli, err := concourseclient.NewConcourseClient(rc.TargetName(inusectx.Target))
		cterror.Check(err)

		jobs, err := cli.Team().ListJobs(inusectx.Pipeline)
		cterror.Check(err)

		err = displayJobs(jobs)
		cterror.Check(err)
	},
}

func displayJobs(jobs []atc.Job) error {
	p, err := io.NewPrinter(outputFormat)
	if err != nil {
		return err
	}

	data := [][]string{}
	for _, job := range jobs {
		data = append(data, []string{strconv.Itoa(job.ID), job.Name})
	}

	p.Display(!outputNoHeader, data, []string{"ID", "Name"})
	return nil
}

func init() {
	flyCmd.AddCommand(flyJobsCmd)
}
