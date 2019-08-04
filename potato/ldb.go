package potato

import (
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
)

func ldb_set(key []byte, data []byte) error {
	err := ldb.Put(key, data, nil)
	if err != nil {
		logger.Error("LDB cannot set key-val")
		return err
	}
	return nil
}

func ldb_get(key []byte) (data []byte, err error) {
	data, err = ldb.Get(key, nil)
	if err != nil {
		logger.Error("LDB cannot get key-val", err)
		return nil, err
	}
	return data, nil
}

func ldb_del(key []byte) error {
	err := ldb.Delete(key, nil)
	if err != nil {
		logger.Error("LDB cannot delete key")
		return err
	}
	return nil
}

func ldb_mset(keys []string, val []byte) error {
	if len(keys) > 0 {
		batch := new(leveldb.Batch)
		for _, v := range keys {
			batch.Put([]byte(v), val)
		}
		err := ldb.Write(batch, nil)
		if err != nil {
			logger.Error("LDB cannot mset key-val")
			return err
		}
	}

	return nil
}

func ldb_mdel(keys []string) error {
	if len(keys) > 0 {
		batch := new(leveldb.Batch)
		for _, v := range keys {
			batch.Delete([]byte(v))
		}
		err := ldb.Write(batch, nil)
		if err != nil {
			logger.Error("LDB cannot mset key-val")
			return err
		}
	}

	return nil
}

func ldb_scan(prefix []byte) ([]string, error) {
	res := 0
	var keys []string
	iter := ldb.NewIterator(&util.Range{Start: prefix, Limit: nil}, nil)
	for iter.Next() {
		if res > 99 {
			break
		}
		keys = append(keys, string(iter.Key()))
		res++
	}
	iter.Release()
	err := iter.Error()
	if err != nil {
		logger.Error("ldb_scan: ", err)
		return nil, err
	}

	return keys, nil
}
