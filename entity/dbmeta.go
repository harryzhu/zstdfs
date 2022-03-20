package entity

import (
	"encoding/json"
	"log"

	"github.com/dgraph-io/badger/v3"
	"github.com/harryzhu/potatofs/util"
)

var (
	MetaDB *badger.DB
)

func OpenMetaDB(metadir string) {
	opts := badger.DefaultOptions(metadir)
	opts.SyncWrites = true

	var err error
	MetaDB, err = badger.Open(opts)
	if err != nil {
		log.Fatal("OpenMetaDB: ", err)
	}
}

func setMeta(key, meta []byte) error {
	if key == nil {
		return util.ErrorString("key cannot be empty")
	}

	if len(meta) > MaxSizeByte {
		return util.ErrorString("entity is oversize")
	}

	return MetaDB.Update(func(txn *badger.Txn) error {
		err := txn.Set(key, meta)
		if err == nil {

			BoltSave([]byte("meta"), key, nil)
		}
		return err
	})

}

func getMeta(k []byte) (em EntityMeta, err error) {
	if k == nil {
		return em, util.ErrorString("key cannot be empty")
	}
	var meta []byte

	MetaDB.View(func(txn *badger.Txn) error {
		item, err := txn.Get(k)
		if err != nil {
			return err
		}
		meta, err = item.ValueCopy(nil)
		if err != nil {
			return err
		}
		return nil
	})

	if meta != nil {
		err := json.Unmarshal(meta, &em)
		if err != nil {
			return em, err
		}
	}
	return em, nil
}

func delMetaKey(key []byte) error {
	if key == nil {
		return util.ErrorString("key cannot be empty")
	}

	return MetaDB.Update(func(txn *badger.Txn) error {
		err := txn.Delete(key)
		if err == nil {
			BoltSave([]byte("delete"), key, nil)
		}
		return err
	})

}
