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
	return &pbv.Message{Key: []byte("healthcheck"), Data: []byte("OK"), Action: "ping", ErrCode: 0}, nil
}

func (vs *VolumeService) ReadFile(ctx context.Context, MessageIn *pbv.Message) (*pbv.Message, error) {
	f := &pbv.Message{}
	if MessageIn.Key != nil {
		data, err := EntityGet(MessageIn.Key)
		if err != nil {
			return nil, err
		}
		f.Key = MessageIn.Key
		f.Data = data
		f.Action = "get"
		f.ErrCode = 0
		return f, nil
	}
	return nil, errors.New("ERROR")
}

func (vs *VolumeService) StreamSendMessage(stream pbv.VolumeService_StreamSendMessageServer) error {
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			logger.Error("StreamSendMessage:", err)
			return err
		}

		key := in.Key
		action := in.Action
		errcode := in.ErrCode
		data := in.Data
		resp := &pbv.Message{}

		logger.Info("StreamSendMessage: key: ", string(key))
		if IsEmpty(key) || IsEmpty(data) || IsEmptyString(action) {
			logger.Error("StreamSendMessage: key/action/errcode/data is invalid.", errcode)
			resp.Key = []byte("")
			resp.Action = "none"
			resp.ErrCode = 400
			resp.Data = []byte("error-empty")
		} else {
			logger.Info("StreamSendMessage: action: ", strings.ToLower(action))
			switch action {
			case "get":
				{
					resp.Key = key
					resp.Action = "none"
					resp.ErrCode = 404
					resp.Data = nil

					if EntityExists(key) == true {
						data, err := EntityGet(key)
						if err == nil {
							resp.Key = key
							resp.Action = "none"
							resp.ErrCode = 0
							resp.Data = data
						}
					}
				}
			case "set":
				{
					resp.Key = key
					resp.Action = "none"
					resp.ErrCode = 0
					resp.Data = nil

					if EntityExists(key) == false {
						err := EntitySet(key, data)
						if err != nil {
							logger.Error("vs:StreamSendMessage:Error EntitySet:", err)
							resp.Key = key
							resp.Action = "none"
							resp.ErrCode = 400
							resp.Data = []byte("error-set")
						} else {
							PeersMark("sync", "set", string(key), "1")
						}
					}

				}
			case "del":
				{
					resp.Key = key
					resp.Action = "none"
					resp.ErrCode = 0
					resp.Data = nil

					if EntityExists(key) == true {
						err := EntityDelete(key)
						if err != nil {
							resp.Key = key
							resp.Action = "none"
							resp.ErrCode = 400
							resp.Data = []byte("error-del")
						} else {
							PeersMark("sync", "del", string(key), "1")
						}
					}
				}

			}

		}

		if err := stream.Send(resp); err != nil {
			continue
		}

	}
}

func StartNodeServer() {
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
