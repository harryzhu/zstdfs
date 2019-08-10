package potato

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	//"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/golang/groupcache"
)

type EntityObject struct {
	Name string
	Mime string
	Size string
	Data []byte
}

type EntityResponse struct {
	URL  string
	Name string
	Mime string
	Size string
}

var r = gin.Default()
var (
	HTTP_SITE_URL        string
	HTTP_TEMP_UPLOAD_DIR string
	FaviconByte          = []byte("")
)

func beforeStartHttpServer() {
	// pre-load favicon file
	if FaviconFile, err := ioutil.ReadFile(cfg.Http.Favicon_file); err == nil {
		FaviconByte = FaviconFile
	} else {
		logger.Warn("cannot read favicon file:", cfg.Http.Favicon_file)
	}

	HTTP_SITE_URL = cfg.Http.Site_url
	HTTP_TEMP_UPLOAD_DIR = cfg.Http.Temp_upload_dir
}

func StartHttpServer() {
	addressHttp := strings.Join([]string{cfg.Http.Ip, cfg.Http.Port}, ":")
	if isDebug == false {
		gin.SetMode(gin.ReleaseMode)
		gin.DisableConsoleColor()
		logfile := cfg.Http.Log_file
		f, _ := os.Create(logfile)
		gin.DefaultWriter = io.MultiWriter(f)
	} else {
		gin.SetMode(gin.DebugMode)

		logfile := cfg.Http.Log_file
		f, _ := os.Create(logfile)
		gin.DefaultWriter = io.MultiWriter(f)

		logger.Info("in DebugMode, log will not flush to disk.")
	}

	if cfg.Http.Cors_enabled == true {
		corsConfig := cors.DefaultConfig()
		corsConfig.AllowOrigins = cfg.Http.Cors_allow_origins
		corsConfig.AllowMethods = cfg.Http.Cors_allow_methods
		corsConfig.AllowHeaders = cfg.Http.Cors_allow_headers
		corsConfig.ExposeHeaders = cfg.Http.Cors_expose_headers
		corsConfig.AllowCredentials = cfg.Http.Cors_allow_credentials
		corsConfig.MaxAge = cfg.Http.Cors_maxage_hours * time.Hour

		r.Use(cors.New(corsConfig))
		logger.Info("CORS is enabled.")
	} else {
		r.Use(cors.Default())
	}

	r.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {

		// your custom format
		return fmt.Sprintf("[%s] - %s %s %s %d %s \"%s\" [%s] \"%s\" \n",
			param.TimeStamp.Format(time.RFC3339),
			param.ClientIP,
			param.Request.Proto,
			param.Method,
			param.StatusCode,
			param.Latency,
			param.Path,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	}))
	r.Use(gin.Recovery())

	//r.Use(Validate())

	r.Use(gzip.Gzip(gzip.DefaultCompression))
	//r.Use(AuthRequired())
	r.LoadHTMLGlob("static/**/*.tmpl")
	r.Use(gin.Recovery())
	r.GET("/", HttpDefaultHome)
	v1 := r.Group("/v1")
	{
		v1.GET("/", HttpHome)
		v1.GET("/ping", HttpPing)
		v1.GET("/k/:key", HttpGroupCache)
		v1.GET("/form-uploads.html", HttpFormFiles)
		v1.GET("/meta-sync-list.html", HttpMetaSyncList)
		v1.GET("/meta-list.html", HttpMetaList)
		v1.GET("/list", HttpList)
		v1.GET("/list-checker", HttpListWithChecker)
		v1.GET("/stats", HttpStats)
		v1.GET("/signin", HttpSignin)
		v1.GET("/checker", HttpChecker)
		v1.GET("/checker/:key", HttpChecker)
		v1.GET("/_groupcache/:key", HttpGroupCache)

		// POST
		v1.POST("/uploads", HttpUpload)
		v1.POST("/auth", HttpAuth)

		// DELETE
		v1.DELETE("/k/:key", HttpDelete)
	}

	r.GET("/favicon.ico", HttpFavicon)

	logger.Info("Endpoint HTTP: ", addressHttp)
	beforeStartHttpServer()
	r.Run(addressHttp)
}

func HttpFavicon(c *gin.Context) {
	c.Data(http.StatusOK, "image/x-icon", FaviconByte)
}

func HttpSignin(c *gin.Context) {
	actionPath := strings.Join([]string{HTTP_SITE_URL, "v1", "auth"}, "/")

	c.Header("Content-Type", "text/html")
	c.HTML(http.StatusOK, "v1/signin.tmpl", gin.H{"action": actionPath})
}

func HttpChecker(c *gin.Context) {
	c.Header("Content-Type", "text/html")

	key := c.Param("key")

	if len(key) > 0 {
		c.HTML(http.StatusOK, "default/checker.tmpl", gin.H{"frmkey": key})
	} else {
		c.HTML(http.StatusOK, "default/checker.tmpl", gin.H{"frmkey": ""})
	}

}

func HttpAuth(c *gin.Context) {
	c.SetCookie("name", "test", 3600, "/", "", false, false)
	c.Redirect(302, "/")
}

func HttpPing(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "pong",
	})
}

func HttpDefaultHome(c *gin.Context) {
	c.HTML(http.StatusOK, "default/index.tmpl", gin.H{})
}

func HttpHome(c *gin.Context) {
	c.HTML(http.StatusOK, "v1/index.tmpl", gin.H{})
}

func HttpDelete(c *gin.Context) {
	key := c.Param("key")
	msg := "key should not be empty"
	if len(key) > 0 {
		err := EntityDelete([]byte(key))
		if err != nil {
			msg = "cannot delete the key."
		} else {
			PeersMark("sync", "del", key, "1")
			msg = "delete successfully."
		}
	}
	c.JSON(200, gin.H{
		"message": msg,
	})
}

