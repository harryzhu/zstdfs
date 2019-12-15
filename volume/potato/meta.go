package potato

import (
	//"errors"
	"strconv"
	"strings"

	"github.com/syndtr/goleveldb/leveldb/util"
)

func MetaGet(key []byte) ([]byte, error) {
	if IsEmpty(key) {
		return nil, ErrKeyIsEmpty
	}
	data, err := ldb_get(key)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func MetaSet(key []byte, val []byte) error {
	if IsEmpty(key) || IsEmpty(val) {
		return ErrKeyIsEmpty
	}
	if err := ldb_set(key, val); err != nil {
		return err
	}
	return nil
}

func MetaDelete(key []byte) error {
	if err := ldb_del(key); err != nil {
		return err
	}
	return nil
}

func MetaMultiDelete(keys []string) error {
	if err := ldb_mdel(keys); err != nil {
		logger.Debug("Meta Multi-Delete: ", err)
		return err
	}
	return nil
}

func MetaScanExists(prefix []byte) bool {
	res := 0
	iter := ldb.NewIterator(&util.Range{Start: prefix, Limit: nil}, nil)
	for iter.Next() {
		if res > 0 {
			break
		}
		res++
	}
	iter.Release()
	err := iter.Error()
	if err != nil {
		logger.Error("MetaScanCount: ", err)
	}
	if res > 0 {
		return true
	}
	return false
}

func MetaScan(prefix []byte, size int) ([]string, error) {
	keys, err := ldb_scan(prefix, size)
	if err != nil {
		return nil, err
	}
	return keys, nil
}

func MetaScanList(prefix []byte, size int) (listHtml string) {
	slaves := cfg.Volume.Peers
	fileKeys := []string{}
	if len(slaves) > 0 {
		res := 0
		iter := ldb.NewIterator(&util.Range{Start: prefix, Limit: nil}, nil)
		for iter.Next() {
			if len(iter.Key()) > 0 {
				fileKeys = append(fileKeys, string(iter.Key()))
				res++
				if res >= size {
					break
				}
			}

		}
		iter.Release()
		err := iter.Error()
		if err != nil {
			logger.Error("MetaScanList: ", err)
		}

	}

	fileKeys_len := len(fileKeys)
	if fileKeys_len == 0 {
		logger.Debug("No Entities Replication Needed.")
		return ""
	}
	listHtml = ""
	for k, v := range fileKeys {
		listHtml = strings.Join([]string{v, " : ", strconv.Itoa(k), "<br/>", listHtml}, "")
	}

	return listHtml
}

func MetaList(size int) (listHtml string) {
	fileKeys := []string{}

	res := 0
	iter := ldb.NewIterator(nil, nil)
	for iter.Next() {
		if len(iter.Key()) > 0 {
			fileKeys = append(fileKeys, string(iter.Key()))
			res++
			if res >= size {
				break
			}
		}

	}
	iter.Release()
	err := iter.Error()
	if err != nil {
		logger.Error("MetaSyncList: ", err)
	}

	fileKeys_len := len(fileKeys)
	if fileKeys_len == 0 {
		logger.Debug("No Entities Replication Needed.")
		return ""
	}
	listHtml = ""
	for k, v := range fileKeys {
		listHtml = strings.Join([]string{v, " : ", strconv.Itoa(k), "<br/>", listHtml}, "")
	}

	return listHtml
}

func metaKeyJoin(cat, rpcaddress, key string) string {
	if len(cat) == 0 || len(rpcaddress) == 0 || len(key) == 0 {
		return ""
	}
	return strings.Join([]string{cat, rpcaddress, key}, "/")
}

func metaKeySplit(metakey string) []string {
	var arr_metakey []string
	if len(metakey) <= 0 {
		return arr_metakey
	}
	arr_mk := strings.Split(metakey, "/")
	if len(arr_mk) == 5 {
		return arr_mk
	}
	return arr_metakey
}
