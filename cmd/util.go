package cmd

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/cespare/xxhash/v2"
	"github.com/klauspost/compress/zstd"
	"github.com/zeebo/blake3"
)

func FatalError(prefix string, err error) {
	if err != nil {
		log.Fatalln(Red("ERROR:"), Red(prefix), err)
	}
}

func PrintError(prefix string, err error) {
	if err != nil {
		log.Println(Red("ERROR:"), Red(prefix), err)
	}
}

func DebugInfo(prefix string, args ...any) {
	if IsDebug {
		var info []string
		for _, arg := range args {
			info = append(info, fmt.Sprintf("%v", arg))
		}
		log.Printf("INFO: %v: %v\n", prefix, strings.Join(info, ""))
	}
}

func DebugWarn(prefix string, args ...any) {
	if IsDebug {
		var info []string
		for _, arg := range args {
			info = append(info, fmt.Sprintf("%v", arg))
		}
		log.Println(Yellow("WARN:"), Yellow(prefix), Yellow(strings.Join(info, "")))
	}
}
func IsAnyEmpty(args ...string) bool {
	for _, arg := range args {
		if arg == "" {
			return true
		}
	}
	return false
}

func IsAnyNil(args ...[]byte) bool {
	for _, arg := range args {
		if arg == nil {
			return true
		}
	}
	return false
}

func PrintPflags() error {
	if IsDebug == false {
		return nil
	}
	var keys []string
	// root
	flagroot := map[string]any{
		"IsDebug":            IsDebug,
		"IsIgnoreError":      IsIgnoreError,
		"IsDatabaseReadOnly": IsDatabaseReadOnly,
		"MaxUploadSizeMB":    MaxUploadSizeMB,
	}
	Pflags["_global"] = flagroot

	// httpd
	flaghttpd := map[string]any{
		"Host":             Host,
		"Port":             Port,
		"UploadDir":        UploadDir,
		"StaticDir":        StaticDir,
		"DiskCacheExpires": DiskCacheExpires,
		"AdminUser":        AdminUser,
		"AdminPassword":    AdminPassword,
	}
	Pflags["httpd"] = flaghttpd

	// import
	flagimport := map[string]any{
		"ImportUser":            ImportUser,
		"ImportGroup":           ImportGroup,
		"ImportDir":             ImportDir,
		"ImportExt":             ImportExt,
		"ImportIsIgnoreDotFile": ImportIsIgnoreDotFile,
	}
	Pflags["import"] = flagimport

	// export
	flagexport := map[string]any{
		"ExportDir": ExportDir,
	}
	Pflags["export"] = flagexport

	// put
	flagput := map[string]any{
		"PutUser":  PutUser,
		"PutGroup": PutGroup,
		"PutKey":   PutKey,
		"PutFile":  PutFile,
	}
	Pflags["put"] = flagput

	// delete
	flagdelete := map[string]any{
		"DeleteUser":  DeleteUser,
		"DeleteGroup": DeleteGroup,
		"DeleteKey":   DeleteKey,
	}
	Pflags["delete"] = flagdelete

	fmt.Println(Cyan("-----------------------------------"))
	fmt.Println(Cyan("Current command flag value:"))
	for k1 := range Pflags {
		keys = append(keys, k1)
	}
	sort.Strings(keys)

	for _, k := range keys {
		fmt.Println(Green(k + ":"))
		for k2, v2 := range Pflags[k] {
			fmt.Printf("    %v=%v\n", k2, v2)
		}
	}
	fmt.Println(Cyan("-----------------------------------"))

	return nil
}

func GetXxhash(b []byte) string {
	return strconv.FormatUint(xxhash.Sum64(b), 10)
}

func SumBlake3(b []byte) []byte {
	h := blake3.New()
	h.Write(b)
	return h.Sum(nil)
}

func Normalize(s string) string {
	ban := []string{`\`, `:`, `*`, `?`, `<`, `>`, `|`, `"`, `^`}
	for _, v := range ban {
		s = strings.ReplaceAll(s, v, "")
	}
	s = strings.Trim(s, " ")
	s = strings.Trim(s, "/")
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, "&", "-")

	DebugInfo("Normalize", s)
	return s
}

func JoinKey(s []string) string {
	var s2 []string
	for _, v := range s {
		s2 = append(s2, Normalize(v))
	}
	return strings.ToLower(strings.Join(s2, "/"))
}

func ZstdBytes(rawin []byte) []byte {
	enc, _ := zstd.NewWriter(nil)
	return enc.EncodeAll(rawin, nil)
}

func UnZstdBytes(zin []byte) []byte {
	dec, _ := zstd.NewReader(nil)
	out, err := dec.DecodeAll(zin, nil)
	if err != nil {
		PrintError("UnZstdBytes:DecodeAll", err)
	}
	return out
}

func MakeDirs(dpath string) error {
	DebugInfo("MakeDirs", dpath)
	_, err := os.Stat(dpath)
	if err != nil {
		err = os.MkdirAll(dpath, os.ModePerm)
		PrintError("MakeDirs:MkdirAll", err)
	}
	return nil
}

func DefaultBase64Asset(fpath string, b64 string) {
	_, err := os.Stat(fpath)
	if err != nil {
		d64, _ := base64.StdEncoding.DecodeString(b64)
		ioutil.WriteFile(fpath, []byte(d64), os.ModePerm)
	}
}

func Int2Int64(n int) int64 {
	s := strconv.Itoa(n)
	m, err := strconv.ParseInt(s, 10, 64)
	PrintError("Int2Int64", err)
	return m
}

func CleanExpires(fpath string, expireSecond float64) error {
	if fpath == "" {
		DebugWarn("CleanExpires", "path cannot be empty")
		return nil
	}
	if strings.HasPrefix(fpath, "/") || strings.Contains(fpath, ":") {
		DebugWarn("CleanExpires", "path cannot start with / or contain :, should be a relative path")
		return nil
	}
	fpath = strings.Trim(fpath, "/")
	DebugInfo("CleanExpires:path", fpath)

	tNow := time.Now()
	filepath.Walk(fpath, func(path string, finfo os.FileInfo, err error) error {
		if finfo.IsDir() {
			return nil
		}

		if strings.HasPrefix(finfo.Name(), ".") {
			return nil
		}

		tAge := tNow.Sub(finfo.ModTime()).Seconds()

		if tAge > expireSecond {
			os.Remove(path)
			DebugInfo("CleanExpires:remove expired file", tAge, ": ", path)
		}
		return nil
	})
	return nil
}
