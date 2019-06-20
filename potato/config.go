package potato

import (
	"os"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/couchbase/moss"
	"github.com/dgraph-io/badger"
	log "github.com/sirupsen/logrus"
)

func Echo() error {
	Logger.Info("Echo is OK.")
	return nil
}

func loadConfigFromFile() error {

	_, err := os.Stat("conf.toml")
	if err != nil {
		Logger.Fatal("cannot find the configuration file: conf.toml")
	}

	if _, err := toml.DecodeFile("conf.toml", &CFG); err != nil {
		Logger.Fatal(err)
	} else {
		Logger.Info("conf.toml", " was loaded.")
		Logger.Info(CFG.Welcome)
		switch strings.ToUpper(CFG.Global.Log_level) {
		case "DEBUG":
			Logger.SetLevel(log.DebugLevel)
		case "INFO":
			Logger.SetLevel(log.InfoLevel)
		case "WARN":
			Logger.SetLevel(log.WarnLevel)
		case "ERROR":
			Logger.SetLevel(log.ErrorLevel)
		default:
			Logger.SetLevel(log.ErrorLevel)
		}

		Logger.SetLevel(log.DebugLevel)
	}
	return nil
}

func openDatabase() error {

	if _, err := os.Stat(CFG.Volume.Db_data_dir); err != nil {
		Logger.Fatal("cannot find the data dir: ", CFG.Volume.Db_data_dir)
	}

	if _, err := os.Stat(CFG.Volume.Db_value_dir); err != nil {
		Logger.Fatal("cannot find the value dir: ", CFG.Volume.Db_value_dir)
	}
	opts := badger.DefaultOptions
	opts.Truncate = true
	opts.MaxTableSize = 64 << 20  // 64MB
	opts.LevelOneSize = 256 << 20 // 256MB
	opts.Dir = CFG.Volume.Db_data_dir
	opts.ValueDir = CFG.Volume.Db_value_dir

	var err error
	DB, err = badger.Open(opts)
	if err != nil {
		Logger.Fatal("db cannot open: ", err)
	}

	return nil
}

func openCacheCollection() error {
	var err error
	COLL, err = moss.NewCollection(moss.CollectionOptions{})
	if err != nil {
		Logger.Fatal("Cache collection cannot open: ", err)
	}
	COLL.Start()
	return nil
}

func smokeTest() error {
	// cache:
	batch, err := COLL.NewBatch(0, 0)
	defer batch.Close()

	batch.Set([]byte("smokeTest-Cache"), []byte("OK"))
	err = COLL.ExecuteBatch(batch, moss.WriteOptions{})

	ss, err := COLL.Snapshot()
	defer ss.Close()

	ropts := moss.ReadOptions{}
	valstc, err := ss.Get([]byte("smokeTest-Cache"), ropts)
	if err != nil {
		Logger.Fatal("smokeTest-Cache: ", err)
	} else {
		Logger.Info("smokeTest(1/2)-Cache: ", string(valstc))
	}

	// DB
	err = DB.Update(func(txn *badger.Txn) error {
		err := txn.Set([]byte("smokeTest-Database"), []byte("OK"))
		return err
	})

	if err != nil {
		Logger.Fatal("smokeTest-Database: Set: ", err)
	}

	var valCopy []byte
	err = DB.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("smokeTest-Database"))
		if err != nil {
			return err
		} else {
			err := item.Value(func(val []byte) error {
				valCopy = append([]byte{}, val...)
				return nil
			})
			if err != nil {
				return err
			}
			return nil
		}
	})

	if err != nil {
		Logger.Fatal("smokeTest-Database: Get: ", err)
	}

	Logger.Info("smokeTest(2/2)-Database: ", string(valCopy))

	return nil
}
