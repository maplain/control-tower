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
	"github.com/maplain/control-tower/pkg/concourseclient"
	"github.com/maplain/control-tower/pkg/config"
	cterror "github.com/maplain/control-tower/pkg/error"
	"github.com/skratchdot/open-golang/open"
	"github.com/spf13/cobra"
)

// openCmd represents the open command
var openCmd = &cobra.Command{
	Use:   "open",
	Short: "open in use fly context",
	Run: func(cmd *cobra.Command, args []string) {
		var c config.Context
		ctx, name, err := config.LoadInUseContext()
		cterror.Check(err)

		if len(args) == 0 {
			c = ctx.Contexts[name]
		} else {
			var ok bool
			c, ok = ctx.Contexts[args[0]]
			if !ok {
				cterror.Check(config.ContextNotFound)
			}
		}

		u, err := concourseclient.GetPipelineURL(c.Target, c.Pipeline)
		cterror.Check(err)

		open.Start(u)
	},
}

func init() {
	rootCmd.AddCommand(openCmd)
}
