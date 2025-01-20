package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	bolt "go.etcd.io/bbolt"
)

const SystemBucket string = "_system"
const MetaBucket string = "_meta"
const FbinBucket string = "_fbin"

const PageSize int = 1000

var db *bolt.DB

func init() {

	//defer pprof.StopCPUProfile()
	MakeDirs("data")
	MakeDirs("data/_sync")
	MakeDirs("data/logs")
	//
	MakeDirs("www")
	MakeDirs("www/uploads")
	MakeDirs("www/export")
	MakeDirs("www/temp")
	MakeDirs("www/assets")
	MakeDirs("www/static")
	//
	DefaultBase64Asset("www/assets/video-js.min.css", videojsmincss)
	DefaultBase64Asset("www/assets/video.min.js", videominjs)
	DefaultBase64Asset("www/assets/style.css", stylecss)
	DefaultBase64Asset("www/assets/favicon.png", faviconpng)
	DefaultBase64Asset("www/assets/video-bg.png", videobgpng)
	//
	// for memory
	videojsmincss, videominjs, stylecss, faviconpng, videobgpng = "", "", "", "", ""
	//
	MaxUploadSize = Int2Int64(MaxUploadSizeMB * MB)
}

func openBolt() {
	var err error

	db, err = bolt.Open("data/data.db", 0600, &bolt.Options{
		Timeout:  3 * time.Second,
		ReadOnly: IsDatabaseReadOnly,
	})
	if IsDatabaseReadOnly {
		DebugWarn("openBolt", "db is in ReadOnly mode")
	}

	FatalError("openBolt", err)

	db.Update(func(tx *bolt.Tx) error {
		_, err = tx.CreateBucketIfNotExists([]byte(SystemBucket))
		if err != nil {
			FatalError("OpenBolt:CreateBucketIfNotExists", err)
			return err
		}

		_, err = tx.CreateBucketIfNotExists([]byte(MetaBucket))
		if err != nil {
			FatalError("OpenBolt:CreateBucketIfNotExists", err)
			return err
		}

		_, err = tx.CreateBucketIfNotExists([]byte(FbinBucket))
		if err != nil {
			FatalError("OpenBolt:CreateBucketIfNotExists", err)
			return err
		}
		return nil
	})
}

func dbSave(user, group, key string, data []byte) string {
	if IsAnyEmpty(user, group, key) || IsAnyNil(data) || len(data) == 0 {
		return ""
	}

	if strings.HasPrefix(user, "_") {
		DebugWarn("dbSave:param invalid", "param --user could not be start with _")
		return ""
	}

	user = strings.ToLower(user)
	group = strings.ToLower(group)

	gkey := JoinKey([]string{group, key})
	fullKey := JoinKey([]string{user, gkey})

	isNewPut := false

	tx, err := db.Begin(true)
	if err != nil {
		return ""
	}
	defer tx.Rollback()

	// Fbin
	fbin := tx.Bucket([]byte(FbinBucket))
	fbinK := SumBlake3(data)
	if fbin.Get(fbinK) == nil {
		fbin.Put(fbinK, ZstdBytes(data))
	}

	bkt, err := tx.CreateBucketIfNotExists([]byte(user))
	if err != nil {
		FatalError("dbSave:CreateBucketIfNotExists", err)
		return ""
	}

	if bkt.Get([]byte(gkey)) != nil {
		isNewPut = false
		DebugWarn("dbSave:Update:SKIP", "key exists")
	} else {
		if err := bkt.Put([]byte(gkey), fbinK); err == nil {
			isNewPut = true
		}
	}

	if err := tx.Commit(); err != nil {
		FatalError("dbSave:Commit", err)
		return ""
	}

	if isNewPut {
		dbUpdateSys(fullKey)
	}

	return fullKey
}

func dbUpdateMeta(k, v string) error {
	db.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte(MetaBucket))
		err := bkt.Put([]byte(k), []byte(v))
		PrintError("dbUpdateMeta", err)
		return err
	})
	return nil
}

