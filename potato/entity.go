package potato

import (
	//"context"
	"strings"
	//pb "github.com/dgraph-io/badger/pb"
)

type Entity struct {
	Key  string
	Data []byte
}

func EntitySet(key string, data []byte) error {
	err := db_set(key, data)
	if err != nil {
		return err
	}
	if SLAVES_LENGTH > 0 {
		for _, slave := range SLAVES {
			prefix := strings.Join([]string{"sync", slave}, ":")
			MetaSet(prefix, key, []byte("0"))
		}
	}

	return nil
}

func EntityGet(key string) ([]byte, error) {
	v, err := db_get(key)
	if err != nil {
		return nil, err
	}
	return v, nil
}

func EntityDelete(key string) error {
	err := db_delete(key)
	if err != nil {
		return err
	}
	return nil
}

func EntityExists(key string) bool {
	_, err := db_get(key)
	if err != nil {
		return false
	}
	return true
}

func EntityCompaction() error {
	Logger.Debug("DB Compaction is starting...")
	db_compact()
	return nil
}
