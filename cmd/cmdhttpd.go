/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"sync"

	"github.com/spf13/cobra"
)

var (
	Host          string
	Port          int
	UploadDir     string
	StaticDir     string
	AdminUser     string
	AdminPassword string
)

// httpdCmd represents the httpd command
var httpdCmd = &cobra.Command{
	Use:              "httpd",
	Short:            "httpd --host= --port= --upload-dir -- static-dir= --admin-user= --admin-password=",
	Long:             ``,
	TraverseChildren: true,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		PrintPflags()
		DebugInfo("IsDatabaseReadOnly", IsDatabaseReadOnly)
		DebugInfo("Param: upload-dir", UploadDir)
		DebugInfo("Param: static-dir", StaticDir)
		openBolt()

	},
	Run: func(cmd *cobra.Command, args []string) {
		wg := sync.WaitGroup{}
		wg.Add(2)
		go func() {
			StartHTTPServer()
		}()

		go func() {
			StartCron()
		}()
		wg.Wait()
	},
}

func init() {
	rootCmd.AddCommand(httpdCmd)
	httpdCmd.PersistentFlags().StringVar(&Host, "host", "0.0.0.0", "host")
	httpdCmd.PersistentFlags().IntVar(&Port, "port", 8080, "port")
	httpdCmd.PersistentFlags().StringVar(&UploadDir, "upload-dir", "www/uploads", "temp dir for uploads")
	httpdCmd.PersistentFlags().StringVar(&StaticDir, "static-dir", "www/static", "static dir for no zstd files")
	httpdCmd.PersistentFlags().StringVar(&AdminUser, "admin-user", "", "for /admin/*")
	httpdCmd.PersistentFlags().StringVar(&AdminPassword, "admin-password", "", "for /admin/*")
	httpdCmd.PersistentFlags().IntVar(&MaxUploadSizeMB, "max-upload-size-mb", 16, "max upload size, default: 16mb")

	httpdCmd.MarkFlagsRequiredTogether("host", "port", "upload-dir")

}
