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
	"strconv"

	"github.com/maplain/control-tower/pkg/config"
	cterror "github.com/maplain/control-tower/pkg/error"
	"github.com/maplain/control-tower/pkg/io"
	"github.com/spf13/cobra"
)

// contextListCmd represents the create command
var contextListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"l"},
	Short:   "list all fly contexts",
	Run: func(cmd *cobra.Command, args []string) {
		ctx, err := config.LoadContexts()
		cterror.Check(err)

		var data [][]string
		for name, c := range ctx.Contexts {
			data = append(data, []string{name, strconv.FormatBool(c.InUse)})
		}

		p, err := io.NewPrinter(outputFormat)
		cterror.Check(err)

		p.Display(!outputNoHeader, data, []string{"Name", "Use"})
	},
}

func init() {
	contextCmd.AddCommand(contextListCmd)
}
