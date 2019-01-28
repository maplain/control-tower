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
	"github.com/maplain/control-tower/templates"
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
	flyOutputJobStatus    string
)

// outputCmd represents the output command
var outputCmd = &cobra.Command{
	Use:     "outputs",
	Aliases: []string{"o"},
	Short:   "get the output of a pipeline",
	Run: func(cmd *cobra.Command, args []string) {
		err := flyOutputCmdValidate()
		cterror.Check(err)

		target := flyOutputTarget
		pipelineName := flyOutputPipelineName
		team := flyOutputTeam
		pipelineType := flyOutputPipelineType

		if target == "" || pipelineName == "" || team == "" || pipelineType == "" {
			ctx, name, _ := config.LoadInUseContext()
			c := ctx.Contexts[name]
			if target == "" {
				target = c.Target
			}
			if pipelineName == "" {
				pipelineName = c.Pipeline
			}
			if team == "" {
				team = c.Team
			}
			if pipelineType == "" {
				pipelineType = c.PipelineType
			}
		}

		cli, err := concourseclient.NewConcourseClient(rc.TargetName(target))
		cterror.Check(err)

		err = ValidateProfileTypes(pipelineType)
		cterror.Check(err)
		outputFunc := profileOutputRegistry[pipelineType]
		outputArgs := templates.Args(map[string]interface{}{})
		outputArgs[templates.JobStatusArgKey] = flyOutputJobStatus
		output, err := outputFunc(team, pipelineName, cli, outputArgs)
		cterror.Check(err)

		d, err := yaml.Marshal(&output)
		cterror.Check(err)

		fmt.Println(string(d))

	},
}

func flyOutputCmdValidate() error {
	err := rootValidate()
	if err != nil {
		if err == config.ContextNotSetError {
			if flyOutputPipelineName != "" && flyOutputTarget != "" && flyOutputTeam != "" && flyOutputPipelineType != "" {
				return nil
			} else {
				return FlyOutputParameterMissingError
			}
		}
		return err
	}
	if flyOutputPipelineType == "" {
		ctx, name, _ := config.LoadInUseContext()
		if ctx.Contexts[name].Target == "" {
			return FlyOutputParameterMissingError
		}
	}
	return nil
}

func init() {
	flyCmd.AddCommand(outputCmd)

	outputCmd.Flags().StringVar(&flyOutputPipelineType, "type", "", "type of the pipeline")
	outputCmd.Flags().StringVarP(&flyOutputPipelineName, "pipeline-name", "n", "", "name of the pipeline")
	outputCmd.Flags().StringVarP(&flyOutputTeam, "team", "m", "", "team that owns the pipeline")
	outputCmd.Flags().StringVarP(&flyOutputTarget, "target", "t", "", "concourse target")
	outputCmd.Flags().StringVar(&flyOutputJobStatus, "job-status", "succeeded", "expected job status")
}
