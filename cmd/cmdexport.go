/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/spf13/cobra"
)

var (
	ExportDir string
)

// exportCmd represents the export command
var exportCmd = &cobra.Command{
	Use:              "export",
	TraverseChildren: true,
	Short:            "export --dir=/path/to/folder/you/want/to/export --ignore-error=true|false",
	Long: `stop the httpd server first, then run export.
	e.g.: ./zstdfs export --dir=/home/harry/test/export
	`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		PrintPflags()
		if IsAnyEmpty(ExportDir) {
			FatalError("exportCmd: --dir= ", ErrParamEmpty)
		}
		openBolt()
	},
	Run: func(cmd *cobra.Command, args []string) {
		exportFiles(ExportDir)
	},
}

func init() {
	rootCmd.AddCommand(exportCmd)

	exportCmd.PersistentFlags().StringVar(&ExportDir, "dir", "www/export", "export root dir")

	exportCmd.MarkFlagRequired("dir")
}
