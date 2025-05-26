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
		ctx.StatusCode(iris.StatusForbidden)
		ctx.Writef("Error: 403 Forbidden")
		return
	}

	DebugInfo("=====", fuser, "::", fid, "::", fmeta)
	//
	user := mysqlAPIKeyLogin(fuser, fapikey, 1)
	if user.APIKey != fapikey {
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
		DebugInfo("apiUploadFiles", "fmeta[sha256] cannot be empty")
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.Writef("Error: 400 BadRequest")
		return
	}

	DebugInfo("=====apiUploadFiles:meta", meta)
	entity := NewEntity(fuser, fid)

	var dest string
	var success bool
	_, fileHeader, err := ctx.FormFile("file")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.WriteString(err.Error())
		return
	}
	dest = filepath.Join(UploadDir, fuser, fileHeader.Filename)
	MakeDirs(filepath.Join(UploadDir, fuser))
	ctx.SaveFormFile(fileHeader, dest)
	entity = entity.WithFile(dest)

	if entity.Data != nil {
		if SHA256Bytes(entity.Data) != meta["fsha256"] {
			ctx.StatusCode(iris.StatusBadRequest)
			ctx.WriteString("Error: meta[fsha256] is invalid")
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

	if success {
		fmt.Println("apiUploadFiles: OK")
		ctx.StatusCode(iris.StatusOK)
		ctx.Writef("OK")
	} else {
		fmt.Println("apiUploadFiles: FAILED")
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
	entity := NewEntity(fuser, fkey).WithFile(dest).WithTags(ftags).WithMeta("fsha256", SHA256File(dest))
	success := entity.Save()
	//
	if !success {
		ctx.Header("Content-Type", "text/plain;charset=utf-8")
		ctx.Writef("Error: %s does not upload:\n %s/%s", fileHeader.Filename, fuser, fkey)
		return
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
