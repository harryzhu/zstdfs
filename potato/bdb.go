package potato

import (
	"errors"
	"sync/atomic"

	"github.com/dgraph-io/badger"
)

func bdb_set(key []byte, data []byte) error {
	if IsEmpty(key) || IsEmpty(data) || IsOversize(data) {
		return errors.New("bdb_set: entity key or data is invalid/oversize.")
	}

	data_zipped := Zip(data)
	err := bdb.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(key), data_zipped)
	})
	if err != nil {
		logger.Debug("failed to set key: ", string(key), ", Error: ", err)
		return err
	}
	atomic.AddUint64(&bdbSetCounter, 1)
	return nil
}

func bdb_get(key []byte) ([]byte, error) {
	if IsEmpty(key) {
		return nil, errors.New("bdb_get: entity key should not be empty.")
	}

	atomic.AddUint64(&bdbGetCounter, 1)
	var valCopy []byte
	err := bdb.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			return err
		}

		valCopy, err = item.ValueCopy(nil)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		logger.Debug("bdb_get: the key does not exist: ", string(key), " ,Error: ", err)
		return nil, err
	}
	return Unzip(valCopy), nil
}

func bdb_exists(key []byte) bool {
	if IsEmpty(key) {
		return false
	}

	err := bdb.View(func(txn *badger.Txn) error {
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

func bdb_delete(key []byte) error {
	if IsEmpty(key) {
		return errors.New("bdb_delete: entity key should not be empty.")
	}

	err := bdb.Update(func(txn *badger.Txn) error {
		return txn.Delete(key)
	})
	if err != nil {
		logger.Debug("failed to delete key: ", string(key), " ,Error: ", err)
		return err
	}
	return nil
}

func bdb_compact() {
	err := bdb.RunValueLogGC(0.7)
	if err != nil {
		logger.Debug("Error while RunValueLogGC: ", err)
	}
}

func bdb_scan() (keys []string) {
	err := bdb.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 100
		it := txn.NewIterator(opts)
		defer it.Close()
		klimit := 0
		for it.Rewind(); it.Valid(); it.Next() {
			if klimit > 100 {
				break
			}
			item := it.Item()
			k := item.Key()
			keys = append(keys, string(k))

			klimit++
		}
		return nil
	})
	if err != nil {
		return nil
	}
	return keys
}

func bdb_key_scan(prefix []byte) (keys []string) {
	err := bdb.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		opts.PrefetchSize = 100
		it := txn.NewIterator(opts)
		defer it.Close()
		klimit := 0

		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			if klimit > 100 {
				break
			}
			item := it.Item()
			k := item.Key()
			keys = append(keys, string(k))

			klimit++
		}

		return nil
	})
	if err != nil {
		return nil
	}
	return keys
}
