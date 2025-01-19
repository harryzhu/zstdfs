package cmd

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"mime"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/middleware/basicauth"
)

var MaxUploadSize int64

func BeforeStart() {
	DebugInfo("MaxUploadSize", MaxUploadSize)
	if MaxUploadSize <= 0 {
		DebugWarn("MaxUploadSize", "<=0, you cannot upload any files")
	}

}

func StartHTTPServer() {
	BeforeStart()

	app := iris.New()
	app.OnErrorCode(iris.StatusNotFound, notFound)
	app.OnErrorCode(iris.StatusInternalServerError, internalServerError)

	tmpl := iris.HTML(embeddedFS, ".html")

	tmpl.RootDir("template")
	tmpl.Delims("{{", "}}")
	tmpl.Reload(IsDebug)
	app.RegisterView(tmpl)

	homeAPI := app.Party("/")
	{
		homeAPI.Use(iris.Compression)

		homeAPI.Get("/", homeIndex)
	}

	bucketsAPI := app.Party("/buckets")
	{
		bucketsAPI.Use(iris.Compression)

		bucketsAPI.Get("/", listBuckets)
		bucketsAPI.Get("/{bucket:string}", listGroups)
		bucketsAPI.Get("/{bucket:string}/{groups:string}/{pagenum:int}", listFiles)
	}

	zstdAPI := app.Party("/z")
	{
		zstdAPI.Use(iris.Compression)

		zstdAPI.Get("/{bucket:string}/{fname:path}", getFiles)
	}

	playAPI := app.Party("/play")
	{
		playAPI.Use(iris.Compression)
		playAPI.HandleDir("/temp", iris.Dir("www/temp"), iris.DirOptions{})

		playAPI.Get("/v/{bucket:string}/{fname:path}", playVideos)
	}

	StaticOptions := iris.DirOptions{
		ShowList: true,
		DirList: iris.DirListRich(iris.DirListRichOptions{
			TmplName: "dirlist.html",
		}),
	}
	app.HandleDir("/static", iris.Dir(StaticDir), StaticOptions)

	adminAPI := app.Party("/admin")

	{
		if AdminUser != "" && AdminPassword != "" {
			auth := basicauth.Default(map[string]string{
				AdminUser: AdminPassword,
			})
			adminAPI.Use(auth)
		}

		adminAPI.Use(iris.Compression)

		adminAPI.Post("/upload", uploadFile)
		adminAPI.Get("/edit", editFiles)
	}

	sysAPI := app.Party("/_stats")

	{
		sysAPI.Use(iris.Compression)
		sysAPI.HandleDir("/_sync", iris.Dir("data/_sync"), StaticOptions)

		sysAPI.Get("/_buckets", sysAllBuckets)
		sysAPI.Get("/_groups/{bucket:string}", sysAllGroups)
		sysAPI.Get("/_keys/{bucket:string}", sysAllBucketKeys)
		sysAPI.Get("/_keys/{bucket:string}/{prefix:string}", sysAllBucketPrefixKeys)
		sysAPI.Get("/_system/{pagenum:int}", sysPagedKvs)
		sysAPI.Get("/_meta/", sysMeta)
	}

	app.Listen(fmt.Sprintf("%s:%d", Host, Port))

}

func notFound(ctx iris.Context) {
	ctx.View("404.html")
}

func internalServerError(ctx iris.Context) {
	ctx.WriteString("Oups something went wrong, try again")
}

func homeIndex(ctx iris.Context) {
	ctx.View("home.html")
}

func listBuckets(ctx iris.Context) {
	var navList []map[string]string
	buckets := getAllBuckets()

	navList = make([]map[string]string, 1)
	for _, bkt := range buckets {
		if bkt != "" && strings.HasPrefix(bkt, "_") != true {
			gb := make(map[string]string, 1)
			gb["api_party"] = "buckets"
			gb["bucket"] = bkt
			navList = append(navList, gb)
		}
	}

	ctx.ViewData("nav_list", navList)
	ctx.View("bucket-list.html")
}

