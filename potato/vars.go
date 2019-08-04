package potato

import (
	"github.com/dgraph-io/badger"
	"github.com/golang/groupcache"
	log "github.com/sirupsen/logrus"
	"github.com/syndtr/goleveldb/leveldb"
)

var (
	logger         = log.New()
	cfg            Config
	bdb            *badger.DB
	ldb            *leveldb.DB
	cacheGroup     *groupcache.Group
	cacheGroupName string = "gcentity"
)

// dynamic
var (
	isBDBValueLogGCNeeded bool = true
	isMaster              bool = true
	volumeSelf            string
	volumePeers           []string
	volumePeersLength     int = 0
	volumePeersLive       map[string]bool
)

// Limits
var (
	grpcMAXMSGSIZE int   = 256 << 20
	entityMaxSize  int   = 64 << 20
	cacheSize      int64 = 1024 << 20
)

var (
	bdbGetCounter uint64 = 0
	bdbSetCounter uint64 = 0
)
