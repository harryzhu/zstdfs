package cmd

import (
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand/v2"
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
	"golang.org/x/crypto/bcrypt"
)

func GetNowUnix() int64 {
	return time.Now().UTC().Unix()
}

func GetNowUnixMillo() int64 {
	return time.Now().UTC().UnixMilli()
}

func ShowRunningTime(t1 int64, label string) {
	t2 := GetNowUnixMillo()
	DebugInfo("RunningTime", label, ": ", fmt.Sprintf("%v", t2-t1), " msec. [", t2, "-", t1, "]")
}

func ToUnixSlash(s string) string {
	// for windows
	return strings.ReplaceAll(s, "\\", "/")
}

func UnixFormat(t int64, format string) string {
	if format == "" {
		//"2006-01-02 15:04:05"
		return time.Unix(t, 0).Format("2006-01-02 15:04:05")
	}
	return time.Unix(t, 0).Format(format)
}

func GetRandomInts(count, min, max int) (ints []int) {
	ts := time.Now()
	ts1, _ := strconv.Atoi(strconv.FormatInt(ts.UTC().Unix(), 10))
	ts2, _ := strconv.Atoi(strconv.FormatInt(ts.UnixNano(), 10))
	r := rand.New(rand.NewPCG(uint64(ts1), uint64(ts2)))
	DebugInfo("GetRandomInts:max", max)
	if max <= 0 {
		return ints
	}
	for i := 0; i < count; i++ {

		ints = append(ints, r.IntN(max))
	}
	return ints
}

func GetSiteURL() string {
	if SiteURL != "" {
		return SiteURL
	}
	if Host == "0.0.0.0" {
		return strings.Join([]string{"http://localhost", Port}, ":")
	}
	return strings.Join([]string{"http://", Host, ":", Port}, "")
}

func GetURI(id string) string {
	if IsAnyEmpty(id) {
		return ""
	}
	uri := GetXxhash([]byte(id))
	fext := strings.ToLower(filepath.Ext(id))
	if fext != "" {
		uri = strings.Join([]string{GetXxhash([]byte(id)), fext}, "")
	}

	return uri
}

func SHA256String(s string) string {
	h := sha256.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}

func SHA256Bytes(b []byte) string {
	h := sha256.New()
	h.Write(b)
	return hex.EncodeToString(h.Sum(nil))
}

func SHA256File(fpath string) string {
	f, err := os.Open(fpath)
	defer f.Close()
	if err != nil {
		return ""
	}
	h := sha256.New()
	_, err = io.Copy(h, f)
	if err != nil {
		return ""
	}
	return hex.EncodeToString(h.Sum(nil))
}

func GetPassword(u, p string) string {
	if IsAnyEmpty(u, p) {
		return ""
	}
	p1 := SHA256String(strings.Join([]string{SHA256String(p), SHA256String(strings.ToLower(u))}, ":"))
	hash, err := bcrypt.GenerateFromPassword([]byte(p1), bcrypt.DefaultCost)
	if err != nil {
		return ""
	}
	return string(hash)
}

func VerifyPassword(h string, u, p string) bool {
	if IsAnyEmpty(h, u, p) {
		return false
	}
	p1 := SHA256String(strings.Join([]string{SHA256String(p), SHA256String(strings.ToLower(u))}, ":"))

	err := bcrypt.CompareHashAndPassword([]byte(h), []byte(p1))
	if err != nil {
		return false
	}
	return true
}

func GenAPIKey(u string) string {
	p1 := strings.Join([]string{UnixFormat(GetNowUnix(), ""), strings.ToLower(u)}, ":")
	return GetXxhash([]byte(p1))
}

func ToKMGTB(n int) string {
	if n > TB {
		return fmt.Sprintf("%.1f TB", float64(n)/float64(TB))
	}

	if n > GB {
		return fmt.Sprintf("%.1f GB", float64(n)/float64(GB))
	}

	if n > MB {
		return fmt.Sprintf("%.1f MB", float64(n)/float64(MB))
	}

	if n > KB {
		return fmt.Sprintf("%.1f KB", float64(n)/float64(KB))
	}

	return fmt.Sprintf("%d Bytes", n)
}

