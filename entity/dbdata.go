package entity

import (
	//"encoding/json"
	"log"
	"strconv"

	"github.com/dgraph-io/badger/v3"
	"github.com/harryzhu/potatofs/util"
)

var (
	DataDB      *badger.DB
	MaxSizeByte int
)

func OpenDataDB(datadir string) {
	opts := badger.DefaultOptions(datadir)
	opts.SyncWrites = true

	var err error
	DataDB, err = badger.Open(opts)
	if err != nil {
		log.Fatal(err)
	}
}

func getData(k []byte) (data []byte, err error) {
	if k == nil {
		return nil, util.ErrorString("getData: key cannot be empty")
	}

	DataDB.View(func(txn *badger.Txn) error {
		item, err := txn.Get(k)
		if err != nil {
			return err
		}
		data, err = item.ValueCopy(nil)
		if err != nil {
			return err
		}
		return nil
	})

	if data != nil {
		return data, nil
	}
	//return nil, nil
	return nil, util.ErrorString("cannot get data")
}

func setData(key, data []byte) error {
	if key == nil || data == nil {
		return nil
	}

	if len(data) > MaxSizeByte {
		return util.ErrorString("entity is oversize, max_size: " + strconv.Itoa(MaxSizeByte))
	}

	err := DataDB.View(func(txn *badger.Txn) error {
		_, err := txn.Get(key)
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return DataDB.Update(func(txn *badger.Txn) error {
			err := txn.Set(key, data)
			if err == nil {
				BoltSave([]byte("data"), key, []byte("-1"))
			}
			return err
		})
	}

	return nil

}

func ExistsData(k []byte) bool {
	if k == nil {
		return false
	}

	err := DataDB.View(func(txn *badger.Txn) error {
		_, err := txn.Get(k)
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return false
	}
	return true
}