func dbUpdateSys(k string) error {
	var idstr string
	db.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte(SystemBucket))

		id, _ := bkt.NextSequence() //uint64
		idstr = strconv.FormatUint(id, 10)
		err := bkt.Put([]byte(idstr), []byte(k))
		PrintError("dbUpdateSys", err)

		return err
	})

	dbUpdateMeta("meta/current_increment_id", idstr)
	return nil
}
func dbPagedSys(pageNum int) []string {
	var res []string

	if pageNum < 1 {
		pageNum = 1
	}
	db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte(SystemBucket)).Cursor()

		prefix := []byte("")
		min := (pageNum - 1) * PageSize
		max := min + PageSize

		for k, v := c.Seek(prefix); k != nil; k, v = c.Next() {
			intk, _ := strconv.Atoi(string(k))
			if intk > min && intk <= max {
				res = append(res, fmt.Sprintf("%s:%s", k, v))
			}
		}

		return nil
	})
	pageFile := fmt.Sprintf("data/_sync/%d.json", pageNum)
	DebugInfo("dbPagedSys:pageFile", pageFile)
	_, err := os.Stat(pageFile)
	if err != nil {
		bres, err := json.Marshal(res)
		PrintError("dbPagedSys:json.Marshal", err)

		err = ioutil.WriteFile(pageFile, bres, os.ModePerm)
		PrintError("dbPagedSys:WriteFile", err)
	}

	return res
}

func dbGet(bkt, key string) []byte {
	var fbinK []byte
	var fdata []byte
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bkt))
		if b == nil {
			return ErrNotExist
		}
		fbinK = b.Get([]byte(key))
		return nil
	})
	PrintError("dbGet", err)

	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(FbinBucket))
		fdata = b.Get(fbinK)
		return nil
	})

	if fdata == nil {
		return nil
	}
	return UnZstdBytes(fdata)
}

func dbDelete(bkt, key string) error {
	if IsAnyEmpty(bkt, key) {
		PrintError("dbDelete", ErrBucketKeyEmpty)
		return ErrBucketKeyEmpty
	}
	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bkt))
		if b != nil && b.Get([]byte(key)) != nil {
			return b.Delete([]byte(key))
		}
		return ErrNotExist
	})
	if err != nil {
		PrintError("dbDelete", err)
	} else {
		DebugInfo("dbDelete", "OK")
	}

	return err
}

func getAllBuckets() []string {
	var bkts []string
	db.View(func(tx *bolt.Tx) error {
		tx.ForEach(func(name []byte, b *bolt.Bucket) error {
			if name != nil && string(name) != "" {
				bkts = append(bkts, string(name))
			}
			return nil
		})
		return nil
	})
	return bkts
}

func getAllGroups(bkt string) []string {
	var groups []string
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bkt))
		if b == nil {
			return nil
		}
		c := b.Cursor()

		prefix := []byte("")
		m := make(map[string]int, 1)
		kstr := ""
		gname := ""
		slashidx := 0
		for k, _ := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, _ = c.Next() {
			kstr = string(k)
			slashidx = strings.Index(kstr, "/")
			if slashidx > 0 {
				gname = kstr[0:slashidx]
				m[gname] = 1
			}
		}

		for k := range m {
			groups = append(groups, k)
		}

		return nil
	})
	sort.Strings(groups)
	return groups
}

func getAllKeys(bkt string, pre string) []string {
	var keys []string
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bkt))
		if b == nil {
			return nil
		}
		c := b.Cursor()

		prefix := []byte(pre)

		for k, _ := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, _ = c.Next() {
			keys = append(keys, string(k))
		}
		return nil
	})
	return keys
}

func getAllFiles(bkt string, grp string, pageNum int) []string {
	var files []string
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bkt))
		if b == nil {
			return nil
		}
		c := b.Cursor()

		prefix := []byte(grp)

		min := (pageNum - 1) * PageSize
		max := min + PageSize
		idx := 0
		for k, _ := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, _ = c.Next() {
			if idx < min {
				idx++
				continue
			}

			if idx > max {
				break
			}
			files = append(files, strings.TrimLeft(strings.TrimLeft(string(k), grp), "/"))
			idx++
		}

		return nil
	})
	return files
}

func getFile(bkt string, fname string) []byte {
	var val []byte
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bkt))
		val = b.Get([]byte(fname))
		return nil
	})
	return val
}

func getMeta(prefix string) (kvs []map[string]string) {
	db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte(MetaBucket)).Cursor()
		if c == nil {
			return nil
		}
		pre := []byte(prefix)
		for k, v := c.Seek(pre); k != nil; k, v = c.Next() {
			kv := make(map[string]string, 1)
			kv["key"] = string(k)
			kv["val"] = string(v)
			kvs = append(kvs, kv)
		}
		return nil
	})
	return kvs
}

