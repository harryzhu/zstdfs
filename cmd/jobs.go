package cmd

import (
	pbm "hazhufs/pb/master_pb"
	pbv "hazhufs/pb/volume_pb"
	//"hazhufs/util"
	"io"
	"strconv"
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
	node_self := strings.Join([]string{CFG.Volume.Ip, CFG.Volume.Port}, ":")
	fileRequests := []*pbm.File{}

	var k, n string
	var s, c uint32
	stmt, err := DBMETA.Prepare("SELECT key,node,size,created FROM data WHERE synced = ? and inmaster = ? and node = ? LIMIT 1000")
	if err != nil {
		Logger.Error(err)
		return
	}

	rows, err := stmt.Query(0, 0, node_self)

	if err != nil {
		Logger.Error(err)
		return
	}

	for rows.Next() {
		err = rows.Scan(&k, &n, &s, &c)
		if err != nil {
			Logger.Error(err)
			continue
		}

		Logger.Info("--", k, n, s, c)

		fileRequests = append(fileRequests, &pbm.File{
			Key:      k,
			Node:     n,
			Size:     s,
			Synced:   0,
			InMaster: 1,
			Created:  c})
	}

	int_length_fileRequests := len(fileRequests)
	Logger.Debug("will sync files to master totally(streaming): ", len(fileRequests))

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
			if in.ErrorCode == 0 && len(in.Key) == 32 {
				VolumeUpdateMetaData(in.Key, in.Node, "inmaster", strconv.Itoa(int(in.InMaster)))
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
