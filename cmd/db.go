package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	//"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	bolt "go.etcd.io/bbolt"
)

const SystemBucket string = "_system"
const MetaBucket string = "_meta"

// const FbinBucket string = "_fbin"
const UserBucket string = "_user"
const aesKeyDefault string = "thisis32bitlongpassphrasedefault"

const PageSize int = 1000

var (
	db          *bolt.DB
	userpass    map[string]string = make(map[string]string)
	currentUser string
	aesKey      []byte
)

func openBolt() {
	var err error

	db, err = bolt.Open("data/data.db", 0600, &bolt.Options{
		Timeout:  3 * time.Second,
		ReadOnly: IsDatabaseReadOnly,
	})
	FatalError("openBolt", err)

	if IsDatabaseReadOnly {
		DebugWarn("openBolt", "db is in ReadOnly mode")
	}

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

		_, err = tx.CreateBucketIfNotExists([]byte(UserBucket))
		if err != nil {
			FatalError("OpenBolt:CreateBucketIfNotExists", err)
			return err
		}

		return nil
	})

	updateUserPass()
	openBadger()
}

func dbSave(user, group, key string, data []byte) string {
	if IsAnyEmpty(user, group, key) || IsAnyNil(data) || len(data) == 0 {
		return EmptyVal
	}

	if strings.HasPrefix(user, "_") {
		DebugWarn("dbSave:param invalid", "param --user could not be start with _")
		return EmptyVal
	}

	user = strings.ToLower(user)
	group = strings.ToLower(group)

	// for windows
	key = ToUnixSlash(key)

	gkey := JoinKey([]string{group, key})
	fullKey := JoinKey([]string{user, gkey})

	tx, err := db.Begin(true)
	if err != nil {
		return EmptyVal
	}
	defer tx.Rollback()

	bktUser, err := tx.CreateBucketIfNotExists([]byte(user))
	if err != nil {
		FatalError("dbSave:CreateBucketIfNotExists", err)
		return EmptyVal
	}

	if bktUser.Get([]byte(gkey)) != nil {
		DebugWarn("dbSave:Update:SKIP", "key exists")
		return fullKey
	}
	// Fbin
	// bktFbin := tx.Bucket([]byte(FbinBucket))
	// fbinK := SumBlake3(data)
	// if bktFbin.Get(fbinK) == nil {
	// 	bktFbin.Put(fbinK, ZstdBytes(data))
	// }
	fbinK := fbinSave(data)
	if fbinK == nil {
		DebugWarn("dbSave:fbinSave", "can not save into badger")
		return EmptyVal
	}

	if err := bktUser.Put([]byte(gkey), fbinK); err != nil {
		FatalError("dbSave:PUT", err)
		return EmptyVal
	}

	var idstr string
	// update system auto_increment_id
	bktSystem := tx.Bucket([]byte(SystemBucket))
	id, _ := bktSystem.NextSequence() //uint64
	idstr = strconv.FormatUint(id, 10)
	bktSystem.Put([]byte(idstr), []byte(fullKey))
	// update max increment_id
	bktMeta := tx.Bucket([]byte(MetaBucket))
	bktMeta.Put([]byte("meta/current_increment_id"), []byte(idstr))

	if err := tx.Commit(); err != nil {
		FatalError("dbSave:Commit", err)
		return EmptyVal
	}

	return fullKey
}

func dbGet(bkt, key string) []byte {
	var fbinK []byte
	var fdata []byte
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bkt))
		if b == nil {
			return ErrNotExist
		}
		fbinK = b.Get([]byte(key))
		return nil
	})
	if fbinK != nil {
		fdata = fbinGet(fbinK)
	}

	return fdata
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

func dbUserAdd(name, pass string) error {
	if IsAnyEmpty(name, pass) {
		DebugWarn("addUser", "name/password cannot be empty")
		return ErrParamEmpty
	}
	name = strings.ToLower(name)
	if AdminUser == name {
		DebugWarn("addUser", "name cannot be same as admin")
		return ErrParamInvalid
	}

	if len(name) <= 2 {
		DebugWarn("addUser", "name should be more than 2 letters")
		return ErrParamInvalid
	}

	err := db.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte(UserBucket))
		if bkt.Get([]byte(name)) != nil {
			DebugWarn("addUser", "name exists")
			return ErrParamInvalid
		}
		encpass := aesEncHex([]byte(pass))
		err := bkt.Put([]byte(name), []byte(encpass))
		PrintError("addUser", err)
		return err
	})

	return err
}

func dbUserDelete(name string) error {
	if IsAnyEmpty(name) {
		DebugWarn("deleteUser", "name cannot be empty")
		return ErrParamEmpty
	}

	if AdminUser == name {
		DebugWarn("deleteUser", "admin cannot be removed")
	}

	err := db.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte(UserBucket))
		if bkt.Get([]byte(name)) == nil {
			DebugWarn("deleteUser", "user does not exist: ", name)
			return ErrNotExist
		}
		err := bkt.Delete([]byte(name))
		PrintError("deleteUser", err)
		return err
	})
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

func getSystemPaged(pageNum int) []string {
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

func updateUserPass() error {
	userpassReset := make(map[string]string)
	userpass = userpassReset

	db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte(UserBucket)).Cursor()
		if c == nil {
			return nil
		}
		pre := []byte("")
		for k, v := c.Seek(pre); k != nil && v != nil; k, v = c.Next() {
			dec := aesDecHex(string(v))
			userpass[string(k)] = dec
		}
		return nil
	})
	if AdminUser != "" && AdminPassword != "" {
		userpass[AdminUser] = AdminPassword
	}

	DebugInfo("updateUserPass:", "===========")
	for k, v := range userpass {
		DebugInfo("updateUserPass:", k, ":", v)
	}

	return nil
}
