/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var (
	IsDebug         bool
	MaxUploadSizeMB int64
	Host            string
	Port            string
	SiteURL         string
	UploadDir       string
	StaticDir       string
	BulkLoadDir     string
	BulkLoadExt     string
	BulkLoadUser    string
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

	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
	//
	BeforeStart()
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&IsDebug, "debug", false, "if print debug info")
	rootCmd.PersistentFlags().Int64Var(&MaxUploadSizeMB, "max-upload-size-mb", 16, "max upload size, default: 16mb")
	rootCmd.PersistentFlags().StringVar(&Host, "host", "0.0.0.0", "host, default: 0.0.0.0")
	rootCmd.PersistentFlags().StringVar(&Port, "port", "9090", "port, default: 9090")
	rootCmd.PersistentFlags().StringVar(&UploadDir, "upload-dir", "", "Upload Dir")
	rootCmd.PersistentFlags().StringVar(&StaticDir, "static-dir", "", "Static Dir")
	rootCmd.PersistentFlags().StringVar(&SiteURL, "site-url", "", "site url")
	rootCmd.PersistentFlags().StringVar(&BulkLoadDir, "bulk-load-dir", "", "BulkLoad dir path")
	rootCmd.PersistentFlags().StringVar(&BulkLoadExt, "bulk-load-ext", "", "BulkLoad file type: extension, i.e.: .mp4")
	rootCmd.PersistentFlags().StringVar(&BulkLoadUser, "bulk-load-user", "", "BulkLoad username")
}
