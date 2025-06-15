package cmd

import (
	stdContext "context"
	"fmt"
	"mime"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"time"

	"strings"

	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/kataras/iris/v12"
)

var app *iris.Application

func StartHTTPServer() {

	app = iris.New()
	//app.IsDebug()

	tmpl := iris.HTML(embeddedFS, ".html")

	tmpl.RootDir("template")
	tmpl.Delims("{{", "}}")
	tmpl.Reload(IsDebug)
	app.RegisterView(tmpl)

	StaticOptions := iris.DirOptions{
		ShowList: true,
		DirList: iris.DirListRich(iris.DirListRichOptions{
			TmplName: "dirlist.html",
		}),
	}
	app.HandleDir("/static", iris.Dir(StaticDir), StaticOptions)

	homeAPI := app.Party("/")
	{
		homeAPI.Use(iris.Compression)
		homeAPI.Get("/", guestIndex)
		homeAPI.Get("/home", homeIndex)
		homeAPI.Get("/signup", signupIndex)
		homeAPI.Post("/usersignup", userSignup)
		homeAPI.Get("/signin", loginIndex)
		homeAPI.Post("/userlogin", userLogin)
		homeAPI.Get("/logout", logoutIndex)
		homeAPI.HandleDir("/assets", iris.Dir(AssetDir), iris.DirOptions{ShowList: false})
	}

	userFileAPI := app.Party("/f/")
	{
		userFileAPI.Use(iris.Compression)
		userFileAPI.Get("/{uname:string}/{key:path}", getFiles)
	}

	thumbAPI := app.Party("/thumb/")
	{
		thumbAPI.Use(iris.Compression)
		thumbAPI.Get("/{uname:string}/{key:path}", thumbFiles)
	}

	shareAPI := app.Party("/s/")
	{
		shareAPI.Use(iris.Compression)
		shareAPI.Get("/{uname:string}/{val:string}", getFiles)
	}

	playAPI := app.Party("/play")
	{
		playAPI.Use(iris.Compression)
		playAPI.HandleDir("/temp", iris.Dir(TempDir), iris.DirOptions{})

		playAPI.Get("/v/{bucket:string}/{fname:path}", playVideos)
	}

	userAPI := app.Party("/user/")
	{
		userAPI.Use(iris.Compression)

		userAPI.Get("/buckets", adminListBuckets)
		userAPI.Get("/buckets/{uname:string}", adminListGroup)
		userAPI.Get("/buckets/{uname:string}/{fkey:path}", adminListKeys)
		//
		userAPI.Post("/upload", uploadFile)
		userAPI.Get("/edit", editFiles)
		userAPI.Get("/likes/{dotcolor:string}", likeFiles)
		userAPI.Get("/tags", tagList)
		userAPI.Post("/tags", tagList)
		userAPI.Get("/tags/{tagname:string}", tagFiles)
		userAPI.Get("/tags/top", topTags)
		userAPI.Get("/caption", captionIndex)
		userAPI.Post("/caption", captionIndex)
		userAPI.Get("/caption/top/{lang:string}", topCaption)
		userAPI.Get("/caption/{lang:string}/{captionword:string}", captionFiles)
		userAPI.Get("/by/{key:string}/{val:path}", byKeyFiles)
		userAPI.Get("/top/{countname:string}/{min:int}/{max:int}", topCountFiles)
		userAPI.Get("/samefiles", adminSameFiles)
		//
		userAPI.Get("/playlist", playList)
		userAPI.Get("/playlist/{prefix:path}", playPrefixList)
		userAPI.Get("/playlikes/{color:string}", playDotColorList)
		//
		userAPI.Get("/dot/{color:string}/{uname:string}/{fname:path}", dotColorFile)
		//
		userAPI.Get("/cache", bcacheItems)

	}

	v1API := app.Party("/api/")
	{
		v1API.Get("/upload/schema.json", apiUploadSchema)
		v1API.Post("/upload", apiUploadFiles)
		v1API.Post("/list/tags", apiListTags)
		v1API.Post("/list/caption", apiListCaption)
	}

	if MaxUploadSizeMB <= 0 {
		DebugWarn("StartHTTPServer: --max-upload-size-mb ", "cannot be 0, will use 16 as default")
		MaxUploadSizeMB = 16
	}

	app.Listen(fmt.Sprintf("%s:%s", Host, Port),
		iris.WithDynamicHandler,
		iris.WithPostMaxMemory((MaxUploadSizeMB+1)<<20))

}

func StopHTTPServer() {
	fmt.Println("stopping the server ...")
	close(chanShell)
	appShutdown()
	sqldb.Close()
	bgrdb.Close()
	bcache.Close()
}

func appShutdown() {
	ctx, cancel := stdContext.WithTimeout(stdContext.Background(), 3*time.Second)
	defer cancel()
	// close all hosts
	err := app.Shutdown(ctx)
	if err != nil {
		PrintError("appShutdown:", err)
	}
}

func notFound(ctx iris.Context) {
	ctx.StatusCode(iris.StatusNotFound)
	ctx.View("404.html")
}

func guestIndex(ctx iris.Context) {
	ctx.View("home.html", iris.Map{})
}

func homeIndex(ctx iris.Context) {
	currentUser := getCurrentUser(ctx)
	DebugInfo("Home:currentUser", currentUser)

	data := iris.Map{
		"current_user": currentUser.Name,
	}

	if currentUser.IsAdmin == 1 {
		data["is_admin"] = true
	}

	ctx.View("home.html", data)
}

