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
	ContextSetCmdParameterMissingError = cterror.Error("context name parameter is missing")
)

// contextSetCmd represents the create command
var contextSetCmd = &cobra.Command{
	Use:   "set",
	Short: "use a fly context",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cterror.Check(ContextSetCmdParameterMissingError)
		}
		contextSetCmdName := args[0]
		ctx, err := config.LoadContexts()
		cterror.Check(err)

		_, ok := ctx.Contexts[contextSetCmdName]
		if !ok {
			cterror.Check(config.ContextNotFound)
		}
		// mutual exclusive
		for key, v := range ctx.Contexts {
			if key == contextSetCmdName {
				v.InUse = true
				ctx.Contexts[key] = v
			} else {
				v.InUse = false
				ctx.Contexts[key] = v
			}
		}

		err = config.SaveContexts(ctx)
		cterror.Check(err)
		fmt.Printf("current context is set to %s\n", contextSetCmdName)
	},
}

func init() {
	contextCmd.AddCommand(contextSetCmd)
}
