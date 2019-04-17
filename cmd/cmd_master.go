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
	pbm "hazhufs/pb/master_pb"
	"net"

	//"strconv"
	"strings"

	//"github.com/hashicorp/memberlist"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

// masterCmd represents the master command
var masterCmd = &cobra.Command{
	Use:   "master",
	Short: "master",
	Long:  `master.`,
	Run: func(cmd *cobra.Command, args []string) {
		PrepareMasterDatabase()

		addressMaster := strings.Join([]string{CFG.Master.Ip, CFG.Master.Port}, ":")
		listening, err := net.Listen("tcp", addressMaster)
		if err != nil {
			Logger.Fatalf("Failed to listen: ", err)
		} else {
			Logger.Info("Start as Master Role: ", addressMaster)

		}

		grpcServerMaster := grpc.NewServer(grpc.MaxMsgSize(GRPCMAXMSGSIZE))
		pbm.RegisterMasterServiceServer(grpcServerMaster, &masterService{})
		grpcServerMaster.Serve(listening)
	},
}

func init() {
	rootCmd.AddCommand(masterCmd)

}
