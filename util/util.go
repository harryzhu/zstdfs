package util

import (
	"crypto/md5"
	//"encoding/hex"
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io/ioutil"
	"log"
	"mime"
	"regexp"

	//"net/url"
	"os"
	"path/filepath"
	"strings"
)

func GetEnv(key string, defaultValue string) (val string) {
	v := os.Getenv(key)
	if v != "" {
		return v
	}
	return defaultValue
}

func MD5(b []byte) string {
	return fmt.Sprintf("%x", md5.Sum(b))
}

func ErrorString(s string) error {
	return errors.New(s)
}

func GetMimeByName(n string) string {
	return mime.TypeByExtension(filepath.Ext(filepath.Base(n)))
}

func GetFileInfoByte(fpath string) (os.FileInfo, []byte, error) {
	finfo, err := os.Stat(fpath)
	if err != nil {
		log.Println(err)
		return nil, nil, err
	}

	fdata, err := ioutil.ReadFile(fpath)
	if err != nil {
		log.Println(err)
		return finfo, nil, err
	}

	return finfo, fdata, nil
}

func SaveFile(fpath string, data []byte) error {
	return ioutil.WriteFile(fpath, data, 0755)
}

func DirWalker(path_dir string, filter string) (files []string) {
	var filelist = make([]string, 0, 10)
	abspath, _ := filepath.Abs(path_dir)
	filepath.Walk(abspath, func(abspath string, f os.FileInfo, err error) error {
		if f == nil {
			return err
		}
		if f.IsDir() {
			return nil
		}
		if strings.HasPrefix(filepath.Base(abspath), ".") {
			return nil
		}
		matched, _ := filepath.Match(filter, filepath.Base(abspath))
		if matched == true {
			filelist = append(filelist, abspath)
		}

		return nil
	})

	return filelist
}

func FormatTags(tagStr string) []string {
	ts := strings.ReplaceAll(tagStr, ";", ",")
	ts = strings.Trim(ts, ",")
	ts = strings.TrimSpace(ts)

	var arrResult []string
	arrTS := strings.Split(ts, ",")
	for _, v := range arrTS {
		if v != "" {
			arrResult = append(arrResult, v)
		}
	}
	return arrResult
}

func GetImageWidthHeight(fpath string) (w, h int) {
	img, _ := os.Open(fpath)

	c, _, err := image.DecodeConfig(img)
	if err != nil {
		return 0, 0
	}
	w = c.Width
	h = c.Height

	img.Close()
	return w, h
}

func ValidateANS(s string) bool {
	reg := regexp.MustCompile(`[\/\\\s]+`)
	return !reg.MatchString(s)
}
