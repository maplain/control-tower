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

	"github.com/maplain/control-tower/pkg/config"
	cterror "github.com/maplain/control-tower/pkg/error"
	client "github.com/maplain/control-tower/pkg/flyclient"
	"github.com/spf13/cobra"
)

var (
	contextCreateCmdTarget       string
	contextCreateCmdPipeline     string
	contextCreateCmdName         string
	contextCreateCmdOverwrite    bool
	contextCreateCmdPipelineType string
)

// contextCreateCmd represents the create command
var contextCreateCmd = &cobra.Command{
	Use:   "c",
	Short: "creates a fly context",
	Run: func(cmd *cobra.Command, args []string) {
		err := contextCreateValidate()
		cterror.Check(err)

		ctx := config.Context{
			Target:   contextCreateCmdTarget,
			Pipeline: contextCreateCmdPipeline,
		}
		target, _ := client.LoadTarget(ctx.Target)
		ctx.Team = target.TeamName

		if contextCreateCmdPipelineType != "" {
			ctx.PipelineType = contextCreateCmdPipelineType
		}
		err = ctx.Save(contextCreateCmdName, contextCreateCmdOverwrite)
		if err == config.ContextAlreadyExist {
			fmt.Printf("context with name %s already exists. use --overwrite if you want\n", contextCreateCmdName)
			return
		}
		cterror.Check(err)

		fmt.Printf("context %s is created\nuse ct context view -n %s to check details\n", contextCreateCmdName, contextCreateCmdName)
	},
}

func contextCreateValidate() error {
	if contextCreateCmdPipelineType != "" {
		err := ValidateProfileTypes(contextCreateCmdPipelineType)
		if err != nil {
			return err
		}
	}
	_, err := client.LoadTarget(contextCreateCmdTarget)
	if err != nil {
		return err
	}
	return nil
}

func init() {
	contextCmd.AddCommand(contextCreateCmd)
	contextCreateCmd.Flags().StringVarP(&contextCreateCmdTarget, "target", "t", "", "fly target name")
	contextCreateCmd.Flags().StringVar(&contextCreateCmdPipeline, "pipeline-name", "", "fly pipeline name")
	contextCreateCmd.Flags().StringVarP(&contextCreateCmdName, "name", "n", "", "context name")
	contextCreateCmd.Flags().StringVar(&contextCreateCmdPipelineType, "pipeline-type", "", "pipeline type")
	contextCreateCmd.Flags().BoolVarP(&contextCreateCmdOverwrite, "overwrite", "o", false, "if overwrite existing context with the same name")

	contextCreateCmd.MarkFlagRequired("target")
	contextCreateCmd.MarkFlagRequired("pipeline")
	contextCreateCmd.MarkFlagRequired("name")
}