func ToKWM(n int) string {
	if n > M {
		return fmt.Sprintf("%.1fM", float64(n)/float64(M))
	}
	if n > W {
		return fmt.Sprintf("%.1fW", float64(n)/float64(W))
	}
	if n > K {
		return fmt.Sprintf("%.1fK", float64(n)/float64(K))
	}
	return Int2Str(n)
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

func PrintSpinner(s string) {
	//if IsDebug == false {
	fmt.Printf("... %5.30s\r", s)
	//}
}

func GetEnv(k string, defaultVal string) string {
	ev := os.Getenv(k)
	if ev == "" {
		return defaultVal
	}
	return ev
}

func Str2Float64(n string) float64 {
	s, err := strconv.ParseFloat(n, 64)
	if err != nil {
		return 0
	}
	return s
}

func Str2Int(n string) int {
	if strings.Index(n, ".") > 0 {
		n = n[:strings.Index(n, ".")]
	}
	s, err := strconv.Atoi(n)
	if err != nil {
		return 0
	}
	return s
}

func Str2Int64(n string) int64 {
	s, err := strconv.ParseInt(n, 10, 64)
	if err != nil {
		return 0
	}
	return s
}

func Str2Strings(line string, separator string) (lines []string) {
	line = TagLineFormat(line)
	lines = strings.Split(line, separator)
	return lines
}

func Str2Ints(line string, separator string) (ints []int) {
	line = TagLineFormat(line)
	lines := strings.Split(line, separator)
	for _, k := range lines {
		if k != "" {
			ints = append(ints, Str2Int(k))
		}
	}
	return ints
}

func Int2Int64(n int) int64 {
	s := strconv.Itoa(n)
	m, err := strconv.ParseInt(s, 10, 64)
	PrintError("Int2Int64", err)
	return m
}

func Int2Str(n int) string {
	return strconv.Itoa(n)
}

func Int64ToString(n int64) string {
	return strconv.FormatInt(n, 10)
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
	_, err := os.Stat(dpath)
	if err != nil {
		DebugInfo("MakeDirs", dpath)
		err = os.MkdirAll(dpath, os.ModePerm)
		PrintError("MakeDirs:MkdirAll", err)
	}
	return nil
}

func GetXxhash(b []byte) string {
	return strconv.FormatUint(xxhash.Sum64(b), 10)
}

func SumBlake3(b []byte) []byte {
	h := blake3.New()
	h.Write(b)
	return []byte(fmt.Sprintf("%x", h.Sum(nil)))
}

func Contains(arr []string, target string) bool {
	for _, val := range arr {
		if val == target {
			return true
		}
	}
	return false
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

func TagLineFormat(s string) string {
	DebugInfo("TagLineFormat:before", s)
	s = strings.ReplaceAll(s, "，", ",")
	s = strings.ReplaceAll(s, ";", ",")
	s = strings.ReplaceAll(s, "；", ",")
	s = strings.ReplaceAll(s, "|", ",")
	s = strings.ReplaceAll(s, "/", ",")
	s = strings.ReplaceAll(s, "、", ",")
	s = strings.ReplaceAll(s, "#", ",")
	s = strings.ReplaceAll(s, " ", ",")
	ss := strings.Split(s, ",")
	var words []string
	for _, ele := range ss {
		ele = strings.TrimLeft(ele, " ")
		if ele != "" {
			words = append(words, ele)
		}
	}
	DebugInfo("TagLineFormat:after", words)
	return strings.Trim(strings.Join(words, ","), ",")
}

func CleanExpires(fpath string, expireSecond float64) error {
	if fpath == "" {
		DebugWarn("CleanExpires", "path cannot be empty")
		return nil
	}

	tNow := time.Now()
	filepath.Walk(fpath, func(path string, finfo os.FileInfo, err error) error {
		if finfo.IsDir() {
			return nil
		}

		if strings.HasPrefix(finfo.Name(), ".") || strings.HasPrefix(finfo.Name(), "..") {
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

func GobDump(fpath string, data any) bool {
	savePath := ToUnixSlash(filepath.Join(CacheDir, "gob", fpath))
	MakeDirs(filepath.Dir(savePath))

	fp, err := os.Create(savePath)
	if err != nil {
		PrintError("GobDump:os.Create", err)
		return false
	}
	enc := gob.NewEncoder(fp)
	err = enc.Encode(data)
	PrintError("GobDump:enc.Encode", err)

	return true
}

func GobLoad(fpath string, data any, expireSeconds int64) bool {
	savePath := ToUnixSlash(filepath.Join(CacheDir, "gob", fpath))
	finfo, err := os.Stat(savePath)
	if err != nil {
		return false
	}
	if GetNowUnix()-finfo.ModTime().Unix() > expireSeconds {
		return false
	}
	fp, err := os.Open(savePath)
	if err != nil {
		PrintError("GobLoad:os.Open", err)
		return false
	}
	dec := gob.NewDecoder(fp)
	err = dec.Decode(data)
	if err != nil {
		PrintError("GobLoad:dec.Decode", err)
		return false
	}

	return true
}

func GobTime(fpath string) int64 {
	savePath := ToUnixSlash(filepath.Join(CacheDir, "gob", fpath))
	finfo, err := os.Stat(savePath)
	if err != nil {
		return 0
	}
	return finfo.ModTime().Unix()
}

func GobRemove(fpath string) bool {
	savePath := ToUnixSlash(filepath.Join(CacheDir, "gob", fpath))
	_, err := os.Stat(savePath)
	if err != nil {
		return true
	}
	err = os.Remove(savePath)
	if err != nil {
		return false
	}

	return true
}

func UniqueInts(elements []int) []int {
	m := make(map[int]int)
	var eleNew []int
	for _, v := range elements {
		m[v] = 0
	}
	for k := range m {
		eleNew = append(eleNew, k)
	}
	sort.Ints(eleNew)
	return eleNew
}

func LoadFileBytes(fpath string) []byte {
	data, err := os.ReadFile(fpath)
	if err != nil {
		PrintError("LoadFileBytes", err)
		return nil
	}
	return data
}

func MP4ToAPNG(src, dst string) error {
	if src == "" || dst == "" {
		return nil
	}
	src = ToUnixSlash(src)
	dst = ToUnixSlash(dst)

	tempAPNGShellPath := dst + ".ffmpeg.bat"
	tempAPNGCommand := strings.Join([]string{
		"ffmpeg -n -i \"",
		src,
		"\" -f apng -an -plays 0 -vf \"fps=6, scale=144:-1\" -ss 00:00:02 -t 2 \"",
		dst, "\""}, "")

	DebugInfo("MP4ToAPNG", tempAPNGShellPath, "=>", tempAPNGCommand)
	MakeDirs(filepath.Dir(dst))

	fp, _ := os.OpenFile(tempAPNGShellPath, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0777)
	fp.WriteString(tempAPNGCommand + "\n")
	chanShell <- tempAPNGShellPath

	return nil
}

func Image2Thumb(src, dst string) error {
	if src == "" || dst == "" {
		return nil
	}
	src = ToUnixSlash(src)
	dst = ToUnixSlash(dst)

	tempMagickShellPath := dst + ".magick.bat"
	tempMagickCommand := strings.Join([]string{
		"magick ",
		src,
		" -resize 144x256 -quality 89 ",
		dst}, "")

	DebugInfo("Image2Thumb", tempMagickShellPath, "=>", tempMagickCommand)
	MakeDirs(filepath.Dir(dst))

	fp, _ := os.OpenFile(tempMagickShellPath, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0777)
	fp.WriteString(tempMagickCommand + "\n")
	chanShell <- tempMagickShellPath

	return nil
}

func IndexOr(ukey string, args ...string) bool {
	ukey = strings.ToLower(ukey)
	for _, arg := range args {
		if strings.Index(ukey, strings.ToLower(arg)) >= 0 {
			return true
		}
	}
	return false
}

func MapKeyOrdered2(maps []map[string]int) (ordmaps []map[string]int) {
	var mkeys []string
	for _, kv := range maps {
		for mkey, _ := range kv {
			mkeys = append(mkeys, mkey)
		}
	}
	sort.Sort(sort.StringSlice(mkeys))
	for _, mk := range mkeys {
		for _, kv := range maps {
			if mval, ok := kv[mk]; ok {
				ordmaps = append(ordmaps, map[string]int{mk: mval})
			}
		}
	}
	return ordmaps
}

func MapKeyOrdered(maps []map[string]int) []map[string]int {
	// 提取所有键并去重
	keySet := make(map[string]struct{})
	for _, m := range maps {
		for k := range m {
			keySet[k] = struct{}{}
		}
	}

	// 将键转换为切片并排序
	keys := make([]string, 0, len(keySet))
	for k := range keySet {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// 按排序后的键收集值
	var result []map[string]int
	for _, k := range keys {
		for _, m := range maps {
			if val, exists := m[k]; exists {
				result = append(result, map[string]int{k: val})
			}
		}
	}

	return result
}
