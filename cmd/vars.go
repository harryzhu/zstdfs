package cmd

import (
	"database/sql"
	//"fmt"
	badger "github.com/dgraph-io/badger/v4"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

const KB int = 1 << 10
const MB int = 1 << 20
const GB int = 1 << 30
const TB int = 1 << 40

const ADMIN string = "_admin"
const AdminBucket string = ADMIN

const DiskCacheExpires float64 = 86400

var (
	Params     map[string]any
	sqldb      *sql.DB
	mgodb      *mongo.Database
	bgrdb      *badger.DB
	DATA_DIR   string
	TEMP_DIR   string
	CACHE_DIR  string
	ASSET_DIR  string
	STATIC_DIR string
	//
	FunctionCacheExpires int64 = 300
)
