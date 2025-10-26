package cmd

import (
	"fmt"
	"io/ioutil"

	"mime"
	"os"
	"path/filepath"
	"strings"

	"github.com/kataras/iris/v12"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type VideoItem struct {
	ID       string
	URI      string
	User     string
	SrcID    string
	SrcURI   string
	Mime     string
	Meta     map[string]string
	Data     []byte
	ViewData iris.Map
}

func NewVideoItem(user, key, val string) (videoitem VideoItem, err error) {
	meta := mongoGetByKey(user, key, val)
	mid, ok := meta["_id"]
	if !ok {
		return videoitem, ErrEmptyMeta
	}
	videoitem.Meta = meta
	videoitem.ID = mid
	videoitem.URI = meta["uri"]
	videoitem.User = user
	videoitem.Mime = meta["mime"]
	videoitem.SrcID = strings.Join([]string{user, mid}, "/")
	videoitem.SrcURI = strings.Join([]string{user, meta["uri"]}, "/")
	videoitem.Data, _ = gcGet([]byte(meta["_fsum"]))

	videoitem.ViewData = iris.Map{
		"current_user": user,
		"id_name":      strings.Join([]string{user, meta["_id"]}, "/"),
		"uri_name":     strings.Join([]string{user, meta["uri"]}, "/"),
		"site_url":     GetSiteURL(),
		"video_mime":   meta["mime"],
		"video_src":    "",
		"active_color": "",
	}
	return videoitem, nil
}

func playVideos(ctx iris.Context) {
	bucket := ctx.Params().Get("bucket")
	fname := ctx.Params().Get("fname")

	currentUser := getCurrentUser(ctx)
	DebugInfo("playVideos:currentUser", currentUser, ":", currentUser.Name, ", path: ", fname)
	if currentUser.Name != bucket {
		return
	}

	videoItem, err := NewVideoItem(bucket, "_id", fname)
	if err != nil {
		DebugWarn("playVideos", err)
		return
	}
	DebugInfo("---", videoItem.Meta)
	fkey := strings.Join([]string{TempDir, bucket, fname}, "/")
	MakeDirs(filepath.Dir(fkey))
	_, err = os.Stat(fkey)
	if err != nil {
		b := videoItem.Data
		if b == nil {
			ctx.NotFound()
			return
		}
		err = ioutil.WriteFile(fkey, b, os.ModePerm)
		PrintError("playVideos:ioutil.WriteFile", err)
	}

	fext := filepath.Ext(fname)
	mimeType := videoItem.Mime

	if fext != "" {
		mimeType = mime.TypeByExtension(fext)
	}

	videosrc := strings.Join([]string{"/play", "temp", bucket, fname}, "/")
	data := videoItem.ViewData
	data["video_src"] = videosrc
	data["video_mime"] = mimeType
	data["id_name"] = fname

	ctx.View("playlist.html", data)
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

	if GobLoad(videoUrlsCacheFile, &videoUrls, FunctionCacheExpires) == false {
		files := mongoRandomGetURI(uname, 30)
		for _, f := range files {
			videoUrls = append(videoUrls, strings.Join([]string{GetSiteURL(), "s", uname, f}, "/"))
		}
		GobDump(videoUrlsCacheFile, videoUrls)
		DebugInfo("-----playList: from", "DB")
	} else {
		DebugInfo("*******playList: from", "CACHE")
	}

	//DebugInfo("playList", videoUrls)
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

	videoSrc := ""
	uriName := ""
	idName := ""
	dotColor := ""
	if len(videoUrls) > 0 {
		playURL := strings.Replace(videoUrls[Str2Int(curVid)], "/s/", "/play/v/", 1)
		DebugInfo("playList:Cookie:curVid", curVid, "::", playURL)

		videoSrc = videoUrls[Str2Int(curVid)]
		uriName = videoSrc[strings.Index(videoSrc, "/s/")+len(uname)+4:]
		DebugInfo("uriName:", uriName)
		videoItem, err := NewVideoItem(uname, "uri", uriName)
		if err != nil {
			return
		}
		dc, ok := videoItem.Meta["dot_color"]
		if ok {
			dotColor = dc
			idName = videoItem.Meta["_id"]
			mimeType = videoItem.Meta["mime"]
		}
	}

	data := iris.Map{
		"current_user": uname,
		"video_src":    videoSrc,
		"video_mime":   mimeType,
		"active_color": dotColor,
		"id_name":      idName,
		"site_url":     GetSiteURL(),
	}

	ctx.View("playlist.html", data)
}

func playPrefixList(ctx iris.Context) {
	prefix := ctx.Params().Get("prefix")

	currentUser := getCurrentUser(ctx)
	if currentUser.Name == "" {
		return
	}

	uname := currentUser.Name
	DebugInfo("playPrefixList:currentUser", currentUser, ":", uname)

	var videoUrls []string
	videoUrlsCacheFile := fmt.Sprintf("%s/playPrefixList/video_%s.dat", uname, GetXxhash([]byte(prefix)))

	if GobLoad(videoUrlsCacheFile, &videoUrls, FunctionCacheExpires) == false {
		_, lines := mongoListFiles(uname, prefix, bson.D{{"mtime", -1}})
		var files []string
		for _, line := range lines {
			DebugInfo("====ext", filepath.Ext(line))
			if filepath.Ext(line) == ".mp4" {
				files = append(files, line)
			}
		}

		for _, f := range files {
			videoUrls = append(videoUrls, strings.Join([]string{GetSiteURL(), "f", uname, prefix, f}, "/"))
		}
		GobDump(videoUrlsCacheFile, videoUrls)
		DebugInfo("-----playPrefixList: from", "DB")
	} else {
		DebugInfo("*******playPrefixList: from", "CACHE")
	}

	//DebugInfo("playList", videoUrls)
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

	videoSrc := ""
	idName := ""
	dotColor := ""
	if len(videoUrls) > 0 {
		playURL := strings.Replace(videoUrls[Str2Int(curVid)], "/f/", "/play/v/", 1)
		DebugInfo("playPrefixList:Cookie:curVid", curVid, "::", playURL)

		videoSrc = videoUrls[Str2Int(curVid)]
		idName = videoSrc[strings.Index(videoSrc, "/f/")+len(uname)+4:]
		DebugInfo("idName:", idName)
		videoItem, err := NewVideoItem(uname, "_id", idName)
		if err != nil {
			return
		}
		dc, ok := videoItem.Meta["dot_color"]
		if ok {
			dotColor = dc
			idName = videoItem.Meta["_id"]
			mimeType = videoItem.Meta["mime"]
		}
	}

	data := iris.Map{
		"current_user": uname,
		"video_src":    videoSrc,
		"video_mime":   mimeType,
		"active_color": dotColor,
		"id_name":      idName,
		"site_url":     GetSiteURL(),
	}

	ctx.View("playlist.html", data)
}

func playDotColorList(ctx iris.Context) {
	color := ctx.Params().Get("color")
	currentUser := getCurrentUser(ctx)
	uname := currentUser.Name
	if IsAnyEmpty(color, uname) {
		return
	}

	var videoUrls []string
	videoUrlsCacheFile := fmt.Sprintf("%s/playDotColorList/dot_color_%s.dat", uname, GetXxhash([]byte(color)))

	if GobLoad(videoUrlsCacheFile, &videoUrls, FunctionCacheExpires) == false {
		lines := mongoAggFilesByKey(uname, "dot_color", color)
		var files []string
		for _, line := range lines {
			DebugInfo("====ext", filepath.Ext(line))
			if filepath.Ext(line) == ".mp4" {
				files = append(files, line)
			}
		}
		for _, f := range files {
			videoUrls = append(videoUrls, strings.Join([]string{GetSiteURL(), "f", uname, f}, "/"))
		}
		GobDump(videoUrlsCacheFile, videoUrls)
		DebugInfo("-----playDotColorList: from", "DB")
	} else {
		DebugInfo("*******playDotColorList: from", "CACHE")
	}

	//DebugInfo("playList", videoUrls)
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

	videoSrc := ""
	idName := ""
	dotColor := ""
	if len(videoUrls) > 0 {
		playURL := strings.Replace(videoUrls[Str2Int(curVid)], "/s/", "/play/v/", 1)
		DebugInfo("playDotColorList:Cookie:curVid", curVid, "::", playURL)

		videoSrc = videoUrls[Str2Int(curVid)]
		idName = videoSrc[strings.Index(videoSrc, "/f/")+len(uname)+4:]
		DebugInfo("idName:", idName)
		videoItem, err := NewVideoItem(uname, "_id", idName)
		if err != nil {
			return
		}
		dc, ok := videoItem.Meta["dot_color"]
		if ok {
			dotColor = dc
			idName = videoItem.Meta["_id"]
			mimeType = videoItem.Meta["mime"]
		}
	}

	data := iris.Map{
		"current_user": uname,
		"video_src":    videoSrc,
		"video_mime":   mimeType,
		"active_color": dotColor,
		"id_name":      idName,
		"site_url":     GetSiteURL(),
	}

	ctx.View("playlist.html", data)
}
