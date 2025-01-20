/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"time"

	"github.com/robfig/cron/v3"
)

func StartCron() {
	c := cron.New()

	c.AddFunc("@every 30m", func() {
		time.Sleep(2 * time.Second)
		DebugInfo("StartCron", "CleanExpires")
		CleanExpires("www/temp", 1800)
		CleanExpires("www/uploads", 1800)
		if UploadDir != "www/uploads" {
			CleanExpires(UploadDir, 1800)
		}
	})
	c.Start()
}
