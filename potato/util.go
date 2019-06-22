package potato

import (
	"crypto/md5"
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
		Logger.Debug("Unzip: snappy cannot uncompress the data.")
		return nil
	}
	return dst
}

func ByteMD5(b []byte) string {
	ctx := md5.New()
	ctx.Write(b)
	return strings.ToLower(hex.EncodeToString(ctx.Sum(nil)))
}