func listGroups(ctx iris.Context) {
	var navList []map[string]string
	bucket := ctx.Params().Get("bucket")

	groups := getAllGroups(bucket)

	navList = make([]map[string]string, 1)
	for _, grp := range groups {
		if grp != "" {
			gb := make(map[string]string, 1)
			gb["api_party"] = "buckets"
			gb["bucket"] = bucket
			gb["group"] = grp
			navList = append(navList, gb)
		}
	}

	ctx.View("group-list.html", iris.Map{
		"bucket":   bucket,
		"nav_list": navList,
	})

}

func listFiles(ctx iris.Context) {
	var navList []map[string]string
	bucket := ctx.Params().Get("bucket")
	group := ctx.Params().Get("groups")
	pagenum, err := strconv.Atoi(ctx.Params().Get("pagenum"))

	if err != nil || pagenum < 1 {
		pagenum = 1
	}
	pageprev := 1
	pagenext := 1

	if pagenum >= 1 {
		pageprev = pagenum - 1
		pagenext = pagenum + 1
	}

	if pageprev < 1 {
		pageprev = 1
	}

	if IsAnyEmpty(bucket, group) {
		return
	}

	files := getAllFiles(bucket, group, pagenum)
	navList = make([]map[string]string, 1)
	for _, f := range files {
		if f != "" {
			gb := make(map[string]string, 1)
			gb["api_party"] = "z"
			gb["bucket"] = bucket
			gb["group"] = group
			gb["fkey"] = f
			navList = append(navList, gb)
		}
	}
	ctx.View("file-list.html", iris.Map{
		"pageprev": pageprev,
		"pagenext": pagenext,
		"bucket":   bucket,
		"group":    group,
		"nav_list": navList,
	})
}

func getFiles(ctx iris.Context) {
	bucket := ctx.Params().Get("bucket")
	fname := ctx.Params().Get("fname")

	b := dbGet(bucket, fname)
	blen := len(b)
	if blen == 0 {
		ctx.NotFound()
		return
	}

	fext := filepath.Ext(fname)
	flen := strconv.Itoa(blen)

	DebugInfo("getFiles", fext, ",size:", flen)
	mimeType := "text/plain"

	if fext != "" {
		mimeType = mime.TypeByExtension(fext)
	}
	ctx.Header("Content-Type", mimeType)
	ctx.Write(b)
}

func playVideos(ctx iris.Context) {
	rand.Seed(time.Now().Unix())
	randClean := rand.Intn(100)
	DebugInfo("playVideos:randClean", randClean)
	if randClean%10 == 0 {
		DebugInfo("playVideos:randClean", "run temp dir clean")
		go func() {
			filepath.Walk("www/temp/", func(path string, finfo os.FileInfo, err error) error {
				if finfo.IsDir() {
					return nil
				}

				if strings.HasPrefix(finfo.Name(), ".") {
					return nil
				}

				tNow := time.Now()
				tAge := tNow.Sub(finfo.ModTime()).Seconds()

				if tAge > 1800.0 {
					os.Remove(path)
					DebugInfo("playVideos:remove expired file", tAge, ": ", path)
				}
				return nil
			})
		}()
	}

	//

	bucket := ctx.Params().Get("bucket")
	fname := ctx.Params().Get("fname")
	b := dbGet(bucket, fname)
	blen := len(b)
	if blen == 0 {
		ctx.NotFound()
		return
	}
	fkey := strings.Join([]string{"www", "temp", bucket, fname}, "/")
	MakeDirs(filepath.Dir(fkey))
	_, err := os.Stat(fkey)
	if err != nil {
		ioutil.WriteFile(fkey, b, os.ModePerm)
	}

	fext := filepath.Ext(fname)
	mimeType := "video/mp4"

	if fext != "" {
		mimeType = mime.TypeByExtension(fext)
	}

	video_src := strings.Join([]string{"/play", "temp", bucket, fname}, "/")
	ctx.View("player.html", iris.Map{
		"video_src":  video_src,
		"video_mime": mimeType,
	})
}

