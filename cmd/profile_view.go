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

	"github.com/spf13/cobra"
	"github.com/maplain/control-tower/pkg/config"
	cterror "github.com/maplain/control-tower/pkg/error"
	"github.com/maplain/control-tower/pkg/io"
	"github.com/maplain/control-tower/pkg/secret"
)

var (
	viewName string
)

// viewCmd represents the view command
var viewCmd = &cobra.Command{
	Use:   "view",
	Short: "view contents of a profile",
	Long: `Examples:

ct profile view -n deploy-kubo --key=1234567891123456`,
	Run: func(cmd *cobra.Command, args []string) {
		d, err := readProfile(viewName)
		cterror.Check(err)

		yamld, err := io.DumpYaml(d)
		cterror.Check(err)

		fmt.Printf("%s", yamld)
	},
}

func readProfile(name string) (string, error) {
	filepath, err := config.GetProfilePath(name)
	if err != nil {
		return "", err
	}

	if io.NotExist(filepath) {
		return "", errors.New(fmt.Sprintf("profile with name %s does not exist. please use `ct profile list` to check available profiles", viewName))
	}

	ed, err := io.ReadFromFile(filepath)
	if err != nil {
		return "", err
	}

	d, err := secret.Decrypt(string(ed[:]), encryptionKey)
	if err != nil {
		return "", err
	}
	return d, nil
}

func init() {
	profileCmd.AddCommand(viewCmd)
	viewCmd.Flags().StringVarP(&viewName, "name", "n", "", "name of the profile")
	viewCmd.MarkFlagRequired("name")
}
