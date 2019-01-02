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

	"github.com/concourse/fly/rc"
	"github.com/maplain/control-tower/pkg/concourseclient"
	"github.com/maplain/control-tower/pkg/config"
	cterror "github.com/maplain/control-tower/pkg/error"
	"github.com/spf13/cobra"
	yaml "gopkg.in/yaml.v2"
)

const (
	FlyOutputParameterMissingError = cterror.Error("pipeline name, team, or target is missing, or provide an valid context name")
)

var (
	flyOutputPipelineType string
	flyOutputPipelineName string
	flyOutputTeam         string
	flyOutputTarget       string
)

// outputCmd represents the output command
var outputCmd = &cobra.Command{
	Use:   "outputs",
	Short: "get the output of a pipeline",
	Run: func(cmd *cobra.Command, args []string) {
		err := flyOutputCmdValidate()
		cterror.Check(err)

		target := flyOutputTarget
		pipelineName := flyOutputPipelineName
		team := flyOutputTeam

		if target == "" || pipelineName == "" || team == "" {
			ctx, err := config.LoadContext(contextName)
			cterror.Check(err)
			if target == "" {
				target = ctx.Target
			}
			if pipelineName == "" {
				pipelineName = ctx.Pipeline
			}
			if team == "" {
				team = ctx.Team
			}
		}

		cli, err := concourseclient.NewConcourseClient(rc.TargetName(target))
		cterror.Check(err)

		err = ValidateProfileTypes(flyOutputPipelineType)
		cterror.Check(err)
		outputFunc := profileOutputRegistry[flyOutputPipelineType]
		output, err := outputFunc(team, pipelineName, cli)
		cterror.Check(err)

		d, err := yaml.Marshal(&output)
		cterror.Check(err)

		fmt.Println(string(d))

	},
}

func flyOutputCmdValidate() error {
	if flyOutputPipelineName == "" || flyOutputTarget == "" || flyOutputTeam == "" {
		if contextName != "" {
			_, err := config.LoadContext(contextName)
			if err != nil {
				return FlyOutputParameterMissingError
			} else {
				return nil
			}
		} else {
			return FlyOutputParameterMissingError
		}
	}
	return nil
}

func init() {
	flyCmd.AddCommand(outputCmd)

	outputCmd.Flags().StringVarP(&flyOutputPipelineType, "type", "p", "", "type of the pipeline")
	outputCmd.Flags().StringVarP(&flyOutputPipelineName, "pipeline-name", "n", "", "name of the pipeline")
	outputCmd.Flags().StringVarP(&flyOutputTeam, "team", "m", "", "team that owns the pipeline")
	outputCmd.Flags().StringVarP(&flyOutputTarget, "target", "t", "", "concourse target")
	outputCmd.MarkFlagRequired("type")
}
