package cmd

import (
	"fmt"
	"path/filepath"

	badger "github.com/dgraph-io/badger/v4"
)

func badgerConnect() *badger.DB {
	data_dir := ToUnixSlash(filepath.Join(DATA_DIR, "fbin"))
	MakeDirs(data_dir)
	DebugInfo("badgerConnect", data_dir)
	opts := badger.DefaultOptions(data_dir)
	opts.Dir = data_dir
	opts.ValueDir = data_dir
	opts.BaseTableSize = 256 << 20
	opts.NumVersionsToKeep = 1
	opts.SyncWrites = false
	opts.ValueThreshold = 256
	opts.CompactL0OnClose = true

	db, err := badger.Open(opts)
	FatalError("badgerConnect", err)
	return db
}

func badgerSave(val []byte) (key []byte) {
	if val == nil {
		DebugWarn("badgerSave", "val cannot be empty")
		return nil
	}
	key = SumBlake3(val)

	err := bgrdb.Update(func(txn *badger.Txn) error {
		_, err := txn.Get(key)
		if err == nil {
			return nil
		}
		err = txn.Set([]byte(key), ZstdBytes(val))
		PrintError("badgerSave", err)
		return err
	})
	if err != nil {
		return nil
	}
	return key
}

func badgerGet(key []byte) (val []byte) {
	if key == nil {
		DebugWarn("badgerGet", "key cannot be empty")
		return nil
	}

	bgrdb.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			DebugWarn("badgerGet", err, ":", string(key))
			return nil
		}
		itemVal, err := item.ValueCopy(nil)
		PrintError("badgerGet", err)
		//DebugInfo("badgerGet", len(itemVal))
		val = UnZstdBytes(itemVal)
		return err
	})

	return val
}

func badgerList(uname string) {

	err := bgrdb.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			k := item.Key()
			err := item.Value(func(v []byte) error {
				DebugInfo("badgerList", fmt.Sprintf("%s", k), ": ", len(v))
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		PrintError("badgerList", err)
	}
}

func badgerExists(key []byte) bool {
	if key == nil {
		DebugWarn("badgerExists", "key cannot be empty")
		return false
	}

	err := bgrdb.View(func(txn *badger.Txn) error {
		_, err := txn.Get(key)
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
