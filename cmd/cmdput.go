/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"strings"
	//"io/ioutil"
	//"path/filepath"

	"github.com/spf13/cobra"
)

var (
	PutUser     string
	PutGroup    string
	PutPrefix   string
	PutFile     string
	PutAuth     string
	PutEndpoint string
)

// putCmd represents the put command
var putCmd = &cobra.Command{
	Use:              "put",
	TraverseChildren: true,
	Short:            "put --endpoint= --auth= --user= --group= --prefix= --file=",
	Long:             `e.g.: ./zstdfs put --endpoint=http://localhost:8080/admin/upload --auth=admin:123 --user=harry --group=web02 --prefix=bootstrap/v3.5 --file=/home/harry/bootstrap/v3.5/bs.min.css`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		PrintPflags()
		if IsAnyEmpty(PutUser, PutGroup, PutFile) {
			FatalError("putCmd: --user= --group= --file= ", ErrParamEmpty)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		formParams := map[string]string{
			"fuser":   PutUser,
			"fgroup":  PutGroup,
			"fprefix": PutPrefix,
		}
		authUserPass := strings.Split(PutAuth, ":")
		if authUserPass[0] == PutUser || authUserPass[0] == AdminUser {
			clientDo("POST", PutEndpoint, PutAuth, PutFile, formParams)
		} else {
			DebugWarn("putCommand", "user cannot put files into others' bucket: ", authUserPass[0], " <=> ", PutUser)
		}
	},
}

func init() {
	rootCmd.AddCommand(putCmd)
	putCmd.PersistentFlags().StringVar(&PutEndpoint, "endpoint", "http://localhost:8080/admin/upload", "server address")
	putCmd.PersistentFlags().StringVar(&PutAuth, "auth", "admin:123", "username+:+password")
	putCmd.PersistentFlags().StringVar(&PutUser, "user", "test", "user")
	putCmd.PersistentFlags().StringVar(&PutGroup, "group", "portal", "group")
	putCmd.PersistentFlags().StringVar(&PutPrefix, "prefix", "", "prefix before file name")
	putCmd.PersistentFlags().StringVar(&PutFile, "file", "", "file path")
	putCmd.PersistentFlags().IntVar(&MaxUploadSizeMB, "max-upload-size-mb", 16, "max upload size, default: 16mb")

	putCmd.MarkFlagRequired("endpoint")
	putCmd.MarkFlagRequired("user")
	putCmd.MarkFlagRequired("group")
	putCmd.MarkFlagRequired("file")

}
