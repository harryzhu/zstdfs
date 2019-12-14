package potato

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"strings"
	"time"

	//"github.com/cespare/xxhash"
	"github.com/golang/snappy"
)

func Zip(src []byte) []byte {
	return snappy.Encode(nil, src)
}

func Unzip(src []byte) []byte {
	dst, err := snappy.Decode(nil, src)
	if err != nil {
		return nil
	}
	return dst
}

func ByteMD5(b []byte) string {
	ctx := md5.New()
	ctx.Write(b)
	return strings.ToLower(hex.EncodeToString(ctx.Sum(nil)))
}

func ByteSHA256(b []byte) string {
	ctx := sha256.New()
	ctx.Write(b)
	return strings.ToLower(hex.EncodeToString(ctx.Sum(nil)))
}

func IsEmpty(b []byte) bool {
	if b == nil || len(b) <= 0 {
		return true
	}
	return false
}

func IsEmptyString(b string) bool {
	if b == "" || len(b) == 0 {
		return true
	}
	return false
}

func IsOversize(b []byte) bool {
	if len(b) > entityMaxSize {
		return true
	}
	return false
}

func TimeNowNanoString() string {
	return strconv.FormatInt(time.Now().UnixNano(), 10)
}

func TimeNowUnixString() string {
	return strconv.FormatInt(time.Now().Unix(), 10)
}

func SecondFormatTimeString(s string) string {
	t, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return ""
	}
	return time.Unix(t, 0).Format("2006-01-02 03:04:05")
}
