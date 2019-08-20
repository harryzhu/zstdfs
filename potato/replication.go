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
			replicate("sync", ip_port)
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

func replicate(cat, ip_port string) error {
	prefix := strings.Join([]string{cat, ip_port}, "/")
	logger.Debug("sync: prefix length: ", len(prefix), ", prefix: ", prefix)

	fileKeys, err := MetaScan([]byte(prefix), 400)
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

			logger.Debug("runStreamSendMessage: Key: ", string(in.Key), string(in.Action), in.ErrCode)
			if in != nil && len(in.Key) > 0 && len(in.Time) > 0 && len(in.Action) > 0 && in.ErrCode == 0 {
				pk := strings.Join([]string{prefix, string(in.Key), in.Time, in.Action}, "/")

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
	timeNano := ""
	frequest := &pbv.Message{}
	for k, fk := range fileKeys {
		logger.Debug("Sync Key Index:", k, ", ", fk)
		arr_fk := strings.Split(fk, "/")

		if len(arr_fk) == 5 {
			action = arr_fk[4]
			fileKey = arr_fk[2]
			timeNano = arr_fk[3]
		}

		if len(action) <= 0 || len(fileKey) <= 0 {
			continue
		}

		frequest.Key = []byte(fileKey)
		frequest.Action = action
		frequest.ErrCode = 0
		frequest.Time = timeNano

		switch action {
		case "get":
			{
				frequest.Data = nil
			}
		case "set":
			{
				keyExistsMsg, err := client.HandleFile(ctx, &pbv.Message{Key: []byte(fileKey), Action: "exists"})
				if err == nil {
					logger.Debug("Sync/Set/ ErrCode: ", keyExistsMsg.ErrCode)
					if keyExistsMsg.ErrCode == 200 {
						logger.Debug("Sync/Set/ had been skipped on: ", ip_port, ", key: ", fk)
						MetaDelete([]byte(fk))
						continue
					}
				} else {
					logger.Debug("Check if key exists, Error: ", err)
				}

				data, err := EntityGet([]byte(fileKey))
				if err != nil || data == nil {
					logger.Debug("Stream_02: EntityGet Error: ", fileKey)
					continue
				} else {
					frequest.Data = data
				}
			}
		case "del":
			{
				frequest.Data = nil
			}
		case "ban":
			{
				frequest.Data = nil
			}
		case "pub":
			{
				frequest.Data = nil
			}
		default:
			{
				continue
			}
		}

		if err := stream.Send(frequest); err != nil {
			logger.Warn("StreamSend: Failed to send a filerequest: ", err)
		}
		logger.Debug("StreamSend: Deleting: action: ", action, " key: ", fileKey)

	}

	stream.CloseSend()
	<-waitc

	return nil
}
