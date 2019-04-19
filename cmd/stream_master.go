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
		synced := in.Synced
		inmaster := in.InMaster
		created := in.Created
		Logger.Debug("masterService: Received File: KEY:", key)
		if len(key) == 32 {
			err_save := MasterCreateNodeFiles(key, node, size, synced, inmaster, created)
			if err_save != nil {
				inmaster = 0
				Logger.Error(err_save)
			} else {
				inmaster = 1
				Logger.Debug("masterService SaveFile OK")
			}
			response_out := &pbm.File{ErrorCode: 0, Key: key, Node: node, Synced: synced, InMaster: inmaster}
			if err := stream.Send(response_out); err != nil {
				return err
			}

		} else {
			Logger.Error("StreamSendFile: ByteMD5 Error:")
			note := &pbm.File{ErrorCode: 500, Key: ""}
			if err := stream.Send(note); err != nil {
				return err
			}
			return errors.New("StreamSendFile: ByteMD5 Error.")
		}

	}
}
