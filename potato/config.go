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
	CWRITER, err = moss.NewCollection(moss.CollectionOptions{})
	if err != nil {
		Logger.Fatal("Cache collection cannot open: ", err)
		return err
	}
	CWRITER.Start()

	CREADER, err = moss.NewCollection(moss.CollectionOptions{})
	if err != nil {
		Logger.Fatal("Cache collection cannot open: ", err)
		return err
	}
	CREADER.Start()

	batchWriter, err = CWRITER.NewBatch(0, 0)
	if err != nil {
		Logger.Fatal("Cache CWRITER cannot NewBatch: ", err)
	}

	batchReader, err = CREADER.NewBatch(0, 0)
	if err != nil {
		Logger.Fatal("Cache CREADER cannot NewBatch: ", err)
	}
	return nil
}

func smokeTest() error {
	const TOTAL_STEPS string = "5"
	// cache writer:
	batchWriter.Set([]byte("smokeTest-CacheWriter"), []byte("OK"))
	CWRITER.ExecuteBatch(batchWriter, moss.WriteOptions{})

	ssWriter, err := CWRITER.Snapshot()
	defer ssWriter.Close()

	ropts_writer := moss.ReadOptions{}
	valstc_writer, err := ssWriter.Get([]byte("smokeTest-CacheWriter"), ropts_writer)
	if err != nil {
		Logger.Fatal("smokeTest-CacheWriter: ", err)
	} else {
		Logger.Info("1/", TOTAL_STEPS, ", smokeTest-CacheWriter: ", string(valstc_writer))
	}

	// cache reader:
	batchReader.Set([]byte("smokeTest-CacheReader"), []byte("OK"))
	CREADER.ExecuteBatch(batchReader, moss.WriteOptions{})

	ssReader, err := CREADER.Snapshot()
	defer ssReader.Close()

	ropts_reader := moss.ReadOptions{}
	valstc_reader, err := ssReader.Get([]byte("smokeTest-CacheReader"), ropts_reader)
	if err != nil {
		Logger.Fatal("smokeTest-CacheReader: ", err)
	} else {
		Logger.Info("2/", TOTAL_STEPS, ", smokeTest-CacheReader: ", string(valstc_reader))
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

	Logger.Info("3/", TOTAL_STEPS, ", smokeTest-Database: ", string(valCopy))

	_, err = os.Stat(CFG.Http.Temp_dir)
	if os.IsNotExist(err) {
		Logger.Fatal("4/", TOTAL_STEPS, ", smokeTest-HTTP TEMP DIR is not writeable: CFG.Http.Temp_dir")
	} else {
		HTTP_TEMP_DIR = CFG.Http.Temp_dir
		Logger.Info("4/", TOTAL_STEPS, ", smokeTest-HTTP_TEMP_DIR: ", HTTP_TEMP_DIR)
	}

	if len(CFG.Http.Site_url) > 0 {
		HTTP_SITE_URL = CFG.Http.Site_url
		Logger.Info("5/", TOTAL_STEPS, ", smokeTest-HTTP_SITE_URL: ", HTTP_SITE_URL)
	} else {
		Logger.Fatal("5/", TOTAL_STEPS, ", smokeTest-HTTP_SITE_URL: CFG.Http.Site_url ")
	}

	return nil
}
