package potato

import (
	//"errors"
	"strconv"
	"strings"

	//"github.com/couchbase/moss"
	"github.com/syndtr/goleveldb/leveldb/util"
)

func MetaGet(prefix string, key string) ([]byte, error) {
	metakey := strings.Join([]string{prefix, key}, ":")
	Logger.Debug("Meta Get:", metakey)
	data, err := ldb_get(metakey)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func MetaSet(prefix string, key string, data []byte) error {
	metakey := strings.Join([]string{prefix, key}, ":")
	Logger.Debug("Meta Set:", metakey)
	if err := ldb_set(metakey, data); err != nil {
		return err
	}
	return nil
}

func MetaDelete(prefix string, key string) error {
	metakey := strings.Join([]string{prefix, key}, ":")
	Logger.Debug("Meta Delete:", metakey)
	if err := ldb_del(metakey); err != nil {
		return err
	}
	return nil
}

func MetaMultiDelete(keys []string) error {
	Logger.Debug("Meta Multi-Delete: ", len(keys))
	if err := ldb_mdel(keys); err != nil {
		return err
	}
	return nil
}

func MetaSyncCount() (res int) {
	res = 0
	iter := LDB.NewIterator(&util.Range{Start: []byte("sync:"), Limit: nil}, nil)
	for iter.Next() {
		res++
		if res > 0 {
			break
		}
	}
	iter.Release()
	err := iter.Error()
	if err != nil {
		Logger.Error("MetaSyncCount: ", err)
		return -1
	}
	return res
}

func MetaScan(prefix string) ([]string, error) {
	keys, err := ldb_scan(prefix)
	if err != nil {
		return nil, err
	}
	return keys, nil
}

func MetaSyncList() (listHtml string) {
	slaves := CFG.Replication.Slaves
	fileKeys := []string{}
	if len(slaves) > 0 {
		res := 0
		iter := LDB.NewIterator(&util.Range{Start: []byte("sync:"), Limit: nil}, nil)
		for iter.Next() {
			if len(iter.Key()) > 0 {
				fileKeys = append(fileKeys, string(iter.Key()))
				res++
				if res > 1000 {
					break
				}
			}

		}
		iter.Release()
		err := iter.Error()
		if err != nil {
			Logger.Error("MetaSyncList: ", err)
		}

	}

	fileKeys_len := len(fileKeys)
	if fileKeys_len == 0 {
		Logger.Debug("No Entities Replication Needed.")
		return ""
	}
	listHtml = ""
	for k, v := range fileKeys {
		listHtml = strings.Join([]string{v, " : ", strconv.Itoa(k), "<br/>", listHtml}, "")
	}

	return listHtml
}

func MetaList() (listHtml string) {
	fileKeys := []string{}

	res := 0
	iter := LDB.NewIterator(nil, nil)
	for iter.Next() {
		if len(iter.Key()) > 0 {
			fileKeys = append(fileKeys, string(iter.Key()))
			res++
			if res > 100 {
				break
			}
		}

	}
	iter.Release()
	err := iter.Error()
	if err != nil {
		Logger.Error("MetaSyncList: ", err)
	}

	fileKeys_len := len(fileKeys)
	if fileKeys_len == 0 {
		Logger.Debug("No Entities Replication Needed.")
		return ""
	}
	listHtml = ""
	for k, v := range fileKeys {
		listHtml = strings.Join([]string{v, " : ", strconv.Itoa(k), "<br/>", listHtml}, "")
	}

	return listHtml
}
