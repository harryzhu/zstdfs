package potato

import (
	"errors"
	"time"

	pbv "./pb/volume"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

func Heartbeat() {
	for _, ip_port := range VOLUME_PEERS {
		if HealthCheck(ip_port) != nil {
			VOLUME_PEERS_LIVE[ip_port] = false
		} else {
			VOLUME_PEERS_LIVE[ip_port] = true
		}
	}
	//Logger.Debug("VOLUME_PEERS_LIVE: ", VOLUME_PEERS_LIVE)
}

func HealthCheck(ip_port string) error {
	conn, err := grpc.Dial(ip_port, grpc.WithInsecure(),
		grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(GRPCMAXMSGSIZE)))
	if err != nil {
		return err
	}
	defer conn.Close()

	client := pbv.NewVolumeServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	NodeStatus, err := client.HealthCheck(ctx, &pbv.Message{Code: 200, Okay: true, Data: []byte("Ping")})
	if err != nil {
		//Logger.Debug("HealthCheck ERROR: ", err)
		return err
	}

	if NodeStatus.Okay == true {
		return nil
	}

	err = errors.New("ERROR")

	return err
}
