/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"io/ioutil"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	PutUser  string
	PutGroup string
	PutKey   string
	PutFile  string
)

// putCmd represents the put command
var putCmd = &cobra.Command{
	Use:              "put",
	TraverseChildren: true,
	Short:            "put --user= --group= --key= --file= ;if --key is empty, will use filename as default key",
	Long:             `e.g.: ./zstdfs put --user=harry --group=web02 --key=bootstrap/v3.5/bs.min.css --file=/home/harry/bootstrap/v3.5/bs.min.css`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		PrintPflags()
		if IsAnyEmpty(PutUser, PutGroup, PutFile) {
			FatalError("putCmd: --user= --group= --file= ", ErrParamEmpty)
		}
		openBolt()
	},
	Run: func(cmd *cobra.Command, args []string) {
		data, err := ioutil.ReadFile(PutFile)
		FatalError("cmdPut", err)
		if PutKey == "" {
			PutKey = filepath.Base(PutFile)
		}

		dbSave(PutUser, PutGroup, PutKey, data)

	},
}

func init() {
	rootCmd.AddCommand(putCmd)

	putCmd.PersistentFlags().StringVar(&PutUser, "user", "test", "user")
	putCmd.PersistentFlags().StringVar(&PutGroup, "group", "portal", "group")
	putCmd.PersistentFlags().StringVar(&PutKey, "key", "", "key name")
	putCmd.PersistentFlags().StringVar(&PutFile, "file", "", "file path")
	putCmd.PersistentFlags().IntVar(&MaxUploadSizeMB, "max-upload-size-mb", 16, "max upload size, default: 16mb")

	putCmd.MarkFlagRequired("user")
	putCmd.MarkFlagRequired("group")
	putCmd.MarkFlagRequired("file")

}
