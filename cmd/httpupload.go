package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

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

func apiUploadFiles(ctx iris.Context) {
	if ctx.Method() != "POST" {
		DebugInfo("apiUploadFiles", "pls use POST")
		return
	}

	fid := ctx.PostValue("fid")
	fuser := Normalize(ctx.PostValue("fuser"))
	fapikey := ctx.PostValue("fapikey")
	fmeta := ctx.PostValue("fmeta")
	if IsAnyEmpty(fuser, fapikey, fid) {
		DebugWarn("apiUploadFiles.10", "fuser/fapikey/fid cannot be empty")
		ctx.StatusCode(iris.StatusForbidden)
		ctx.Writef("Error: 403 Forbidden")
		return
	}

	_, fileHeader, err := ctx.FormFile("file")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.WriteString(err.Error())
		return
	}

	var dest string
	var success bool
	dest = filepath.Join(UploadDir, fuser, fileHeader.Filename)
	MakeDirs(filepath.Join(UploadDir, fuser))
	ctx.SaveFormFile(fileHeader, dest)
	//
	finfo, err := os.Stat(dest)
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.WriteString(err.Error())
		return
	}
	if finfo.Size() > MaxUploadSizeMB<<20 {
		errSize := fmt.Sprintf("File is too large. Max size limit: %v, But Uploaded size: %v", MaxUploadSizeMB<<20, finfo.Size())
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.WriteString(errSize)
		return
	}

	DebugInfo("=====", fuser, "::", fid, "::", fmeta)
	//
	user := mysqlAPIKeyLogin(fuser, fapikey, 1)
	if user.APIKey != fapikey {
		DebugWarn("apiUploadFiles.20", "user.APIKey: ", user.APIKey, ", fapikey: ", fapikey)
		ctx.StatusCode(iris.StatusForbidden)
		ctx.Writef("Error: 403 Forbidden")
		return
	}

	//var success bool
	var meta map[string]string
	if err := json.Unmarshal([]byte(fmeta), &meta); err != nil {
		PrintError("apiUploadFiles:json.Unmarshal(fmeta)", err)
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.Writef("Error: 400 BadRequest")
		return
	}
	if meta["fsha256"] == "" {
		DebugWarn("apiUploadFiles", "fmeta[sha256] cannot be empty")
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.Writef("Error: 400 BadRequest")
		return
	}

	DebugInfo("=====apiUploadFiles:meta", meta)
	entity := NewEntity(fuser, fid)

	entity = entity.WithFile(dest)

	if entity.Data != nil {
		if SHA256Bytes(entity.Data) != meta["fsha256"] {
			ctx.StatusCode(iris.StatusBadRequest)
			ctx.WriteString("Error: meta[fsha256] is invalid")
			return
		}
	}

	for k, v := range meta {
		entity = entity.WithMeta(k, fmt.Sprintf("%v", v))
	}

	if entity.Data != nil {
		success = entity.Save()
	}

	if success {
		DebugInfo("apiUploadFiles:Upload", "OK")
		ctx.StatusCode(iris.StatusOK)
		ctx.Writef("OK")
	} else {
		DebugWarn("apiUploadFiles:Upload", "FAILED")
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.Writef("Error")
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

	data := iris.Map{}

	if userlogin != fuser {
		data["error_message"] = "Error: bucket name must be same as username."
		ctx.View("message-upload.html", data)
		return
	}

	if IsAnyEmpty(fuser, fgroup) {
		data["error_message"] = "Error: username and group cannot be empty, or file size exceeds limit."
		ctx.View("message-upload.html", data)
		return
	}

	if strings.HasPrefix(fuser, "_") {
		data["error_message"] = "Error: username cannot start with _"
		ctx.View("message-upload.html", data)
		return
	}

	_, fileHeader, err := ctx.FormFile("file")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		data["error_message"] = err.Error()
		ctx.View("message-upload.html", data)
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

	finfo, err := os.Stat(dest)
	if err != nil {
		data["error_message"] = err.Error()
		ctx.View("message-upload.html", data)
		return
	}

	if finfo.Size() > MaxUploadSizeMB<<20 {
		//currentPostMaxSize := (ctx.Application().ConfigurationReadOnly().GetPostMaxMemory())
		//DebugWarn("uploadFile:currentPostMaxSize", currentPostMaxSize)
		data["error_message"] = fmt.Sprintf("File is too large. Maxsize limit: %v, But Uploaded size: %v", MaxUploadSizeMB<<20, finfo.Size())
		ctx.View("message-upload.html", data)
		return
	}

	fbasename := Normalize(filepath.Base(dest))
	fkey := ToUnixSlash(filepath.Join(fgroup, fbasename))
	if fprefix != "" {
		fkey = ToUnixSlash(filepath.Join(fgroup, fprefix, fbasename))
	}
	DebugInfo("uploadFile:key", fkey)
	//
	entity := NewEntity(fuser, fkey).WithFile(dest).WithTags(ftags).WithMeta("fsha256", SHA256File(dest))
	success := entity.Save()
	//
	ctx.Header("Content-Type", "text/html;charset=utf-8")

	if !success {
		data["error_message"] = strings.Join([]string{
			"Error: ",
			fileHeader.Filename,
			" does not upload: ",
			fuser,
			"/",
			fkey}, "")
		ctx.View("message-upload.html", data)
		return
	}

	data["url_file"] = fmt.Sprintf("/f/%s/%s", fuser, fkey)
	data["url_dir"] = fmt.Sprintf("/user/buckets/%s/%s", fuser, ToUnixSlash(filepath.Dir(fkey)))
	data["url_play"] = fmt.Sprintf("/play/v/%s/%s", fuser, fkey)

	data["site_url"] = GetSiteURL()

	ctx.View("message-upload.html", data)
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
