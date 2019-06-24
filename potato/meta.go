package potato

import (
	//"errors"
	"strings"
	//"github.com/couchbase/moss"
)

func MetaGet(prefix string, key string) ([]byte, error) {
	metakey := strings.Join([]string{prefix, key}, ":")
	Logger.Debug("Meta Get:", metakey)
	data, err := cm_get(metakey)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func MetaSet(prefix string, key string, data []byte) error {
	metakey := strings.Join([]string{prefix, key}, ":")
	Logger.Debug("Meta Set:", metakey)
	if err := cm_set(metakey, data); err != nil {
		return err
	}
	return nil
}

func MetaDelete(prefix string, key string) error {
	metakey := strings.Join([]string{prefix, key}, ":")
	Logger.Debug("Meta Delete:", metakey)
	if err := cm_del(metakey); err != nil {
		return err
	}
	return nil
}
