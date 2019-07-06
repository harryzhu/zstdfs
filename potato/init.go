package potato

import (
	"os"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/dgraph-io/badger"
	"github.com/golang/groupcache"
	log "github.com/sirupsen/logrus"
	"github.com/syndtr/goleveldb/leveldb"
)

func init() {
	loadConfigFromFile()

	openDatabase()
	smokeTest()
	getSlavesLength()
	IsDBValueLogGCNeeded = true
	setEntityMaxSize()
	setCacheMaxSize()
	setIsMaster()
	openMeta()
	openGroupCache()
}

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

func smokeTest() error {
	const TOTAL_STEPS string = "5"
	// cache writer:

	// DB
	err := DB.Update(func(txn *badger.Txn) error {
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

func openMeta() {
	var err error
	var meta_dir string
	_, err = os.Stat(CFG.Volume.Meta_dir)
	if os.IsNotExist(err) {
		Logger.Fatal("META DIR is not writeable: CFG.Volume.Meta_dir")
	} else {
		meta_dir = CFG.Volume.Meta_dir
		Logger.Debug("meta_dir:", meta_dir)
	}

	LDB, err = leveldb.OpenFile(CFG.Volume.Meta_dir, nil)
	if err != nil {
		Logger.Fatal("Error while opening LDB")
	}

}

func openGroupCache() {
	csf := CFG.Volume.Cache_self
	cps := CFG.Volume.Cache_peers
	if cps == nil || len(cps) < 1 {
		Logger.Fatal("Cache Peers Configuration Error.")
	}

	opts := groupcache.HTTPPoolOptions{BasePath: "/_groupcache/"}
	CACHE_PEERS = groupcache.NewHTTPPoolOpts(csf, &opts)

	cps_str := strings.Join(cps, ",")
	Logger.Debug("Cache Peers:", cps_str)
	CACHE_PEERS.Set(cps_str)

	CACHE_GROUP = groupcache.NewGroup(CACHEGROUPNAME, CACHEGROUPSIZE, groupcache.GetterFunc(
		func(ctx groupcache.Context, key string, dest groupcache.Sink) error {
			//Logger.Info("groupcache.NewGroup: ", csf)
			data, err := EntityGet(key)
			if err != nil {
				return err
			}

			dest.SetBytes(data)
			return nil
		}))

}
