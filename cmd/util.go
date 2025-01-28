package cmd

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
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
		if IsIgnoreError == false {
			log.Fatalln(Red("ERROR:"), Red(prefix), err)
		} else {
			log.Println(Red("ERROR:"), Red(prefix), err)
		}

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
		log.Println(Yellow("WARN:"), Yellow(prefix+":"), Yellow(strings.Join(info, "")))
	}
}

func PrintlnInfo(prefix string, args ...any) {

	var info []string
	for _, arg := range args {
		info = append(info, fmt.Sprintf("%v", arg))
	}
	log.Printf("INFO: %v: %v\n", prefix, strings.Join(info, ""))

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
		"PutEndpoint": PutEndpoint,
		"PutUser":     PutUser,
		"PutGroup":    PutGroup,
		"PutPrefix":   PutPrefix,
		"PutFile":     PutFile,
		"PutAuth":     PutAuth,
	}
	Pflags["put"] = flagput

	// delete
	flagdelete := map[string]any{
		"DeleteEndpoint": DeleteEndpoint,
		"DeleteUser":     DeleteUser,
		"DeleteGroup":    DeleteGroup,
		"DeleteKey":      DeleteKey,
		"DeleteAuth":     DeleteAuth,
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
	// RFC1123
	ban := []string{`\`, `:`, `*`, `?`, `<`, `>`, `|`, `^`}
	for _, v := range ban {
		s = strings.ReplaceAll(s, v, "")
	}
	s = strings.ReplaceAll(s, "/.", "/")
	s = strings.ReplaceAll(s, "./", "/")
	s = strings.ReplaceAll(s, "./.", "/")
	s = strings.Trim(s, " ")
	s = strings.Trim(s, "/")
	s = strings.Trim(s, "/")
	s = strings.Trim(s, ".")

	manySpaces := regexp.MustCompile(`[\s]{2,}`)
	s = manySpaces.ReplaceAllString(s, " ")

	s = strings.ReplaceAll(s, " ", "-")
	s = strings.ReplaceAll(s, "&", "-")

	DebugInfo("Normalize", s)
	return s
}

func JoinKey(s []string) string {
	var s2 []string
	for _, v := range s {
		s2 = append(s2, v)
	}
	return strings.Join(s2, "/")
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

func PrintSpinner(s string) {
	if IsDebug == false {
		fmt.Printf("... %5.30s\r", s)
	}

}

func ToUnixSlash(s string) string {
	// for windows
	return strings.ReplaceAll(s, "\\", "/")
}

func DefaultAsset(dest string, src string) {
	_, err := os.Stat(dest)
	if err != nil {
		b, err := embeddedFS.ReadFile(src)
		if err != nil {
			DebugWarn("DefaultAsset", err)
		} else {
			ioutil.WriteFile(dest, b, os.ModePerm)
		}
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

func GetEnv(k string, defaultVal string) string {
	ev := os.Getenv(k)
	if ev == "" {
		return defaultVal
	}
	return ev
}

func SHA256String(s string) string {
	h := sha256.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}

func aesEncHex(src []byte) string {
	return hexByte2Str(EncryptAES(src))
}

func aesDecHex(enc string) string {
	return string(DecryptAES(hexStr2Byte(enc)))
}

func hexByte2Str(b []byte) string {
	return hex.EncodeToString(b)
}

func hexStr2Byte(s string) []byte {
	decoded, err := hex.DecodeString(s)
	if err != nil {
		PrintError("hexStr2Byte", err)
	}

	return decoded
}

func EncryptAES(plaintext []byte) (encrypted []byte) {
	block, _ := aes.NewCipher(aesKey)
	blockSize := block.BlockSize()
	plaintext = pkcs5Padding(plaintext, blockSize)
	blockMode := cipher.NewCBCEncrypter(block, aesKey[:blockSize])
	encrypted = make([]byte, len(plaintext))
	blockMode.CryptBlocks(encrypted, plaintext)

	return encrypted
}

func DecryptAES(encrypted []byte) (decrypted []byte) {
	block, _ := aes.NewCipher(aesKey)
	blockSize := block.BlockSize()
	blockMode := cipher.NewCBCDecrypter(block, aesKey[:blockSize])
	decrypted = make([]byte, len(encrypted))
	blockMode.CryptBlocks(decrypted, encrypted)
	decrypted = pkcs5UnPadding(decrypted)
	return decrypted
}

func pkcs5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}
func pkcs5UnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}
