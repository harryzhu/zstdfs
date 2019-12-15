package potato

import (
	"context"
	"sync"

	//pbv "./pb/volume"
	//"github.com/dgraph-io/badger"
	"github.com/dgraph-io/badger/pb"
)

func BDBSubscribe() {
	var wg_sub sync.WaitGroup
	wg_sub.Add(1)
	go func() {
		err := bdb.Subscribe(context.Background(), func(kvs *pb.KVList) {
			for _, kv := range kvs.GetKv() {

				logger.Info("Subscribe:", string(kv.Key))
			}

		},
			[]byte("a"), []byte("b"), []byte("c"), []byte("d"),
			[]byte("e"), []byte("f"), []byte("0"), []byte("1"),
			[]byte("2"), []byte("3"), []byte("4"), []byte("5"),
			[]byte("6"), []byte("7"), []byte("8"), []byte("9"))

		if err != nil {
			logger.Error("Subscribe:", err)
		}
	}()
	wg_sub.Wait()

}
