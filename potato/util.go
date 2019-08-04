package potato

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"strings"

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
