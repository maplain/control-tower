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

var (
	varsGetVarFiles  []string
	varsGetKey       string
	varsGetOnlyValue bool
)

var varsGetCmd = &cobra.Command{
	Use:   "get",
	Short: "get value of key in a var file",
	Run: func(cmd *cobra.Command, args []string) {
		if varsGetKey == "" {
			fmt.Println("key can not be empty")
			os.Exit(1)
		}

		res := make(map[string]interface{})
		for _, f := range varsGetVarFiles {
			v, err := config.GetValue(f, varsGetKey)
			if err != config.KeyNotFoundError {
				cterror.Check(err)
			}
			if err == nil {
				res[f] = v
			}
		}
		for k, v := range res {
			if varsGetOnlyValue {
				fmt.Printf("%s\n", v)
			} else {
				fmt.Printf("%s\n%s\n", k, v)
			}
		}
	},
}

func init() {
	varsCmd.AddCommand(varsGetCmd)
	varsGetCmd.Flags().StringSliceVarP(&varsGetVarFiles, "var-file", "v", []string{}, "var files")
	varsGetCmd.Flags().StringVarP(&varsGetKey, "key", "k", "", "key")
	varsGetCmd.Flags().BoolVar(&varsGetOnlyValue, "only-value", true, "only print out values")
	varsGetCmd.MarkFlagRequired("var-file")
	varsGetCmd.MarkFlagRequired("key")
}
