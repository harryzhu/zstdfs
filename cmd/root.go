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
	"hazhufs/util"
	"os"

	"github.com/BurntSushi/toml"
	//log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "hazhufs",
	Short: "hazhufs is a simple file system for web image hosting.",
	Long:  `hazhufs is a simple file system for web image hosting.`,
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
	fmt.Println("Log Level: ", Logger.Level)
	Logger.SetLevel(util.GetLogLevel())

	cf, err := util.GetConfigFile()
	if err != nil {
		Logger.Fatal("cannot load the config file.")
	}

	if _, err := toml.DecodeFile(cf, &CFG); err != nil {
		Logger.Fatal(err)
		return
	} else {
		Logger.Info(cf, " was loaded.")
		Logger.Info(CFG.Welcome)
	}
	//fmt.Println(CFG)
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
