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
	"os"

	"github.com/maplain/control-tower/pkg/config"
	cterror "github.com/maplain/control-tower/pkg/error"
	"github.com/spf13/cobra"
)

var varsKeysCmd = &cobra.Command{
	Use:   "keys",
	Short: "prints out keys of a var file",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("usage: ct vars keys [varfile path]")
			os.Exit(1)
		}
		path := args[0]
		keys, err := config.AllKeys(path)
		cterror.Check(err)

		for _, k := range keys {
			fmt.Println(k)
		}
	},
}

func init() {
	varsCmd.AddCommand(varsKeysCmd)
}
