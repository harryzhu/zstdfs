/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"runtime"
	"zstdfs/cmd"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	cmd.Execute()

}
