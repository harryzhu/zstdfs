package potato

import (
	"errors"

	"github.com/couchbase/moss"
)

// Writer
func cm_set(key string, data []byte) error {
	batchMeta, err := CMETA.NewBatch(0, 0)
	if err != nil {
		Logger.Error("Cache CMETA cannot NewBatch: ", err)
		return err
	}
	err = batchMeta.Set([]byte(key), data)
	CMETA.ExecuteBatch(batchMeta, moss.WriteOptions{})
	if err != nil {
		return err
	}
	return nil
}

func cm_get(key string) ([]byte, error) {
	ssMeta, err := CMETA.Snapshot()
	if err == nil {
		defer ssMeta.Close()

		result, err := ssMeta.Get([]byte(key), moss.ReadOptions{})
		if err != nil {
			return nil, errors.New("failed to get key from cache")
		}
		return result, nil
	}
	return nil, errors.New("failed to get key from cache")
}

func cm_del(key string) error {
	batchMeta, err := CMETA.NewBatch(0, 0)
	if err != nil {
		Logger.Error("Cache CMETA cannot NewBatch: ", err)
		return err
	}
	err = batchMeta.Del([]byte(key))
	if err != nil {
		return err
	}
	err = CMETA.ExecuteBatch(batchMeta, moss.WriteOptions{})
	if err != nil {
		return err
	}
	return nil
}
