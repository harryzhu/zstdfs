package cmd

import (
	"errors"
	pbv "hazhufs/pb/volume_pb"
	"io"
	"strings"

	"golang.org/x/net/context"
)

type volumeService struct {
}

func (vs *volumeService) Ping(ctx context.Context, PingIn *pbv.Info) (*pbv.Info, error) {
	Logger.Debug("Ping Test...")
	return &pbv.Info{Action: PingIn.Action, ErrorCode: "200", ErrorMessage: "OK"}, nil
}

func (vs *volumeService) ReadFile(ctx context.Context, FileIn *pbv.File) (*pbv.File, error) {
	Logger.Info("ReadFile, Key: ", FileIn.Key)
	if len(FileIn.Key) != 32 {
		Logger.Error("ReadFile, Key length is not 32. ")
		return nil, errors.New("Key length is invalid")
	}
	k := strings.ToLower(FileIn.Key)
	did := k[0:1]
	file := GetFile(DBDATA[did], k)
	//file := &pbv.File{Key: k, Meta: []byte("fff"), Data: []byte("fff")}
	if file == nil {
		return nil, errors.New("cannot get the file")
	}
	Logger.Info("GetFile, Key: ", file.Key)
	return &pbv.File{Key: file.Key, Meta: file.Meta, Data: file.Data}, nil

}

func (vs *volumeService) TransactionStart(ctx context.Context, FileIn *pbv.Empty) (*pbv.Empty, error) {
	DBDATABULKMODEL = true
	TransBegin()
	return &pbv.Empty{}, nil
}

func (vs *volumeService) TransactionEnd(ctx context.Context, FileIn *pbv.Empty) (*pbv.Empty, error) {
	TransCommit()
	DBDATABULKMODEL = false
	return &pbv.Empty{}, nil
}

func (vs *volumeService) StreamSendFile(stream pbv.VolumeService_StreamSendFileServer) error {
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			Logger.Error("StreamSendFile:", err)
			return err
		}

		k := in.Key
		Logger.Info("Received File: KEY:", k)
		if len(k) == 32 {
			did := k[0:1]
			Logger.Debug("StreamSendFile: DBID:", did)
			err_save := SaveFile(DBDATA[did], k, in.Meta, in.Data)
			if err_save != nil {
				Logger.Error(err_save)
			} else {
				Logger.Info("SaveFile OOOOK")
			}
			note := &pbv.File{Key: k}
			if err := stream.Send(note); err != nil {
				return err
			}

		} else {
			Logger.Error("StreamSendFile: ByteMD5 Error:")
			note := &pbv.File{Key: ""}
			if err := stream.Send(note); err != nil {
				return err
			}
			return errors.New("StreamSendFile: ByteMD5 Error.")
		}

	}
	return nil
}
