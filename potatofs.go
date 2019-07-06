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
	wg.Add(3)
	go func() {
		cronVolume := cron.New()
		cronVolume.AddFunc("*/300 * * * * *", func() { potato.EntityCompaction() })
		cronVolume.AddFunc("*/300 * * * * *", func() { potato.RunReplicateParallel() })
		cronVolume.Start()
	}()

	go func() {
		potato.StartNodeServer()
	}()

	go func() {
		potato.StartHttpServer()
	}()

	potato.Echo()

	wg.Wait()

}
