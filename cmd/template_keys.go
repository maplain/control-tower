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
	"github.com/pkg/errors"

	cterror "github.com/maplain/control-tower/pkg/error"
	"github.com/maplain/control-tower/pkg/io"
	"github.com/maplain/control-tower/templates"
	"github.com/spf13/cobra"
)

var (
	templateKeysCmdTemplateFilePath string
	templateKeyTemplateFormat       string
)

const (
	templateKeysUnsupportedOutputTypeError = cterror.Error("unsupported output type")
)

var templateKeysCmd = &cobra.Command{
	Use:   "keys",
	Short: "prints out required keys for a template",
	Run: func(cmd *cobra.Command, args []string) {
		if io.NotExist(templateKeysCmdTemplateFilePath) {
			cterror.Check(errors.WithMessage(templates.TemplateFileNotFoundError, templateKeysCmdTemplateFilePath))
		}
		data, err := io.ReadFromFile(templateKeysCmdTemplateFilePath)
		cterror.Check(err)

		keys, err := templates.AllUniqueKeys(string(data), templates.TemplateType(templateKeyTemplateFormat))
		cterror.Check(err)

		p, err := io.NewPrinter(outputFormat)
		cterror.Check(err)

		var res [][]string
		for _, key := range keys {
			res = append(res, []string{key})
		}
		p.Display(!outputNoHeader, res, []string{"keys"})
	},
}

func init() {
	templateCmd.AddCommand(templateKeysCmd)
	templateKeysCmd.Flags().StringVarP(&templateKeysCmdTemplateFilePath, "template", "t", "", "template file")
	templateKeysCmd.Flags().StringVar(&templateKeyTemplateFormat, "template-format", "bosh", "template file format")
	templateKeysCmd.MarkFlagRequired("template")
}
