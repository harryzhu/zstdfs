package cmd

import (
	"encoding/json"
	"fmt"

	//"io/ioutil"
	//"mime"
	"os"
	"path/filepath"

	//"sort"
	//"strconv"
	"strings"

	//"github.com/kataras/iris/v12/sessions"
	"github.com/kataras/iris/v12"
)

func apiHasFile(ctx iris.Context) {
	blake3sum := ctx.Params().Get("blake3sum")

	if badgerExists([]byte(blake3sum)) {
		ctx.JSON(iris.Map{"status": 1})
	} else {
		ctx.JSON(iris.Map{"status": 0})
	}

}

func apiUploadFile(ctx iris.Context) {
	if ctx.Method() != "POST" {
		DebugInfo("apiUploadFile", "pls use POST")
		return
	}

	fuser := Normalize(ctx.PostValue("fuser"))
	fapikey := ctx.PostValue("fapikey")
	//
	fid := ctx.PostValue("fid")
	fsum := ctx.PostValue("fsum")
	fmeta := ctx.PostValue("fmeta")
	//
	fgroup := Normalize(ctx.PostValue("fgroup"))
	fprefix := Normalize(ctx.PostValue("fprefix"))
	ftags := Normalize(ctx.PostValue("ftags"))

	DebugInfo("=====", fuser, "::", fgroup, "::", fprefix, "::", ftags, "::", fmeta, "::", fsum)
	//
	user := mysqlApiKeyLogin(fuser, fapikey, 1)
	if user.ApiKey != fapikey {
		ctx.StatusCode(iris.StatusForbidden)
		ctx.Writef("Error: 403 Forbidden")
		return
	}

	var fkey string
	var dest string

	if fsum == "" {
		_, fileHeader, err := ctx.FormFile("file")
		if err != nil {
			ctx.StatusCode(iris.StatusBadRequest)
			return
		}
		//
		dest = filepath.Join(UploadDir, fuser, fileHeader.Filename)
		MakeDirs(filepath.Join(UploadDir, fuser))
		ctx.SaveFormFile(fileHeader, dest)
		//
		fbasename := Normalize(filepath.Base(dest))
		fkey = ToUnixSlash(filepath.Join(fgroup, fbasename))
		if fprefix != "" {
			fkey = ToUnixSlash(filepath.Join(fgroup, fprefix, fbasename))
		}
	}

	if fid != "" {
		fkey = fid
	}

	DebugInfo("apiUploadFile:key", fkey)
	//
	entity := NewEntity(fuser, fkey)
	if fsum == "" && dest != "" {
		_, err := os.Stat(dest)
		if err == nil {
			entity = entity.WithFile(dest)
		}
	}

	var success bool = false
	var meta map[string]any
	err := json.Unmarshal([]byte(fmeta), &meta)
	DebugInfo("=====apiUploadFile:meta", meta)
	if err != nil {
		PrintError("apiUploadFile:json.Unmarshal(fmeta)", err)
	} else {
		for k, v := range meta {
			DebugInfo("apiUploadFile:entity.WithMeta", k, "=", v)
			entity = entity.WithMeta(k, fmt.Sprintf("%v", v))
		}
	}
	DebugInfo("apiUploadFile:fid", fid, ":ID:", entity.ID)

	entity = entity.WithTags(ftags)

	if fsum != "" {
		entity = entity.WithMeta("fsum", fsum)
		success = entity.SaveWithoutData()
	} else {
		success = entity.Save()
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

	ctx.SetMaxRequestBodySize(Params["MaxUploadSize"].(int64))
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
		DebugWarn("MaxUploadSize", Params["MaxUploadSize"].(int64))
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
