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

	"github.com/pkg/errors"

	"github.com/concourse/fly/rc"
	"github.com/maplain/control-tower/pkg/config"
	cterror "github.com/maplain/control-tower/pkg/error"
	client "github.com/maplain/control-tower/pkg/flyclient"
	"github.com/maplain/control-tower/pkg/io"
	"github.com/maplain/control-tower/templates"
	"github.com/spf13/cobra"
)

const (
	setPipelineCmd              = "set-pipeline"
	flyPipelineFlag             = "--pipeline"
	flyPipelineConfigFlag       = "--config"
	flyPipelineLoadVarsFromFlag = "--load-vars-from"

	EmptyParameterError = cterror.Error("parameter empty")
)

var (
	templatePath string
	templateType string

	deployProfileNames   []string
	deployProfilePaths   []string
	deployCmdProfileTags []string

	deployTarget string
	pipelineName string

	deployCmdEncryptionKey string
)

// deployCmd represents the deploy command
var deployCmd = &cobra.Command{
	Use:     "deploy",
	Aliases: []string{"d", "dp"},
	Short:   "deploys a pipeline with template specified and parameters saved in a profile",
	Long: `Examples:

ct deploy -t deploy-kubo -p deploy-kubo`,
	Run: func(cmd *cobra.Command, args []string) {
		err := deployCmdValidate()
		cterror.Check(err)

		if deployTarget == "" || pipelineName == "" {
			ctx, name, err := config.LoadInUseContext()
			cterror.Check(err)

			if deployTarget == "" {
				deployTarget = ctx.Contexts[name].Target
			}
			if pipelineName == "" {
				pipelineName = ctx.Contexts[name].Pipeline
			}
		}

		if templatePath == "" {
			if templateType != "" {
				template, found := templates.SupportedTemplateType[templateType]
				if !found {
					cterror.Check(errors.Wrap(templates.TemplateNotSupportedError, templateType))
				}
				tmpTemplateFile, remover, err := io.WriteToTempFile(template, "", "ct-deploy")
				defer remover()
				cterror.Check(err)

				// set templatePath to this temporary file path
				templatePath = tmpTemplateFile

			} else {
				cterror.Check(errors.Wrap(EmptyParameterError, "either --template or --template-type has to be provided"))
			}
		} else {
			if templateType != "" {
				cterror.Warnf("--template-type %s is overwritten by --template %s", templateType, templatePath)
			}
		}
		dcmd := client.NewFlyCmd().
			WithTarget(rc.TargetName(deployTarget)).
			WithSubCommand(setPipelineCmd).
			WithArg(flyPipelineFlag, pipelineName).
			WithArg(flyPipelineConfigFlag, templatePath)

		profiles, err := config.LoadProfileControlInfo()
		cterror.Check(err)

		profileNames := deployProfileNames

		for _, t := range deployCmdProfileTags {
			ps := profiles.GetProfilesByTag(t)
			for _, p := range ps {
				profileNames = append(profileNames, p.Name)
			}
		}

		for _, name := range profileNames {
			profileInfo, err := profiles.GetProfileInfoByName(name)
			cterror.Check(err)

			profileData, err := profiles.LoadProfileByName(name, deployCmdEncryptionKey)
			cterror.Check(err)

			if profileInfo.IsTemplate() {
				newProfileInfo, vars, err := profileInfo.PopulateTemplate()
				cterror.Check(err)
				fmt.Println(vars)
				profileData, err = templates.InterpolateContent(profileData, []string{vars})
				profiles.SaveProfileWithKey(newProfileInfo, false, profileData, deployCmdEncryptionKey)
				profiles.Save()
			}

			tmpfile, remover, err := io.WriteToTempFile(profileData, "", "profiles")
			defer remover()
			cterror.Check(err)

			dcmd.WithArg(flyPipelineLoadVarsFromFlag, tmpfile)
		}

		for _, path := range deployProfilePaths {
			dcmd.WithArg(flyPipelineLoadVarsFromFlag, path)
		}

		err = dcmd.Run()
		if err != nil {
			err = errors.Wrap(err, "run fly deploy cmd failed")
		}
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

	deployCmd.Flags().StringVarP(&templateType, "template-type", "t", "", "type of pipeline template")
	deployCmd.Flags().StringVarP(&templatePath, "template", "m", "", "path to pipeline template")
	deployCmd.Flags().StringVar(&deployTarget, "target", "", "fly target name")
	deployCmd.Flags().StringSliceVarP(&deployProfileNames, "profile-name", "p", deployProfileNames, "profile name, can be used multiple times to specify many profiles to be used")
	deployCmd.Flags().StringSliceVar(&deployCmdProfileTags, "profile-tag", deployCmdProfileTags, "tag of profile, can be used multiple times to specify many profiles to be used")
	deployCmd.Flags().StringSliceVar(&deployProfilePaths, "profile-path", nil, "profile path, can be used multiple times to specify many profile paths to be used")
	deployCmd.Flags().StringVarP(&pipelineName, "pipeline-name", "n", "", "pipeline name you want to set")
	deployCmd.Flags().StringVarP(&deployCmdEncryptionKey, "key", "k", config.DefaultEncryptionKey, "a key to encrypt templates and profiles, which has to be in length of 16, 24 or 32 bytes")
}
