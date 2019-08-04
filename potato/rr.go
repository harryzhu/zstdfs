package potato

// import (
// 	//"context"
// 	"errors"
// 	//"io"
// 	"strings"
// 	"time"

// 	pbv "./pb/volume"
// 	"golang.org/x/net/context"
// 	"google.golang.org/grpc"
// )

// func EntityGetRoundRobin(key string) ([]byte, error) {
// 	Logger.Info("EntityGetRoundRobin:", key)
// 	var data []byte
// 	var err_return error = errors.New("not exist")

// 	for ip_port, is_live := range VOLUME_PEERS_LIVE {
// 		if err_return == nil {
// 			break
// 		}

// 		if is_live == false {
// 			continue
// 		}
// 		logger.Info("EntityGetRoundRobin:ip_port:", ip_port)
// 		conn, err := grpc.Dial(ip_port, grpc.WithInsecure(),
// 			grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(GRPCMAXMSGSIZE), grpc.MaxCallRecvMsgSize(GRPCMAXMSGSIZE)))
// 		if err != nil {
// 			continue
// 		}
// 		defer conn.Close()

// 		client := pbv.NewVolumeServiceClient(conn)
// 		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
// 		defer cancel()

// 		fileRead, err := client.ReadFile(ctx, &pbv.File{Key: key})
// 		if err != nil {
// 			continue
// 		}
// 		data = fileRead.Data
// 		err_return = nil
// 		break

// 	}
// 	return data, err_return
// }
