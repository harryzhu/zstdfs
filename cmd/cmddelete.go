/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var (
	DeleteUser  string
	DeleteGroup string
	DeleteKey   string
)

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:              "delete",
	TraverseChildren: true,
	Short:            "delete --user= --group= --key=",
	Long: `e.g.: 
	./zstdfs delete --user=harry --group=web02 --key=bootstrap/v3.5/bs.min.css,
	will delete the key[web02/bootstrap/v3.5/bs.min.css] from the bucket[harry].
	`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		PrintPflags()
		if IsAnyEmpty(DeleteUser, DeleteGroup, DeleteKey) {
			FatalError("deleteCmd:--user= --group= --key=", ErrParamEmpty)
		} else {
			DebugInfo("will delete key", "[", DeleteGroup, "/", DeleteKey, "] from bucket [", DeleteUser, "]")
		}
		openBolt()
	},
	Run: func(cmd *cobra.Command, args []string) {
		fkey := strings.Join([]string{DeleteGroup, DeleteKey}, "/")
		DebugInfo("DeleteCmd:Key", DeleteUser, fkey)
		err := dbDelete(DeleteUser, fkey)
		if err != nil {
			PrintError("deleteCmd", err)
		} else {
			fmt.Println("OK")
		}
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
	deleteCmd.PersistentFlags().StringVar(&DeleteUser, "user", "", "user")
	deleteCmd.PersistentFlags().StringVar(&DeleteGroup, "group", "", "group")
	deleteCmd.PersistentFlags().StringVar(&DeleteKey, "key", "", "key name")

	deleteCmd.MarkFlagRequired("user")
	deleteCmd.MarkFlagRequired("group")
	deleteCmd.MarkFlagRequired("key")
}
