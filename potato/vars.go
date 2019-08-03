package potato

import (

	//"github.com/BurntSushi/toml"
	//"github.com/couchbase/moss"
	"github.com/dgraph-io/badger"
	"github.com/golang/groupcache"
	log "github.com/sirupsen/logrus"
	"github.com/syndtr/goleveldb/leveldb"
)

const (
	GRPCMAXMSGSIZE int    = 256 << 20
	CACHEGROUPSIZE int64  = 1024 << 20
	CACHEGROUPNAME string = "groupcache"
)

var (
	CFG                  Config
	Logger               = log.New()
	DB                   *badger.DB
	IsDBValueLogGCNeeded bool
	IsReplicationNeeded  bool = true
	HTTP_TEMP_DIR        string
	HTTP_SITE_URL        string
	IsMaster             bool
	SLAVES               []string
	SLAVES_LENGTH        int
	CACHE_GROUP          *groupcache.Group
	CACHE_PEERS          *groupcache.HTTPPool
	CACHE_BASEPATH       string = "/_groupcache/"
)

var (
	MODE            string = "PRODUCTION"
	ENTITY_MAX_SIZE int    = 32 << 20
	CACHE_MAX_SIZE  int    = 1 << 20
)

var (
	LDB *leveldb.DB
)

var (
	DBGetCounter uint64 = 0
	DBSetCounter uint64 = 0
)

var (
	VOLUME_SELF       map[string]string
	VOLUME_PEERS      []string
	VOLUME_PEERS_LIVE map[string]bool
)