func HttpUpload(c *gin.Context) {
	form, _ := c.MultipartForm()
	files := form.File["uploads[]"]
	logger.Debug("files:", len(files))
	var resp []*EntityResponse
	for _, file := range files {
		logger.Debug("uploading: ", file.Filename)
		ftemp := strings.Join([]string{HTTP_TEMP_UPLOAD_DIR, file.Filename}, "/")
		c.SaveUploadedFile(file, ftemp)

		fileData, err := ioutil.ReadFile(ftemp)
		if err == nil {
			if len(fileData) > entityMaxSize {
				continue
			}

			fname := file.Filename
			fsize := strconv.Itoa(len(fileData))
			fmime := mime.TypeByExtension(path.Ext(ftemp))
			if fmime == "" {
				fmime = "application/octet-stream"
			}

			sb := &EntityObject{
				Name: fname,
				Size: fsize,
				Mime: fmime,
				Data: fileData,
			}

			ett := &EntityResponse{
				URL:  "",
				Name: "",
				Size: "",
				Mime: "",
			}

			fileExt := path.Ext(path.Base(file.Filename))
			sb_key := ByteSHA256(fileData)
			sb_key = strings.ToLower(strings.Join([]string{sb_key, fileExt}, ""))

			if EntityExists([]byte(sb_key)) == true {
				ett.URL = strings.Join([]string{HTTP_SITE_URL, "v1", "k", sb_key}, "/")
				ett.Name = sb_key
				ett.Size = fsize
				ett.Mime = fmime

				resp = append(resp, ett)
			} else {
				byteEntityObject, err := json.Marshal(sb)

				if err == nil {
					err := EntitySet([]byte(sb_key), byteEntityObject)
					//PeersMark("sync", "set", sb_key, "1")
					if err != nil {
						logger.Debug("Error while EntitySet: ", sb_key)
					} else {
						PeersMark("sync", "set", sb_key, "1")
						logger.Debug("OK while EntitySet: ", sb_key)
						ett.URL = strings.Join([]string{HTTP_SITE_URL, "v1", "k", sb_key}, "/")
						ett.Name = sb_key
						ett.Size = fsize
						ett.Mime = fmime

						resp = append(resp, ett)
					}

				}
			}

		}

	}
	jsonResponse, err := json.Marshal(resp)
	if err != nil {
		jsonResponse = []byte("")
	}
	c.Header("Content-Type", "text/json")
	c.String(http.StatusOK, string(jsonResponse))
}

func HttpFormFiles(c *gin.Context) {
	formAction := strings.Join([]string{HTTP_SITE_URL, "v1", "uploads"}, "/")
	c.Header("Content-Type", "text/html")
	c.HTML(http.StatusOK, "v1/form-uploads.tmpl", gin.H{"action": formAction})
}

func HttpMetaSyncList(c *gin.Context) {
	listHtml := MetaScanList([]byte("sync/"), 1000)
	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, listHtml)
}

func HttpMetaList(c *gin.Context) {
	listHtml := MetaList(100)
	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, listHtml)
}

func HttpList(c *gin.Context) {
	listHtml := EntityScan("fd")
	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, listHtml)
}

func HttpListWithChecker(c *gin.Context) {
	listHtml := EntityScanHtmlChecker("fd")
	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, listHtml)
}

func HttpStats(c *gin.Context) {
	var stats = make(map[string]string)
	stats["Gets"] = cacheGroup.Stats.Gets.String()
	stats["DBGetCounter"] = strconv.FormatUint(atomic.LoadUint64(&bdbGetCounter), 10)
	stats["DBSetCounter"] = strconv.FormatUint(atomic.LoadUint64(&bdbSetCounter), 10)
	stats["CacheHits"] = cacheGroup.Stats.CacheHits.String()
	stats["PeerLoads"] = cacheGroup.Stats.PeerLoads.String()
	stats["PeerErrors"] = cacheGroup.Stats.PeerErrors.String()
	stats["Loads"] = cacheGroup.Stats.Loads.String()
	stats["LoadsDeduped"] = cacheGroup.Stats.LoadsDeduped.String()
	stats["LocalLoads"] = cacheGroup.Stats.LocalLoads.String()
	stats["LocalLoadErrs"] = cacheGroup.Stats.LocalLoadErrs.String()
	stats["ServerRequests"] = cacheGroup.Stats.ServerRequests.String()
	c.Header("Content-Type", "text/html")
	c.HTML(http.StatusOK, "v1/stats.tmpl", gin.H{"Stats": stats})

}

func HttpGroupCache(c *gin.Context) {
	key := c.Param("key")

	var data []byte

	err := cacheGroup.Get(c, key, groupcache.AllocatingByteSliceSink(&data))

	if err != nil {
		c.Data(http.StatusNotFound, "text/html", []byte("404 NOT Found"))
	}

	var ettobj EntityObject
	erru := json.Unmarshal(data, &ettobj)

	var eo_name, eo_mime, eo_size string
	var eo_data []byte

	if erru == nil {
		eo_name = ettobj.Name
		eo_mime = ettobj.Mime
		eo_size = ettobj.Size
		eo_data = ettobj.Data

		if eo_mime == "" {
			eo_mime = "application/octet-stream"
		}
		c.Header("X-PotatoFS-Name", eo_name)
		c.Header("Content-Length", eo_size)
		c.Data(http.StatusOK, eo_mime, eo_data)

	} else {
		c.Data(http.StatusNotFound, "text/html", []byte("404 NOT Found"))
	}
}
