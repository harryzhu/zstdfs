package cmd

import (
	//"encoding/json"
	"fmt"
	"io/ioutil"
	"mime"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	//"github.com/kataras/iris/v12/sessions"
	"github.com/kataras/iris/v12"
)

func StartHTTPServer() {

	app := iris.New()
	//app.IsDebug()

	tmpl := iris.HTML(embeddedFS, ".html")

	tmpl.RootDir("template")
	tmpl.Delims("{{", "}}")
	tmpl.Reload(IsDebug)
	app.RegisterView(tmpl)

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
		homeAPI.HandleDir("/assets", iris.Dir(ASSET_DIR), iris.DirOptions{ShowList: false})
	}

	StaticOptions := iris.DirOptions{
		ShowList: true,
		DirList: iris.DirListRich(iris.DirListRichOptions{
			TmplName: "dirlist.html",
		}),
	}
	app.HandleDir("/static", iris.Dir(STATIC_DIR), StaticOptions)

	userFileAPI := app.Party("/f/")
	{
		userFileAPI.Use(iris.Compression)
		userFileAPI.Get("/{uname:string}/{key:path}", getFiles)
		//
	}

	playAPI := app.Party("/play")
	{
		playAPI.Use(iris.Compression)
		playAPI.HandleDir("/temp", iris.Dir(TEMP_DIR), iris.DirOptions{})

		playAPI.Get("/v/{bucket:string}/{fname:path}", playVideos)
	}

	userAPI := app.Party("/user/")
	{
		userAPI.Use(iris.Compression)

		userAPI.Get("/buckets", adminListBuckets)
		userAPI.Get("/buckets/{uname:string}", adminListFiles)
		userAPI.Get("/buckets/{uname:string}/{fkey:path}", adminListKeys)
		//
		userAPI.Post("/upload", uploadFile)
		userAPI.Get("/edit", editFiles)
		userAPI.Get("/likes/{dotcolor:string}", likeFiles)
		userAPI.Get("/tags", tagList)
		userAPI.Post("/tags", tagList)
		userAPI.Get("/tags/{tagname:string}", tagFiles)
		userAPI.Get("/tags/top", topTags)
		userAPI.Get("/by/{key:string}/{val:path}", ByKeyFiles)
		userAPI.Get("/top/{countname:string}/{min:int}/{max:int}", topCountFiles)
		userAPI.Get("/playlist", playList)

		//
	}

	v1API := app.Party("/api/")
	{
		v1API.Post("/upload", apiUploadFile)
		v1API.Get("/has/{blake3sum:string}", apiHasFile)
	}

	app.Listen(fmt.Sprintf("%s:%s", Host, Port), iris.WithDynamicHandler)

}

func StopHTTPServer() {
	fmt.Println("stopping the server ...")
	sqldb.Close()
	bgrdb.Close()
}

func notFound(ctx iris.Context) {
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
	}

	if len(files) > 0 {
		data["file_count"] = Int2Str(len(files))
	}

	if currentUser.IsAdmin == 1 {
		data["is_admin"] = true
	}

	ctx.View("top-list.html", data)
}

