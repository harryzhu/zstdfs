/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var (
	IsDebug            bool
	IsIgnoreError      bool
	IsDatabaseReadOnly bool
	DataDir            string
	MaxUploadSizeMB    int

	Pflags map[string]map[string]any = make(map[string]map[string]any)
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:              "zstdfs",
	Short:            "--debug=true|false",
	TraverseChildren: true,
	Long: `1)zstdfs httpd [flags] is a http server for uploading and accessing files.
all files will be compressed by zstd algorithm, zstd is effective for ascii files,
e.g.: .css, .js, .txt, .md, .html, .py ...
for image/video/binary-files, you could put them in a static folder(via subcommand httpd flag
--static-folder=/path/to/your/folder), then url is: http://host:port/static/your-file-path.
2)zstdfs import [flags]: batch import specified files into zstdfs from a folder.
3)zstdfs export [flags]: export all files from zstdfs into a local folder.
4)zstdfs put [flags]: put a single file into zstdfs.
5)zstdfs delete [flags]: delete a single file from zstdfs.
6)readonly mode: you can use --readonly to protect the database from any update.
-----
if you want to run subcommand [export/import/put/delete], please stop subcommand [httpd] server first.
-----
1)url: http://host:port/z/{bucket}/{group}/... is for files stored in database;
2)url: http://host:port/static/... is for non-ascii files stored in the disk;
-----
if you want to run subcommand [export/import/put/delete], please stop subcommand [httpd] server first.
`,

	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		DebugInfo("zstdfs", "Thanks for choosing zstdfs!")
		PrintPflags()
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
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&IsDebug, "debug", false, "if print debug info")
	rootCmd.PersistentFlags().IntVar(&MaxUploadSizeMB, "max-upload-size-mb", 16, "max upload size, default: 16mb")
	rootCmd.PersistentFlags().BoolVar(&IsIgnoreError, "ignore-error", false, "if stop when error")
	rootCmd.PersistentFlags().BoolVar(&IsDatabaseReadOnly, "readonly", false, "if set database in ReadOnly mode")
}
