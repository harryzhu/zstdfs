package potato

import (
	// 	//"github.com/BurntSushi/toml"
	"github.com/syndtr/goleveldb/leveldb"
)

func ldb_set(key string, data []byte) error {
	err := LDB.Put([]byte(key), data, nil)
	if err != nil {
		Logger.Error("LDB cannot set key-val")
		return err
	}
	return nil
}

func ldb_get(key string) (data []byte, err error) {
	data, err = LDB.Get([]byte(key), nil)
	if err != nil {
		Logger.Error("LDB cannot get key-val", err)
		return nil, err
	}
	return data, nil
}

func ldb_del(key string) error {
	err := LDB.Delete([]byte(key), nil)
	if err != nil {
		Logger.Error("LDB cannot delete key")
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
		err := LDB.Write(batch, nil)
		if err != nil {
			Logger.Error("LDB cannot mset key-val")
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
		err := LDB.Write(batch, nil)
		if err != nil {
			Logger.Error("LDB cannot mset key-val")
			return err
		}
	}

	return nil
}
