package cmd

import (
	"encoding/json"
	//"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	//"html/template"
	//"io/ioutil"
	"strconv"
	"strings"
	"time"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/harryzhu/potatofs/entity"
	//"go.uber.org/zap"
	//"github.com/harryzhu/potatofs/util"
)

func IndexHandler(ctx *gin.Context) {
	ctx.String(http.StatusOK, "ok")
}

func ListHandler(ctx *gin.Context) {
	ctx.Header("Content-Type", "text/html")

	page := ctx.DefaultQuery("page", "1")
	keys := entity.GetMetaKeyListByPrefix([]byte(""), page)

	liPage := GenPagination(page)
	li := GenURLList(keys, liPage)

	pgHTML := strings.Join([]string{pgHeader, li, pgFooter}, "")
	ctx.String(http.StatusOK, pgHTML)
}

func ListByUserHandler(ctx *gin.Context) {
	ctx.Header("Content-Type", "text/html")
	user := ctx.Param("user")
	page := ctx.DefaultQuery("page", "1")
	prefix := ""

	if user == "" {
		ctx.Abort()
	} else {
		prefix = strings.Join([]string{user, "/"}, "")
	}

	keys := entity.GetMetaKeyListByPrefix([]byte(prefix), page)
	liPage := GenPagination(page)
	li := GenURLList(keys, liPage)

	pgHTML := strings.Join([]string{pgHeader, li, pgFooter}, "")
	ctx.String(http.StatusOK, pgHTML)
}

func ListByUserAlbumHandler(ctx *gin.Context) {
	ctx.Header("Content-Type", "text/html")
	user := ctx.Param("user")
	album := ctx.Param("album")
	page := ctx.DefaultQuery("page", "1")
	log.Println("Page: ", page)

	prefix := ""

	if user == "" || album == "" {
		ctx.Abort()
	}

	prefix = strings.Join([]string{user, album}, "/")

	keys := entity.GetMetaKeyListByPrefix([]byte(prefix), page)
	liPage := GenPagination(page)
	li := GenURLList(keys, liPage)

	pgHTML := strings.Join([]string{pgHeader, li, pgFooter}, "")
	ctx.String(http.StatusOK, pgHTML)
}

func GenPagination(pageStr string) string {
	pageNum, _ := strconv.Atoi(pageStr)
	if pageNum < 1 {
		pageNum = 1
	}

	pagePrev := pageNum - 1
	pageNext := pageNum + 1

	if pagePrev < 1 {
		pagePrev = 1
	}

	liRoot := strings.Join([]string{"<a class=\"btn-nav-page\" href=\"", Config.SiteURL, "/list", "\">List All</a>"}, "")

	aPrev := strings.Join([]string{"<a class=\"btn-nav-page\" href=?page=", strconv.Itoa(pagePrev), "> Prev </a>"}, "")
	aNext := strings.Join([]string{"<a class=\"btn-nav-page\" href=?page=", strconv.Itoa(pageNext), "> Next </a>"}, "")
	aUpload := strings.Join([]string{"<a class=\"btn-nav-page\" href=\"", Config.SiteURL, "/upload-files\"", "> Upload </a>"}, "")

	liPage := strings.Join([]string{liRoot, aPrev, aNext, aUpload}, "  |  ")

	liPage = strings.Join([]string{"<div class=\"nav\">", strings.Trim(liPage, "|"), "</div>"}, "")
	return liPage
}

func GenURLList(keys []string, navList string) string {
	li := ""
	theURL := ""
	for _, key := range keys {
		theURL = GenURLByKey(key)
		if theURL != "" {
			li = strings.Join([]string{li, "<li class=\"item\">", theURL, "</li>"}, "")
		}
	}

	liList := strings.Join([]string{navList, "<ol class=\"key-list\">", li, "</ol>"}, "")

	return liList
}

