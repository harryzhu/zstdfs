package util

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"hash/crc32"
	"io"
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
)

var (
	HAZHUFS_LOGLEVEL   string
	HAZHUFS_CONFIGFILE string
)

func init() {
	HAZHUFS_LOGLEVEL = os.Getenv("HAZHUFS_LOGLEVEL")
	HAZHUFS_CONFIGFILE = os.Getenv("HAZHUFS_CONFIGFILE")
}

func GetLogLevel() log.Level {
	HAZHUFS_LOGLEVEL = os.Getenv("HAZHUFS_LOGLEVEL")
	switch HAZHUFS_LOGLEVEL {
	case "DEBUG":
		return log.DebugLevel
	case "INFO":
		return log.InfoLevel
	case "WARN":
		return log.WarnLevel
	default:
		return log.ErrorLevel
	}
}

func GetConfigFile() (configpath string, err error) {
	configpath = os.Getenv("HAZHUFS_CONFIGFILE")
	_, err = os.Stat(HAZHUFS_CONFIGFILE)
	if err != nil {
		return "", err
	}
	return configpath, nil
}

func ByteCRC32(fdata []byte) uint32 {
	crc32q := crc32.MakeTable(0xD5828281)
	return crc32.Checksum(fdata, crc32q)
}

func FileMD5(fpath string) string {
	f, _ := os.Open(fpath)
	defer f.Close()
	md5hash := md5.New()
	if _, err := io.Copy(md5hash, f); err != nil {
		return ""
	}
	return hex.EncodeToString(md5hash.Sum(nil))
}

func TextMD5(text string) string {
	ctx := md5.New()
	ctx.Write([]byte(text))
	return hex.EncodeToString(ctx.Sum(nil))
}

func ByteMD5(b []byte) string {
	ctx := md5.New()
	ctx.Write(b)
	return strings.ToLower(hex.EncodeToString(ctx.Sum(nil)))
}

func DirWalker(path string, filter string) (files []string) {
	var filelist = make([]string, 0, 10)
	abspath, _ := filepath.Abs(path)
	err := filepath.Walk(abspath, func(abspath string, f os.FileInfo, err error) error {
		if f == nil {
			return err
		}
		if f.IsDir() {
			return nil
		}
		matched, _ := filepath.Match(filter, filepath.Base(abspath))
		if matched == true {
			filelist = append(filelist, abspath)
		}

		return nil
	})
	if err != nil {
		fmt.Printf("filepath.Walk() returned %v\n", err)
	}
	return filelist
}

func Echo() error {
	return nil
}
