package cmd

import (
	pbm "hazhufs/pb/master_pb"
	pbv "hazhufs/pb/volume_pb"
	"hazhufs/util"
	"io"
	"strings"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

func Heartbeat() error {
	cfgPeers := CFG.Volume.Peers
	Logger.Debug("cfgPeers: ", cfgPeers)
	arr_addresses := strings.Split(cfgPeers, ";")
	for _, addr := range arr_addresses {
		//Logger.Debug("Heartbeat Ping:", addr)
		conn, err := grpc.Dial(addr, grpc.WithInsecure())
		defer conn.Close()
		if err != nil {
			Logger.Warn("Ping Connection Error: ", err)
		} else {
			client := pbv.NewVolumeServiceClient(conn)
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			_, err := client.Ping(ctx, &pbv.Info{Action: "heartbeat"})

			if err != nil {
				//Logger.Warn("Ping Error: ", err)
			} else {
				//Logger.Debug("Ping OK: ", cping.Message)
				VOLUMEPEERSONLINE = append(VOLUMEPEERSONLINE, addr)
			}
		}
	}

	return nil
}

func Job_VolumeMasterStreamSendFile(client pbm.MasterServiceClient) {

	fileRequests := []*pbm.File{}

	for _, d := range []string{"1", "2", "3"} {
		key := util.TextMD5(d)
		fileRequests = append(fileRequests, &pbm.File{Key: key, Node: key, Size: 23, Created: 23232})

	}
	int_length_fileRequests := len(fileRequests)
	Logger.Info("will sync files to master totally(streaming): ", len(fileRequests))

	ctx, cancel := context.WithTimeout(context.Background(), 24*3600*time.Second)
	defer cancel()

	stream, err := client.StreamSendFile(ctx)
	if err != nil {
		Logger.Warn("%v.StreamSendFile(_) = _, %v", client, err)
		return
	}
	waitc := make(chan struct{})
	int_response := 0
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
			int_response++
			Logger.Info("Got message response key: ", int_response, "/", int_length_fileRequests, ": ", in.Key)
		}
	}()
	for _, filerequest := range fileRequests {
		if err := stream.Send(filerequest); err != nil {
			Logger.Warn("Failed to send a filerequest: ", err)
		}
	}
	stream.CloseSend()
	<-waitc

}
