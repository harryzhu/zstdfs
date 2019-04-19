package cmd

import (
	"errors"
	// 	"os"
	// 	"strconv"
	"strings"
	// 	"time"
	// 	"github.com/boltdb/bolt"
	_ "github.com/mattn/go-sqlite3"
)

func VolumeMetaDataExists(key string, node string) bool {
	//Logger.Debug("MasterNodeFileExists")
	if len(key) != 32 || len(node) < 1 {
		return false
	}
	var b_return bool = false
	var k, n string
	stmt, err := DBMETA.Prepare("SELECT key,node FROM data WHERE key = ? and node = ? LIMIT 1")
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

func VolumeUpdateMetaData(key string, node string, field string, field_value string) error {
	Logger.Info(VolumeUpdateMetaData, key, node, field, field_value)
	if len(key) != 32 || len(node) < 1 || len(field) < 1 || len(field_value) < 1 {
		return errors.New("nodefile field value is invalid.")
	}

	if VolumeMetaDataExists(key, node) == true {
		var q_sql = ""

		switch field {
		case "inmaster":
			q_sql = strings.Join([]string{"UPDATE data SET inmaster = ? where key = ? and node = ?"}, "")
		case "synced":
			q_sql = strings.Join([]string{"UPDATE data SET synced = ? where key = ? and node = ?"}, "")
		}

		stmt, err := DBMETA.Prepare(q_sql)
		if err != nil {
			Logger.Error(err)
			return errors.New("DBMASTER Prepare failed.")
		}

		_, err = stmt.Exec(field_value, key, node)
		if err != nil {
			Logger.Error("VolumeUpdateMetaData: stmt_Exec: ", err)
			return errors.New("DBMASTER INSERT failed.")
		}
	}

	return nil
}
