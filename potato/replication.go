package potato

import (
	"io"
	//"strconv"
	//"reflect"
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

	if IsMaster == false {
		return nil
	}

	if IsReplicationNeeded == false {
		Logger.Debug("====no replication needed==== IsReplicationNeeded:", IsReplicationNeeded)
		return nil
	}

	msc := MetaSyncCount()
	if msc <= 0 {
		Logger.Debug("====no replication needed==== MetaSyncCount:", msc)
		return nil
	}

	slaves := CFG.Replication.Slaves
	if len(slaves) > 0 {
		//Logger.Debug("Begin: IsReplicationNeeded: ", IsReplicationNeeded)
		IsReplicationNeeded = false
		wg := sync.WaitGroup{}

		for _, slave := range slaves {
			conn, err := grpc.Dial(slave, grpc.WithInsecure(),
				grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(GRPCMAXMSGSIZE)))
			if err != nil {
				Logger.Error("Slave Conn Error:", err)
				return nil
			}
			defer conn.Close()

			Logger.Debug("Slave Connection State: ", slave, " : ", conn.GetState().String())
			if conn.GetState().String() == "IDLE" || conn.GetState().String() == "CONNECTING" || conn.GetState().String() == "READY" {
				wg.Add(1)
				go func() {
					defer wg.Done()
					Logger.Debug("Thread Start: replicate to: ", slave)
					replicate(slave)
					Logger.Debug("Thread End: ", slave)
					time.Sleep(1 * time.Second)

				}()
				time.Sleep(1 * time.Second)
			}

		}

		wg.Wait()
		//Logger.Debug("Complete: IsReplicationNeeded: ", IsReplicationNeeded)
		IsReplicationNeeded = true
	}
	return nil
}

func replicate(ip_port string) error {

	fileKeys := []string{}

	ssm, err := CMETA.Snapshot()
	if err != nil {
		Logger.Error("expected ssm ok")
	}
	prefix := strings.Join([]string{"sync:", ip_port, ":"}, "")
	prefix_length := len(prefix)
	Logger.Debug("sync: prefix_length: ", prefix_length, ", prefix: ", prefix)
	iter, err := ssm.StartIterator([]byte(prefix), nil, moss.IteratorOptions{})
	if err != nil || iter == nil {
		Logger.Error("expected iter")
	}

	for i := 0; i < 50; i++ {
		k, v, err := iter.Current()
		if err != nil {
			continue
		}
		if k != nil && v != nil {
			k_raw := string(k)[prefix_length:]
			Logger.Debug("add to sync list:", k_raw)
			fileKeys = append(fileKeys, k_raw)
		}

		err = iter.Next()
		if err != nil {
			break
		}
	}

	ssm.Close()

	fileKeys_len := len(fileKeys)
	if fileKeys_len == 0 {
		Logger.Debug("No Entities Replication Needed.")
		return nil
	}

	conn, err := grpc.Dial(ip_port, grpc.WithInsecure(),
		grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(GRPCMAXMSGSIZE)))
	if err != nil {
		Logger.Error("Client Conn Error:", err)
		return nil
	}
	defer conn.Close()

	Logger.Debug("Client Connection State: ", conn.GetState().String())
	if conn.GetState().String() == "IDLE" || conn.GetState().String() == "CONNECTING" || conn.GetState().String() == "READY" {
		c := pbv.NewVolumeServiceClient(conn)

		runStreamSendFile(c, ip_port, prefix, fileKeys)
	}

	return nil
}

func runStreamSendFile(client pbv.VolumeServiceClient, ip_port string, prefix string, fileKeys []string) error {
	Logger.Debug("Start Replication..........")
	fileKeys_len := len(fileKeys)
	if fileKeys_len == 0 {
		return nil
	}
	Logger.Debug("fileKeys length: ", fileKeys_len)

	//for i := 0; i < fileKeys_len; i++ {
	//ctx, cancel := context.WithTimeout(context.Background(), 24*3600*time.Second)
	ctx, _ := context.WithTimeout(context.Background(), 24*3600*time.Second)
	//defer cancel()

	stream, err := client.StreamSendFile(ctx)
	if err != nil {
		Logger.Warn("Stream_01 StreamSendFile ERROR: ", err)
		return err
	}
	var lock sync.Mutex
	waitc := make(chan struct{})
	go func() {
		var deleteKeys []string
		lock.Lock()
		for {
			in, err := stream.Recv()
			if err == io.EOF {
				// read done.
				close(waitc)
				Logger.Debug("break loop: stream.Recv, close waitc: ", err)
				break
			}
			if err != nil {
				Logger.Warn("Stream_01: Failed to receive a filerequest: ", err)
			}
			if len(in.Key) > 0 {
				pk := strings.Join([]string{prefix, in.Key}, "")

				deleteKeys = append(deleteKeys, pk)

				Logger.Debug("synced key: ", in.Key, ", MetaDelete key: ", pk)
			}
		}
		lock.Unlock()

		if len(deleteKeys) > 0 {
			Logger.Info("MultiDelete Keys.")
			MetaMultiDelete(deleteKeys)
		}

		return
	}()

	for k, fk := range fileKeys {
		Logger.Debug("Sync Key Index:", k, ", ", fk)
		data, err := EntityGet(fk)
		if err != nil || data == nil {
			continue
		} else {
			frequest := &pbv.File{Key: fk, Data: data}
			if err := stream.Send(frequest); err != nil {
				Logger.Warn("Stream_02: Failed to send a filerequest: ", err)
			}
		}
	}

	stream.CloseSend()
	<-waitc

	return nil
}
