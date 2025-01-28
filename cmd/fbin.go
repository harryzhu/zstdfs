package cmd

import (
	"io/ioutil"
	"path/filepath"

	badger "github.com/dgraph-io/badger/v4"
)

var fbin *badger.DB

func openBadger() {
	var err error
	opts := badger.DefaultOptions("data/fbin")
	opts.BaseTableSize = 16 << 20
	opts.NumVersionsToKeep = 1
	opts.SyncWrites = false
	opts.ValueThreshold = 256
	opts.CompactL0OnClose = true

	fbin, err = badger.Open(opts)
	FatalError("openBadger", err)
}

func fbinSave(val []byte) (key []byte) {
	if val == nil {
		DebugWarn("fbinSave", "val cannot be empty")
		return nil
	}
	fbinK := SumBlake3(val)

	err := fbin.Update(func(txn *badger.Txn) error {
		_, err := txn.Get(fbinK)
		if err == nil {
			return nil
		}
		err = txn.Set([]byte(fbinK), ZstdBytes(val))
		FatalError("fbinSave", err)
		return err
	})
	if err != nil {
		return nil
	}
	return fbinK
}

func fbinGet(key []byte) (val []byte) {
	if key == nil {
		DebugWarn("fbinGet", "key cannot be empty")
		return nil
	}
	fbin.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			return nil
		}
		itemVal, err := item.ValueCopy(nil)
		PrintError("fbinGet", err)
		val = UnZstdBytes(itemVal)
		return err
	})
	return val
}

func fbinBatchSave(files []string, dpath string) error {
	wb := fbin.NewWriteBatch()
	defer wb.Cancel()

	fbin.View(func(txn *badger.Txn) error {
		for _, fpath := range files {
			relPath, err := filepath.Rel(dpath, fpath)
			PrintError("fbinBatchSave:relPath", err)
			relPath = ToUnixSlash(relPath)

			if filepath.Base(relPath) != filepath.Base(fpath) {
				FatalError("fbinBatchSave:filepath.Base", ErrFileNameNotMatch)
			}

			fdata, err := ioutil.ReadFile(fpath)
			if err != nil {
				PrintError("fbinBatchSave:ReadFile", err)
				continue
			}
			if fdata == nil {
				DebugWarn("fbinBatchSave", "filesize is 0, SKIP")
				continue
			}
			// Fbin
			fbinK := SumBlake3(fdata)

			_, err = txn.Get(fbinK)
			if err != nil {
				wb.SetEntry(badger.NewEntry(fbinK, ZstdBytes(fdata)))
			}
		}
		return nil
	})

	if err := wb.Flush(); err != nil {
		FatalError("fbinBatchSave:wb.Flush", err)
	}
	return nil
}
