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
	"fmt"

	"github.com/maplain/control-tower/pkg/secret"
	"github.com/spf13/cobra"
)

var (
	secretLength int
)

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:     "generate",
	Aliases: []string{"g"},
	Short:   "generate a random secret as key for encryption/decryption",
	Long: `examples:

ct secret generate
ct secret generate -l 24`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("%s", secret.RandStringByBytes(secretLength))
	},
}

func init() {
	secretCmd.AddCommand(generateCmd)
	generateCmd.Flags().IntVarP(&secretLength, "length", "l", 16, "length of generated secret")
}
