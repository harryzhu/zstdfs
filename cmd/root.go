/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"os"

	//"strconv"

	"github.com/harryzhu/potatofs/entity"
	"github.com/harryzhu/potatofs/util"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	logger *zap.Logger
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "potatofs",
	Short: "A brief description of your application",
	Long:  ``,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	LoadConfigurationFile()
	fmt.Println(Config.Welcome)
	if Config.MaxSizeMB == 0 {
		Config.MaxSizeMB = 32
	}
	entity.MaxSizeByte = Config.MaxSizeMB << 20
	fmt.Println("max entity size: ", entity.MaxSizeByte)

	z := util.Zapper{
		Level: Config.LogLevel,
		File:  Config.LogFile,
	}
	logger = z.GetLogger().Logger

	logger.Info("logger init successfully")
}
