package potato

import (
	"io"
	//"strconv"
	"strings"
	"sync"
	"time"

	pbv "./pb/volume"
	"github.com/couchbase/moss"

	//"github.com/dgraph-io/badger"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

func RunReplicateParallel() error {
	if SLAVES_LENGTH == 0 {
		return nil
	}
	if CFG.Replication.Is_master == false {
		return nil
	}

	if IsReplicationNeeded == false {
		Logger.Debug("***IsReplicationNeeded***: ", IsReplicationNeeded)
		return nil
	}

	slaves := CFG.Replication.Slaves
	if len(slaves) > 0 {
		IsReplicationNeeded = false
		wg := sync.WaitGroup{}
		wg.Add(len(slaves))
		for _, slave := range slaves {
			Logger.Debug("slave: ", slave)
			go func() {
				Logger.Debug("Start: replicate to: ", slave)
				replicate(slave)
				Logger.Debug("End: replicate to: ", slave)
				time.Sleep(3 * time.Second)
				wg.Done()

			}()
			time.Sleep(1 * time.Second)
		}
		wg.Wait()
		IsReplicationNeeded = true
	}
	return nil
}

func replicate(ip_port string) error {

	conn, err := grpc.Dial(ip_port, grpc.WithInsecure(),
		grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(GRPCMAXMSGSIZE)))
	if err != nil {
		Logger.Error("Replication Error:", err)
		return nil
	}
	defer conn.Close()

	c := pbv.NewVolumeServiceClient(conn)

	runStreamSendFile(c, ip_port)

	return nil
}

func runStreamSendFile(client pbv.VolumeServiceClient, ip_port string) {

	Logger.Info("Start Replication..........")
	fileRequests := []*pbv.File{}

	ssm, err := CMETA.Snapshot()
	if err != nil {
		Logger.Error("expected ssm ok")
	}
	prefix := strings.Join([]string{"sync:", ip_port, ":"}, "")
	prefix_length := len(prefix)
	Logger.Debug("sync prefix: ", prefix_length, ", ", prefix)
	iter, err := ssm.StartIterator([]byte(prefix), nil, moss.IteratorOptions{})
	if err != nil || iter == nil {
		Logger.Error("expected iter")
	}

	for i := 0; i < 100; i++ {
		k, v, err := iter.Current()
		if err != nil {
			continue
		}
		if k != nil && v != nil {
			k_raw := string(k)[prefix_length:]
			Logger.Debug("will replicate:", k_raw)
			data, err := EntityGet(k_raw)
			if err != nil {
				Logger.Debug("will sync:", k_raw)
				continue
			}
			fileRequests = append(fileRequests, &pbv.File{Key: k_raw, Data: data})
		}

		err = iter.Next()
		if err != nil {
			break
		}
	}

	ssm.Close()

	fileRequests_len := len(fileRequests)
	Logger.Info("fileRequests length: ", fileRequests_len)

	for i := 0; i < fileRequests_len; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 24*3600*time.Second)
		defer cancel()

		stream, err := client.StreamSendFile(ctx)
		if err != nil {
			Logger.Warn("Client StreamSendFile ERROR: %v", err)
		}
		waitc := make(chan struct{})
		go func() {
			prefix_without_colon := prefix[0 : len(prefix)-1]
			for {
				in, err := stream.Recv()
				if err == io.EOF {
					// read done.
					close(waitc)
					return
				}
				if err != nil {
					Logger.Warn("Failed to receive a filerequest : %v", err)
				}
				MetaDelete(prefix_without_colon, in.Key)
				Logger.Info("Got message response key: ", "/", ": ", in.Key)
			}
		}()
		for _, filerequest := range fileRequests {
			if err := stream.Send(filerequest); err != nil {
				Logger.Warn("Failed to send a filerequest: ", err)
			}
		}
		stream.CloseSend()
		<-waitc

	}

}