func topCountFiles(ctx iris.Context) {
	min := ctx.Params().Get("min")
	max := ctx.Params().Get("max")
	countname := ctx.Params().Get("countname")
	if IsAnyEmpty(countname, min, max) {
		return
	}

	currentUser := getCurrentUser(ctx)
	DebugInfo("topDiggFiles:currentUser", currentUser)

	uname := currentUser.Name

	files := mongoAggCountByKey(uname, countname, Str2Int(min), Str2Int(max))

	navFileList := genNavFileList(files, "", uname)

	var navBreadcrumb []map[string]string
	bc := make(map[string]string)
	bckey := strings.Join([]string{"Top(", countname, ")[ ", min, " ~ ", max, " ]"}, "")
	bc[bckey] = fmt.Sprintf("/user/top/%s/%s/%s", countname, min, max)
	navBreadcrumb = append(navBreadcrumb, bc)

	DebugInfo("topDiggFiles", navFileList)

	data := iris.Map{
		"nav_file_list":  navFileList,
		"nav_breadcrumb": navBreadcrumb,
		"current_user":   uname,
		"current_prefix": "",
		"site_url":       GetSiteURL(),
	}

	if len(files) > 0 {
		data["file_count"] = Int2Str(len(files))
	}

	if len(files) < 2001 && len(files) > 0 {
		data["enable_thumbnail"] = true
	}

	if currentUser.IsAdmin == 1 {
		data["is_admin"] = true
	}

	ctx.View("top-list.html", data)
}

func getFiles(ctx iris.Context) {
	uname := ctx.Params().Get("uname")
	key := ctx.Params().Get("key")
	val := ctx.Params().Get("val")
	DebugInfo("getFiles:key", key)
	DebugInfo("getFiles:val", val)

	if val != "" && key == "" {
		meta := mongoGetByKey(uname, "uri", val)

		keyID, ok := meta["_id"]
		if !ok {
			DebugInfo("getURIFiles:fsum is empty: uri", val)
			ctx.StatusCode(iris.StatusNotFound)
			ctx.View("404.html")
			return
		}
		key = keyID
	}

	fext := filepath.Ext(key)
	fname := filepath.Join("f", uname, key)

	entity := NewEntity(uname, key).Head()
	//DebugInfo("getFiles", "entity.meta", entity.Meta)
	if Str2Int(entity.Meta["is_public"]) == 0 {
		currentUser := getCurrentUser(ctx)
		if currentUser.Name != uname || currentUser.IsAdmin != 1 {
			ctx.StatusCode(iris.StatusForbidden)
			ctx.View("403.html")
			return
		}
	}

	if Str2Int(entity.Meta["is_ban"]) == 0 {
		DebugInfo("getFiles:is_ban==1", fname)
		currentUser := getCurrentUser(ctx)
		if currentUser.IsAdmin != 1 {
			ctx.StatusCode(iris.StatusNotFound)
			ctx.View("404.html")
			return
		}
	}

	if entity.Meta["_fsum"] == "" {
		DebugInfo("getFiles:fsum is empty", fname)
		ctx.StatusCode(iris.StatusNotFound)
		ctx.View("404.html")
		return
	}

	entity = entity.Get()
	if entity.Data == nil {
		DebugInfo("getFiles: Data is empty", fname)
	}

	mimeType := "application/octet-stream"
	if entity.Meta["mime"] != "" {
		mimeType = entity.Meta["mime"]
	} else if fext != "" {
		mimeType = mime.TypeByExtension(fext)
	}

	ctx.Header("X-Powered-By", "zstdfs")
	if IsDebug {
		ctx.Header("Cache-Control", "public, max-age=0")
	} else {
		ctx.Header("Cache-Control", "public, max-age=86400")
	}

	ctx.Header("Content-Type", mimeType)
	ctx.StatusCode(iris.StatusOK)
	ctx.Write(entity.Data)
}

func thumbFiles(ctx iris.Context) {
	uname := ctx.Params().Get("uname")
	key := ctx.Params().Get("key")
	DebugInfo("thumbFiles:key", key)

	ctx.Header("X-Powered-By", "zstdfs")

	if uname == "" || key == "" {
		ctx.StatusCode(iris.StatusNotFound)
		ctx.View("404.html")
		return
	}

	fext := strings.ToLower(filepath.Ext(key))
	fname := filepath.Join("thumb", uname, key)

	entity := NewEntity(uname, key).Head()
	//DebugInfo("thumbFiles", "entity.meta", entity.Meta)

	if Str2Int(entity.Meta["is_public"]) == 0 {
		currentUser := getCurrentUser(ctx)
		if currentUser.Name != uname || currentUser.IsAdmin != 1 {
			ctx.StatusCode(iris.StatusForbidden)
			ctx.Header("Content-Type", "image/png")
			ctx.Header("Cache-Control", "public, max-age=0")
			ctx.Write(bin403Logo)
			return
		}
	}

	if Str2Int(entity.Meta["is_ban"]) == 1 {
		DebugInfo("thumbFiles:is_ban==1", fname)
		currentUser := getCurrentUser(ctx)
		if currentUser.IsAdmin != 1 {
			ctx.StatusCode(iris.StatusForbidden)
			ctx.Header("Content-Type", "image/png")
			ctx.Header("Cache-Control", "public, max-age=0")
			ctx.Write(binBannedLogo)
			return
		}
	}

	if entity.Meta["_fsum"] == "" {
		DebugInfo("thumbFiles:fsum is empty", fname)
		ctx.StatusCode(iris.StatusInternalServerError)
		ctx.Header("Content-Type", "image/png")
		ctx.Header("Cache-Control", "public, max-age=0")
		ctx.Write(bin500Logo)
		return
	}

	mimeType := "application/octet-stream"
	extImages := []string{".jpg", ".jpeg", ".png", ".bmp", ".gif", ".webp"}
	if fext == ".mp4" {
		mimeType = "image/apng"
	} else {
		mimeType = entity.Meta["mime"]
	}

	thumbImage := filepath.Join(ThumbDir, uname, key)
	thumbImage = ToUnixSlash(thumbImage)
	DebugInfo(thumbImage)
	_, err := os.Stat(thumbImage)
	if err != nil {
		fextDefaultLogo := strings.Join([]string{"thumb_logo_", strings.Trim(fext, "."), ".png"}, "")
		fextDefaultPath := filepath.Join(AssetDir, fextDefaultLogo)
		DebugInfo("thumbFiles:fextDefaultPath", fextDefaultPath)
		if !Contains(extImages, fext) && fext != ".mp4" {
			ctx.Header("Content-Type", "image/png")
			_, err := os.Stat(fextDefaultPath)
			if err != nil {
				ctx.Write(binFileDocumentLogo)
			} else {
				ctx.Write(LoadFileBytes(fextDefaultPath))
			}

			return
		}
		entity = entity.Get()
		if entity.Data == nil {
			DebugInfo("thumbFiles: Data is empty", fname)
			ctx.Header("Content-Type", "image/png")
			ctx.Header("Cache-Control", "public, max-age=0")
			ctx.Write(bin404Logo)
			return
		}

		mp4Temp := filepath.Join(TempDir, "thumb", uname, key)
		mp4Temp = ToUnixSlash(mp4Temp)

		MakeDirs(filepath.Dir(thumbImage))
		MakeDirs(filepath.Dir(mp4Temp))

		fp, _ := os.OpenFile(mp4Temp, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0777)
		fp.Write(entity.Data)
		fp.Close()
		if fext == ".mp4" {
			MP4ToAPNG(mp4Temp, thumbImage)
		}

		if Contains(extImages, fext) {
			Image2Thumb(mp4Temp, thumbImage)
		}
	}

	ctx.Header("Content-Type", mimeType)
	ctx.StatusCode(iris.StatusOK)
	data, err := os.ReadFile(thumbImage)
	if err != nil {
		DebugWarn("thumbFiles:", thumbImage)
		DebugWarn("thumbFiles", err.Error())
		ctx.Header("Cache-Control", "public, max-age=0")
		ctx.Write(bin404Logo)
		return
	}

	if IsDebug {
		ctx.Header("Cache-Control", "public, max-age=0")
	} else {
		ctx.Header("Cache-Control", "public, max-age=3600")
	}

	if len(data) == 0 {
		ctx.Write(binEmptyLogo)
	} else {
		ctx.Write(data)
	}

	return
}

