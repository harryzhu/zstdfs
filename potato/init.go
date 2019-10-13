package potato

import (
	"os"
	"strings"

	//"time"

	//"github.com/BurntSushi/toml"
	"github.com/coocood/freecache"
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

	smokeTest()

}

func openBDB() {
	if _, err := os.Stat(cfg.Volume.Db_data_dir); err != nil {
		logger.Fatal("cannot find the data dir: ", cfg.Volume.Db_data_dir)
	}

	if _, err := os.Stat(cfg.Volume.Db_value_dir); err != nil {
		logger.Fatal("cannot find the value dir: ", cfg.Volume.Db_value_dir)
	}

	opts := badger.DefaultOptions(cfg.Volume.Db_data_dir)
	opts.SyncWrites = isBDBSyncWrites
	opts.Truncate = true
	opts.MaxTableSize = 256 << 20  // 64MB
	opts.LevelOneSize = 1024 << 20 // 256MB
	opts.Dir = cfg.Volume.Db_data_dir
	opts.ValueDir = cfg.Volume.Db_value_dir

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

	if cfg.Volume.Is_master != false {
		isMaster = true
	}
	logger.Info("Volume is Master: ", isMaster)

	cv_max_size_mb := cfg.Volume.Max_size_mb
	if cv_max_size_mb > 0 {
		entityMaxSize = cv_max_size_mb * 1024 * 1024
	}
	logger.Info("Limits: entity Max Size: ", entityMaxSize)

	if grpcMAXMSGSIZE < entityMaxSize*16 {
		grpcMAXMSGSIZE = entityMaxSize * 16
	}
	logger.Info("Limits: rpc Message Max Size: ", grpcMAXMSGSIZE)

	cv_cache_size_mb := cfg.Volume.Cache_size_mb
	cv_cache_max_entity_size_byte := cfg.Volume.Cache_max_entity_size_byte

	if cv_cache_size_mb <= 16 {
		cacheSize = 16 << 20
	} else {
		cacheSize = cv_cache_size_mb << 20
	}
	maxCacheValueLen = cacheSize/1024 - freecache.ENTRY_HDR_SIZE - 64 - 1

	if cv_cache_max_entity_size_byte <= maxCacheValueLen {
		logger.Info("max cache value size: HARD LIMIT(1/1024 of cache size - 81): ", maxCacheValueLen)
		logger.Info("Config setting: volume.cache_max_entity_size_byte: ", cv_cache_max_entity_size_byte, " <= ", maxCacheValueLen)
		maxCacheValueLen = cv_cache_max_entity_size_byte
	} else {
		logger.Warn("Config setting: volume.cache_max_entity_size_byte should be <= ", maxCacheValueLen)
	}

	// base on the manual test, when cache_size_mb = 1024, if maxCacheValueLen > (cacheSize/1024 - freecache.ENTRY_HDR_SIZE - KEY_LENGTH) - 1,
	// freecache consider it as LargeEntry and will not cache it.
	logger.Info("Limits: cache Size: ", cacheSize, ", ENTRY_HDR_SIZE: ", freecache.ENTRY_HDR_SIZE)
	logger.Info("Limits: max Cache Value Size: ", maxCacheValueLen, " bytes(or ", maxCacheValueLen/1024, " kb).")

	cacheFree = freecache.NewCache(cacheSize)

	if cfg.Volume.Ip == "" || cfg.Volume.Port == "" {
		logger.Fatal("Volume IP/Port cannot be empty.")
	} else {
		volumeSelf = strings.Join([]string{cfg.Volume.Ip, cfg.Volume.Port}, ":")
		logger.Info("Volume RPC(Local): ", volumeSelf)
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

	if cfg.Volume.Db_syncwrites == false {
		isBDBSyncWrites = false
	}
	logger.Info("Volume isSyncWrites: ", isBDBSyncWrites)

}
