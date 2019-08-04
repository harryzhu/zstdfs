package main

import (
	//"net/http"
	//_ "net/http/pprof"

	"sync"

	"./potato"
	"github.com/robfig/cron"
)

func main() {
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		cronVolume := cron.New()
		//cronVolume.AddFunc("* */8 * * * *", func() { potato.BdbCompaction() })
		cronVolume.AddFunc("*/5 * * * * *", func() { potato.Heartbeat() })
		//cronVolume.AddFunc("*/3 * * * * *", func() { potato.RunReplicateParallel() })
		cronVolume.Start()
	}()

	go func() {
		potato.StartNodeServer()
	}()

	// go func() {
	// 	potato.StartHttpServer()
	// }()

	potato.OnReady()

	wg.Wait()

}