func editFiles(ctx iris.Context) {
	currentUser := getCurrentUser(ctx)
	DebugInfo("editFiles:currentUser", currentUser)
	userlogin := currentUser.Name

	if userlogin == "" {
		return
	}

	ckfgroup := ""
	if ctx.GetCookie("ck_fgroup") != "" {
		ckfgroup = ctx.GetCookie("ck_fgroup")
	}

	ckfprefix := ""
	if ctx.GetCookie("ck_fprefix") != "" {
		ckfprefix = ctx.GetCookie("ck_fprefix")
	}

	ckftags := ""
	if ctx.GetCookie("ck_ftags") != "" {
		ckftags = ctx.GetCookie("ck_ftags")
	}

	currentPostMaxSize := (ctx.Application().ConfigurationReadOnly().GetPostMaxMemory())
	DebugInfo("editFiles:currentPostMaxSize", currentPostMaxSize)

	data := iris.Map{
		"form_action":             "/user/upload",
		"form_max_upload_size_mb": (currentPostMaxSize >> 20) - 1,
		"current_user":            currentUser.Name,
		"ck_fuser":                userlogin,
		"ck_fgroup":               ckfgroup,
		"ck_fprefix":              ckfprefix,
		"ck_ftags":                ckftags,
	}

	if currentUser.IsAdmin == 1 {
		data["is_admin"] = true
	}

	ctx.View("upload-form.html", data)

}

func topTags(ctx iris.Context) {
	currentUser := getCurrentUser(ctx)
	DebugInfo("topTags:currentUser", currentUser)
	var uname string
	if currentUser.Name == "" {
		return
	}
	uname = currentUser.Name
	var counts []int
	countsCacheFile := fmt.Sprintf("%s/topTags/counts.dat", uname)

	nameNum := make(map[string]int)
	nameNumCacheFile := fmt.Sprintf("%s/topTags/nameNum.dat", uname)

	cacheTime := ""

	t1 := GetNowUnix()

	if GobLoad(countsCacheFile, &counts, FunctionCacheExpires) == false || GobLoad(nameNumCacheFile, &nameNum, FunctionCacheExpires) == false {
		userTagCount := mongoTagList(uname, "")
		for ukey, uval := range userTagCount {
			if uval > 9 {
				nameNum[ukey] = uval
				counts = append(counts, uval)
			}

		}
		sort.Ints(counts)
		GobDump(countsCacheFile, counts)
		GobDump(nameNumCacheFile, nameNum)
	} else {
		if GobTime(countsCacheFile) != 0 {
			cacheCurrent := UnixFormat(GobTime(countsCacheFile), "01-02 15:04")
			cacheRefreshSeconds := FunctionCacheExpires - (GetNowUnix() - GobTime(countsCacheFile))
			cacheTime = fmt.Sprintf("Update: %s. Will refresh after %v seconds", cacheCurrent, cacheRefreshSeconds)
		}

	}

	t2 := GetNowUnix()

	//DebugWarn("counts", counts, " ===> Elapse: ", (t2 - t1), " seconds")
	//DebugWarn("nameNum", nameNum, " ===> Elapse: ", (t2 - t1), " seconds")
	uqcounts := UniqueInts(counts)
	DebugInfo("topTags:uqCounts", uqcounts, "; length: ", len(uqcounts), " ===> Elapse: ", (t2 - t1), " seconds")
	DebugInfo("topTags:cacheTime", cacheTime)

	slots := len(uqcounts) / 10
	DebugWarn("topTags:slots", slots)

	viewData := iris.Map{}

	if slots > 9 {
		var groups []int
		for i := 0; i < 10; i++ {
			if i == 9 {
				groups = append(groups, uqcounts[len(uqcounts)-1])
			} else {
				groups = append(groups, uqcounts[i*slots])
			}

		}
		DebugWarn("topTags:groups", groups)

		g0 := make(map[string]int)
		g12 := make(map[string]int)
		g345 := make(map[string]int)
		g678 := make(map[string]int)
		g910 := make(map[string]int)

		if groups[0] < 10 && groups[2] > 10 {
			groups[0] = 9
		}

		for k, v := range nameNum {
			if v <= groups[0] {
				g0[k] = v
			}
			if v > groups[0] && v <= groups[2] {
				g12[k] = v
			}
			if v > groups[2] && v <= groups[5] {
				g345[k] = v
			}
			if v > groups[5] && v <= groups[8] {
				g678[k] = v
			}
			if v > groups[8] && v <= groups[9] {
				g910[k] = v
			}
		}
		// DebugWarn("------topTags:g0", g0)
		// DebugWarn("------topTags:g12", g12)
		// DebugWarn("------topTags:g345", g345)
		// DebugWarn("------topTags:g678", g678)
		// DebugWarn("------topTags:g910", g910)

		viewData = iris.Map{
			"nav_tags_g0":   g0,
			"nav_tags_g12":  g12,
			"nav_tags_g345": g345,
			"nav_tags_g678": g678,
			"nav_tags_g910": g910,
			"current_user":  uname,
		}
	} else {
		gall := make(map[string]int)
		for k, v := range nameNum {
			gall[k] = v
		}
		viewData = iris.Map{
			"nav_tags_gall": gall,
			"current_user":  uname,
		}
	}

	if cacheTime != "" {
		viewData["cache_time"] = cacheTime
	}

	ctx.View("top_tags.html", viewData)
}

