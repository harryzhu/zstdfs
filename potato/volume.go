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
		resp := &pbv.Message{Key: []byte(""), Data: []byte("error"), Action: "none", ErrCode: 400}

		if IsEmpty(key) || IsEmpty(data) || IsEmptyString(action) {
			logger.Error("StreamSendMessage: key/action/errcode/data is invalid.", errcode)
			resp = &pbv.Message{Key: []byte(""), Action: "none", ErrCode: 400, Data: []byte("error")}
		} else {
			switch strings.ToLower(action) {
			case "get":
				{
					if EntityExists(key) == true {
						data, err := EntityGet(key)
						if err != nil {
							resp = &pbv.Message{Key: key, Action: "none", ErrCode: 404, Data: []byte("error-get")}
						}
						resp = &pbv.Message{Key: key, Action: "none", ErrCode: 0, Data: data}
					}
				}
			case "set":
				{
					if EntityExists(key) == false {
						err := EntitySet(key, data)
						if err != nil {
							resp = &pbv.Message{Key: key, Action: "none", ErrCode: 400, Data: []byte("error-set")}
						} else {
							PeersMark("sync", "set", string(key))
							resp = &pbv.Message{Key: key, Action: "none", ErrCode: 0, Data: nil}
						}
					}
				}
			case "del":
				{
					if EntityExists(key) == true {
						err := EntityDelete(key)
						if err != nil {
							resp = &pbv.Message{Key: key, Action: "none", ErrCode: 400, Data: []byte("error-del")}
						} else {
							PeersMark("sync", "del", string(key))
							resp = &pbv.Message{Key: key, Action: "none", ErrCode: 0, Data: nil}
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
