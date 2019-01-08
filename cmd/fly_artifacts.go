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

	"github.com/maplain/control-tower/pkg/config"
	cterror "github.com/maplain/control-tower/pkg/error"
	"github.com/spf13/cobra"
)

// flyArtifactsCmd represents the fly artifacts command
var flyArtifactsCmd = &cobra.Command{
	Use:   "artifacts",
	Short: "get the artifacts of a pipeline",
	Run: func(cmd *cobra.Command, args []string) {
		ctx, c, err := config.LoadInUseContext()
		cterror.Check(err)

		inuseContext := ctx.Contexts[c]
		err = ValidateProfileTypes(inuseContext.PipelineType)
		cterror.Check(err)

		artifactsFunc := profileArtifactsRegistry[inuseContext.PipelineType]
		dir, err := artifactsFunc(inuseContext.Target, inuseContext.Pipeline)
		fmt.Printf("ls -al %s to check your downloaded artifacts\n", dir)
		cterror.Check(err)
	},
}

func init() {
	flyCmd.AddCommand(flyArtifactsCmd)
}
