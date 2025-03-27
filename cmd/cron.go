/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"strings"

	"github.com/robfig/cron/v3"
)

func StartCron() {
	c := cron.New()

	c.AddFunc("@every 5m", func() {
		DebugInfo("StartCron", "CleanExpires")
		DebugInfo("StartCron", "DiskCacheExpires: ", DiskCacheExpires)
		DebugInfo("StartCron", "TEMP_DIR: ", TEMP_DIR)
		DebugInfo("StartCron", "UploadDir: ", UploadDir)

		if strings.Index(TEMP_DIR, "www/temp") > 0 {
			DebugInfo("StartCron", "cleaning:", TEMP_DIR)
			CleanExpires(TEMP_DIR, DiskCacheExpires)
		}

		if strings.Index(UploadDir, "www/uploads") > 0 {
			DebugInfo("StartCron", "cleaning:", UploadDir)
			CleanExpires(UploadDir, DiskCacheExpires)
		}

	})
	c.Start()
}
