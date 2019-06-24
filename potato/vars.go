package potato

import (
	//"github.com/BurntSushi/toml"
	"github.com/couchbase/moss"
	"github.com/dgraph-io/badger"
	log "github.com/sirupsen/logrus"
)

const (
	GRPCMAXMSGSIZE int = 1024 * 1024 * 64
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
	SLAVES               []string
	SLAVES_LENGTH        int
)

var (
	MODE string = "PRODUCTION"
)

var (
	batchWriter moss.Batch
	batchReader moss.Batch
)
