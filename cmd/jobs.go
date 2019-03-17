package cmd

import (
	pbv "hazhufs/pb/volume_pb"
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
		Logger.Debug("Heartbeat Ping:", addr)
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
				Logger.Warn("Ping Error: ", err)
			} else {
				//Logger.Debug("Ping OK: ", cping.Message)
				VOLUMEPEERSONLINE = append(VOLUMEPEERSONLINE, addr)
			}
		}
	}

	return nil
}
