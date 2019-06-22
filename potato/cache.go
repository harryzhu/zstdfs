package potato

import (
	"errors"

	"github.com/couchbase/moss"
)

func CacheGet(key string) ([]byte, error) {
	ssReader, err := CREADER.Snapshot()
	if err == nil {
		defer ssReader.Close()

		result, err := ssReader.Get([]byte(key), moss.ReadOptions{})
		if err != nil {
			return nil, errors.New("failed to get key from cache")
		}
		return result, nil
	}
	return nil, errors.New("failed to get key from cache")
}

func CacheSet(key string, data []byte) error {
	batchReader, err := CREADER.NewBatch(0, 0)
	if err != nil {
		Logger.Error("Cache CREADER cannot NewBatch: ", err)
		return err
	}
	err = batchReader.Set([]byte(key), data)
	CREADER.ExecuteBatch(batchReader, moss.WriteOptions{})
	if err != nil {
		return err
	}
	return nil
}