func exportFiles(dpath string) {
	MakeDirs(dpath)
	buckets := getAllBuckets()
	var notExportFiles []string

	if len(buckets) > 0 {
		for _, bucket := range buckets {
			if bucket == "" || strings.HasPrefix(bucket, "_") {
				continue
			}
			DebugInfo("exportFiles:Bucket", bucket)
			db.View(func(tx *bolt.Tx) error {
				c := tx.Bucket([]byte(bucket)).Cursor()

				prefix := []byte("")

				for k, v := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, v = c.Next() {

					exportPath := strings.Join([]string{dpath, bucket, string(k)}, "/")
					DebugInfo("exportFiles:exportPath", exportPath)
					MakeDirs(filepath.Dir(exportPath))
					DebugInfo("exportFiles:Size:", strconv.Itoa(len(v)))
					ioutil.WriteFile(exportPath, UnZstdBytes(v), os.ModePerm)
					finfo, err := os.Stat(exportPath)

					if IsIgnoreError {
						PrintError("exportFiles:os.Stat", err)
					} else {
						FatalError("exportFiles:os.Stat", err)
					}

					if finfo.Size() == 0 {
						if IsIgnoreError {
							PrintError("exportFiles:size", ErrFileSizeZero)
						} else {
							FatalError("exportFiles:size", ErrFileSizeZero)
						}
					}

					if err != nil || finfo.Size() == 0 {
						notExportFiles = append(notExportFiles, strings.Join([]string{"EXPORT ERROR:", bucket, string(k)}, ":"))
					}
				}

				return nil
			})
		}
	}

	if len(notExportFiles) > 0 {
		err := ioutil.WriteFile("data/logs/export_error_files-"+GetXxhash([]byte(dpath))+".log", []byte(strings.Join(notExportFiles, "\r\n")), os.ModePerm)
		PrintError("ExportFiles:write log", err)
	}
}

func ImportFiles(dpath, ext, user, group string) error {
	DebugInfo("ImportFiles:dir", dpath)
	DebugInfo("ImportFiles:ext", ext)
	DebugInfo("ImportFiles:user", user)
	DebugInfo("ImportFiles:group", group)
	DebugInfo("ImportFiles:max-upload-size", MaxUploadSize)
	if dpath == "" || ext == "" || user == "" || group == "" {
		DebugWarn("ImportFiles:Param", "dpath, ext, user, group cannot be empty")
		return nil
	}
	var relPath string
	var fdata []byte
	var ignoreFiles []string

	filepath.Walk(dpath, func(path string, finfo os.FileInfo, err error) error {
		if finfo.IsDir() {
			return nil
		}
		if ImportIsIgnoreDotFile {
			if strings.HasPrefix(finfo.Name(), ".") {
				ignoreFiles = append(ignoreFiles, "IGNORE:dot-file: "+path)
				return nil
			}
		}

		if ext != "*" {
			if filepath.Ext(strings.ToLower(finfo.Name())) != strings.ToLower(ext) {
				return nil
			}
		}

		if finfo.Size() > MaxUploadSize {
			ignoreFiles = append(ignoreFiles, "IGNORE:oversize: "+path)
			return nil
		}

		if finfo.Size() == 0 {
			ignoreFiles = append(ignoreFiles, "IGNORE:0 size: "+path)
			return nil
		}

		relPath = strings.TrimPrefix(path, dpath)
		relPath = strings.Trim(relPath, "/")
		if filepath.Base(relPath) != filepath.Base(path) {
			if IsIgnoreError {
				PrintError("ImportFiles:filepath.Base", ErrFileNameNotMatch)
			} else {
				FatalError("ImportFiles:filepath.Base", ErrFileNameNotMatch)
			}
		}
		fdata, err = ioutil.ReadFile(path)
		if err != nil {
			PrintError("ImportFiles:ReadFile", err)
			return nil
		}
		k := dbSave(user, group, relPath, fdata)
		DebugInfo("Saved", strings.Join([]string{path, relPath, k}, "=>"))
		return nil
	})

	if len(ignoreFiles) > 0 {
		loghash := strings.Join([]string{dpath, ext, user, group}, ":")
		err := ioutil.WriteFile("data/logs/import_ignore_files-"+GetXxhash([]byte(loghash))+".log", []byte(strings.Join(ignoreFiles, "\r\n")), os.ModePerm)
		PrintError("ImportFiles:write log", err)
	}

	return nil
}
