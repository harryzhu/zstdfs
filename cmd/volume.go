// Copyright © 2019 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	pbv "hazhufs/pb/volume_pb"
	"net"
	"strings"

	"github.com/robfig/cron"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

// volumeCmd represents the volume command
var volumeCmd = &cobra.Command{
	Use:   "volume",
	Short: "volume",
	Long:  `volume`,
	Run: func(cmd *cobra.Command, args []string) {
		//fmt.Println("fff")
		PrepareVolumeDatabases()

		go func() {
			cronVolume := cron.New()
			cronVolume.AddFunc("*/3 * * * * *", func() { runSync() })
			cronVolume.Start()
		}()

		addressVolume := strings.Join([]string{CFG.Volume.Ip, CFG.Volume.Port}, ":")
		listening, err := net.Listen("tcp", addressVolume)
		if err != nil {
			Logger.Fatalf("Failed to listen: ", err)
		} else {
			Logger.Info("Start as Volume Role: ", addressVolume)

		}

		grpcServerVolume := grpc.NewServer(grpc.MaxMsgSize(GRPCMAXMSGSIZE))
		pbv.RegisterVolumeServiceServer(grpcServerVolume, &volumeService{})
		grpcServerVolume.Serve(listening)
	},
}

func init() {
	rootCmd.AddCommand(volumeCmd)

	//
	Logger.Info("volume LOGLEVEL: ", Logger.Level)
	//fmt.Println("f22ff")
}
