/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/spf13/cobra"
)

var (
	ImportUser            string
	ImportGroup           string
	ImportExt             string
	ImportDir             string
	ImportIsIgnoreDotFile bool
)

// importCmd represents the import command
var importCmd = &cobra.Command{
	Use:              "import",
	TraverseChildren: true,
	Short:            "import --user= --group= --ext=.css|.js|.txt|* --dir=/path/to/folder/you/want/to/import",
	Long:             `e.g.: ./zstdfs import --user=harry --group=web01 --ext=.css --dir=/home/harry/www/static/web01`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		PrintPflags()
		if IsAnyEmpty(ImportUser, ImportGroup, ImportDir, ImportExt) {
			FatalError("importCmd: --user= --group= --dir= --ext=", ErrParamEmpty)
		}
		openBolt()
	},
	Run: func(cmd *cobra.Command, args []string) {
		ImportFiles(ImportDir, ImportExt, ImportUser, ImportGroup)
	},
}

func init() {
	rootCmd.AddCommand(importCmd)

	importCmd.PersistentFlags().StringVar(&ImportUser, "user", "", "user")
	importCmd.PersistentFlags().StringVar(&ImportGroup, "group", "", "group")
	importCmd.PersistentFlags().StringVar(&ImportDir, "dir", "", "dir")
	importCmd.PersistentFlags().StringVar(&ImportExt, "ext", ".css", "file type, i.e: .css or .js, * means all except dot-file")
	importCmd.PersistentFlags().BoolVar(&ImportIsIgnoreDotFile, "ignore-dot-file", true, "if ignore dot file, i.e: .DS_Store")
	importCmd.PersistentFlags().IntVar(&MaxUploadSizeMB, "max-upload-size-mb", 16, "max upload size, default: 16mb")

	importCmd.MarkFlagsRequiredTogether("user", "group", "dir", "ext")
}
