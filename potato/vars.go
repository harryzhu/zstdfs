package potato

import (
	//"github.com/BurntSushi/toml"
	"github.com/couchbase/moss"
	"github.com/dgraph-io/badger"
	log "github.com/sirupsen/logrus"
)

const (
	GRPCMAXMSGSIZE int = 1024 * 1024 * 16
)

var (
	CFG    Config
	Logger = log.New()
	DB     *badger.DB
	COLL   moss.Collection
)

func init() {
	loadConfigFromFile()
	openDatabase()
	openCacheCollection()
	smokeTest()
}
