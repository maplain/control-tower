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

	"github.com/pkg/errors"

	cterror "github.com/maplain/control-tower/pkg/error"
	"github.com/maplain/control-tower/pkg/io"
	"github.com/maplain/control-tower/templates"
	"github.com/spf13/cobra"
)

var (
	templateConvertCmdFrom     string
	templateConvertCmdTo       string
	templateConvertCmdTemplate string
)

var templateConvertCmd = &cobra.Command{
	Use:   "convert",
	Short: "converts template from one format to the other one",
	Run: func(cmd *cobra.Command, args []string) {
		if io.NotExist(templateConvertCmdTemplate) {
			cterror.Check(errors.WithMessage(templates.TemplateFileNotFoundError, templateConvertCmdTemplate))
		}
		data, err := io.ReadFromFile(templateConvertCmdTemplate)
		cterror.Check(err)

		res, err := templates.ConvertTemplate(string(data), templates.TemplateType(templateConvertCmdFrom), templates.TemplateType(templateConvertCmdTo))
		cterror.Check(err)

		fmt.Println(res)
	},
}

func init() {
	templateCmd.AddCommand(templateConvertCmd)
	templateConvertCmd.Flags().StringVarP(&templateConvertCmdTemplate, "template", "t", templateConvertCmdTemplate, "template file")
	templateConvertCmd.Flags().StringVar(&templateConvertCmdFrom, "from", string(templates.RubyTemplateType), "source template format")
	templateConvertCmd.Flags().StringVar(&templateConvertCmdTo, "to", string(templates.BoshTemplateType), "target template format")
	templateConvertCmd.MarkFlagRequired("template")
}
