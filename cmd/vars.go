package cmd

import (
	"database/sql"
	"errors"

	badger "github.com/dgraph-io/badger/v4"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

const (
	KB int = 1 << 10
	MB int = 1 << 20
	GB int = 1 << 30
	TB int = 1 << 40
)

const (
	K int = 1000
	W int = 10000
	M int = 1000000
)

const (
	AdminBucket      string  = "_admin"
	DiskCacheExpires float64 = 3600
)

var (
	sqldb    *sql.DB
	mgodb    *mongo.Database
	bgrdb    *badger.DB
	DataDir  string
	TempDir  string
	CacheDir string
	AssetDir string
	//
	FunctionCacheExpires int64 = 600
	// Statistics
	minDiggCount     int = 100000
	minCommentCount  int = 20000
	minCollectCount  int = 10000
	minShareCount    int = 10000
	minDownloadCount int = 10000
)

var (
	testUser string = "harry"
	testKey  string = "sample_group/sample/prefix/test.jpg"
)

var (
	ErrEmptyMeta error = errors.New("meta is empty")
)
