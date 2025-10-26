/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
)

var (
	Host      string
	Port      string
	rpcServer string

	SiteURL       string
	UploadDir     string
	StaticDir     string
	ThumbDir      string
	MinTopCaption int

	BulkLoadDir  string
	BulkLoadExt  string
	BulkLoadUser string
)

// httpCmd represents the http command
var httpCmd = &cobra.Command{
	Use:   "http",
	Short: "http server",
	Long:  ``,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		BeforeHTTPStart()
		if BulkLoadDir != "" && BulkLoadExt != "" && BulkLoadUser != "" {
			badgerBulkLoad(BulkLoadDir, BulkLoadExt)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {

		wg.Add(8)

		go func() {
			StartHTTPServer()
		}()

		go func() {
			TaskRunShellCommandInChannel()
		}()

		go func() {
			TaskRunShellCommandInChannel()
		}()

		go func() {
			TaskRunShellCommandInChannel()
		}()

		go func() {
			TaskRunShellCommandInChannel()
		}()

		go func() {
			TaskDeleteFilesInFilesToBeRemoved()
		}()

		go func() {
			StartCron()
		}()

		go func() {
			onExit()
		}()
	},
}

func init() {
	rootCmd.AddCommand(httpCmd)
	httpCmd.PersistentFlags().StringVar(&Host, "host", "0.0.0.0", "host, default: 0.0.0.0")
	httpCmd.PersistentFlags().StringVar(&Port, "port", "9090", "port, default: 9090")

	httpCmd.PersistentFlags().StringVar(&rpcServer, "rpc-server", "127.0.0.1:8282", "rpcServer address, ip:port")
	httpCmd.PersistentFlags().Int64Var(&MaxUploadSizeMB, "max-upload-size-mb", 16, "max upload size, default: 16")
	httpCmd.PersistentFlags().IntVar(&MaxCacheSizeMB, "max-cache-size-mb", 256, "max size for memory cache")
	httpCmd.PersistentFlags().IntVar(&MinTopCaption, "min-top-caption", 10, "min count of top caption, will show on the page")

	httpCmd.PersistentFlags().StringVar(&UploadDir, "upload-dir", "", "Upload Dir")
	httpCmd.PersistentFlags().StringVar(&StaticDir, "static-dir", "", "Static Dir")
	httpCmd.PersistentFlags().StringVar(&ThumbDir, "thumb-dir", "", "Thumbnail Dir")
	httpCmd.PersistentFlags().StringVar(&SiteURL, "site-url", "", "site url")

	httpCmd.PersistentFlags().StringVar(&BulkLoadDir, "bulk-load-dir", "", "BulkLoad dir path")
	httpCmd.PersistentFlags().StringVar(&BulkLoadExt, "bulk-load-ext", "", "BulkLoad file type: extension, i.e.: .mp4")
	httpCmd.PersistentFlags().StringVar(&BulkLoadUser, "bulk-load-user", "", "BulkLoad username")
}

func onExit() {
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		if IsHTTPServerOnline {
			StopHTTPServer()
		}

		time.Sleep(time.Second * 1)
		os.Exit(0)
	}()
}
