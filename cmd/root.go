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
	"os"

	"github.com/maplain/control-tower/pkg/config"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile        string
	outputFormat   string
	outputNoHeader bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "ct",
	Short: "Control Tower that makes fly easier",
	Long: `ct manages profiles which are configurations for pipeline templates. You can create profile from key-value pairs or a variable file.

For example:
ct profile list
ct profile create  --vars="a=b,c=d" --name=deploy-kubo
ct profile create --var-file ../secrets/pks-nsx-t-release-secrets.yml  --name=pks-nsx-t-release-secrets.

You can also create profile interactively for supported type:
ct profile create --type deploy-kubo --name=test1

ct is able to fly pipeline based on provided template and profiles. For example:
ct deploy -p pks-nsx-t-release-secrets -p deploy-kubo -p common-secrets -t npks -m templates/deploy-kubo.yml -t test-pipeline`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//	Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.control-tower.yaml)")
	rootCmd.PersistentFlags().StringVar(&outputFormat, "output-format", "table", "output format")
	rootCmd.PersistentFlags().BoolVar(&outputNoHeader, "no-header", false, "if print out output header")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".control-tower" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(config.ConfigFilename)
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	//if err := viper.ReadInConfig(); err == nil {
	//		fmt.Println("Using config file:", viper.ConfigFileUsed())
	//	}
}

func rootValidate() error {
	_, _, err := config.LoadInUseContext()
	if err != nil {
		if err == config.NoContextInUseFoundError {
			return config.ContextNotSetError
		}
		return err
	}
	return nil
}
