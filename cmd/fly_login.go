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
	client "github.com/maplain/control-tower/pkg/flyclient"
	"github.com/spf13/cobra"
)

var (
	flyLoginCmdTarget string
)

// flyLoginCmd represents the log command
var flyLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "login current target",
	Run: func(cmd *cobra.Command, args []string) {
		err := flyLoginCmdValidate()
		cterror.Check(err)

		target := flyLoginCmdTarget
		if target == "" {
			ctx, name, err := config.LoadInUseContext()
			if err == nil {
				target = ctx.Contexts[name].Target
			}
		}

		err = client.Login(target)
		cterror.Check(err)

		fmt.Printf(fmt.Sprintf("log in target %s\n", target))
	},
}

func flyLoginCmdValidate() error {
	if flyLoginCmdTarget == "" {
		_, _, err := config.LoadInUseContext()
		if err != nil {
			return err
		}
	}
	return nil
}

func init() {
	flyCmd.AddCommand(flyLoginCmd)

	flyLoginCmd.Flags().StringVarP(&flyLoginCmdTarget, "target", "t", "", "fly target")
}
