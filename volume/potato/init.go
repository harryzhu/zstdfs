package potato

import (
	"os"
	"strings"

	//"time"

	//"github.com/BurntSushi/toml"
	"github.com/dgraph-io/badger"

	//log "github.com/sirupsen/logrus"
	"github.com/syndtr/goleveldb/leveldb"
)

func OnReady() error {
	return nil
}

func init() {
	loadConfigFromFile()
	useConfig()

	openBDB()
	openLDB()

	//smokeTest()
	//go func() { Bdb_Subscribe() }()
}

func openBDB() {
	if _, err := os.Stat(cfg.Volume.Data_dir); err != nil {
		logger.Fatal("cannot find the data dir: ", cfg.Volume.Data_dir)
	}

	opts := badger.DefaultOptions(cfg.Volume.Data_dir)
	opts.SyncWrites = isSyncWrites
	opts.Truncate = true
	opts.MaxTableSize = 256 << 20  // 64MB
	opts.LevelOneSize = 1024 << 20 // 256MB
	opts.Dir = cfg.Volume.Data_dir
	opts.ValueDir = cfg.Volume.Data_dir

	var err error
	bdb, err = badger.Open(opts)
	if err != nil {
		logger.Fatal("db cannot open: ", err)
	}
}

func openLDB() {
	if _, err := os.Stat(cfg.Volume.Meta_dir); err != nil {
		logger.Fatal("cannot find the meta dir: ", cfg.Volume.Meta_dir)
	}

	var err error
	ldb, err = leveldb.OpenFile(cfg.Volume.Meta_dir, nil)
	if err != nil {
		logger.Fatal("Error while opening ldb")
	}

}

func useConfig() {
	if cfg.Global.Is_debug == true {
		isDebug = true
	} else {
		isDebug = false
	}
	logger.Info("node is running in debug mode: ", isDebug)

	cv_max_size_mb := cfg.Volume.Max_size_mb
	if cv_max_size_mb > 0 {
		entityMaxSize = cv_max_size_mb * 1024 * 1024
	}
	logger.Info("Limits: entity Max Size: ", entityMaxSize)

	if grpcMAXMSGSIZE < entityMaxSize*16 {
		grpcMAXMSGSIZE = entityMaxSize * 16
	}
	logger.Info("Limits: rpc Message Max Size: ", grpcMAXMSGSIZE)

	if len(cfg.Volume.Self) > 0 {
		volumeSelf = cfg.Volume.Self
	} else {
		logger.Fatal("cfg.Volume.Self is invalid: ", cfg.Volume.Self)
	}

	if len(cfg.Volume.Peers) > 0 {
		cvp := cfg.Volume.Peers
		for _, addr := range cvp {
			if len(addr) > 0 && addr != volumeSelf {
				volumePeers = append(volumePeers, addr)
			}
		}
		volumePeersLength = len(volumePeers)
	}
	logger.Info("Volume Peers: ", strings.Join(volumePeers, "; "))
	logger.Info("Volume Peers Length: ", volumePeersLength)

	volumePeersLive = make(map[string]bool, volumePeersLength)
	for _, vp := range volumePeers {
		volumePeersLive[vp] = false
	}

	if cfg.Volume.Is_syncwrites == false {
		isSyncWrites = false
	}
	logger.Info("Volume isSyncWrites: ", isSyncWrites)

}
