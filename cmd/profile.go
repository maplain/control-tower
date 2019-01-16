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
	"github.com/pkg/errors"

	"github.com/maplain/control-tower/pkg/config"
	cterror "github.com/maplain/control-tower/pkg/error"
	"github.com/maplain/control-tower/pkg/io"
	"github.com/maplain/control-tower/templates"
	"github.com/spf13/cobra"
)

var (
	encryptionKey string
)

// profileCmd represents the profile command
var profileCmd = &cobra.Command{
	Use:     "profile",
	Aliases: []string{"p"},
	Short:   "manages profiles, i.e: configurations for pipelines",
}

func init() {
	rootCmd.AddCommand(profileCmd)
	profileCmd.PersistentFlags().StringVarP(&encryptionKey, "key", "k", config.DefaultEncryptionKey, "a key to encrypt templates and profiles, which has to be in length of 16, 24 or 32 bytes")
	profileCmd.MarkFlagRequired("key")
}

const (
	DeployKuboProfileType         = "kubo"
	NsxAcceptanceTestsProfileType = "nsx-acceptance-tests"

	TypeNotSupportedError = cterror.Error("pipeline type is not supported")
)

var profileRegistry map[string]io.Values = map[string]io.Values{
	DeployKuboProfileType: map[string]io.UserInput{
		"kubeconfig-bucket":                io.UserInput{Value: "vmw-pks-pipeline-store"},
		"kubeconfig-folder":                io.UserInput{Value: "pks-networking-kubeconfigs"},
		"pks-lock-branch":                  io.UserInput{Value: "master"},
		"pks-lock-pool":                    io.UserInput{Value: "nsx"},
		"pks-nsx-t-release-branch":         io.UserInput{Value: "master"},
		"pks-nsx-t-release-tarball-bucket": io.UserInput{Value: "vmw-pks-pipeline-store"},
		"pks-nsx-t-release-tarball-path":   io.UserInput{Value: "pks-nsx-t/pks-nsx-t-(.*).tgz"},
		//"lock-name":                        io.UserInput{}, // required
	},
	NsxAcceptanceTestsProfileType: map[string]io.UserInput{
		"kubeconfig-bucket":        io.UserInput{Value: "vmw-pks-pipeline-store"},
		"kubeconfig-path":          io.UserInput{Value: "pks-networking-kubeconfigs/kubeconfig-(.*).tgz"},
		"pks-nsx-t-release-branch": io.UserInput{Value: "ci-improvements-proto"},
		"lock-name":                io.UserInput{}, // required
	},
}

func ValidateProfileTypes(t string) error {
	_, ok := profileRegistry[t]
	if !ok {
		return errors.WithMessage(errors.Wrap(TypeNotSupportedError, "in profile registry"), t)
	}
	_, ok = profileOutputRegistry[t]
	if !ok {
		return errors.WithMessage(errors.Wrap(TypeNotSupportedError, "in output registry"), t)
	}
	return nil
}

var profileOutputRegistry map[string]templates.PipelineFetchOutputFunc = map[string]templates.PipelineFetchOutputFunc{
	DeployKuboProfileType:         templates.DeployKuboPipelineFetchOutputFunc,
	NsxAcceptanceTestsProfileType: templates.NsxAcceptanceTestsPipelineFetchOutputFunc,
}

var profileArtifactsRegistry map[string]templates.PipelineGetArtifactsFunc = map[string]templates.PipelineGetArtifactsFunc{
	DeployKuboProfileType: templates.DeployKuboPipelineGetArtifactsFunc,
}
