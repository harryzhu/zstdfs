package potato

import (
	"io"
	"strings"
	"time"

	pbv "./pb/volume"
	"github.com/couchbase/moss"

	//"github.com/dgraph-io/badger"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

func Replicate() error {
	if CFG.Replication.Is_master == false {
		return nil
	}
	addr := strings.Join([]string{CFG.Replication.Slave_ip, CFG.Replication.Slave_port}, ":")
	conn, err := grpc.Dial(addr, grpc.WithInsecure(),
		grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(GRPCMAXMSGSIZE)))
	if err != nil {
		Logger.Error("Replication Error:", err)
	}
	defer conn.Close()

	c := pbv.NewVolumeServiceClient(conn)

	runStreamSendFile(c)

	return nil
}

func runStreamSendFile(client pbv.VolumeServiceClient) {
	fileRequests := []*pbv.File{}
	Logger.Info("Start Replication..........")
	ssm, err := CMETA.Snapshot()
	if err != nil {
		Logger.Error("expected ssm ok")
	}
	iter, err := ssm.StartIterator([]byte("sync:"), nil, moss.IteratorOptions{})
	if err != nil || iter == nil {
		Logger.Error("expected iter")
	}

	for i := 0; i < 10; i++ {
		k, v, err := iter.Current()
		if err != nil {
			continue
		}
		if k != nil && v != nil {
			k_raw := string(k)[5:]
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

	// d, _ := MetaGet("sync", "15ba977eca49f2ebcd6a0ed7ae384b83")
	// Logger.Info("metageteeeest:", string(d))

	// Logger.Info("runStre***")
	// return

	// DB.View(func(txn *badger.Txn) error {
	// 	opts := badger.DefaultIteratorOptions
	// 	opts.PrefetchSize = 100
	// 	it := txn.NewIterator(opts)
	// 	defer it.Close()
	// 	for it.Rewind(); it.Valid(); it.Next() {
	// 		item := it.Item()
	// 		k := item.Key()
	// 		err := item.Value(func(v []byte) error {
	// 			//Logger.Info("key: ", string(k))
	// 			fileRequests = append(fileRequests, &pbv.File{Key: string(k), Data: Unzip(v)})

	// 			return nil
	// 		})
	// 		if err != nil {
	// 			return err
	// 		}
	// 	}
	// 	return nil
	// })

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
				MetaDelete("sync", in.Key)
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
