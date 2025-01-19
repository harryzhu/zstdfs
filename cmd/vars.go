package cmd

import (
	"errors"
)

const MB int = 1 << 20

var (
	ErrParamEmpty       error = errors.New("args could not be empty")
	ErrNotExist         error = errors.New("does not exist")
	ErrBucketKeyEmpty   error = errors.New("bucket and key cannot be empty")
	ErrFileSizeZero     error = errors.New("filesize is 0")
	ErrFileNameNotMatch error = errors.New("basename of path and relPath are different")
)

// func dbGet2(bkt, key string) []byte {
// 	var val []byte
// 	err := db.View(func(tx *bolt.Tx) error {
// 		b := tx.Bucket([]byte(bkt))
// 		if b == nil {
// 			return ErrNotExist
// 		}
// 		val = b.Get([]byte(key))
// 		return nil
// 	})
// 	PrintError("dbGet", err)

// 	if err != nil {
// 		return nil
// 	}
// 	return UnZstdBytes(val)
// }

// func dbSave2(user, group, key string, data []byte) string {
// 	if IsAnyEmpty(user, group, key) || IsAnyNil(data) || len(data) == 0 {
// 		return ""
// 	}

// 	if strings.HasPrefix(user, "_") {
// 		DebugWarn("dbSave:param invalid", "param --user could not be start with _")
// 		return ""
// 	}

// 	user = strings.ToLower(user)
// 	group = strings.ToLower(group)

// 	gkey := JoinKey([]string{group, key})
// 	fullKey := JoinKey([]string{user, gkey})

// 	isNewPut := false

// 	err2 := db.Update(func(tx *bolt.Tx) error {
// 		bkt, err := tx.CreateBucketIfNotExists([]byte(user))
// 		if err != nil {
// 			FatalError("dbSave:CreateBucketIfNotExists", err)
// 			return err
// 		}

// 		if bkt.Get([]byte(gkey)) != nil {
// 			isNewPut = false
// 			DebugWarn("dbSave:Update:SKIP", "key is existing")
// 			return nil
// 		}
// 		err = bkt.Put([]byte(gkey), ZstdBytes(data))
// 		isNewPut = true
// 		PrintError("dbSave:Put", err)

// 		return err
// 	})

// 	if err2 != nil {
// 		return ""
// 	}
// 	if isNewPut {
// 		dbUpdateSys(fullKey)
// 	}

// 	DebugInfo("dbSave:fullKey", fullKey)

// 	return fullKey
// }
