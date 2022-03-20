/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	//"time"
	"strings"

	"github.com/harryzhu/potatofs/entity"

	//"github.com/harryzhu/potatofs/util"
	"github.com/spf13/cobra"
)

var (
	WithFullResync bool = false
)

var volumeCmd = &cobra.Command{
	Use:   "volume",
	Short: "A brief description of your command",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("volume started")

		entity.OpenMetaDB(Config.MetaDir)
		entity.OpenDataDB(Config.DataDir)
		entity.OpenBoltDB()

		wg.Add(4)

		go func() {
			StartVolumeServer()
		}()

		if Config.IsMaster == true && Config.SyncSlave != "" {
			if strings.Join([]string{Config.IP, Config.VolumePort}, ":") != Config.SyncSlave {
				go func() {
					for {
						StartSync()
					}

				}()
			} else {
				fmt.Println("master's ip:port should be different from slave's ip:port")
			}

		}

		go func() {
			StartHttpServer()
		}()

		go func() {
			if WithFullResync == true {
				logger.Info("Start Full Resyncing...")
				entity.WalkMetaKeyListToBoltMetaBucket()
			}
			wg.Done()
		}()

		wg.Wait()
	},
}

func init() {
	rootCmd.AddCommand(volumeCmd)
	volumeCmd.Flags().BoolVar(&WithFullResync, "with-full-resync", false, "--with-full-resync=true|false")

}
