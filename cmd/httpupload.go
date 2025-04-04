package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"

	"io"
	"mime/multipart"
	"os"
	"path/filepath"

	//"sort"
	//"strconv"
	"strings"

	//"github.com/kataras/iris/v12/sessions"
	"github.com/kataras/iris/v12"
)

func apiUploadSchema(ctx iris.Context) {
	var ueSchema UploadEntitySchema
	bf := bytes.NewBuffer([]byte{})
	jsonEncoder := json.NewEncoder(bf)
	jsonEncoder.SetEscapeHTML(false)
	jsonEncoder.Encode(ueSchema)

	DebugInfo("apiUploadSchema", bf.String())
	ctx.Header("Content-Type", "application/json")
	ctx.WriteString(bf.String())
}

func apiHasFiles(ctx iris.Context) {
	if ctx.Method() != "POST" {
		DebugInfo("apiHasFiles", "pls use POST")
		return
	}

	fsums := ctx.PostValue("fsums")
	var reslts []map[string]string

	var jfsums []map[string]string
	err := json.Unmarshal([]byte(fsums), &jfsums)
	if err != nil {
		PrintError("apiHasFiles:json.UnMarshal", err)
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.JSON(reslts)
		return
	}

	for _, kv := range jfsums {
		for k, v := range kv {
			DebugInfo("key", k)
			if badgerExists([]byte(k)) {
				DebugInfo("status", "1")
				m := make(map[string]string)
				m[v] = "1"
				reslts = append(reslts, m)
			} else {
				DebugInfo("status", "0")
				m := make(map[string]string)
				m[v] = "0"
				reslts = append(reslts, m)
			}
		}

	}

	ctx.JSON(reslts)
}

func apiUploadFile(ctx iris.Context) {
	if ctx.Method() != "POST" {
		DebugInfo("apiUploadFile", "pls use POST")
		return
	}

	fid := ctx.PostValue("fid")
	fuser := Normalize(ctx.PostValue("fuser"))
	fapikey := ctx.PostValue("fapikey")
	fmeta := ctx.PostValue("fmeta")
	if IsAnyEmpty(fuser, fapikey, fid) {
		ctx.StatusCode(iris.StatusForbidden)
		ctx.Writef("Error: 403 Forbidden")
		return
	}

	DebugInfo("=====", fuser, "::", fid, "::", fmeta)
	//
	user := mysqlApiKeyLogin(fuser, fapikey, 1)
	if user.ApiKey != fapikey {
		ctx.StatusCode(iris.StatusForbidden)
		ctx.Writef("Error: 403 Forbidden")
		return
	}

	//var success bool
	var meta map[string]string
	if err := json.Unmarshal([]byte(fmeta), &meta); err != nil {
		PrintError("apiUploadFile:json.Unmarshal(fmeta)", err)
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.Writef("Error: 400 BadRequest")
		return
	}

	DebugInfo("=====apiUploadFile:meta", meta, " ==> fsum: ", meta["fsum"])
	entity := NewEntity(fuser, fid)

	var dest string
	var success bool
	_, fileHeader, err := ctx.FormFile("file")
	if err != nil {
		if meta["fsum"] == "" {
			ctx.StatusCode(iris.StatusBadRequest)
			ctx.WriteString(err.Error())
			return
		}
	} else {
		dest = filepath.Join(UploadDir, fuser, fileHeader.Filename)
		MakeDirs(filepath.Join(UploadDir, fuser))
		ctx.SaveFormFile(fileHeader, dest)
		//
		entity = entity.WithFile(dest)
	}

	if entity.Data != nil && meta["fsum"] != "" {
		if string(SumBlake3(entity.Data)) != meta["fsum"] {
			ctx.StatusCode(iris.StatusBadRequest)
			ctx.WriteString("Error: meta[fsum] is invalid")
			return
		}
	}

	for k, v := range meta {
		//DebugInfo("apiUploadFile:entity.WithMeta", k, "=", v)
		entity = entity.WithMeta(k, fmt.Sprintf("%v", v))
	}

	if entity.Data != nil {
		success = entity.Save()
	}

	if meta["fsum"] != "" && entity.Data == nil {
		success = entity.SaveWithoutData()
	}

	if success {
		fmt.Println("apiUploadFile: OK")
	} else {
		fmt.Println("apiUploadFile: FAILED")
	}

}

