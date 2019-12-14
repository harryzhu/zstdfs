package potato

import (
	"errors"

	"github.com/dgraph-io/badger"
	log "github.com/sirupsen/logrus"
	"github.com/syndtr/goleveldb/leveldb"
)

var (
	logger = log.New()
	cfg    Config
	bdb    *badger.DB
	ldb    *leveldb.DB
)

// dynamic
var (
	isDebug      bool = true
	isSyncWrites bool = true
	profilePath  string

	volumeSelf        string
	volumePeers       []string
	volumePeersLength int = 0
	volumePeersLive   map[string]bool
)

// Limits
var (
	grpcMAXMSGSIZE int = 256 << 20
	entityMaxSize  int = 64 << 20
)

// Error
var (
	ErrInGeneral   error = errors.New("error: in general.")
	ErrKeyNotFound error = errors.New("error: key does not found.")
	ErrKeyIsEmpty  error = errors.New("error: key can not be empty.")
	ErrDelFailed   error = errors.New("error: delete failed.")
)
