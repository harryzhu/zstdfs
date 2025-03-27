/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"zstdfs/cmd"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	cmd.Execute()

	wg := sync.WaitGroup{}
	wg.Add(3)

	go func() {
		cmd.StartHTTPServer()
	}()

	go func() {
		cmd.StartCron()
	}()

	go func() {
		onExit()
	}()

	wg.Wait()

}

func onExit() {
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		cmd.StopHTTPServer()
		os.Exit(0)
	}()
}
