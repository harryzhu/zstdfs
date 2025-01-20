/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"runtime"
	"time"

	"github.com/robfig/cron/v3"
)

func StartCron() {
	c := cron.New()

	c.AddFunc("@every 5m", func() {
		time.Sleep(time.Second)
		DebugInfo("StartCron", "CleanExpires")
		CleanExpires("www/temp", DiskCacheExpires)
		CleanExpires("www/uploads", DiskCacheExpires)
		if UploadDir != "www/uploads" {
			CleanExpires(UploadDir, DiskCacheExpires)
		}
		runtime.GC()
	})
	c.Start()
}
