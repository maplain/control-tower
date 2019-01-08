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

	cterror "github.com/maplain/control-tower/pkg/error"
	"github.com/maplain/control-tower/templates"
	"github.com/spf13/cobra"
)

var (
	templateIntTemplatePath string
	templateIntVarPaths     []string
)

var templateIntCmd = &cobra.Command{
	Use:   "int",
	Short: "prints out required keys for a template",
	Run: func(cmd *cobra.Command, args []string) {
		res, err := templates.Interpolate(templateIntTemplatePath, templateIntVarPaths)
		cterror.Check(err)

		fmt.Printf("%s", res)
	},
}

func init() {
	templateCmd.AddCommand(templateIntCmd)
	templateIntCmd.Flags().StringVarP(&templateIntTemplatePath, "template", "t", "", "template file you want to interpolate")
	templateIntCmd.Flags().StringSliceVarP(&templateIntVarPaths, "load-vars-from", "l", templateIntVarPaths, "key value yaml files")
	templateIntCmd.MarkFlagRequired("template")
}
