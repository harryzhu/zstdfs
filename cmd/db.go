// Copyright © 2018 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	//"errors"
	//"strconv"
	"database/sql"
	"os"
	"strings"
	"time"

	"github.com/boltdb/bolt"
	_ "github.com/mattn/go-sqlite3"
)

func PrepareVolumeDatabases() error {
	arr_dbindex := []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "a", "b", "c", "d", "e", "f"}
	arr_dbdata := CFG.Volume.Dirs_data
	if len(arr_dbdata) != 16 {
		Logger.Fatal("CFG.Volume.Dirs_data should be 16.")
	} else {
		Logger.Info("preparing databases : open data(0-F) database. Start...")

		DBDATA = make(map[string]*bolt.DB)
		var errIsDBOpened error
		for i, _ := range arr_dbdata {
			dbpath := strings.Join([]string{arr_dbdata[i], "/data", arr_dbindex[i], ".db"}, "")
			Logger.Debug(dbpath)
			DBDATA[arr_dbindex[i]], errIsDBOpened = bolt.Open(dbpath, 0600, &bolt.Options{Timeout: 3 * time.Second})
			if errIsDBOpened != nil {
				Logger.Fatal("ERROR while opening:", dbpath)
			}
		}
		Logger.Info("preparing databases : open data(0-F) database. Done.")

	}

	dir_dbmeta := CFG.Volume.Dir_meta
	_, err := os.Stat(dir_dbmeta)
	if err != nil {
		Logger.Fatal("ERROR while opening META database:", dir_dbmeta)
	} else {
		metadbpath := strings.Join([]string{dir_dbmeta, "/meta", ".db"}, "")
		Logger.Info("preparing META database : open ", metadbpath)
		// DBMETA, err := sql.Open("sqlite3", metadbpath)
		// defer DBMETA.Close()
		// if err != nil {
		// 	Logger.Fatal("cannot open META database:", err)
		// }

		DBMETA = GetSqlite3(metadbpath)
		sql_table_data := `
		CREATE TABLE IF NOT EXISTS "data"(
		"key" CHARACTER(32) NOT NULL primary key,
		"size" integer NOT NULL,
		"dbid" CHARACTER(2) NOT NULL default "-",
		"node" CHARACTER(64) NOT NULL default "-",
		"synced" integer NOT NULL default 0,
		"enabled" integer NOT NULL default 1,
		"deleted" integer NOT NULL default 0,
		"created" integer NOT NULL default 0
		);
		create index if not exists idxkey on data(key);
		`

		Logger.Info(sql_table_data)
		DBMETA.Exec(sql_table_data)

		var size int
		var key string
		err = DBMETA.QueryRow("SELECT key,size FROM data where key=?", "db-meta-is-ok").Scan(&key, &size)

		if size > 0 {
			Logger.Info("key: ", key, ", size: ", size)
		} else {
			stmt, err := DBMETA.Prepare("INSERT INTO data(key,size) values(?,?)")
			if err != nil {
				Logger.Fatal(err)
			}
			result, err := stmt.Exec("db-meta-is-ok", 1024)
			if err != nil {
				Logger.Fatal(err)
			} else {
				lastID, err := result.LastInsertId()
				if err != nil {
					Logger.Fatal(err)
				} else {
					Logger.Info("Last Insert ID: ", lastID)
				}
			}
		}

	}
	return nil
}

func PrepareMasterDatabase() error {
	dir_dbmeta := CFG.Master.Dir_meta
	_, err := os.Stat(dir_dbmeta)
	if err != nil {
		Logger.Fatal("ERROR while opening Master database:", dir_dbmeta)
	} else {
		metadbpath := strings.Join([]string{dir_dbmeta, "/master", ".db"}, "")
		Logger.Info("preparing Master database : open ", metadbpath)
		DBMASTER = GetSqlite3(metadbpath)
		if err != nil {
			Logger.Fatal("cannot open Master database:", err)
		}

		sql_table_data := `
		PRAGMA default_cache_size = 8000;
		CREATE TABLE IF NOT EXISTS "data"(
		"key" CHARACTER(32) NOT NULL primary key,
		"size" integer NOT NULL,
		"dbid" CHARACTER(2) NOT NULL default "-",
		"node" CHARACTER(64) NOT NULL default "-",
		"synced" integer NOT NULL default 0,
		"enabled" integer NOT NULL default 1,
		"deleted" integer NOT NULL default 0,
		"created" integer NOT NULL default 0
		);
		create index if not exists idxkey on data(key);
		`

		//Logger.Info(sql_table_data)
		DBMASTER.Exec(sql_table_data)

		var size int
		var key string
		err = DBMASTER.QueryRow("SELECT key,size FROM data where key=?", "db-master-is-ok").Scan(&key, &size)

		if size > 0 {
			Logger.Info("key: ", key, ", size: ", size)
		} else {
			stmt, err := DBMASTER.Prepare("INSERT INTO data(key,size,name) values(?,?,?)")
			if err != nil {
				Logger.Fatal(err)
			}
			result, err := stmt.Exec("db-master-is-ok", 1024, "db-master-is-ok")
			if err != nil {
				Logger.Fatal(err)
			} else {
				lastID, err := result.LastInsertId()
				if err != nil {
					Logger.Fatal(err)
				} else {
					Logger.Info("Last Insert ID: ", lastID)
				}
			}
		}

	}
	return nil
}

func GetSqlite3(dbpath string) *sql.DB {
	db, err := sql.Open("sqlite3", dbpath)
	if err != nil {
		Logger.Fatal("cannot open sqlite3 database:", err)
	}

	return db
}

func PrepareHttpCacheDatabase() error {
	str_dbhttp := CFG.Http.Dir_http
	dbpath := strings.Join([]string{str_dbhttp, "/http.db"}, "")
	Logger.Debug("HTTP DB: ", dbpath)
	var err error
	DBHTTP, err = bolt.Open(dbpath, 0600, &bolt.Options{Timeout: 3 * time.Second})
	if err != nil {
		Logger.Fatal("ERROR while opening:", dbpath, err)
		return err
	}
	return nil
}
