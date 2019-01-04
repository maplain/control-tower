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
	"io/ioutil"

	"github.com/maplain/control-tower/pkg/config"
	cterror "github.com/maplain/control-tower/pkg/error"
	client "github.com/maplain/control-tower/pkg/flyclient"
	"github.com/maplain/control-tower/pkg/flyclient/flaghelpers"
	"github.com/spf13/cobra"
)

// logCmd represents the log command
var flyExecuteCmd = &cobra.Command{
	Use:   "execute",
	Short: "execute a fly task",
	Run: func(cmd *cobra.Command, args []string) {
		ctx, c, err := config.LoadInUseContext()
		cterror.Check(err)

		inuseContext := ctx.Contexts[c]

		dir, err := ioutil.TempDir("", "fly")
		cterror.Check(err)

		outputs := []flaghelpers.OutputPairFlag{
			flaghelpers.OutputPairFlag{
				Name: "artifacts",
				Path: dir,
			},
		}
		path := "/Users/fangyuanl/development/go/src/gitlab.eng.vmware.com/PKS/pks-nsx-t-release/ci/tasks/deploy-kubo-outputs.yml"
		err = client.Execute(path, inuseContext.Target, inuseContext.Pipeline, "outputs", outputs)
		cterror.Check(err)

		fmt.Printf("ls -al %s to check your downloaded artifacts\n", dir)
	},
}

func init() {
	flyCmd.AddCommand(flyExecuteCmd)
}
