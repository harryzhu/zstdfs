package cmd

import (
	"context"
	"time"

	pb "zstdfs/pbs"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	gConn       *grpc.ClientConn
	gClient     pb.BadgerClient
	gMaxMsgSize int = 1 << 30
)

func SetGrpcClient() {
	var err error
	gConn, err = grpc.NewClient(rpcServer,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(gMaxMsgSize),
			grpc.MaxCallSendMsgSize(gMaxMsgSize),
		))
	PrintError("SetClient", err)
	gClient = pb.NewBadgerClient(gConn)
}

func gcSet(data []byte) ([]byte, error) {
	ctx, _ := context.WithTimeout(context.Background(), time.Second*90)
	r, err := gClient.Set(ctx, &pb.Item{Data: data, Sum64: GetXxhash(data)})
	if err != nil {
		PrintError("gcSet", err)
		return nil, err
	}

	if r.GetErrcode() != 0 {
		return nil, NewError(string(r.GetStatus()))
	}

	return r.GetKey(), nil
}

func gcGet(key []byte) ([]byte, error) {
	ctx, _ := context.WithTimeout(context.Background(), time.Second*90)
	r, err := gClient.Get(ctx, &pb.Item{Key: key})
	if err != nil {
		PrintError("gcGet", err)
		return nil, err
	}
	rData := r.Data
	//DebugInfo("gcGet", string(rData))
	return rData, nil
}

func gcList(prefix string, pagenum int) []string {
	ctx, _ := context.WithTimeout(context.Background(), time.Second*90)
	res, err := gClient.List(ctx, &pb.ListFilter{Prefix: prefix, Pagenum: int32(pagenum)})
	if err != nil {
		PrintError("gcGet", err)
		return []string{}
	}
	//DebugInfo("gcList", res)
	return res.Keys
}

func gcExists(key []byte) bool {
	ctx, _ := context.WithTimeout(context.Background(), time.Second*90)
	r, err := gClient.Exists(ctx, &pb.Item{Key: key})
	if err != nil {
		PrintError("gcExists", err)
		return false
	}
	rStatus := r.Status
	if string(rStatus) == "1" {
		return true
	}
	return false
}
