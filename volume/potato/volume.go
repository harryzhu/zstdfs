package potato

import (
	//"encoding/json"
	//"errors"
	"io"
	"net"

	//"strings"

	pbv "./pb/volume"
	"golang.org/x/net/context"

	"google.golang.org/grpc"
)

type VolumeService struct{}

var MessageDefault = &pbv.Message{}

func (vs *VolumeService) HealthCheck(ctx context.Context, MessageIn *pbv.Message) (*pbv.Message, error) {
	return &pbv.Message{}, nil
}

func (vs *VolumeService) GetFile(ctx context.Context, MessageIn *pbv.Message) (*pbv.Message, error) {
	f := &pbv.Message{}
	return f, nil
}

func (vs *VolumeService) SetFile(ctx context.Context, MessageIn *pbv.Message) (*pbv.Message, error) {
	f := &pbv.Message{}
	return f, nil
}

func (vs *VolumeService) DelFile(ctx context.Context, MessageIn *pbv.Message) (*pbv.Message, error) {
	f := &pbv.Message{}
	return f, nil
}

func (vs *VolumeService) BanFile(ctx context.Context, MessageIn *pbv.Message) (*pbv.Message, error) {
	f := &pbv.Message{}
	return f, nil
}

func (vs *VolumeService) StreamSetMessage(stream pbv.VolumeService_StreamSetMessageServer) error {
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			logger.Error("StreamSetMessage:", err)
			return err
		}

		if in.Key == nil || in.Data == nil {
			//logger.Warn("StreamSetMessage: key/data is invalid.")
			continue
		}
		resp := MessageDefault
		resp.Key = in.Key
		resp.ErrCode = 0
		resp.Data = nil

		if EntityExists(in.Key) == false {
			err := EntitySet(in.Key, in.Data)
			if err != nil {
				logger.Error("StreamSetMessage:Error EntitySet:", err)
				resp.ErrCode = 500
			} else {
				//PeersMark("sync", "set", string(key), "1")
			}
		}

		if err := stream.Send(resp); err != nil {
			continue
		}

	}
}

func (vs *VolumeService) StreamGetMessage(stream pbv.VolumeService_StreamGetMessageServer) error {
	return nil
}

func (vs *VolumeService) StreamDelMessage(stream pbv.VolumeService_StreamDelMessageServer) error {
	return nil
}

func (vs *VolumeService) StreamBanMessage(stream pbv.VolumeService_StreamBanMessageServer) error {
	return nil
}

func StartVolumeServer() {
	listening, err := net.Listen("tcp", volumeSelf)
	if err != nil {
		logger.Fatalf("Failed to listen: ", err)
	} else {
		logger.Info("Endpoint RPC: ", volumeSelf)

	}

	grpcServerVolume := grpc.NewServer(grpc.MaxMsgSize(grpcMAXMSGSIZE))
	pbv.RegisterVolumeServiceServer(grpcServerVolume, &VolumeService{})

	grpcServerVolume.Serve(listening)
}
