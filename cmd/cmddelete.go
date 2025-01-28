/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"strings"

	"github.com/spf13/cobra"
)

var (
	DeleteUser     string
	DeleteGroup    string
	DeleteKey      string
	DeleteAuth     string
	DeleteEndpoint string
)

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:              "delete",
	TraverseChildren: true,
	Short:            "delete --endpoint= --user= --group= --key= --auth=",
	Long: `e.g.: 
	./zstdfs delete --endpoint=http://localhost:8080/admin/delete --user=harry --group=web02 --key=bootstrap/v3.5/bs.min.css --auth=admin:123,
	will delete the key[web02/bootstrap/v3.5/bs.min.css] from the bucket[harry].
	`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		PrintPflags()
		if IsAnyEmpty(DeleteUser, DeleteGroup, DeleteKey) {
			FatalError("deleteCmd:--user= --group= --key=", ErrParamEmpty)
		} else {
			DebugInfo("will delete key", "[", DeleteGroup, "/", DeleteKey, "] from bucket [", DeleteUser, "]")
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		deleteURL := strings.Join([]string{DeleteEndpoint, DeleteUser, DeleteGroup, DeleteKey}, "/")
		DebugInfo("DeleteCmd:url", deleteURL)
		formParams := map[string]string{}

		authUserPass := strings.Split(DeleteAuth, ":")
		if authUserPass[0] == DeleteUser || authUserPass[0] == AdminUser {
			clientDo("DELETE", deleteURL, DeleteAuth, "", formParams)
		} else {
			DebugWarn("DeleteCmd", "user cannot delete files of others:", authUserPass[0], " <=> ", DeleteUser)
		}

	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
	deleteCmd.PersistentFlags().StringVar(&DeleteEndpoint, "endpoint", "http://localhost:8080/admin/delete", "server address")
	deleteCmd.PersistentFlags().StringVar(&DeleteAuth, "auth", "", "format: username:password")

	deleteCmd.PersistentFlags().StringVar(&DeleteUser, "user", "", "user")
	deleteCmd.PersistentFlags().StringVar(&DeleteGroup, "group", "", "group")
	deleteCmd.PersistentFlags().StringVar(&DeleteKey, "key", "", "key name")

	deleteCmd.MarkFlagRequired("user")
	deleteCmd.MarkFlagRequired("group")
	deleteCmd.MarkFlagRequired("key")
}
