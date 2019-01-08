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
)

var (
	contextDeleteName string
)

// contextDeleteCmd represents the view command
var contextDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "delete a specific context",
	Run: func(cmd *cobra.Command, args []string) {
		ctx, err := config.LoadContexts()
		cterror.Check(err)

		_, ok := ctx.Contexts[contextDeleteName]
		if !ok {
			fmt.Printf("context %s does not exist\n", contextDeleteName)
			return
		}

		delete(ctx.Contexts, contextDeleteName)
		err = config.SaveContexts(ctx)
		cterror.Check(err)
		fmt.Printf("context %s is deleted successfully\n", contextDeleteName)
	},
}

func init() {
	contextCmd.AddCommand(contextDeleteCmd)

	contextDeleteCmd.Flags().StringVarP(&contextDeleteName, "name", "n", "", "name of the context")
	contextDeleteCmd.MarkFlagRequired("name")
}
