/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	//"io/ioutil"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	PutUser     string
	PutGroup    string
	PutKey      string
	PutFile     string
	PutAuth     string
	PutEndpoint string
)

// putCmd represents the put command
var putCmd = &cobra.Command{
	Use:              "put",
	TraverseChildren: true,
	Short:            "put --endpoint= --auth=admin:123 --user= --group= --key= --file= ;if --key is empty, will use filename as default key",
	Long:             `e.g.: ./zstdfs put --user=harry --group=web02 --key=bootstrap/v3.5/bs.min.css --file=/home/harry/bootstrap/v3.5/bs.min.css`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		PrintPflags()
		if IsAnyEmpty(PutUser, PutGroup, PutFile) {
			FatalError("putCmd: --user= --group= --file= ", ErrParamEmpty)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		if PutKey == "" {
			PutKey = filepath.Base(PutFile)
		}

		clientPostFile(PutUser, PutGroup, PutKey, PutFile, PutEndpoint, PutAuth)
	},
}

func init() {
	rootCmd.AddCommand(putCmd)
	putCmd.PersistentFlags().StringVar(&PutEndpoint, "endpoint", "http://localhost:8080/admin/upload", "server address")
	putCmd.PersistentFlags().StringVar(&PutAuth, "auth", "admin:123", "username+:+password")
	putCmd.PersistentFlags().StringVar(&PutUser, "user", "test", "user")
	putCmd.PersistentFlags().StringVar(&PutGroup, "group", "portal", "group")
	putCmd.PersistentFlags().StringVar(&PutKey, "key", "", "key name")
	putCmd.PersistentFlags().StringVar(&PutFile, "file", "", "file path")
	putCmd.PersistentFlags().IntVar(&MaxUploadSizeMB, "max-upload-size-mb", 16, "max upload size, default: 16mb")

	putCmd.MarkFlagRequired("endpoint")
	putCmd.MarkFlagRequired("user")
	putCmd.MarkFlagRequired("group")
	putCmd.MarkFlagRequired("file")

}
