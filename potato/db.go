package potato

import (
	"errors"
	"sync/atomic"

	"github.com/dgraph-io/badger"
)

func db_set(key string, data []byte) error {
	if len(data) > ENTITY_MAX_SIZE {
		return errors.New("set: entity is too large.")
	}

	if len(key) > 0 {
		data_zipped := Zip(data)
		err := DB.Update(func(txn *badger.Txn) error {
			return txn.Set([]byte(key), data_zipped)
		})
		if err != nil {
			Logger.Debug("failed to set key: ", key, ", Error: ", err)
			return err
		}
		atomic.AddUint64(&DBSetCounter, 1)
		return nil
	}

	return errors.New("set: key should not be empty.")
}

func db_get(key string) ([]byte, error) {
	atomic.AddUint64(&DBGetCounter, 1)
	var valCopy []byte
	err := DB.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
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
		Logger.Debug("failed to get key or the key does not exist: ", key, " ,Error: ", err)
		return nil, err
	}
	return Unzip(valCopy), nil
}

func db_delete(key string) error {
	if len(key) > 0 {
		err := DB.Update(func(txn *badger.Txn) error {
			return txn.Delete([]byte(key))
		})
		if err != nil {
			Logger.Debug("failed to delete key: ", key, " ,Error: ", err)
			return err
		}
		return nil
	}

	return errors.New("delete: key should not be empty.")
}

func db_compact() error {
	if IsDBValueLogGCNeeded == true {
		IsDBValueLogGCNeeded = false
		err := DB.RunValueLogGC(0.7)
		if err != nil {
			Logger.Debug("Error while RunValueLogGC: ", err)
		}
		IsDBValueLogGCNeeded = true
	}
	return nil
}

func db_scan() (keys []string) {
	err := DB.View(func(txn *badger.Txn) error {
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
			//Logger.Info("key=", k)
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
