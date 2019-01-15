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
	cterror "github.com/maplain/control-tower/pkg/error"
	"github.com/maplain/control-tower/pkg/io"
	"github.com/spf13/cobra"
)

// typesCmd represents the types command
var typesCmd = &cobra.Command{
	Use:   "types",
	Short: "list out supported built in profile types",
	Run: func(cmd *cobra.Command, args []string) {
		res := [][]string{}
		for name, _ := range profileRegistry {
			res = append(res, []string{name})
		}

		p, err := io.NewPrinter(outputFormat)
		cterror.Check(err)

		p.Display(!outputNoHeader, res, []string{"Type"})
	},
}

func init() {
	profileCmd.AddCommand(typesCmd)
}
