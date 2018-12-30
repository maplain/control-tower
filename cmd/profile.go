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

	"github.com/maplain/control-tower/pkg/io"
	"github.com/spf13/cobra"
)

var (
	encryptionKey string
)

// profileCmd represents the profile command
var profileCmd = &cobra.Command{
	Use:   "profile",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("profile called")
	},
}

func init() {
	rootCmd.AddCommand(profileCmd)
	profileCmd.PersistentFlags().StringVarP(&encryptionKey, "key", "k", "1234567891123456", "a key to encrypt templates and profiles, which has to be in length of 16, 24 or 32 bytes")
	profileCmd.MarkFlagRequired("key")
}

const (
	deployKuboProfileType = "deploy-kubo"
)

var profileRegistry map[string]io.Values = map[string]io.Values{
	deployKuboProfileType: map[string]string{
		"kubeconfig-bucket":        "vmw-pks-pipeline-store",
		"kubeconfig-folder":        "pks-networking-kubeconfigs",
		"pks-lock-branch":          "master",
		"pks-lock-pool":            "nsx",
		"pks-nsx-t-release-branch": "ci-improvements-proto",
	},
}

func ValidateProfileTypes(t string) error {
	switch t {
	case deployKuboProfileType:
		return nil
	default:
		return errors.New(fmt.Sprintf("%s profile type is not supported", t))
	}
	return nil
}
