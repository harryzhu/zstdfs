package cmd

import (
	"database/sql"

	"github.com/boltdb/bolt"
	_ "github.com/mattn/go-sqlite3"
	"github.com/nsqio/go-nsq"
	log "github.com/sirupsen/logrus"
)

var (
	CFG    Config
	Logger = log.New()
)

var (
	DBMASTER *sql.DB
	DBMETA   *sql.DB
	DBDATA   map[string]*bolt.DB
	DBHTTP   *bolt.DB
)

//grpc
const (
	GRPCMAXMSGSIZE int = 1024 * 1024 * 64
)

//NSQ
var (
	NSQP *nsq.Producer
	NSQC *nsq.Consumer
)