func uploadFile(ctx iris.Context) {
	if ctx.Method() != "POST" {
		DebugInfo("uploadFile", "pls use POST")
		return
	}
	//
	currentUser := getCurrentUser(ctx)
	DebugInfo("uploadFile:currentUser", currentUser)
	userlogin := currentUser.Name

	fuser := Normalize(ctx.PostValue("fuser"))
	fgroup := Normalize(ctx.PostValue("fgroup"))
	fprefix := Normalize(ctx.PostValue("fprefix"))
	ftags := Normalize(ctx.PostValue("ftags"))

	if userlogin != fuser {
		ctx.Writef("Error: bucket name must be same as username.")
		return
	}

	if IsAnyEmpty(fuser, fgroup) {
		currentPostMaxSize := (ctx.Application().ConfigurationReadOnly().GetPostMaxMemory())
		DebugInfo("uploadFile:currentPostMaxSize", currentPostMaxSize)

		ctx.Writef("Error: username and group cannot be empty, or file size exceeds limit.")
		return
	}

	if strings.HasPrefix(fuser, "_") {
		ctx.Writef("Error: username cannot start with _")
		return
	}

	_, fileHeader, err := ctx.FormFile("file")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.WriteString(err.Error())
		return
	}

	ctx.SetCookieKV("ck_fuser", fuser)
	ctx.SetCookieKV("ck_fgroup", fgroup)
	ctx.SetCookieKV("ck_fprefix", fprefix)
	ctx.SetCookieKV("ck_ftags", ftags)

	// Upload the file to specific destination.
	dest := filepath.Join(UploadDir, userlogin, fileHeader.Filename)
	MakeDirs(filepath.Join(UploadDir, userlogin))
	ctx.SaveFormFile(fileHeader, dest)

	fbasename := Normalize(filepath.Base(dest))
	fkey := ToUnixSlash(filepath.Join(fgroup, fbasename))
	if fprefix != "" {
		fkey = ToUnixSlash(filepath.Join(fgroup, fprefix, fbasename))
	}
	DebugInfo("uploadFile:key", fkey)
	//
	entity := NewEntity(fuser, fkey).WithFile(dest).WithTags(ftags)
	success := entity.Save()
	//
	if !success {
		ctx.Header("Content-Type", "text/plain;charset=utf-8")
		ctx.Writef("Error: %s does not upload:\n %s/%s", fileHeader.Filename, fuser, fkey)
	}

	res := fmt.Sprintf("OK: File: (%s) was uploaded:<br/><br/>File:<br/><a href=\"/f/%s/%s\">/f/%s/%s</a>",
		fileHeader.Filename, fuser, fkey, fuser, fkey)
	res = strings.Join([]string{res, fmt.Sprintf(" | <a href=\"/f/%s/%s\" download>[Download]</a>",
		fuser, fkey)}, "")
	res = strings.Join([]string{res, fmt.Sprintf("<br/><br/>Dir:<br/><a href=\"/user/buckets/%s/%s\">%s/%s</a>",
		fuser, ToUnixSlash(filepath.Dir(fkey)), fuser, ToUnixSlash(filepath.Dir(fkey)))}, "")
	res = strings.Join([]string{res, fmt.Sprintf("For Video Play:<br/><a href=\"/play/v/%s/%s\">/play/v/%s/%s</a>",
		fuser, fkey, fuser, fkey)}, "<br/><br/>")

	ctx.Header("Content-Type", "text/html;charset=utf-8")
	ctx.Write([]byte(res))
}

func apiBatchImport(ctx iris.Context) {

	if ctx.Method() != "POST" {
		DebugInfo("apiBatchImport", "pls use POST")
		return
	}

	fuser := Normalize(ctx.PostValue("fuser"))
	fapikey := ctx.PostValue("fapikey")

	result := iris.Map{}

	DebugInfo("=====", fuser)
	if IsAnyEmpty(fuser, fapikey) {
		ctx.StatusCode(iris.StatusForbidden)
		ctx.JSON(result)
		return
	}
	//
	user := mysqlApiKeyLogin(fuser, fapikey, 1)
	if user.ApiKey != fapikey || user.IsAdmin != 1 {
		ctx.StatusCode(iris.StatusForbidden)
		ctx.JSON(result)
		return
	}
	//ctx.SetMaxRequestBodySize((MaxUploadSizeMB * 11) << 20)
	maxSize := ctx.Application().ConfigurationReadOnly().GetPostMaxMemory()
	DebugInfo("apiBatchImport:maxSize", maxSize)
	err := ctx.Request().ParseMultipartForm(maxSize)
	if err != nil {
		FatalError("apiBatchImport", err)
		ctx.StatusCode(iris.StatusInternalServerError)
		ctx.WriteString(err.Error())
		return
	}

	saveDir := ToUnixSlash(filepath.Join(UploadDir, fuser))
	MakeDirs(saveDir)
	form := ctx.Request().MultipartForm
	files := form.File

	DebugInfo("files count", len(files))
	failures := 0
	var errFiles []string
	var okFiles []string
	for _, file := range files {
		_, err = saveUploadedFile(file[0], saveDir)
		if err != nil {
			failures++
			errFiles = append(errFiles, file[0].Filename)
		} else {
			savePath := strings.Join([]string{saveDir, file[0].Filename}, "/")
			//DebugInfo("OK", savePath)
			okFiles = append(okFiles, savePath)
		}
	}

	if failures > 0 {
		ctx.Writef("%s: failed to upload: %s\n", failures, strings.Join(errFiles, "\n"))
		return
	}

	var rowsOk []iris.Map
	var rowsErr []iris.Map
	for _, okFile := range okFiles {
		fbinData, err := os.ReadFile(okFile)
		if err != nil {
			PrintError("apiBatchImport:os.ReadFile(okFile)", err)
			rowsErr = append(rowsErr, iris.Map{
				"error":  okFile,
				"status": 0})
			continue
		}

		bkey := badgerSave(fbinData)
		if bkey != nil {
			rowsOk = append(rowsOk, iris.Map{
				"blake3sum": string(bkey),
				"status":    1})
			//os.Remove(okFile)
		} else {
			rowsErr = append(rowsErr, iris.Map{
				"error":  okFile,
				"status": 0})
		}

	}

	rows := make(map[string][]iris.Map)
	rows["ok"] = rowsOk
	rows["error"] = rowsErr

	ctx.JSON(rows)

}

func saveUploadedFile(fh *multipart.FileHeader, destDir string) (int64, error) {
	src, err := fh.Open()
	if err != nil {
		return 0, err
	}
	defer src.Close()
	out, err := os.OpenFile(filepath.Join(destDir, fh.Filename),
		os.O_WRONLY|os.O_CREATE, os.ModePerm)
	if err != nil {
		return 0, err
	}
	defer out.Close()
	wb, err := io.Copy(out, src)
	if err != nil {
		return wb, err
	}
	return wb, nil
}
