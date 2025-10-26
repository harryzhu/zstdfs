package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func badgerBulkLoad(dpath string, fext string) bool {
	DebugInfo("badgerBulkLoad:", "Start ...")

	tsStart := GetNowUnix()
	var counter int
	var batchFiles []string

	if MaxUploadSizeMB <= 0 {
		DebugWarn("badgerBulkLoad", "cannot be 0, will use 16 as default")
		MaxUploadSizeMB = 16
	}

	filepath.Walk(dpath, func(path string, finfo os.FileInfo, err error) error {
		PrintSpinner(Int2Str(counter))

		if finfo.IsDir() {
			return nil
		}
		if fext != "*" {
			if strings.ToLower(filepath.Ext(finfo.Name())) != strings.ToLower(fext) {
				return nil
			}
		}
		if strings.HasPrefix(finfo.Name(), ".") {
			return nil
		}
		if finfo.Size() == 0 || finfo.Size() > (MaxUploadSizeMB<<20) {
			return nil
		}

		DebugInfo("badgerBulkLoad", path)
		if len(batchFiles) < 10 {
			batchFiles = append(batchFiles, path)
			counter++
		}
		if len(batchFiles) >= 10 {
			//batchWriteFiles(batchFiles)
			mongoBatchWriteFiles(batchFiles)
			batchFiles = []string{}
		}

		return nil
	})

	//batchWriteFiles(batchFiles)
	mongoBatchWriteFiles(batchFiles)

	DebugInfo("badgerBulkLoad:", "Done! files: ", counter)
	DebugInfo("badgerBulkLoad", fmt.Sprintf("Elapse: %v seconds", (GetNowUnix()-tsStart)))
	return true
}

func batchWriteFiles2(files []string) bool {

	var batchTotalSize int
	var err error
	var fp *os.File
	var val []byte

	for _, file := range files {
		fp, err = os.Open(file)
		if err != nil {
			PrintError("BatchWriteFiles", err)
			return false
		}
		val, err = io.ReadAll(fp)
		if err != nil {
			PrintError("BatchWriteFiles", err)
			return false
		}
		fp.Close()

		_, err := gcSet(val)

		if err != nil {
			PrintError("BatchWriteFiles", err)
			return false
		}
		batchTotalSize += len(val)
	}

	DebugInfo("BatchWriteFiles: files: ", len(files), ", size:", batchTotalSize>>20, " MB")

	return true
}