func GenURLByKey(key string) string {
	if key == "" {
		return ""
	}
	uak := strings.Split(key, "/")
	if len(uak) != 3 {
		return ""
	}
	u := uak[0]
	a := uak[1]
	k := uak[2]

	listUserURL := strings.Join([]string{"<a class=\"list-user\" href=\"", Config.SiteURL, "/list/", u, "\">", u, "</a>"}, "")
	listUserAlbumURL := strings.Join([]string{"<a class=\"list-album\" href=\"", Config.SiteURL, "/list/", u, "/", a, "\">", a, "</a>"}, "")
	keyURL := strings.Join([]string{"<a class=\"list-key\" href=\"", Config.SiteURL, "/s/", key, "\">", k, "</a>"}, "")
	deleteKeyURL := strings.Join([]string{"<a class=\"btn btn-delete\" href=\"", Config.SiteURL, "/delete-meta/", key, "\">", "DELETE</a>"}, "")

	itemURL := strings.Join([]string{listUserURL, listUserAlbumURL, keyURL}, "/")
	totalURL := strings.Join([]string{"<span class=\"list-user-album-key\">", itemURL, "</span>", deleteKeyURL}, "")
	return totalURL
}

func GetHandler(ctx *gin.Context) {
	tsStart := time.Now()
	key := ctx.Param("key")
	user := ctx.Param("user")
	album := ctx.Param("album")
	if key == "" || user == "" || album == "" {
		ctx.Data(http.StatusOK, "text/plain; charset=utf-8", []byte("user,album,key cannot be empty"))
		ctx.Abort()
		return
	}

	userAlbumKey := strings.Join([]string{user, album, key}, "/")
	log.Printf("KEY: %v", userAlbumKey)
	mimeDefault := "text/plain; charset=utf-8"
	dataDefault := []byte("")
	ett := entity.NewEntity([]byte(userAlbumKey), nil, nil)

	metadata, err := ett.Get()
	var contentLength int64 = 0

	if err != nil {
		ctx.String(http.StatusOK, strings.Join([]string{"Entity GET Error", err.Error(), userAlbumKey}, ":"))
		ctx.Abort()
		return
	} else {
		var meta entity.EntityMeta
		err := json.Unmarshal(metadata.Meta, &meta)
		if err != nil {
			ctx.String(http.StatusOK, err.Error())
			ctx.Abort()
			return
		} else {
			mimeDefault = meta.Mime
			dataDefault = metadata.Data
			contentLength = meta.Size
		}

	}
	tsDuration := time.Since(tsStart).String()
	ctx.Header("Content-Type", mimeDefault)
	ctx.Header("Content-Length", strconv.FormatInt(contentLength, 10))
	ctx.Header("X-Pfs-Duration", tsDuration)
	ctx.Data(http.StatusOK, mimeDefault, dataDefault)

}

func UploadHandler(ctx *gin.Context) {
	// Multipart form
	result := make(map[string]interface{}, 3)

	uploadDir := Config.UploadDir

	form, _ := ctx.MultipartForm()
	username := ctx.PostForm("username")
	album := ctx.PostForm("album")
	files := form.File["files[]"]
	if username == "" || album == "" || len(files) == 0 {
		result["success"] = []string{}
		result["failure"] = []string{}
		result["error"] = "username/album/files cannot be empty"
		ctx.JSON(http.StatusOK, result)
		ctx.Abort()
		return
	}

	var dst, key, fname string

	successList := []string{}
	failedList := []string{}
	for _, file := range files {
		fname = file.Filename
		fname = strings.ReplaceAll(fname, " ", "-")
		dst = filepath.Join(uploadDir, fname)
		log.Println(file.Filename, ":", dst)
		ctx.SaveUploadedFile(file, dst)

		_, err := os.Stat(dst)
		if err != nil {
			continue
		}

		key = strings.Join([]string{username, album, fname}, "/")
		ett, err := entity.NewEntityByFile(dst, key, "", "")
		if err != nil {
			continue
		}

		ett.Get()
		err = ett.Save()
		if err != nil {
			failedList = append(failedList, strings.Join([]string{key, err.Error()}, ":"))
		} else {
			successList = append(successList, key)
		}

	}

	result["success"] = successList
	result["failure"] = failedList
	if len(successList) == len(files) {
		result["error"] = 0
	} else {
		result["error"] = len(failedList)
	}

	ctx.JSON(http.StatusOK, result)

}