func tagList(ctx iris.Context) {
	frmtaglike := ctx.PostValue("frmtaglike")
	frmtaglike = strings.Replace(frmtaglike, "：：", "::", 1)
	currentUser := getCurrentUser(ctx)
	DebugInfo("tagFiles:currentUser", currentUser)
	var uname string
	if currentUser.Name == "" {
		return
	}
	uname = currentUser.Name
	var tags []string
	var filteredTags []string
	var idprefix string
	var tagname string

	if frmtaglike == "" {
		tagCount := mongoTagList(uname, "")
		for k, _ := range tagCount {
			if k != "" {
				tags = append(tags, k)
			}
		}

	} else {
		if strings.Index(frmtaglike, "::") > 0 && strings.Index(frmtaglike, "::") <= len(frmtaglike) {
			pretag := strings.Split(frmtaglike, "::")
			if len(pretag) == 2 {
				if pretag[0] != "" {
					idprefix = pretag[0]
				}
				if pretag[1] != "" {
					tagname = pretag[1]
				}
			}
		} else {
			tagname = frmtaglike
		}

		if idprefix != "" {
			tagCount := mongoTagList(uname, idprefix)
			for k, _ := range tagCount {
				if k != "" {
					tags = append(tags, k)
				}
			}
		} else {
			tagCount := mongoTagList(uname, "")
			for k, _ := range tagCount {
				if k != "" {
					tags = append(tags, k)
				}
			}
		}
		DebugInfo("====", "idprefix:", idprefix, ", tagname:", tagname, ", tags:", tags)
		reg, err := regexp.Compile(fmt.Sprintf("(%s)", tagname))
		PrintError("tagList:regexp.Compile", err)
		for _, tag := range tags {
			if reg.MatchString(tag) == true {
				filteredTags = append(filteredTags, tag)
			}
		}
		tags = filteredTags
	}

	//DebugInfo("===========", filteredTags)

	sort.Strings(tags)

	data := iris.Map{
		"nav_tag_list":   tags,
		"current_user":   uname,
		"current_prefix": frmtaglike,
		"tagname":        tagname,
		"form_action":    "/user/tags",
	}

	if idprefix != "" {
		data["idprefix"] = idprefix
	}

	if len(tags) > 0 {
		data["file_count"] = len(tags)
	}

	if currentUser.IsAdmin == 1 {
		data["is_admin"] = true
	}

	ctx.View("tags.html", data)
}

func tagFiles(ctx iris.Context) {
	tagname := ctx.Params().Get("tagname")
	tagname = strings.Replace(tagname, "：：", "::", 1)
	DebugInfo("tagFiles:tagname=", tagname)
	//
	currentUser := getCurrentUser(ctx)
	DebugInfo("tagFiles:currentUser", currentUser)
	var uname string
	if currentUser.Name == "" {
		return
	}

	uname = currentUser.Name

	var navBreadcrumb []map[string]string

	var files []string
	files = mongoTagFiles(uname, tagname)

	fileCount := len(files)
	navFileList := genNavFileList(files, "", uname)

	bc := make(map[string]string)
	bc["tag"] = fmt.Sprintf("%s", tagname)
	navBreadcrumb = append(navBreadcrumb, bc)

	//DebugInfo("tagFiles:navFileList", navFileList)
	//DebugInfo("tagFiles:navBreadcrumb", navBreadcrumb)
	//DebugInfo("tagFiles:fileCount", fileCount)

	data := iris.Map{
		"nav_file_list":  navFileList,
		"file_count":     Int2Str(fileCount),
		"nav_breadcrumb": navBreadcrumb,
		"current_user":   uname,
		"current_prefix": tagname,
		"site_url":       GetSiteURL(),
	}

	if len(files) < 2001 && len(files) > 0 {
		data["enable_thumbnail"] = true
	}

	ctx.View("tag-files.html", data)
}

