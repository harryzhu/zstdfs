package cmd

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	bolt "go.etcd.io/bbolt"
)

func exportFiles(dpath string) {
	MakeDirs(dpath)
	buckets := getAllBuckets()
	var notExportFiles []string

	if len(buckets) > 0 {
		for _, bucket := range buckets {
			if bucket == "" || strings.HasPrefix(bucket, "_") {
				continue
			}
			DebugInfo("exportFiles:Bucket", bucket)
			db.View(func(tx *bolt.Tx) error {
				c := tx.Bucket([]byte(bucket)).Cursor()

				prefix := []byte("")
				var i int

				for k, v := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, v = c.Next() {

					PrintSpinner(strconv.Itoa(i))

					exportPath := strings.Join([]string{dpath, bucket, string(k)}, "/")
					DebugInfo("exportFiles:exportPath", exportPath)
					MakeDirs(filepath.Dir(exportPath))
					fbinVal := fbinGet(v)
					if fbinVal != nil {
						ioutil.WriteFile(exportPath, fbinVal, os.ModePerm)
					}

					finfo, err := os.Stat(exportPath)
					PrintError("exportFiles:os.Stat", err)
					if err == nil && finfo.Size() == 0 {
						PrintError("exportFiles:size", ErrFileSizeZero)
					}

					if err != nil || finfo.Size() == 0 {
						notExportFiles = append(notExportFiles, strings.Join([]string{"EXPORT ERROR:", bucket, string(k)}, ":"))
					}

					i++
				}

				return nil
			})
		}

		// for bucket start with _

		for _, bucket := range buckets {
			if bucket == "" || bucket == "_fbin" || strings.HasPrefix(bucket, "_") == false {
				continue
			}
			DebugInfo("exportFiles:Bucket", bucket)
			fbucket := ToUnixSlash(filepath.Join(dpath, bucket+".csv"))
			fpb, err := os.OpenFile(fbucket, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.ModePerm)
			FatalError("exportFiles", err)
			csvwriter := csv.NewWriter(fpb)

			db.View(func(tx *bolt.Tx) error {
				c := tx.Bucket([]byte(bucket)).Cursor()
				prefix := []byte("")
				var i int

				for k, v := c.Seek(prefix); k != nil && v != nil; k, v = c.Next() {
					PrintSpinner(strconv.Itoa(i))
					line := []string{string(k), string(v)}
					csvwriter.Write(line)

					i++
				}
				return nil
			})
			csvwriter.Flush()
			fpb.Close()

		}
	}

	if len(notExportFiles) > 0 {
		err := ioutil.WriteFile("data/logs/export_error_files-"+GetXxhash([]byte(dpath))+".log", []byte(strings.Join(notExportFiles, "\r\n")), os.ModePerm)
		PrintError("ExportFiles:write log", err)
	}
}

func ImportFiles(dpath, ext, user, group string) error {
	DebugInfo("ImportFiles:dir", dpath)
	DebugInfo("ImportFiles:ext", ext)
	DebugInfo("ImportFiles:user", user)
	DebugInfo("ImportFiles:group", group)
	DebugInfo("ImportFiles:max-upload-size", MaxUploadSize)
	if dpath == "" || ext == "" || user == "" || group == "" {
		DebugWarn("ImportFiles:Param", "dpath, ext, user, group cannot be empty")
		return nil
	}
	var relPath string
	var fdata []byte
	var ignoreFiles []string
	var readyFiles []string
	// for windows
	dpath = ToUnixSlash(dpath)

	filepath.Walk(dpath, func(path string, finfo os.FileInfo, err error) error {
		if finfo.IsDir() {
			return nil
		}
		// for windows
		path = ToUnixSlash(path)

		if ImportIsIgnoreDotFile {
			if strings.HasPrefix(finfo.Name(), ".") {
				ignoreFiles = append(ignoreFiles, "IGNORE:dot-file: "+path)
				return nil
			}
		}

		if ext != "*" {
			if filepath.Ext(strings.ToLower(finfo.Name())) != strings.ToLower(ext) {
				return nil
			}
		}

		if finfo.Size() > MaxUploadSize {
			ignoreFiles = append(ignoreFiles, "IGNORE:oversize: "+path)
			return nil
		}

		if finfo.Size() == 0 {
			ignoreFiles = append(ignoreFiles, "IGNORE:0 size: "+path)
			return nil
		}

		readyFiles = append(readyFiles, path)

		return nil
	})

	if len(ignoreFiles) > 0 {
		loghash := strings.Join([]string{dpath, ext, user, group}, ":")
		err := ioutil.WriteFile("data/logs/import_ignore_files-"+GetXxhash([]byte(loghash))+".log", []byte(strings.Join(ignoreFiles, "\r\n")), os.ModePerm)
		PrintError("ImportFiles:write log", err)
	}

	fcount := len(readyFiles)

	if fcount == 0 {
		DebugInfo("ImportFiles", "no files will be imported")
		return nil
	}

	fmt.Println("ImportFiles:count:", fcount)

	// batch fbin for turbo
	var batchSize int = 500
	if fcount <= batchSize {
		batchSize = fcount
	}

	var epoch int = fcount/batchSize + 1
	var epochFiles []string

	fmt.Println("ImportFiles:epoch:", epoch)
	for i := 0; i < epoch; i++ {
		PrintSpinner(strconv.Itoa(i))

		idFrom := i * batchSize
		idTo := i*batchSize + batchSize
		if idTo > fcount {
			idTo = fcount
		}
		epochFiles = readyFiles[idFrom:idTo]
		if len(epochFiles) > 0 {
			fbinBatchSave(epochFiles, dpath)
		}
		DebugWarn("-----", i, "-----", len(epochFiles), "-----")

	}

	//
	var err error
	for i, fpath := range readyFiles {
		PrintSpinner(strconv.Itoa(i))
		relPath, err = filepath.Rel(dpath, fpath)
		PrintError("ImportFiles:relPath", err)
		relPath = ToUnixSlash(relPath)

		fdata, err = ioutil.ReadFile(fpath)
		if err != nil {
			PrintError("ImportFiles:ReadFile", err)
			continue
		}

		k := dbSave(user, group, relPath, fdata)
		DebugInfo("Saved", strings.Join([]string{fpath, relPath, k}, "=>"))
	}

	fbin.Close()
	return nil
}
