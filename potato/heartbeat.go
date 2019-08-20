package potato

import (
	"errors"
	"time"

	pbv "./pb/volume"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

func Heartbeat() {
	for _, ip_port := range volumePeers {
		if HealthCheck(ip_port) != nil {
			volumePeersLive[ip_port] = false
		} else {
			volumePeersLive[ip_port] = true
		}
	}
}

func HealthCheck(ip_port string) error {
	conn, err := grpc.Dial(ip_port, grpc.WithInsecure(),
		grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(grpcMAXMSGSIZE)))
	if err != nil {
		return err
	}
	defer conn.Close()

	client := pbv.NewVolumeServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	nodeMessage, err := client.HealthCheck(ctx, &pbv.Message{ErrCode: 0, Action: "ping", Key: []byte("healthcheck"), Data: []byte("Ping")})
	if err != nil {
		return err
	}

	if nodeMessage.ErrCode == 0 {
		return nil
	}

	return errors.New("ERROR")
}
