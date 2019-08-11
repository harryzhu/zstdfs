package potato

import (
	//"context"
	"errors"
	//"io"
	//"strings"
	"time"

	pbv "./pb/volume"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

func EntityGetRoundRobin(key []byte) ([]byte, error) {
	logger.Debug("EntityGetRoundRobin:", string(key))
	var data []byte
	var err_return error = errors.New("not exist")

	for ip_port, is_live := range volumePeersLive {
		if err_return == nil {
			break
		}

		if is_live == false {
			continue
		}
		logger.Debug("EntityGetRoundRobin:ip_port:", ip_port)
		conn, err := grpc.Dial(ip_port, grpc.WithInsecure(),
			grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(grpcMAXMSGSIZE), grpc.MaxCallRecvMsgSize(grpcMAXMSGSIZE)))
		if err != nil {
			continue
		}
		defer conn.Close()

		client := pbv.NewVolumeServiceClient(conn)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		fileRead, err := client.ReadFile(ctx, &pbv.Message{Key: key})
		if err != nil {
			continue
		}
		data = fileRead.Data
		err_return = nil
		break

	}
	return data, err_return
}
