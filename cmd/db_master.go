package cmd

import (
	"errors"
	// 	"os"
	// 	"strconv"
	// 	"strings"
	// 	"time"
	// 	"github.com/boltdb/bolt"
	_ "github.com/mattn/go-sqlite3"
)

type NodeFiles struct {
	Key      string
	Node     string
	Size     uint32
	Synced   uint32
	InMaster uint32
	Created  uint32
}

func MasterNodeFileExists(key string, node string) bool {
	//Logger.Debug("MasterNodeFileExists")
	if len(key) != 32 || len(node) < 1 {
		return false
	}
	var b_return bool = false
	var k, n string
	stmt, err := DBMASTER.Prepare("SELECT key,node FROM nodefiles WHERE key = ? and node = ? LIMIT 1")
	if err != nil {
		Logger.Error(err)
	}

	err = stmt.QueryRow(key, node).Scan(&k, &n)

	if err != nil {
		return false
	}
	if len(k) > 0 && len(n) > 0 {
		b_return = true
	}
	return b_return
}

func MasterCreateNodeFiles(key string, node string, size uint32, synced uint32, inmaster uint32, created uint32) error {
	if len(key) != 32 || len(node) < 1 || size < 1 || created < 1 {
		return errors.New("nodefile field value is invalid.")
	}
	if MasterNodeFileExists(key, node) == false {
		stmt, err := DBMASTER.Prepare("INSERT INTO nodefiles(key,node,size,synced,inmaster,created) values(?,?,?,?,?,?)")
		if err != nil {
			Logger.Error(err)
			return errors.New("DBMASTER Prepare failed.")
		}

		_, err = stmt.Exec(key, node, size, synced, inmaster, created)
		if err != nil {
			Logger.Error(err)
			return errors.New("DBMASTER INSERT failed.")
		}
	}

	return nil
}
