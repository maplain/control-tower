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
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/concourse/fly/rc"
	"github.com/maplain/control-tower/pkg/config"
	cterror "github.com/maplain/control-tower/pkg/error"
	client "github.com/maplain/control-tower/pkg/flyclient"
	"github.com/maplain/control-tower/pkg/io"
	"github.com/spf13/cobra"
)

const (
	setPipelineCmd              = "set-pipeline"
	flyPipelineFlag             = "--pipeline"
	flyPipelineConfigFlag       = "--config"
	flyPipelineLoadVarsFromFlag = "--load-vars-from"
)

var (
	templatePath       string
	deployProfileNames []string
	deployProfilePaths []string
	deployTarget       string
	pipelineName       string
)

// deployCmd represents the deploy command
var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "deploys a pipeline with template specified and parameters saved in a profile",
	Long: `Examples:

ct deploy -t deploy-kubo -p deploy-kubo`,
	Run: func(cmd *cobra.Command, args []string) {
		err := deployCmdValidate()
		cterror.Check(err)

		dcmd := client.NewFlyCmd()
		if deployTarget != "" {
			dcmd.WithTarget(rc.TargetName(deployTarget))
		} else {
			ctx, name, err := config.LoadInUseContext()
			if err == nil {
				dcmd.WithTarget(rc.TargetName(ctx.Contexts[name].Target))
			}
		}
		dcmd.WithSubCommand(setPipelineCmd).
			WithArg(flyPipelineFlag, pipelineName).
			WithArg(flyPipelineConfigFlag, templatePath)

		for _, name := range deployProfileNames {
			profileData, err := readProfile(name)
			cterror.Check(err)

			tmpfile, err := ioutil.TempFile("", "profiles")
			cterror.Check(err)
			// clean up
			defer os.Remove(tmpfile.Name())

			err = io.WriteToFile(profileData, tmpfile.Name())
			cterror.Check(err)
			err = tmpfile.Close()
			cterror.Check(err)

			dcmd.WithArg(flyPipelineLoadVarsFromFlag, tmpfile.Name())
		}

		for _, path := range deployProfilePaths {
			dcmd.WithArg(flyPipelineLoadVarsFromFlag, path)
		}

		err = dcmd.Run()
		cterror.Check(err)
	},
}

func deployCmdValidate() error {
	for _, path := range deployProfilePaths {
		if io.NotExist(path) {
			return errors.New(fmt.Sprintf("%s does not exist", path))
		}
	}
	return nil
}

func init() {
	rootCmd.AddCommand(deployCmd)

	deployCmd.Flags().StringVarP(&templatePath, "template", "m", "", "path to pipeline template")
	deployCmd.Flags().StringVarP(&deployTarget, "target", "t", "", "fly target name")
	deployCmd.Flags().StringArrayVarP(&deployProfileNames, "profile-name", "p", nil, "profile name, can be used multiple times to specify many profiles to be used")
	deployCmd.Flags().StringArrayVar(&deployProfilePaths, "profile-path", nil, "profile path, can be used multiple times to specify many profile paths to be used")
	deployCmd.Flags().StringVarP(&pipelineName, "pipeline-name", "n", "", "pipeline name you want to set")
	deployCmd.MarkFlagRequired("template")
	deployCmd.MarkFlagRequired("pipeline-name")
}
