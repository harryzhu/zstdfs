package cmd

import (
	"io"
	"log"

	//"path/filepath"
	//"strings"
	"time"

	"github.com/harryzhu/potatofs/entity"
	pb "github.com/harryzhu/potatofs/pb"

	//"github.com/harryzhu/potatofs/util"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

func syncGetBoltList() (fileRequests []*pb.Message) {
	metaKeys := entity.BoltList([]byte("meta"), Config.SyncPageSize)

	keysTotal := len(metaKeys)
	if keysTotal < 1 {
		return nil
	} else {
		log.Println("sync keys: ", keysTotal)
	}

	var ett entity.Entity
	var err error
	for _, key := range metaKeys {
		ett, err = entity.Entity{Key: key, Meta: nil, Data: nil}.Get()
		if err != nil {
			log.Println(err)
		}

		//log.Println(string(ett.Key), " : ", string(ett.Meta))
		if ett.Key != nil && ett.Meta != nil {
			fileRequests = append(fileRequests, &pb.Message{
				Key:  ett.Key,
				Meta: ett.Meta,
				Data: ett.Data,
			})
		}
	}

	return fileRequests
}

func syncClientStreamSendFile(client pb.VolumeServiceClient, fileRequests []*pb.Message) error {
	ctx, _ := context.WithTimeout(context.Background(), 24*3600*time.Second)

	stream, err := client.StreamSendMessage(ctx)

	if err != nil {
		log.Printf("cannot connect slave node,will not start sync: %v", err)
		return err
	}
	waitc := make(chan struct{})
	go func() {
		for {
			in, err := stream.Recv()
			if err == io.EOF {
				// read done.
				close(waitc)
				return
			}
			if err != nil {
				log.Println(err)
			} else {
				if err = entity.BoltDelete([]byte("meta"), in.Key); err != nil {
					log.Println(err)
				}

				log.Println("Synced: ", string(in.Key))
			}

		}
	}()

	for _, filerequest := range fileRequests {
		if err := stream.Send(filerequest); err != nil {
			log.Printf("Failed to Sync: %v", err)
		}
	}
	stream.CloseSend()
	<-waitc

	return nil
}

func StartSync() error {
	listening_ip_port := Config.SyncSlave
	conn, err := grpc.Dial(listening_ip_port, grpc.WithInsecure(),
		grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(MaxMessageSize)))
	defer conn.Close()
	if err != nil {
		log.Printf("cannot connect: %v, %v", listening_ip_port, err)
		return err
	}

	c := pb.NewVolumeServiceClient(conn)
	t_start := time.Now()

	fileRequests := syncGetBoltList()

	if len(fileRequests) > 0 {
		if err = syncClientStreamSendFile(c, fileRequests); err != nil {
			time.Sleep(time.Second * 3)
		}
		log.Println("StartSync Elapsed: upload: ", time.Since(t_start))
	} else {
		log.Println("no keys sync needed")
		time.Sleep(time.Second * 3)
	}

	return nil
}
