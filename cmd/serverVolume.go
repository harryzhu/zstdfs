package cmd

import (
	"io"
	"log"
	"net"
	"strings"

	"github.com/harryzhu/potatofs/entity"
	pb "github.com/harryzhu/potatofs/pb"
	"golang.org/x/net/context"

	//"github.com/harryzhu/potatofs/util"
	"google.golang.org/grpc"
)

type VolumeService struct{}

var EmptyMessage = &pb.Message{Key: nil, Meta: nil, Data: nil}
var EmptyEntity entity.Entity
var ett entity.Entity

func (vs *VolumeService) StreamSendMessage(stream pb.VolumeService_StreamSendMessageServer) error {
	resp := EmptyMessage
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			log.Println("io.EOF")
			return nil
		}
		if err != nil {
			log.Println("StreamSendMessage:", err)
			return err
		}

		if in.Key == nil {
			continue
		}

		if len(in.Data) > entity.MaxSizeByte {
			continue
		}

		if len(in.Meta) > entity.MaxSizeByte {
			continue
		}

		resp.Key = in.Key
		resp.Meta = nil
		resp.Data = nil

		ett = EmptyEntity

		ett.Key = in.Key
		ett.Meta = in.Meta
		ett.Data = in.Data

		if err := ett.Save(); err != nil {
			resp.Meta = []byte(err.Error())
			//return err
		}

		if err := stream.Send(resp); err != nil {
			log.Println("StreamSendMessage(Response):", err)
			return err
		}
		return nil
	}

}

func (vs *VolumeService) GetMessage(ctx context.Context, MessageIn *pb.Message) (*pb.Message, error) {
	resp := &pb.Message{Key: MessageIn.Key, Meta: nil, Data: nil}

	ett := entity.NewEntity(MessageIn.Key, nil, nil)
	metadata, err := ett.Get()
	if err != nil {
		return nil, err
	}
	resp.Meta = metadata.Meta
	resp.Data = metadata.Data
	return resp, nil
}

func StartVolumeServer() {
	listening_ip_port := strings.Join([]string{Config.IP, Config.VolumePort}, ":")
	listening, err := net.Listen("tcp", listening_ip_port)
	if err != nil {
		log.Println("Failed to listen: ", err)
	} else {
		log.Println("Endpoint RPC: ", listening_ip_port)
	}

	grpcServerVolume := grpc.NewServer(
		grpc.MaxMsgSize(MaxMessageSize),
		grpc.MaxRecvMsgSize(MaxMessageSize),
		grpc.MaxSendMsgSize(MaxMessageSize))

	pb.RegisterVolumeServiceServer(grpcServerVolume, &VolumeService{})

	grpcServerVolume.Serve(listening)
}
