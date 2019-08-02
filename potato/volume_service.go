package potato

import (
	"errors"
	"io"
	"net"
	"strings"

	pbv "./pb/volume"
	"golang.org/x/net/context"

	"google.golang.org/grpc"
)

type VolumeService struct{}

func (vs *VolumeService) HealthCheck(ctx context.Context, MessageIn *pbv.Message) (*pbv.Message, error) {
	return &pbv.Message{Code: 200, Okay: true, Data: []byte("OK")}, nil
}

func (vs *VolumeService) StreamSendFile(stream pbv.VolumeService_StreamSendFileServer) error {
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			Logger.Error("StreamSendFile:", err)
			return err
		}

		key := strings.ToLower(in.Key)
		if len(key) > 0 {
			resp := &pbv.File{Key: ""}
			if EntityExists(key) == true {
				resp.Key = key
			} else {
				err := EntitySet(key, in.Data)
				if err == nil {
					resp.Key = key
					Logger.Debug("SET: ", resp.Key)
				} else {
					return err
				}
			}

			err := stream.Send(resp)
			if err != nil {
				return err
			}

		} else {
			Logger.Error("StreamSendFile: key is empty.")
			resp := &pbv.File{Key: ""}
			if err := stream.Send(resp); err != nil {
				return err
			}
			return errors.New("StreamSendFile: error while sending.")
		}

	}
}

func StartNodeServer() {
	addressVolume := strings.Join([]string{CFG.Volume.Ip, CFG.Volume.Port}, ":")
	listening, err := net.Listen("tcp", addressVolume)
	if err != nil {
		Logger.Fatalf("Failed to listen: ", err)
	} else {
		Logger.Info("Endpoint RPC: ", addressVolume)

	}

	grpcServerVolume := grpc.NewServer(grpc.MaxMsgSize(GRPCMAXMSGSIZE))
	pbv.RegisterVolumeServiceServer(grpcServerVolume, &VolumeService{})

	grpcServerVolume.Serve(listening)
}

func init() {
	// EntitySet("sdfsdfs", []byte("gg"))

	// d, _ := EntityGet("sdfsdfs")
	// Logger.Info("Gettttest:", d)

	// EntityDelete("sdfsdfs")

	// c, _ := EntityGet("sdfsdfs")
	// Logger.Info("Gettttest2:", c)
}
