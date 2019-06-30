package potato

import (
	"errors"

	"github.com/couchbase/moss"
)

// Writer
func cm_set(key string, data []byte) error {
	totalOps := 32
	totalKeyValBytes := 32 << 20
	batchMeta, err := CMETA.NewBatch(totalOps, totalKeyValBytes)
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

	if err != nil || ssMeta == nil {
		return nil, errors.New("failed to get key from cache")
	}

	defer ssMeta.Close()
	result, err := ssMeta.Get([]byte(key), moss.ReadOptions{})
	if err != nil {
		return nil, errors.New("failed to get key from cache")
	}

	if result == nil {
		return nil, errors.New("failed to get key from cache")
	}

	return result, nil
}

func cm_kget(key string) ([]byte, error) {
	ssMeta, err := CMETA.Snapshot()

	if err != nil || ssMeta == nil {
		return nil, errors.New("failed to get key from cache")
	}

	defer ssMeta.Close()
	result, err := ssMeta.Get([]byte(key), moss.ReadOptions{NoCopyValue: true})
	if err != nil {
		return nil, errors.New("failed to get key from cache")
	}

	if result == nil {
		return nil, errors.New("failed to get key from cache")
	}

	return result, nil
}

func cm_del(key string) error {
	batchMeta, err := CMETA.NewBatch(0, 0)
	if err != nil {
		Logger.Error("Cache CMETA cannot NewBatch: ", err)
		return err
	}
	err = batchMeta.Del([]byte(key))
	if err != nil {
		Logger.Error("Cache CMETA cannot Delete: ", key, ", ", err)
		return err
	}
	err = CMETA.ExecuteBatch(batchMeta, moss.WriteOptions{})
	if err != nil {
		return err
	}
	return nil
}

func cm_mdel(keys []string) error {
	batchMeta, err := CMETA.NewBatch(0, 0)
	if err != nil {
		Logger.Error("Cache CMETA cannot NewBatch: ", err)
		return err
	}
	for _, key := range keys {
		if key != "" {
			err = batchMeta.Del([]byte(key))
			if err != nil {
				Logger.Error("Cache CMETA cannot Delete: ", key, ", ", err)
				return err
			}
		}
	}

	err = CMETA.ExecuteBatch(batchMeta, moss.WriteOptions{})
	if err != nil {
		return err
	}
	return nil
}
