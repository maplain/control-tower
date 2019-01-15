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
	"strings"

	"github.com/concourse/fly/rc"
	"github.com/maplain/control-tower/pkg/concourseclient"
	"github.com/maplain/control-tower/pkg/config"
	cterror "github.com/maplain/control-tower/pkg/error"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	yaml "gopkg.in/yaml.v2"
)

const (
	JobNotFoundError = cterror.Error("job not found")
)

var (
	flyGetConfigCmdJobName string
)

var flyGetConfigCmd = &cobra.Command{
	Use:   "c",
	Short: "get pipeline config",
	Run: func(cmd *cobra.Command, args []string) {
		ctx, c, err := config.LoadInUseContext()
		cterror.Check(err)

		inusectx := ctx.Contexts[c]

		cli, err := concourseclient.NewConcourseClient(rc.TargetName(inusectx.Target))
		cterror.Check(err)

		config, _, _, _, err := cli.Team().PipelineConfig(inusectx.Pipeline)
		cterror.Check(err)

		var d []byte
		if flyGetConfigCmdJobName == "" {
			pconfig, err := yaml.Marshal(config)
			cterror.Check(err)
			d = pconfig
		} else {
			for _, j := range config.Jobs {
				if j.Name == flyGetConfigCmdJobName {
					jconfig, err := yaml.Marshal(j)
					cterror.Check(err)
					d = jconfig
				}
			}
			if strings.TrimSpace(string(d[:])) == "" {
				cterror.Check(errors.Wrap(JobNotFoundError, flyGetConfigCmdJobName))
			}
		}

		fmt.Printf("%s", string(d[:]))
	},
}

func init() {
	flyCmd.AddCommand(flyGetConfigCmd)
	flyGetConfigCmd.Flags().StringVarP(&flyGetConfigCmdJobName, "job", "j", "", "job name")
}
