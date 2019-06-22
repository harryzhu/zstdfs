package potato

import (
	"github.com/couchbase/moss"
)

// Reader
func cr_set(k []byte, v []byte) error {
	batch, _ := CREADER.NewBatch(0, 0)
	batch.Set(k, v)

	err := CREADER.ExecuteBatch(batch, moss.WriteOptions{})
	batch.Close()
	if err != nil {
		return err
	}
	return nil
}

func cr_get(k []byte, v []byte) error {

	return nil
}

func cr_delete(k []byte, v []byte) error {

	return nil
}