func topCaption(ctx iris.Context) {
	captionlang := ctx.Params().Get("lang")
	currentUser := getCurrentUser(ctx)
	DebugInfo("topCaption:currentUser", currentUser)
	var uname string
	if currentUser.Name == "" {
		return
	}
	uname = currentUser.Name
	var counts []int
	countsCacheFile := fmt.Sprintf("%s/topCaption/counts_%s.dat", uname, captionlang)

	nameNum := make(map[string]int)
	nameNumCacheFile := fmt.Sprintf("%s/topCaption/nameNum_%s.dat", uname, captionlang)

	cacheTime := ""

	t1 := GetNowUnix()

	if GobLoad(countsCacheFile, &counts, FunctionCacheExpires) == false || GobLoad(nameNumCacheFile, &nameNum, FunctionCacheExpires) == false {
		userTagCount := mongoCaptionList(uname, captionlang)
		for ukey, uval := range userTagCount {
			if uval > 9 {
				nameNum[ukey] = uval
				counts = append(counts, uval)
			}
		}
		sort.Ints(counts)
		GobDump(countsCacheFile, counts)
		GobDump(nameNumCacheFile, nameNum)
	} else {
		if GobTime(countsCacheFile) != 0 {
			cacheCurrent := UnixFormat(GobTime(countsCacheFile), "01-02 15:04")
			cacheRefreshSeconds := FunctionCacheExpires - (GetNowUnix() - GobTime(countsCacheFile))
			cacheTime = fmt.Sprintf("Update: %s. Will refresh after %v seconds", cacheCurrent, cacheRefreshSeconds)
		}

	}

	t2 := GetNowUnix()

	//DebugWarn("counts", counts, " ===> Elapse: ", (t2 - t1), " seconds")
	//DebugWarn("nameNum", nameNum, " ===> Elapse: ", (t2 - t1), " seconds")
	uqcounts := UniqueInts(counts)
	DebugInfo("topCaption:uqCounts", uqcounts, "; length: ", len(uqcounts), " ===> Elapse: ", (t2 - t1), " seconds")
	DebugInfo("topCaption:cacheTime", cacheTime)

	slots := len(uqcounts) / 10
	DebugWarn("topCaption:slots", slots)

	viewData := iris.Map{}

	if slots > 9 {
		var groups []int
		for i := 0; i < 10; i++ {
			if i == 9 {
				groups = append(groups, uqcounts[len(uqcounts)-1])
			} else {
				groups = append(groups, uqcounts[i*slots])
			}

		}
		DebugWarn("topCaption:groups", groups)

		g0 := make(map[string]int)
		g12 := make(map[string]int)
		g345 := make(map[string]int)
		g678 := make(map[string]int)
		g910 := make(map[string]int)

		if groups[0] < 10 && groups[2] > 10 {
			groups[0] = 9
		}

		for k, v := range nameNum {
			if v <= groups[0] {
				g0[k] = v
			}
			if v > groups[0] && v <= groups[2] {
				g12[k] = v
			}
			if v > groups[2] && v <= groups[5] {
				g345[k] = v
			}
			if v > groups[5] && v <= groups[8] {
				g678[k] = v
			}
			if v > groups[8] && v <= groups[9] {
				g910[k] = v
			}
		}
		// DebugWarn("------topTags:g0", g0)
		// DebugWarn("------topTags:g12", g12)
		// DebugWarn("------topTags:g345", g345)
		// DebugWarn("------topTags:g678", g678)
		// DebugWarn("------topTags:g910", g910)

		viewData = iris.Map{
			"nav_tags_g0":     g0,
			"nav_tags_g12":    g12,
			"nav_tags_g345":   g345,
			"nav_tags_g678":   g678,
			"nav_tags_g910":   g910,
			"current_user":    uname,
			"current_caplang": captionlang,
		}
	} else {
		gall := make(map[string]int)
		for k, v := range nameNum {
			gall[k] = v
		}
		viewData = iris.Map{
			"nav_tags_gall":   gall,
			"current_user":    uname,
			"current_caplang": captionlang,
		}
	}

	if cacheTime != "" {
		viewData["cache_time"] = cacheTime
	}

	ctx.View("top_caption.html", viewData)
}

