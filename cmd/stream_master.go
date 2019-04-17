package cmd

import (
	"errors"
	pbm "hazhufs/pb/master_pb"
	"io"
	//"strings"
	//"golang.org/x/net/context"
)

type masterService struct {
}

func (ms *masterService) StreamSendFile(stream pbm.MasterService_StreamSendFileServer) error {
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			Logger.Error("masterService: StreamSendFile:", err)
			return err
		}

		key := in.Key
		node := in.Node
		size := in.Size
		created := in.Created
		Logger.Debug("masterService: Received File: KEY:", key)
		if len(key) == 32 {
			err_save := MasterSaveNodeFiles(key, node, size, created)
			if err_save != nil {
				Logger.Error(err_save)
			} else {
				Logger.Debug("masterService SaveFile OK")
			}
			note := &pbm.File{Key: key}
			if err := stream.Send(note); err != nil {
				return err
			}

		} else {
			Logger.Error("StreamSendFile: ByteMD5 Error:")
			note := &pbm.File{Key: ""}
			if err := stream.Send(note); err != nil {
				return err
			}
			return errors.New("StreamSendFile: ByteMD5 Error.")
		}

	}
}
