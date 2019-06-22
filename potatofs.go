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
		cronVolume.AddFunc("*/30 * * * * *", func() { potato.EntityCompaction() })

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
