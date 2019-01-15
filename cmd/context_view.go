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
	"github.com/spf13/cobra"
	yaml "gopkg.in/yaml.v2"
)

var (
	contextViewName string
)

// contextViewCmd represents the view command
var contextViewCmd = &cobra.Command{
	Use:   "v",
	Short: "view a specific context. if name is not provided, view in use context",
	Run: func(cmd *cobra.Command, args []string) {
		if contextViewName == "" {
			_, name, err := config.LoadInUseContext()
			cterror.Check(err)

			contextViewName = name
		}
		ctx, err := config.LoadContext(contextViewName)
		cterror.Check(err)

		data, err := yaml.Marshal(&ctx)
		cterror.Check(err)

		fmt.Printf("%s", string(data))
	},
}

func init() {
	contextCmd.AddCommand(contextViewCmd)

	contextViewCmd.Flags().StringVarP(&contextViewName, "name", "n", "", "name of the context")
}
