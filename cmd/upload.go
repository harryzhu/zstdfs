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
	"fmt"
	pbv "hazhufs/pb/volume_pb"
	"hazhufs/util"
	"io"
	"strings"

	//"net"
	//"bytes"
	"io/ioutil"
	"mime"
	"path"
	"strconv"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/spf13/cobra"
)

type uploadConfig struct {
	volume string
	udir   string
	filter string
}

var uc = &uploadConfig{}

// uploadCmd represents the upload command
var uploadCmd = &cobra.Command{
	Use:   "upload",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {

		addr := uc.volume
		conn, err := grpc.Dial(addr, grpc.WithInsecure(),
			grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(GRPCMAXMSGSIZE)))
		defer conn.Close()
		if err != nil {
			Logger.Fatal(fmt.Sprintf("Failed to connect: %v", err))
		} else {
			Logger.Info(fmt.Sprintf("start as upload role. ip: %v, port: %v, udir: %v, filter: %v \n",
				uc.volume, uc.udir, uc.filter))
		}

		//uploadLogger.Debug("Function Elapsed: upload: ", TimerStop())
		c := pbv.NewVolumeServiceClient(conn)
		t_start := time.Now()
		runStreamSendFile(c)
		Logger.Info("Function Elapsed: upload: ", time.Since(t_start))

	},
}

func init() {
	rootCmd.AddCommand(uploadCmd)

	uploadCmd.Flags().StringVarP(&uc.volume, "volume", "", "172.16.32.45:9537", "volume IP address.")
	uploadCmd.Flags().StringVarP(&uc.udir, "udir", "", "/home/ops/data/images/", "images data dir.")
	uploadCmd.Flags().StringVarP(&uc.filter, "filter", "", "*.jpg", "Filter.")

}

func runStreamSendFile(client pbv.VolumeServiceClient) {

	fileRequests := []*pbv.File{}

	files := util.DirWalker(uc.udir, uc.filter)

	Logger.Info("start listing files in dir: ", uc.udir)
	for _, v := range files {
		//fmd5 := util.FileMD5(v)
		fileData, _ := ioutil.ReadFile(v)
		fsize := strconv.Itoa(len([]byte(fileData)))
		fmime := mime.TypeByExtension(path.Ext(v))
		fmeta := strings.Join([]string{fmime, fsize}, ";")
		Logger.Info("fmeta:", fmeta)

		if fileData != nil && fmime != "" {

			fileRequests = append(fileRequests, &pbv.File{Key: util.ByteMD5(fileData), Meta: []byte(fmeta), Data: fileData})
		}
	}
	Logger.Info("will upload files totally(streaming): ", len(fileRequests))

	ctx, cancel := context.WithTimeout(context.Background(), 24*3600*time.Second)
	defer cancel()
	client.TransactionStart(ctx, &pbv.Empty{})
	stream, err := client.StreamSendFile(ctx)
	if err != nil {
		Logger.Warn("%v.StreamSendFile(_) = _, %v", client, err)
	}
	waitc := make(chan struct{})
	go func() {
		for {
			in, err := stream.Recv()
			if err == io.EOF {
				// read done.
				close(waitc)
				return
			}
			if err != nil {
				Logger.Warn("Failed to receive a filerequest : %v", err)
			}
			Logger.Info("Got message response key: ", in.Key)
		}
	}()
	for _, filerequest := range fileRequests {
		if err := stream.Send(filerequest); err != nil {
			Logger.Warn("Failed to send a filerequest: ", err)
		}
	}
	stream.CloseSend()
	<-waitc

	client.TransactionEnd(ctx, &pbv.Empty{})
}