func UploadFilesHandler(ctx *gin.Context) {

	ctx.HTML(http.StatusOK, "upload-files.html", gin.H{})
}

func GetMetaHandler(ctx *gin.Context) {
	tsStart := time.Now()
	key := ctx.Param("key")
	user := ctx.Param("user")
	album := ctx.Param("album")
	if key == "" || user == "" || album == "" {
		ctx.Abort()
		return
	}

	userAlbumKey := strings.Join([]string{user, album, key}, "/")
	log.Printf("KEY: %v", userAlbumKey)
	mimeDefault := "application/json; charset=utf-8"
	ett := entity.NewEntity([]byte(userAlbumKey), nil, nil)

	metadata, err := ett.Get()

	if err != nil {
		ctx.String(http.StatusOK, strings.Join([]string{"Entity GETMETA Error", err.Error(), userAlbumKey}, ":"))
		ctx.Abort()
	}
	tsDuration := time.Since(tsStart).String()
	ctx.Header("Content-Type", mimeDefault)
	ctx.Header("X-Pfs-Duration", tsDuration)
	ctx.String(http.StatusOK, string(metadata.Meta))

}

func DeleteMetaHandler(ctx *gin.Context) {
	tsStart := time.Now()
	key := ctx.Param("key")
	user := ctx.Param("user")
	album := ctx.Param("album")

	result := make(map[string]interface{}, 3)
	result["action"] = "delete"

	mimeDefault := "application/json; charset=utf-8"
	ctx.Header("Content-Type", mimeDefault)

	if key == "" || user == "" || album == "" {
		result["key"] = ""
		result["error"] = "username/album/files cannot be empty"
		ctx.JSON(http.StatusOK, result)
		ctx.Abort()
		return
	}

	userAlbumKey := strings.Join([]string{user, album, key}, "/")
	log.Printf("DELETE KEY: %v", userAlbumKey)

	err := entity.DeleteMetaKey([]byte(userAlbumKey))

	result["key"] = userAlbumKey

	if err != nil {
		result["error"] = err.Error()
	} else {
		result["error"] = 0
	}

	tsDuration := time.Since(tsStart).String()
	ctx.Header("X-Pfs-Duration", tsDuration)
	ctx.JSON(http.StatusOK, result)

}

func ExistsDataHandler(ctx *gin.Context) {
	tsStart := time.Now()
	key := ctx.Param("key")
	if key == "" {
		ctx.Abort()
	}

	mimeDefault := "application/json; charset=utf-8"
	exists := entity.ExistsData([]byte(key))

	res := make(map[string]string, 2)
	res["key"] = key
	if exists {
		res["exists"] = "1"
	} else {
		res["exists"] = "0"
	}

	strRes, err := json.Marshal(res)

	if err != nil {
		ctx.String(http.StatusOK, err.Error())
		ctx.Abort()
	}

	tsDuration := time.Since(tsStart).String()
	ctx.Header("Content-Type", mimeDefault)
	ctx.Header("X-Pfs-Duration", tsDuration)
	ctx.String(http.StatusOK, string(strRes))

}

func StartHttpServer() {
	listening_ip_port := strings.Join([]string{Config.IP, Config.HttpPort}, ":")
	if Config.Debug == true {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	r.LoadHTMLGlob("asset/templates/*")

	r.Use(gzip.Gzip(gzip.DefaultCompression))

	r.GET("/", IndexHandler)

	rList := r.Group("list")
	{
		rList.GET("/", ListHandler)
		rList.GET("/:user", ListByUserHandler)
		rList.GET("/:user/:album", ListByUserAlbumHandler)
	}

	r.GET("/s/:user/:album/:key", GetHandler)
	r.GET("/meta/:user/:album/:key", GetMetaHandler)
	r.GET("/delete-meta/:user/:album/:key", DeleteMetaHandler)
	r.GET("/exists/:key", ExistsDataHandler)
	r.GET("/upload-files/", UploadFilesHandler)

	r.POST("/upload", UploadHandler)

	r.StaticFS("/asset", http.Dir("./asset"))

	r.Run(listening_ip_port)
}