func captionIndex(ctx iris.Context) {
	frmcaplang := ctx.PostValue("frmcaplang")
	frmtaglike := ctx.PostValue("frmtaglike")
	frmtaglike = strings.Replace(frmtaglike, "：：", "::", 1)
	if frmtaglike != "" && frmcaplang != "" {
		ctx.Redirect("/user/caption/" + frmcaplang + "/" + frmtaglike)
	}

	currentUser := getCurrentUser(ctx)
	DebugInfo("captionIndex:currentUser", currentUser)
	// en
	captionlangen := "en"
	nameNumEn := make(map[string]int)
	nameNumEnCacheFile := fmt.Sprintf("%s/captionIndex/nameNum_%s.dat", currentUser.Name, captionlangen)

	if GobLoad(nameNumEnCacheFile, &nameNumEn, FunctionCacheExpires) == false {
		userCaptionCount := mongoCaptionList(currentUser.Name, captionlangen)
		for ukey, uval := range userCaptionCount {
			if strings.Count(ukey, " ") < 3 && uval > 9 {
				nameNumEn[ukey] = uval
			}
		}
		GobDump(nameNumEnCacheFile, nameNumEn)
	}

	// cn
	captionlangcn := "cn"
	nameNumCn := make(map[string]int)
	nameNumCnCacheFile := fmt.Sprintf("%s/captionIndex/nameNum_%s.dat", currentUser.Name, captionlangcn)

	if GobLoad(nameNumCnCacheFile, &nameNumCn, FunctionCacheExpires) == false {
		userCaptionCount := mongoCaptionList(currentUser.Name, captionlangcn)
		for ukey, uval := range userCaptionCount {
			if strings.Count(ukey, " ") < 3 && uval > 9 {
				nameNumCn[ukey] = uval
			}
		}
		GobDump(nameNumCnCacheFile, nameNumCn)
	}

	var groupSkirt, groupSkirtCn []map[string]int
	var groupDress, groupDressCn []map[string]int
	var groupHair, groupHairCn []map[string]int
	var groupShirt, groupShirtCn []map[string]int
	var groupBody, groupBodyCn []map[string]int
	var groupOther, groupOtherCn []map[string]int

	for ukey, uval := range nameNumEn {
		ukey = strings.ToLower(ukey)
		// if strings.Index(ukey, "skirt") >= 0 || strings.Index(ukey, "shorts") >= 0 {
		// 	groupSkirt = append(groupSkirt, map[string]int{ukey: uval})
		// }
		if IndexOr(ukey, "skirt", "shorts") {
			groupSkirt = append(groupSkirt, map[string]int{ukey: uval})
		}
		if IndexOr(ukey, "dress") {
			groupDress = append(groupDress, map[string]int{ukey: uval})
		}
		if IndexOr(ukey, "hair", "smile", "cry") {
			groupHair = append(groupHair, map[string]int{ukey: uval})
		}
		if IndexOr(ukey, "shirt", "shoulder", "top", "tank", "jacket") {
			groupShirt = append(groupShirt, map[string]int{ukey: uval})
		}
		if IndexOr(ukey, "eye", "leg", "toe", "waist", "shoes") {
			groupBody = append(groupBody, map[string]int{ukey: uval})
		}
		//
		if IndexOr(ukey, "jean", "slacks", "pants", "overalls", "trousers", "socks") {
			groupOther = append(groupOther, map[string]int{ukey: uval})
		}

	}

	for ukey, uval := range nameNumCn {
		ukey = strings.ToLower(ukey)
		if strings.Index(ukey, "裙") >= 0 && strings.Index(ukey, "连衣裙") < 0 {
			groupSkirtCn = append(groupSkirtCn, map[string]int{ukey: uval})
		}

		if IndexOr(ukey, "连衣裙") {
			groupDressCn = append(groupDressCn, map[string]int{ukey: uval})
		}
		if IndexOr(ukey, "发", "眼", "笑", "哭") {
			groupHairCn = append(groupHairCn, map[string]int{ukey: uval})
		}
		if IndexOr(ukey, "衫", "恤", "肩", "夹克") {
			groupShirtCn = append(groupShirtCn, map[string]int{ukey: uval})
		}
		if IndexOr(ukey, "腿", "脚", "手", "腰") {
			groupBodyCn = append(groupBodyCn, map[string]int{ukey: uval})
		}
		if IndexOr(ukey, "裤", "袜") {
			groupOtherCn = append(groupOtherCn, map[string]int{ukey: uval})
		}
	}

	groupSkirtOrdered := MapKeyOrdered(groupSkirt)
	groupSkirtCnOrdered := MapKeyOrdered(groupSkirtCn)
	groupDressOrdered := MapKeyOrdered(groupDress)
	groupDressCnOrdered := MapKeyOrdered(groupDressCn)
	groupHairOrdered := MapKeyOrdered(groupHair)
	groupHairCnOrdered := MapKeyOrdered(groupHairCn)
	groupShirtOrdered := MapKeyOrdered(groupShirt)
	groupShirtCnOrdered := MapKeyOrdered(groupShirtCn)
	groupBodyOrdered := MapKeyOrdered(groupBody)
	groupBodyCnOrdered := MapKeyOrdered(groupBodyCn)
	groupOtherOrdered := MapKeyOrdered(groupOther)
	groupOtherCnOrdered := MapKeyOrdered(groupOtherCn)
	//DebugInfo("captionIndex", groupSkirt)

	data := iris.Map{
		"current_user":       currentUser.Name,
		"form_action":        "/user/caption",
		"cap_group_skirt":    groupSkirtOrdered,
		"cap_group_skirt_cn": groupSkirtCnOrdered,
		"cap_group_dress":    groupDressOrdered,
		"cap_group_dress_cn": groupDressCnOrdered,
		"cap_group_hair":     groupHairOrdered,
		"cap_group_hair_cn":  groupHairCnOrdered,
		"cap_group_shirt":    groupShirtOrdered,
		"cap_group_shirt_cn": groupShirtCnOrdered,
		"cap_group_body":     groupBodyOrdered,
		"cap_group_body_cn":  groupBodyCnOrdered,
		"cap_group_other":    groupOtherOrdered,
		"cap_group_other_cn": groupOtherCnOrdered,
	}

	if currentUser.IsAdmin == 1 {
		data["is_admin"] = true
	}

	ctx.View("caption.html", data)
}

func captionFiles(ctx iris.Context) {
	captionlang := ctx.Params().Get("lang")
	captionword := ctx.Params().Get("captionword")
	captionword = strings.Replace(captionword, "：：", "::", 1)
	DebugInfo("captionFiles:captionlang:", captionlang, ": ", captionword)
	//
	currentUser := getCurrentUser(ctx)
	DebugInfo("captionFiles:currentUser", currentUser)
	var uname string
	if currentUser.Name == "" {
		return
	}

	uname = currentUser.Name

	var navBreadcrumb []map[string]string

	var files []string
	files = mongoCaptionFiles(uname, captionlang, captionword)

	fileCount := len(files)
	navFileList := genNavFileList(files, "", uname)

	bc := make(map[string]string)
	bc["tag"] = fmt.Sprintf("%s", captionword)
	navBreadcrumb = append(navBreadcrumb, bc)

	data := iris.Map{
		"form_action":     "/user/caption",
		"nav_file_list":   navFileList,
		"file_count":      Int2Str(fileCount),
		"nav_breadcrumb":  navBreadcrumb,
		"current_user":    uname,
		"current_prefix":  captionword,
		"site_url":        GetSiteURL(),
		"current_caplang": captionlang,
	}

	if captionlang == "en" {
		data["current_caplang_en"] = captionlang
	}
	if captionlang == "cn" {
		data["current_caplang_cn"] = captionlang
	}

	if len(files) < 2001 && len(files) > 0 {
		data["enable_thumbnail"] = true
	}

	ctx.View("caption-files.html", data)
}

