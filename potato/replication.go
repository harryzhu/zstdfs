package potato

import (
	"io"
	//"strconv"
	//"reflect"
	"strings"
	"sync"
	"time"

	pbv "./pb/volume"
	//"github.com/syndtr/goleveldb/leveldb"

	//"github.com/dgraph-io/badger"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

func RunReplicateParallel() error {
	if volumePeersLength == 0 {
		return nil
	}

	if isMaster == false {
		return nil
	}

	have_volume_peers_live := false
	for _, b_live := range volumePeersLive {
		if b_live == true {
			have_volume_peers_live = true
			break
		}
	}

	if have_volume_peers_live == false {
		//logger.Debug("====no replication needed==== no volume peers online")
		return nil
	}

	if isReplicationNeeded == false {
		//logger.Debug("====no replication needed==== isReplicationNeeded: false")
		return nil
	}

	if MetaScanExists([]byte("sync/")) == false {
		//logger.Debug("====no replication needed==== MetaScanExists: false")
		return nil
	}

	logger.Debug("Begin: Replication.")
	isReplicationNeeded = false
	wg_rep := sync.WaitGroup{}

	for ip_port, online := range volumePeersLive {
		if online == false {
			continue
		}

		wg_rep.Add(1)
		go func() {
			defer wg_rep.Done()
			//logger.Debug("Thread Start: replicate to: ", ip_port)
			replicate("sync", "del", ip_port)
			time.Sleep(50 * time.Millisecond)
			replicate("sync", "set", ip_port)
			//logger.Debug("Thread End: ", ip_port)
			time.Sleep(50 * time.Millisecond)

		}()

		time.Sleep(50 * time.Millisecond)

	}

	wg_rep.Wait()
	isReplicationNeeded = true
	logger.Debug("Complete: Replication.")

	return nil
}

func replicate(cat, action, ip_port string) error {
	prefix := strings.Join([]string{cat, action, ip_port}, "/")
	logger.Debug("sync: prefix length: ", len(prefix), ", prefix: ", prefix)

	fileKeys, err := MetaScan([]byte(prefix), 100)
	if err != nil {
		logger.Debug("MetaScan Error.")
		return err
	}
	fileKeys_len := len(fileKeys)
	if fileKeys_len == 0 {
		logger.Debug("No Entities Replication Needed.")
		return nil
	}

	conn, err := grpc.Dial(ip_port, grpc.WithInsecure(),
		grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(grpcMAXMSGSIZE), grpc.MaxCallRecvMsgSize(grpcMAXMSGSIZE)))
	if err != nil {
		logger.Error("Client Conn Error:", err)
		return nil
	}
	//defer conn.Close()

	c := pbv.NewVolumeServiceClient(conn)

	runStreamSendFile(c, ip_port, prefix, fileKeys)

	return nil
}

func runStreamSendFile(client pbv.VolumeServiceClient, ip_port string, prefix string, fileKeys []string) error {
	logger.Debug("Start Replication from: ", volumeSelf, " to ", ip_port, ", Prefix: ", prefix, " ..........")
	fileKeys_len := len(fileKeys)
	if fileKeys_len == 0 {
		return nil
	}
	logger.Debug("fileKeys length: ", fileKeys_len)

	ctx, _ := context.WithTimeout(context.Background(), 24*3600*time.Second)
	//ctx, cancel() := context.WithTimeout(context.Background(), 24*3600*time.Second)
	//defer cancel()

	stream, err := client.StreamSendMessage(ctx)
	if err != nil {
		logger.Warn("runStreamSendFile ERROR: ", err)
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
				logger.Debug("runStreamSendFile: break loop: stream.Recv, close waitc: ", err)
				break
			}
			if err != nil {
				logger.Warn("runStreamSendFile: Failed to receive a file response: ", err)
				continue
			}

			logger.Debug("runStreamSendMessage: Key: ", string(in.Key), string(in.Action), in.ErrCode, string(in.Data))
			if in != nil && len(in.Key) > 0 && in.ErrCode == 0 {
				pk := strings.Join([]string{prefix, string(in.Key)}, "/")

				deleteKeys = append(deleteKeys, pk)

				logger.Debug("OK: runStreamSendFile key: ", string(in.Key), ", MetaDelete key: ", pk)
			}
		}
		lock.Unlock()

		if len(deleteKeys) > 0 {
			logger.Debug("MultiDelete Keys.")
			MetaMultiDelete(deleteKeys)
		}

		//return
	}()

	action := ""
	fileKey := ""
	frequest := &pbv.Message{}
	for k, fk := range fileKeys {
		logger.Debug("Sync Key Index:", k, ", ", fk)
		arr_fk := strings.Split(fk, "/")

		if len(arr_fk) == 4 {
			action = arr_fk[1]
			fileKey = arr_fk[3]
		}

		if len(action) <= 0 || len(fileKey) <= 0 {
			continue
		}

		if action == "set" {
			data, err := EntityGet([]byte(fileKey))
			if err != nil || data == nil {
				logger.Debug("Stream_02: EntityGet Error: ", fileKey)
				continue
			} else {
				frequest.Key = []byte(fileKey)
				frequest.Data = data
				frequest.Action = action
				frequest.ErrCode = 0
				if err := stream.Send(frequest); err != nil {
					logger.Warn("Stream_02: Failed to send a filerequest: ", err)
				}
				logger.Debug("Stream_02: Sending: action: ", action, " key: ", fileKey)
			}
		}

		if action == "del" {
			frequest.Key = []byte(fileKey)
			frequest.Data = nil
			frequest.Action = action
			frequest.ErrCode = 0
			if err := stream.Send(frequest); err != nil {
				logger.Warn("Stream_02: Failed to send a filerequest: ", err)
			}
			logger.Debug("Stream_02: Deleting: action: ", action, " key: ", fileKey)
		}

	}

	stream.CloseSend()
	<-waitc

	return nil
}
