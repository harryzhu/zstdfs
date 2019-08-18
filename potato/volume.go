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
	return &pbv.Message{Key: []byte("healthcheck"), Data: []byte("OK"), Action: "ping", ErrCode: 0, Time: TimeNowUnixString()}, nil
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
		f.Time = TimeNowUnixString()
		return f, nil
	}
	return nil, errors.New("ERROR")
}

func (vs *VolumeService) HeadFile(ctx context.Context, MessageIn *pbv.Message) (*pbv.Message, error) {

	key := MessageIn.Key
	action := MessageIn.Action
	if key == nil || len(action) <= 0 {
		return nil, errors.New("ERROR: key and action cannot be empty.")
	}

	f := &pbv.Message{}
	f.Key = MessageIn.Key
	f.Action = action
	f.Time = TimeNowUnixString()

	switch action {
	case "exists":
		{
			f.Data = nil
			if EntityExists(key) == true {
				f.ErrCode = 200
			} else {
				f.ErrCode = 404
			}
		}
	case "meta":
		{
			_, err := EntityGet(key)
			if err != nil {
				return nil, err
			}
			f.Data = nil
			f.ErrCode = 0
		}
	default:
		{
			f.ErrCode = 500
			f.Data = nil
		}
	}

	return f, nil
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
		timenano := in.Time
		data := in.Data
		resp := &pbv.Message{Key: key, Action: action, ErrCode: 404, Data: nil, Time: timenano}

		if IsEmpty(key) || IsEmptyString(action) || IsEmptyString(timenano) {
			logger.Warn("StreamSendMessage: key/action/timenano/errcode is invalid.", errcode)
			continue
		}

		logger.Debug("StreamSendMessage: action: ", strings.ToLower(action))

		switch action {
		case "get":
			{
				if EntityExists(key) == true {
					data, err := EntityGet(key)
					if err == nil {
						resp.ErrCode = 0
						resp.Data = data
					}
				}
			}
		case "set":
			{
				resp.ErrCode = 0

				if EntityExists(key) == false {
					err := EntitySet(key, data)
					if err != nil {
						logger.Error("vs:StreamSendMessage:Error EntitySet:", err)
						resp.ErrCode = 400
					} else {
						PeersMark("sync", "set", string(key), "1")
					}
				}

			}
		case "del":
			{
				resp.ErrCode = 0

				if EntityExists(key) == true {
					err := EntityDelete(key)
					if err != nil {
						resp.ErrCode = 400
					} else {
						PeersMark("sync", "del", string(key), "1")
					}
				}
			}
		case "ban":
			{
				resp.ErrCode = 0

				err := FileBan(key)
				if err != nil {
					resp.ErrCode = 400
				}
			}
		case "pub":
			{
				resp.ErrCode = 0

				err := FilePub(key)
				if err != nil {
					resp.ErrCode = 400
				}
			}
		default:
			{
				continue
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