func adminListBuckets(ctx iris.Context) {
	currentUser := getCurrentUser(ctx)
	DebugInfo("adminListBuckets:currentUser", currentUser)

	userlogin := currentUser.Name

	var collList []string
	navList := make(map[string]map[string]string)
	buckets := mongoAdminListCollections()

	if currentUser.IsAdmin != 1 {
		if Contains(buckets, userlogin) {
			collList = append(collList, userlogin)
		}
	} else {
		for _, bkt := range buckets {
			if strings.HasPrefix(bkt, "system.") || strings.HasPrefix(bkt, "_") {
				continue
			}
			collList = append(collList, bkt)
		}

	}

	for _, v := range collList {
		DebugInfo("adminListBuckets:collList", v)
		stats := mongoUserStats(v)
		vstats := make(map[string]string)
		val, ok := stats["doc_count"]
		if ok {
			vstats["doc_count"] = val
			vstats["unique_doc_count"] = stats["unique_doc_count"]
			vstats["total_size"] = stats["total_size"]
		}
		navList[v] = vstats
	}

	DebugInfo("navList", navList)

	data := iris.Map{
		"nav_list":     navList,
		"current_user": userlogin,
		"site_url":     GetSiteURL(),
	}

	if len(collList) > 0 {
		data["file_count"] = len(collList)
	}

	if currentUser.IsAdmin == 1 {
		data["is_admin"] = true
	}

	ctx.View("bucket-list.html", data)
}

func adminListGroup(ctx iris.Context) {
	uname := ctx.Params().Get("uname")
	DebugInfo("adminListGroup:uname=", uname)

	//
	currentUser := getCurrentUser(ctx)
	DebugInfo("adminListGroup:currentUser", currentUser)
	if currentUser.IsAdmin != 1 && currentUser.Name != uname {
		return
	}

	dirs, files := mongoListFiles(uname, "", bson.D{{"mtime", -1}})

	navDirList := genNavDirList(dirs, "", uname)
	navFileList := genNavFileList(files, "", uname)

	//DebugInfo("navList", dirs)

	data := iris.Map{
		"nav_dir_list":  navDirList,
		"nav_file_list": navFileList,
		"current_user":  uname,
		"site_url":      GetSiteURL(),
	}
	if len(dirs) > 0 || len(files) > 0 {
		data["file_count"] = Int2Str(len(files) + len(dirs))
	}

	if len(files) < 2001 && len(files) > 0 {
		data["enable_thumbnail"] = true
	}

	if currentUser.IsAdmin == 1 {
		data["is_admin"] = true
	}

	ctx.View("key-list.html", data)
}

func adminListKeys(ctx iris.Context) {
	uname := ctx.Params().Get("uname")
	fkey := strings.Trim(ctx.Params().Get("fkey"), "/")
	if uname == "" || fkey == "" {
		return
	}
	DebugInfo("adminListKeys:uname=", uname, " :fkey=", fkey)
	//
	currentUser := getCurrentUser(ctx)
	DebugInfo("adminListKeys:currentUser", currentUser)
	if currentUser.IsAdmin != 1 && currentUser.Name != uname {
		return
	}

	//navFileList := make(map[string]map[string]any)
	var navBreadcrumb []map[string]string

	dirs, files := mongoListFiles(uname, fkey, bson.D{{"mtime", -1}})

	navDirList := genNavDirList(dirs, fkey, uname)
	navFileList := genNavFileList(files, fkey, uname)

	breads := strings.Split(fkey, "/")
	for idx, bread := range breads {
		if bread != "" {
			fkeyLeft := strings.Join(breads[:idx+1], "/")
			DebugInfo("adminListKeys:fkey_left", bread, "<==", fkeyLeft)
			bc := make(map[string]string)
			bc[bread] = fmt.Sprintf("%s/%s", uname, fkeyLeft)
			navBreadcrumb = append(navBreadcrumb, bc)
		}
	}

	//DebugInfo("adminListKeys:navDirList", navDirList)
	//DebugInfo("adminListKeys:navFileList", navFileList)

	data := iris.Map{
		"nav_dir_list":    navDirList,
		"nav_file_list":   navFileList,
		"nav_breadcrumb":  navBreadcrumb,
		"current_user":    uname,
		"current_prefix":  fkey,
		"url_play_prefix": strings.Join([]string{"/user/playlist", fkey}, "/"),
		"site_url":        GetSiteURL(),
	}

	if len(files) > 0 || len(dirs) > 0 {
		data["file_count"] = Int2Str(len(files) + len(dirs))
	}

	if len(files) < 2001 && len(files) > 0 {
		data["enable_thumbnail"] = true
	}

	if currentUser.IsAdmin == 1 {
		data["is_admin"] = true
	}

	ctx.View("key-list.html", data)

}

func adminSameFiles(ctx iris.Context) {
	currentUser := getCurrentUser(ctx)
	DebugInfo("adminSameFiles:currentUser", currentUser)
	if currentUser.IsAdmin != 1 || currentUser.Name == "" {
		return
	}

	uname := currentUser.Name

	mongoAdminResetKeyStats(uname, "fsha256")

	mongoAdminUpdateKeyStats(uname, "fsha256")

	urls := mongoAdminGetKeyStats(uname, "fsha256")

	data := iris.Map{
		"urls":         urls,
		"urls_count":   len(urls),
		"current_user": uname,
		"site_url":     GetSiteURL(),
	}

	if currentUser.IsAdmin == 1 {
		data["is_admin"] = true
	}

	ctx.View("same-files.html", data)

}

func likeFiles(ctx iris.Context) {
	dotcolor := ctx.Params().Get("dotcolor")

	currentUser := getCurrentUser(ctx)
	uname := currentUser.Name
	DebugInfo("likeFiles:currentUser", currentUser, ":", uname)
	if uname == "" || dotcolor == "" {
		return
	}

	//likesList := make(map[string]string)

	files := mongoAggFilesByKey(uname, "dot_color", dotcolor)
	navFileList := genNavFileList(files, "", uname)
	//DebugInfo("likeFiles", navFileList)

	data := iris.Map{
		"nav_file_list":      navFileList,
		"current_dotcolor":   dotcolor,
		"current_user":       uname,
		"url_play_dot_color": strings.Join([]string{GetSiteURL(), "user/playlikes", dotcolor}, "/"),
		"site_url":           GetSiteURL(),
	}

	if len(files) > 0 {
		data["file_count"] = Int2Str(len(files))
	}

	if len(files) < 2001 && len(files) > 0 {
		data["enable_thumbnail"] = true
	}

	if currentUser.IsAdmin == 1 {
		data["is_admin"] = true
	}

	ctx.View("like-list.html", data)
}

