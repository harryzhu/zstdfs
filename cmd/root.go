/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"
	"sync"

	"github.com/spf13/cobra"
)

var (
	IsDebug            bool
	IsGrpcServerOnline bool
	IsHTTPServerOnline bool
	MaxUploadSizeMB    int64
	MaxCacheSizeMB     int
	wg                 sync.WaitGroup
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "zstdfs",
	Short: "",
	Long:  ``,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		DebugInfo("zstdfs", "Thanks for choosing zstdfs!")

	},
	Run: func(cmd *cobra.Command, args []string) {

	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {

		wg.Wait()
	},
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
	rootCmd.PersistentFlags().BoolVar(&IsDebug, "debug", true, "if print debug info")
	rootCmd.PersistentFlags().Int64Var(&MaxUploadSizeMB, "max-upload-size-mb", 16, "Max Upload Size(MB), default: 16")

	wg = sync.WaitGroup{}
}
