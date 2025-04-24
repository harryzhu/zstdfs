package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	badger "github.com/dgraph-io/badger/v4"
)

func badgerConnect() *badger.DB {
	datadir := ToUnixSlash(filepath.Join(DataDir, "fbin"))
	MakeDirs(datadir)
	DebugInfo("badgerConnect", datadir)
	opts := badger.DefaultOptions(datadir)
	opts.Dir = datadir
	opts.ValueDir = datadir
	opts.BaseTableSize = 256 << 20
	opts.NumVersionsToKeep = 1
	opts.SyncWrites = false
	opts.ValueThreshold = 16
	opts.CompactL0OnClose = true

	db, err := badger.Open(opts)
	FatalError("badgerConnect", err)
	return db
}

func badgerSave(val []byte) (key []byte) {
	if val == nil {
		DebugWarn("badgerSave", "val cannot be empty")
		return nil
	}
	key = SumBlake3(val)

	err := bgrdb.Update(func(txn *badger.Txn) error {
		_, err := txn.Get(key)
		if err == nil {
			return nil
		}
		err = txn.Set([]byte(key), ZstdBytes(val))
		PrintError("badgerSave", err)
		return err
	})
	if err != nil {
		return nil
	}
	return key
}

func badgerGet(key []byte) (val []byte) {
	if key == nil {
		DebugWarn("badgerGet", "key cannot be empty")
		return nil
	}

	bgrdb.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			DebugWarn("badgerGet", err, ":", string(key))
			return nil
		}
		itemVal, err := item.ValueCopy(nil)
		PrintError("badgerGet", err)
		DebugInfo("badgerGet", len(itemVal), " :", string(key))
		val = UnZstdBytes(itemVal)
		return err
	})

	//DebugInfo("badgerGet:val", val)
	return val
}

func badgerList(uname string) {

	err := bgrdb.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			k := item.Key()
			err := item.Value(func(v []byte) error {
				DebugInfo("badgerList", fmt.Sprintf("%s", k), ": ", len(v))
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		PrintError("badgerList", err)
	}
}

func badgerExists(key []byte) bool {
	if key == nil {
		DebugWarn("badgerExists", "key cannot be empty")
		return false
	}

	err := bgrdb.View(func(txn *badger.Txn) error {
		_, err := txn.Get(key)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return false
	}

	return true
}

func badgerBulkLoad(dpath string, fext string) bool {
	DebugInfo("badgerBulkLoad:", "Start ...")

	tsStart := GetNowUnix()
	var counter int
	var batchFiles []string

	filepath.Walk(dpath, func(path string, finfo os.FileInfo, err error) error {
		PrintSpinner(Int2Str(counter))

		if finfo.IsDir() {
			return nil
		}
		if fext != "*" {
			if strings.ToLower(filepath.Ext(finfo.Name())) != strings.ToLower(fext) {
				return nil
			}
		}
		if strings.HasPrefix(finfo.Name(), ".") {
			return nil
		}
		if finfo.Size() == 0 || finfo.Size() > (MaxUploadSizeMB<<20) {
			return nil
		}
		//DebugInfo("badgerBulkLoad", path)
		if len(batchFiles) < 10 {
			batchFiles = append(batchFiles, path)
			counter++
		}
		if len(batchFiles) >= 10 {
			batchWriteFiles(batchFiles)
			mongoBatchWriteFiles(batchFiles)
			batchFiles = []string{}
		}

		return nil
	})

	batchWriteFiles(batchFiles)
	mongoBatchWriteFiles(batchFiles)

	DebugInfo("badgerBulkLoad:", "Done! files: ", counter)
	DebugInfo("badgerBulkLoad", fmt.Sprintf("Elapse: %v seconds", (GetNowUnix()-tsStart)))
	return true
}

func batchWriteFiles(files []string) bool {
	txn := bgrdb.NewTransaction(true)
	defer txn.Discard()
	var batchTotalSize int

	for _, file := range files {
		fp, err := os.Open(file)
		if err != nil {
			PrintError("BatchWriteFiles", err)
			return false
		}
		val, err := io.ReadAll(fp)
		if err != nil {
			PrintError("BatchWriteFiles", err)
			return false
		}
		fp.Close()

		err = txn.Set(SumBlake3(val), ZstdBytes(val))
		if err != nil {
			PrintError("BatchWriteFiles", err)
			return false
		}
		batchTotalSize += len(val)
	}
	if err := txn.Commit(); err != nil {
		FatalError("BatchWriteFiles", err)
		return false
	}
	DebugInfo("BatchWriteFiles: files: ", len(files), ", size:", batchTotalSize>>20, " MB")

	bgrdb.Sync()
	return true
}
