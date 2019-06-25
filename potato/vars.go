package potato

import (
	//"github.com/BurntSushi/toml"
	"github.com/couchbase/moss"
	"github.com/dgraph-io/badger"
	log "github.com/sirupsen/logrus"
)

const (
	GRPCMAXMSGSIZE int = 256 << 20
)

var (
	CFG                  Config
	Logger               = log.New()
	DB                   *badger.DB
	IsDBValueLogGCNeeded bool
	IsReplicationNeeded  bool = true
	CMETA                moss.Collection
	CREADER              moss.Collection
	HTTP_TEMP_DIR        string
	HTTP_SITE_URL        string
	IsMaster             bool
	SLAVES               []string
	SLAVES_LENGTH        int
)

var (
	MODE            string = "PRODUCTION"
	ENTITY_MAX_SIZE int    = 32 << 20
	CACHE_MAX_SIZE  int    = 1 << 20
)

var (
	batchWriter moss.Batch
	batchReader moss.Batch
)
