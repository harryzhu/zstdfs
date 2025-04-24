package cmd

import (
	"strings"

	"github.com/robfig/cron/v3"
)

func StartCron() {
	c := cron.New()

	c.AddFunc("@every 15m", func() {
		DebugInfo("StartCron", "CleanExpires")
		DebugInfo("StartCron", "DiskCacheExpires: ", DiskCacheExpires)
		DebugInfo("StartCron", "TempDir: ", TempDir)
		DebugInfo("StartCron", "UploadDir: ", UploadDir)

		if strings.Index(TempDir, "www/temp") > 0 {
			DebugInfo("StartCron", "cleaning:", TempDir)
			CleanExpires(TempDir, DiskCacheExpires)
		}

		if strings.Index(UploadDir, "www/uploads") > 0 {
			DebugInfo("StartCron", "cleaning:", UploadDir)
			CleanExpires(UploadDir, DiskCacheExpires)
		}

	})
	c.Start()
}
