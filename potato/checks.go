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

func getRunMode() {
	if CFG.Global.Is_debug == true {
		MODE = "DEBUG"
	} else {
		MODE = "PRODUCTION"
	}
}

func getSlavesLength() {
	if len(CFG.Replication.Slaves) >= 0 {
		SLAVES = CFG.Replication.Slaves
	} else {
		Logger.Fatal("Relication Slaves should be like: [\"12.34.56.78:910\",\"1.2.3.4:5678\"]")
	}

	self_ip_port := strings.Join([]string{CFG.Volume.Ip, CFG.Volume.Port}, ":")

	if len(SLAVES) > 0 {
		for _, v := range SLAVES {
			if v == self_ip_port {
				Logger.Fatal("Relication Slaves can not include(myself): ", self_ip_port)
			}
		}
		SLAVES_LENGTH = len(SLAVES)
	} else {
		SLAVES_LENGTH = 0
	}
}

func setEntityMaxSize() {
	if CFG.Volume.Max_size > 0 {
		ENTITY_MAX_SIZE = CFG.Volume.Max_size
	}
}

func setCacheMaxSize() {
	if CFG.Volume.Max_cache_size > 0 {
		CACHE_MAX_SIZE = CFG.Volume.Max_cache_size
	}
}

func setIsMaster() {
	if CFG.Replication.Is_master == true {
		IsMaster = true
	} else {
		IsMaster = false
	}

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

func openMetaCollection() error {
	var err error
	var meta_dir string
	_, err = os.Stat(CFG.Volume.Meta_dir)
	if os.IsNotExist(err) {
		Logger.Fatal("META DIR is not writeable: CFG.Volume.Meta_dir")
	} else {
		meta_dir = CFG.Volume.Meta_dir
	}

	so := moss.DefaultStoreOptions
	so.CollectionOptions.MinMergePercentage = 0.5
	so.CompactionPercentage = 0.7
	so.CompactionSync = true
	so.KeepFiles = false
	spo := moss.StorePersistOptions{CompactionConcern: moss.CompactionAllow}
	var store *moss.Store
	store, CMETA, err = moss.OpenStoreCollection(meta_dir, so, spo)
	//CMETA, err = moss.NewCollection(moss.CollectionOptions{})
	if err != nil || store == nil || CMETA == nil {
		Logger.Fatal("Cache collection cannot open: ", err)
		return err
	}
	CMETA.Start()
	return nil
}

func openCacheCollection() error {
	var err error
	CREADER, err = moss.NewCollection(moss.CollectionOptions{})
	if err != nil {
		Logger.Fatal("Cache collection cannot open: ", err)
		return err
	}
	CREADER.Start()

	batchWriter, err = CMETA.NewBatch(0, 0)
	if err != nil {
		Logger.Fatal("Cache CMETA cannot NewBatch: ", err)
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
	batchWriter.Set([]byte("smokeTest-CacheMETA"), []byte("OK"))
	CMETA.ExecuteBatch(batchWriter, moss.WriteOptions{})

	ssWriter, err := CMETA.Snapshot()
	defer ssWriter.Close()

	ropts_writer := moss.ReadOptions{}
	valstc_writer, err := ssWriter.Get([]byte("smokeTest-CacheMETA"), ropts_writer)
	if err != nil {
		Logger.Fatal("smokeTest-CacheWriter: ", err)
	} else {
		Logger.Info("1/", TOTAL_STEPS, ", smokeTest-CacheMETA: ", string(valstc_writer))
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
		err := txn.Set([]byte("smokeTest-Database"), Zip([]byte("OK")))
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

	Logger.Info("3/", TOTAL_STEPS, ", smokeTest-Database: ", string(Unzip(valCopy)))

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

func testOptions() {
	opt_volume_max_size := WriteInt(256 << 20)
	opt_volume_max_cache_size := WriteInt(256 << 20)

	conf := NewOption(opt_volume_max_size, opt_volume_max_cache_size)
	Logger.Info(conf.i)
}
