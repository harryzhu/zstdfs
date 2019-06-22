package potato

import (
	"github.com/couchbase/moss"
)

// Writer
func cw_set(k []byte, v []byte) error {
	batch, _ := CWRITER.NewBatch(0, 0)
	batch.Set(k, v)

	err := CWRITER.ExecuteBatch(batch, moss.WriteOptions{})
	batch.Close()
	if err != nil {
		return err
	}
	return nil
}

func cw_get(k []byte, v []byte) error {

	return nil
}

func cw_delete(k []byte, v []byte) error {

	return nil
}