func byKeyFiles(ctx iris.Context) {
	key := ctx.Params().Get("key")
	val := ctx.Params().Get("val")

	currentUser := getCurrentUser(ctx)
	uname := currentUser.Name
	DebugInfo("ByKeyFiles:currentUser", currentUser, ":", uname)
	if uname == "" || key == "" {
		return
	}

	files := mongoAggFilesByKey(uname, key, val)
	navFileList := genNavFileList(files, "", uname)
	DebugInfo("ByKeyFiles:navFileList", navFileList)

	data := iris.Map{
		"nav_file_list": navFileList,
		"nav_by_key":    key,
		"nav_by_val":    val,
		"current_user":  uname,
		"site_url":      GetSiteURL(),
	}

	if len(files) > 0 {
		data["file_count"] = Int2Str(len(files))
	}

	if len(files) < 2001 && len(files) > 0 {
		data["enable_thumbnail"] = true
	}

	if currentUser.IsAdmin == 1 {
		data["is_admin"] = true
	}

	ctx.View("by-list.html", data)
}

func dotColorFile(ctx iris.Context) {
	color := ctx.Params().Get("color")
	uname := ctx.Params().Get("uname")
	fname := ctx.Params().Get("fname")
	if IsAnyEmpty(color, uname, fname) {
		ctx.JSON(iris.Map{
			"dot_color": "",
			"error":     "color, user, file cannot be empty.",
		})
		return
	}

	currentUser := getCurrentUser(ctx)
	if uname != currentUser.Name {
		ctx.JSON(iris.Map{
			"dot_color": "",
			"error":     "user is invalid.",
		})
		return
	}

	frmDotColor := ""
	allowColors := []string{"red", "green", "gold", "black", "blue", "orange", "purple", "empty"}
	if !Contains(allowColors, color) {
		ctx.JSON(iris.Map{
			"dot_color": "",
			"error":     "color is invalid.",
		})
		return
	}

	frmDotColor = strings.Join([]string{"dot", color}, "-")
	DebugInfo("fname:", fname, " <== ", frmDotColor)

	if frmDotColor != "" {
		if frmDotColor == "dot-empty" {
			frmDotColor = ""
		}
		success := mongoSave(uname, fname, "dot_color", frmDotColor)
		if success == false {
			ctx.JSON(iris.Map{
				"dot_color": "",
				"error":     "data cannot be saved.",
			})
			return
		}
	}

	meta := mongoGet(uname, fname)
	currentDotColor, ok := meta["dot_color"]
	if !ok {
		ctx.JSON(iris.Map{
			"dot_color": "",
			"error":     "cannot find dot_color field from meta.",
		})
		return
	}
	DebugInfo("frmDotColor:", currentDotColor)

	ctx.JSON(iris.Map{
		"dot_color": currentDotColor,
		"error":     "",
	})
}

func bcacheItems(ctx iris.Context) {
	currentUser := getCurrentUser(ctx)
	uname := currentUser.Name

	if uname == "" && currentUser.IsAdmin != 1 {
		return
	}

	data := iris.Map{
		"bcacheItems":  bcacheScan(uname),
		"current_user": uname,
		"site_url":     GetSiteURL(),
	}
	if currentUser.IsAdmin == 1 {
		data["is_admin"] = true
	}
	ctx.View("bcache-items.html", data)
}

func apiListTags(ctx iris.Context) {
	if ctx.Method() != "POST" {
		DebugInfo("apiListTags", "pls use POST")
		return
	}
	ctx.Header("Content-Type", "text/json;charset=utf-8")

	fuser := ctx.PostValue("fuser")
	fapikey := ctx.PostValue("fapikey")

	if IsAnyEmpty(fuser, fapikey) {
		DebugWarn("apiListTags.10", "fuser/fapikey cannot be empty")
		ctx.StatusCode(iris.StatusForbidden)
		ctx.Writef("Error: 403 Forbidden")
		return
	}

	user := mysqlAPIKeyLogin(fuser, fapikey, 1)
	if user.APIKey != fapikey {
		DebugWarn("apiListTags.20", "user.APIKey: ", user.APIKey, ", fapikey: ", fapikey)
		ctx.StatusCode(iris.StatusForbidden)
		ctx.Writef("Error: 403 Forbidden")
		return
	}

	tagCount := mongoTagList(fuser, "")
	ctx.JSON(tagCount)
}

func apiListCaption(ctx iris.Context) {
	if ctx.Method() != "POST" {
		DebugInfo("apiListCaption", "pls use POST")
		return
	}
	ctx.Header("Content-Type", "text/json;charset=utf-8")

	fuser := ctx.PostValue("fuser")
	fapikey := ctx.PostValue("fapikey")
	flanguage := ctx.PostValue("flanguage")

	if IsAnyEmpty(fuser, fapikey, flanguage) {
		DebugWarn("apiListCaption.10", "fuser/fapikey cannot be empty")
		ctx.StatusCode(iris.StatusForbidden)
		ctx.Writef("Error: 403 Forbidden")
		return
	}

	user := mysqlAPIKeyLogin(fuser, fapikey, 1)
	if user.APIKey != fapikey {
		DebugWarn("apiListCaption.20", "user.APIKey: ", user.APIKey, ", fapikey: ", fapikey)
		ctx.StatusCode(iris.StatusForbidden)
		ctx.Writef("Error: 403 Forbidden")
		return
	}

	captionCount := mongoCaptionList(fuser, flanguage)
	ctx.JSON(captionCount)
}
