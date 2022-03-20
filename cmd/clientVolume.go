package cmd

import (
	"io"
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/harryzhu/potatofs/entity"
	pb "github.com/harryzhu/potatofs/pb"
	"github.com/harryzhu/potatofs/util"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

func clientStreamSendFile(client pb.VolumeServiceClient) {
	files := util.DirWalker(DIR, FILTER)
	files_total := len(files)
	log.Println("total files: ", files_total)

	fileRequests := []*pb.Message{}

	fkey := ""
	ftags := ""
	fcomment := ""
	for _, fname := range files {
		fkey = strings.Join([]string{USERNAME, ALBUM, filepath.Base(fname)}, "/")
		if TAGS != "" {
			ftags = TAGS
		}

		if COMMENT != "" {
			fcomment = COMMENT
		} else {
			fcomment = strings.ReplaceAll(filepath.Dir(fname), filepath.Dir(filepath.Dir(fname)), "")
			fcomment = strings.Trim(fcomment, "/")
			fcomment = strings.Trim(fcomment, "\\")
		}

		ett, err := entity.NewEntityByFile(fname, fkey, ftags, fcomment)
		if err != nil {
			continue
		}
		ett.Key = []byte(fkey)
		fileRequests = append(fileRequests, &pb.Message{
			Key:  ett.Key,
			Meta: ett.Meta,
			Data: ett.Data,
		})

		ctx, _ := context.WithTimeout(context.Background(), 24*3600*time.Second)
		stream, err := client.StreamSendMessage(ctx)
		//stream, err := client.StreamBatchSendMessage(ctx)
		if err != nil {
			log.Printf("Client StreamSendFile ERROR: %v", err)
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
					log.Printf("Failed to receive a filerequest : %v", err)
				}
				log.Println("SUCCESS: ", string(in.Key))
			}
		}()
		for _, filerequest := range fileRequests {
			if err := stream.Send(filerequest); err != nil {
				log.Printf("Failed to send: %v", err)
			}
		}
		stream.CloseSend()
		<-waitc
		fileRequests = fileRequests[0:0]
	}

}

func clientStreamSendFile2(client pb.VolumeServiceClient) {
	files := util.DirWalker(DIR, FILTER)
	files_total := len(files)
	log.Println("total files: ", files_total)

	fileRequest := &pb.Message{}

	fkey := ""
	ftags := ""
	fcomment := ""

	ctx, _ := context.WithTimeout(context.Background(), 24*3600*time.Second)
	stream, err := client.StreamSendMessage(ctx)

	if err != nil {
		log.Printf("Client StreamSendFile ERROR: %v", err)
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
				log.Printf("Failed to receive a filerequest : %v", err)
			}
			log.Println("SUCCESS: ", string(in.Key))
		}
	}()

	for _, fname := range files {
		fkey = strings.Join([]string{USERNAME, ALBUM, filepath.Base(fname)}, "/")
		if TAGS != "" {
			ftags = TAGS
		}

		if COMMENT != "" {
			fcomment = COMMENT
		} else {
			fcomment = strings.ReplaceAll(filepath.Dir(fname), filepath.Dir(filepath.Dir(fname)), "")
			fcomment = strings.Trim(fcomment, "/")
			fcomment = strings.Trim(fcomment, "\\")
		}

		ett, err := entity.NewEntityByFile(fname, fkey, ftags, fcomment)
		if err != nil {
			continue
		}
		ett.Key = []byte(fkey)

		fileRequest = &pb.Message{
			Key:  ett.Key,
			Meta: ett.Meta,
			Data: ett.Data,
		}

		if err := stream.Send(fileRequest); err != nil {
			log.Printf("Failed to send: %v", err)
		}

	}

	stream.CloseSend()
	<-waitc

}

func Fput() {
	listening_ip_port := strings.Join([]string{Config.IP, Config.VolumePort}, ":")
	conn, err := grpc.Dial(listening_ip_port, grpc.WithInsecure(),
		grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(MaxMessageSize)))
	defer conn.Close()
	if err != nil {
		log.Printf("Failed to connect: %v", err)
	} else {
		log.Printf("start as upload role.")
	}

	c := pb.NewVolumeServiceClient(conn)
	t_start := time.Now()
	clientStreamSendFile(c)
	log.Print("Function Elapsed: upload: ", time.Since(t_start))
}