func getFiles(ctx iris.Context) {
	uname := ctx.Params().Get("uname")
	key := ctx.Params().Get("key")

	fext := filepath.Ext(key)
	fname := filepath.Join("f", uname, key)

	entity := NewEntity(uname, key).GetMeta()
	DebugInfo("getFiles:entity:", entity.Meta)
	if entity.Meta["is_public"] != "1" {
		currentUser := getCurrentUser(ctx)
		if currentUser.Name != uname || currentUser.IsAdmin != 1 {
			ctx.View("403.html")
			return
		}
	}

	if entity.Meta["is_ban"] == "1" {
		DebugInfo("getFiles:is_ban==1", fname)
		currentUser := getCurrentUser(ctx)
		if currentUser.IsAdmin != 1 {
			ctx.View("404.html")
			return
		}
	}

	if entity.Meta["fsum"] == "" {
		DebugInfo("getFiles:fsum is empty", fname)
		ctx.View("404.html")
		return
	}

	entity = entity.Get()
	if entity.Data == nil {
		DebugInfo("getFiles: Data is empty", fname)
		ctx.View("404.html")
		return
	}

	mimeType := "text/plain"

	if fext != "" {
		mimeType = mime.TypeByExtension(fext)
	}

	ctx.Header("X-Powered-By", "zstdfs")
	ctx.Header("Cache-Control", "public, max-age=0")
	ctx.Header("Content-Type", mimeType)
	ctx.Write(entity.Data)
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

	data := iris.Map{
		"form_action":             "/user/upload",
		"form_max_upload_size_mb": strconv.Itoa(MaxUploadSizeMB),
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

func playVideos(ctx iris.Context) {
	bucket := ctx.Params().Get("bucket")
	fname := ctx.Params().Get("fname")

	fkey := strings.Join([]string{TEMP_DIR, bucket, fname}, "/")
	MakeDirs(filepath.Dir(fkey))
	_, err := os.Stat(fkey)
	if err != nil {
		meta := mongoGet(bucket, fname)
		b := badgerGet([]byte(meta["fsum"]))
		if len(b) == 0 {
			ctx.NotFound()
			return
		}
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
		userTags := mongoTagList(uname)
		for _, utag := range userTags {
			nameCount := mongoTagCount(uname, utag)
			c, ok := nameCount[utag]
			if ok {
				nameNum[utag] = c
				counts = append(counts, c)
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

	currentUser := getCurrentUser(ctx)
	DebugInfo("tagFiles:currentUser", currentUser)
	var uname string
	if currentUser.Name == "" {
		return
	}
	uname = currentUser.Name
	var tags []string

	if currentUser.IsAdmin == 1 {
		buckets := mongoAdminListCollections()
		for _, bucket := range buckets {
			usertags := mongoTagList(bucket)
			for _, t := range usertags {
				if Contains(tags, t) {
					continue
				}
				tags = append(tags, t)
			}
		}
	} else {
		usertags := mongoTagList(uname)
		tags = usertags
	}

	var filteredTags []string
	if frmtaglike != "" {
		reg, err := regexp.Compile(fmt.Sprintf("(%s)", frmtaglike))
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
		"nav_tag_list": tags,
		"current_user": uname,
		"form_action":  "/user/tags",
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
	if currentUser.IsAdmin == 1 {
		buckets := mongoAdminListCollections()
		for _, bucket := range buckets {
			userfiles := mongoTagFiles(bucket, tagname)
			for _, t := range userfiles {
				files = append(files, t)
			}
		}
	} else {
		files = mongoTagFiles(uname, tagname)
	}

	fileCount := len(files)
	navFileList := genNavFileList(files, "", uname)

	bc := make(map[string]string)
	bc["tag"] = fmt.Sprintf("%s", tagname)
	navBreadcrumb = append(navBreadcrumb, bc)

	//DebugInfo("tagFiles:navFileList", navFileList)
	//DebugInfo("tagFiles:navBreadcrumb", navBreadcrumb)
	//DebugInfo("tagFiles:fileCount", fileCount)

	ctx.View("tag-files.html", iris.Map{
		"nav_file_list":  navFileList,
		"file_count":     Int2Str(fileCount),
		"nav_breadcrumb": navBreadcrumb,
		"current_user":   uname,
		"current_prefix": tagname,
	})
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
		v_stats := make(map[string]string)
		val, ok := stats["doc_count"]
		if ok {
			v_stats["doc_count"] = val
		}
		navList[v] = v_stats
	}

	DebugInfo("navList", navList)

	data := iris.Map{
		"nav_list":     navList,
		"current_user": userlogin,
	}

	if len(collList) > 0 {
		data["file_count"] = len(collList)
	}

	if currentUser.IsAdmin == 1 {
		data["is_admin"] = true
	}

	ctx.View("bucket-list.html", data)
}

func adminListFiles(ctx iris.Context) {
	uname := ctx.Params().Get("uname")
	DebugInfo("adminListFiles:uname=", uname)
	//
	currentUser := getCurrentUser(ctx)
	DebugInfo("adminListFiles:currentUser", currentUser)
	if currentUser.IsAdmin != 1 && currentUser.Name != uname {
		return
	}

	dirs, files := mongoListFiles(uname, "")

	navDirList := genNavDirList(dirs, "", uname)
	navFileList := genNavFileList(files, "", uname)

	DebugInfo("navList", navDirList)

	data := iris.Map{
		"nav_dir_list":  navDirList,
		"nav_file_list": navFileList,
		"current_user":  uname,
	}
	if len(dirs) > 0 || len(files) > 0 {
		data["file_count"] = Int2Str(len(files) + len(dirs))
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

	dirs, files := mongoListFiles(uname, fkey)

	navDirList := genNavDirList(dirs, fkey, uname)
	navFileList := genNavFileList(files, fkey, uname)

	breads := strings.Split(fkey, "/")
	for idx, bread := range breads {
		if bread != "" {
			fkey_left := strings.Join(breads[:idx+1], "/")
			DebugInfo("adminListKeys:fkey_left", bread, "<==", fkey_left)
			bc := make(map[string]string)
			bc[bread] = fmt.Sprintf("%s/%s", uname, fkey_left)
			navBreadcrumb = append(navBreadcrumb, bc)
		}
	}

	DebugInfo("adminListKeys:navDirList", navDirList)
	DebugInfo("adminListKeys:navFileList", navFileList)

	data := iris.Map{
		"nav_dir_list":   navDirList,
		"nav_file_list":  navFileList,
		"nav_breadcrumb": navBreadcrumb,
		"current_user":   uname,
		"current_prefix": fkey,
	}

	if len(files) > 0 || len(dirs) > 0 {
		data["file_count"] = Int2Str(len(files) + len(dirs))
	}

	if currentUser.IsAdmin == 1 {
		data["is_admin"] = true
	}

	ctx.View("key-list.html", data)

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
		"nav_file_list":    navFileList,
		"current_dotcolor": dotcolor,
		"current_user":     uname,
	}

	if len(files) > 0 {
		data["file_count"] = Int2Str(len(files))
	}

	if currentUser.IsAdmin == 1 {
		data["is_admin"] = true
	}

	ctx.View("like-list.html", data)
}

func playList(ctx iris.Context) {
	currentUser := getCurrentUser(ctx)
	uname := currentUser.Name
	DebugInfo("playList:currentUser", currentUser, ":", uname)
	if uname == "" {
		return
	}
	var videoUrls []string
	videoUrlsCacheFile := fmt.Sprintf("%s/playList/video_urls.dat", uname)

	if GobLoad(videoUrlsCacheFile, &videoUrls, 60) == false {
		files := mongoRandomGet(uname, 30)
		for _, f := range files {
			videoUrls = append(videoUrls, strings.Join([]string{"http://192.168.0.100:9090/f", uname, f}, "/"))
		}
		GobDump(videoUrlsCacheFile, videoUrls)
		DebugInfo("-----playList: from", "DB")
	} else {
		DebugInfo("*******playList: from", "CACHE")
	}

	//DebugInfo("playList", videoUrls)
	// fext := filepath.Ext(fname)
	mimeType := "video/mp4"

	curVid := ctx.GetCookie("cur_vid")

	if curVid == "" {
		curVid = "0"
		ctx.SetCookieKV("cur_vid", curVid)
	}

	if Str2Int(curVid) >= len(videoUrls) {
		curVid = "0"
		ctx.SetCookieKV("cur_vid", curVid)
	} else {
		ctx.SetCookieKV("cur_vid", Int2Str(Str2Int(curVid)+1))
	}

	playURL := strings.Replace(videoUrls[Str2Int(curVid)], "/f/", "/play/v/", 1)
	DebugInfo("playList:Cookie:curVid", curVid, "::", playURL)

	data := iris.Map{
		"current_user": uname,
		"video_src":    videoUrls[Str2Int(curVid)],
		"video_mime":   mimeType,
	}

	ctx.View("playlist.html", data)
}

func ByKeyFiles(ctx iris.Context) {
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
	}

	if len(files) > 0 {
		data["file_count"] = Int2Str(len(files))
	}

	if currentUser.IsAdmin == 1 {
		data["is_admin"] = true
	}

	ctx.View("by-list.html", data)
}
