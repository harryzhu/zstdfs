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
	CMETA                moss.Collection
	CREADER              moss.Collection
	HTTP_TEMP_DIR        string
	HTTP_SITE_URL        string
)

var (
	batchWriter moss.Batch
	batchReader moss.Batch
)

func init() {
	loadConfigFromFile()
	openDatabase()
	openMetaCollection()
	openCacheCollection()
	smokeTest()
	IsDBValueLogGCNeeded = true

}
