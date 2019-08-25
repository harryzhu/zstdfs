package potato

import (
	"encoding/json"
	//"errors"
	"io"
	"net"

	//"strings"

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
	return nil, ErrInGeneral
}

func (vs *VolumeService) HandleFile(ctx context.Context, MessageIn *pbv.Message) (*pbv.Message, error) {

	key := MessageIn.Key
	action := MessageIn.Action
	if key == nil || len(action) <= 0 {
		return nil, ErrKeyIsEmpty
	}

	f := &pbv.Message{}
	f.Key = MessageIn.Key
	f.Data = nil
	f.Action = action
	f.Time = TimeNowUnixString()

	switch action {
	case "exists":
		{
			if EntityExists(key) == true {
				f.ErrCode = 200
			} else {
				f.ErrCode = 404
			}
		}
	case "del":
		{
			f.ErrCode = 200

			if EntityDelete(key) != nil {
				f.ErrCode = 400
			}
			cacheDelete(key)
		}
	case "ban":
		{
			f.ErrCode = 200

			if FileBan(key) != nil {
				f.ErrCode = 400
			}
			cacheDelete(key)
		}
	case "pub":
		{
			f.ErrCode = 200

			if FilePub(key) != nil {
				f.ErrCode = 400
			}
			cacheDelete(key)
		}
	case "head":
		{
			f.ErrCode = 404

			data, err := EntityGet(key)
			if err != nil {
				return nil, err
			}
			var fileobj FileObject
			erru := json.Unmarshal(data, &fileobj)
			if erru == nil {
				fileobj_meta := &FileObject{
					Ver:  fileobj.Ver,
					Stat: fileobj.Stat,
					Csec: fileobj.Csec,
					Msec: fileobj.Msec,
					Name: fileobj.Name,
					Size: fileobj.Size,
					Mime: fileobj.Mime,
				}

				byteFileObjectMeta, err := json.Marshal(fileobj_meta)
				if err == nil {
					f.Data = byteFileObjectMeta
					f.ErrCode = 0
				}
			}
		}
	default:
		{
			f.ErrCode = 500
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
