package entity

import (
	//"encoding/json"

	//"fmt"
	"log"
	"strconv"

	"github.com/dgraph-io/badger/v3"
	"github.com/harryzhu/potatofs/util"
)

type Entity struct {
	Key  []byte
	Meta []byte
	Data []byte
}

func NewEntity(key, meta, data []byte) Entity {
	return Entity{
		Key:  key,
		Meta: meta,
		Data: data,
	}

}

func NewEntityByFile(fpath string, key string, tags string, comment string) (Entity, error) {
	var ett Entity
	if key == "" {
		return ett, util.ErrorString("key cannot be empty")
	}
	ett.Key = []byte(key)

	finfo, fdata, err := util.GetFileInfoByte(fpath)
	if err != nil {
		return ett, err
	}

	w, h := util.GetImageWidthHeight(fpath)

	meta := EntityMeta{}
	meta = meta.WithName(finfo.Name()).WithMime(util.GetMimeByName(finfo.Name())).WithSize(finfo.Size())
	meta = meta.WithWidth(w).WithHeight(h)
	meta = meta.WithComment(comment).WithTags(tags).WithRef(util.MD5(fdata))

	jsonMeta, err := meta.Marshal()
	if err != nil {
		log.Println(err)
		return ett, err
	}

	ett.Data = fdata
	ett.Meta = jsonMeta

	return ett, nil
}

func (ett Entity) Save() error {
	if ett.Key == nil {
		return util.ErrorString("key cannot be empty")
	}
	var errMeta, errData error

	if ett.Data != nil {
		k := []byte(util.MD5(ett.Data))
		errData = setData(k, ett.Data)
	}

	if errData != nil {
		return errData
	}

	errMeta = setMeta(ett.Key, ett.Meta)

	if errMeta != nil {
		return errMeta
	}
	return nil
}

func (ett Entity) Get() (Entity, error) {
	ettQuery := Entity{}
	if ett.Key == nil {
		return ettQuery, util.ErrorString("key cannot be empty")
	}

	ettQuery.Key = ett.Key

	meta, err := getMeta(ett.Key)
	if err != nil {
		return ettQuery, err
	}

	ettQuery.Meta, err = meta.Marshal()
	if err != nil {
		return ettQuery, err
	}

	data, err := getData([]byte(meta.Ref))
	if err != nil {
		ettQuery.Data = nil
		return ettQuery, err
	}

	ettQuery.Data = data

	return ettQuery, nil
}

func GetMetaKeyList(page int) []string {
	if page < 1 {
		page = 1
	}

	pageSize := 1000
	offset := (page - 1) * pageSize
	idx := 0

	var keys []string
	err := MetaDB.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			if idx < offset {
				idx++
				continue
			}
			if idx >= offset+pageSize {
				break
			}
			item := it.Item()
			k := item.Key()
			keys = append(keys, string(k))
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
	return keys
}

func GetMetaKeyListByPrefix(prefix []byte, pageStr string) []string {
	page, err := strconv.Atoi(pageStr)
	if err != nil {
		page = 1
	}

	if page < 1 {
		page = 1
	}
	pageSize := 100
	offset := (page - 1) * pageSize
	idx := 0
	var keys []string
	err = MetaDB.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			if idx < offset {
				idx++
				continue
			}
			if idx >= offset+pageSize {
				break
			}
			item := it.Item()
			k := item.Key()
			keys = append(keys, string(k))

			idx++
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
	return keys
}

func DeleteMetaKey(k []byte) error {
	if k == nil {
		return nil
	}

	return delMetaKey(k)
}

func WalkMetaKeyListToBoltMetaBucket() error {
	var boltMetaDB []byte = []byte("meta")
	MetaDB.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			k := item.Key()
			BoltSave(boltMetaDB, k, nil)
		}
		return nil
	})

	return nil
}
