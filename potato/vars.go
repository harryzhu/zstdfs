package potato

import (
	"errors"

	"github.com/coocood/freecache"
	"github.com/dgraph-io/badger"
	log "github.com/sirupsen/logrus"
	"github.com/syndtr/goleveldb/leveldb"
)

var (
	logger    = log.New()
	cfg       Config
	bdb       *badger.DB
	ldb       *leveldb.DB
	cacheFree *freecache.Cache
)

// dynamic
var (
	isDebug               bool = true
	isBDBValueLogGCNeeded bool = true
	isBDBSyncWrites       bool = true
	isReplicationNeeded   bool = true
	isMaster              bool = true

	volumeSelf        string
	volumePeers       []string
	volumePeersLength int = 0
	volumePeersLive   map[string]bool
)

// Limits
var (
	grpcMAXMSGSIZE  int = 256 << 20
	entityMaxSize   int = 64 << 20
	cacheSize       int = 1024 << 20
	cacheExpiration int = 3600
)

var (
	bdbGetCounter   uint64 = 0
	bdbSetCounter   uint64 = 0
	cacheSetCounter uint64 = 0
)

var (
	maxCacheValueLen int = 0
)

// Error
var (
	ErrInGeneral   error = errors.New("error: in general.")
	ErrKeyNotFound error = errors.New("error: key does not found.")
	ErrKeyIsEmpty  error = errors.New("error: key can not be empty.")
	ErrDelFailed   error = errors.New("error: delete failed.")
)