func uploadFile(ctx iris.Context) {
	ctx.SetMaxRequestBodySize(MaxUploadSize)
	fuser := Normalize(ctx.PostValue("fuser"))
	fgroup := Normalize(ctx.PostValue("fgroup"))
	fprefix := Normalize(ctx.PostValue("fprefix"))

	if IsAnyEmpty(fuser, fgroup) {
		DebugWarn("MaxUploadSize", MaxUploadSize)
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

	// Upload the file to specific destination.
	dest := filepath.Join(UploadDir, fileHeader.Filename)
	ctx.SaveFormFile(fileHeader, dest)

	data, err := ioutil.ReadFile(dest)
	PrintError("uploadFile", err)

	fkey := Normalize(filepath.Base(dest))
	if fprefix != "" {
		fkey = strings.Join([]string{fprefix, fkey}, "/")
	}
	if IsAnyEmpty(fuser, fgroup, fkey) == false {
		insertKey := dbSave(fuser, fgroup, fkey, data)
		res := "ERROR: cannot save into db, or the db is in READ-ONLY mode."
		if insertKey != "" {
			res = fmt.Sprintf("OK: File: (%s) was uploaded:<br/><br/><br/><a href=\"/z/%s\">%s</a>",
				fileHeader.Filename, insertKey, "/z/"+insertKey)
			res = strings.Join([]string{res, fmt.Sprintf(" | <a href=\"/z/%s\" download>[Download]</a>",
				"/z/"+insertKey)}, "")

			res = strings.Join([]string{res, fmt.Sprintf("For Video Play:<br/><a href=\"/play/v/%s\">%s</a>",
				insertKey, "/play/v/"+insertKey)}, "<br/><br/>")
		}
		go func(rmpath string) {
			os.Remove(rmpath)
		}(dest)
		ctx.Header("Content-Type", "text/html;charset=utf-8")
		ctx.Write([]byte(res))
	} else {
		ctx.Header("Content-Type", "text/plain;charset=utf-8")
		ctx.Writef("Error: %s does not upload:\n %s/%s/%s", fileHeader.Filename, fuser, fgroup, fkey)
	}

}

func editFiles(ctx iris.Context) {
	ckfuser := ""
	if ctx.GetCookie("ck_fuser") != "" {
		ckfuser = ctx.GetCookie("ck_fuser")
	}

	ckfgroup := ""
	if ctx.GetCookie("ck_fgroup") != "" {
		ckfgroup = ctx.GetCookie("ck_fgroup")
	}

	ckfprefix := ""
	if ctx.GetCookie("ck_fprefix") != "" {
		ckfprefix = ctx.GetCookie("ck_fprefix")
	}

	ctx.View("upload-form.html", iris.Map{
		"form_action":             "/admin/upload",
		"form_max_upload_size_mb": strconv.Itoa(MaxUploadSizeMB),
		"ck_fuser":                ckfuser,
		"ck_fgroup":               ckfgroup,
		"ck_fprefix":              ckfprefix,
	})
}

func sysAllBuckets(ctx iris.Context) {
	buckets := getAllBuckets()

	ctx.JSON(buckets)
}

func sysAllGroups(ctx iris.Context) {
	bkt := ctx.Params().Get("bucket")
	groups := getAllGroups(bkt)

	ctx.JSON(groups)
}

func sysAllBucketKeys(ctx iris.Context) {
	bkt := ctx.Params().Get("bucket")
	if bkt == "" {
		ctx.Writef("Error: %s can not be empty", "param bucket cannot be empty")
		return
	}

	ctx.JSON(getAllKeys(bkt, ""))
}

func sysAllBucketPrefixKeys(ctx iris.Context) {
	bkt := ctx.Params().Get("bucket")
	pre := ctx.Params().Get("prefix")
	if IsAnyEmpty(bkt, pre) {
		ctx.Writef("Error: %s can not be empty", "param bucket and prefix cannot be empty")
		return
	}

	ctx.JSON(getAllKeys(bkt, pre))
}

func sysPagedKvs(ctx iris.Context) {
	pagenum, err := strconv.Atoi(ctx.Params().Get("pagenum"))
	PrintError("sysPagedKvs:pagenum", err)
	kvs := dbPagedSys(pagenum)
	ctx.JSON(kvs)
}

func sysMeta(ctx iris.Context) {
	kvs := getMeta("")
	ctx.JSON(kvs)
}
