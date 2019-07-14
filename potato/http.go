package potato

import (
	"encoding/json"
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

var r = gin.Default()

func StartHttpServer() {
	addressHttp := strings.Join([]string{CFG.Http.Ip, CFG.Http.Port}, ":")
	if MODE == "PRODUCTION" {
		gin.SetMode(gin.ReleaseMode)
		gin.DisableConsoleColor()
		logfile := CFG.Http.Log_file
		f, _ := os.Create(logfile)
		gin.DefaultWriter = io.MultiWriter(f)
	} else {
		gin.SetMode(gin.DebugMode)
		Logger.Info("in DebugMode, log will not flush to disk.")
	}

	if CFG.Http.Cors_enabled == true {
		corsConfig := cors.DefaultConfig()
		corsConfig.AllowOrigins = CFG.Http.Cors_allow_origins
		corsConfig.AllowMethods = CFG.Http.Cors_allow_methods
		corsConfig.AllowHeaders = CFG.Http.Cors_allow_headers
		corsConfig.ExposeHeaders = CFG.Http.Cors_expose_headers
		corsConfig.AllowCredentials = CFG.Http.Cors_allow_credentials
		corsConfig.MaxAge = CFG.Http.Cors_maxage_hours * time.Hour

		r.Use(cors.New(corsConfig))
		Logger.Info("CORS is enabled.")
	} else {
		r.Use(cors.Default())
	}

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
		//v1.GET("/k/:key", HttpGet)
		v1.GET("/k/:key", HttpGroupCache)
		v1.GET("/form-uploads.html", HttpFormFiles)
		v1.GET("/meta-sync-list.html", HttpMetaSyncList)
		v1.GET("/meta-list.html", HttpMetaList)
		v1.GET("/list", HttpList)
		v1.GET("/stats", HttpStats)
		v1.GET("/signin", HttpSignin)
		v1.GET("/_groupcache/:key", HttpGroupCache)

		// POST
		v1.POST("/uploads", HttpUpload)
		v1.POST("/auth", HttpAuth)
	}

	r.GET("/favicon.ico", HttpFavicon)
	r.Run(addressHttp)
}

// func Validate() gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		val, _ := c.Cookie("name")
// 		Logger.Info("cookie:", val)
// 		username := c.Query("user01")
// 		password := c.Query("pswd01")

// 		if username == "user01" && password == "pswd01" {
// 			// session := sessions.Default(context)
// 			// session.Set("session_01", "hahaha")
// 			// session.Save()

// 			c.Next()
// 		} else {
// 			//c.Abort()
// 			//c.JSON(http.StatusUnauthorized, gin.H{"message": "authentication failed."})
// 		}

// 	}
// }

func HttpFavicon(c *gin.Context) {
	fileData, err := ioutil.ReadFile(CFG.Http.Favicon_file)
	if err != nil {
		fileData = []byte("")
	}
	c.Data(http.StatusOK, "image/x-icon", fileData)
}

func HttpSignin(c *gin.Context) {
	actionPath := strings.Join([]string{HTTP_SITE_URL, "v1", "auth"}, "/")

	c.Header("Content-Type", "text/html")
	c.HTML(http.StatusOK, "v1/signin.tmpl", gin.H{"action": actionPath})
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
	c.Data(http.StatusOK, "text/html", []byte("OK"))
}

func HttpHome(c *gin.Context) {
	c.HTML(http.StatusOK, "v1/index.tmpl", gin.H{})
}

func HttpGet(c *gin.Context) {

	timer_start := time.Now()
	key := c.Param("key")

	data, err := CacheGet(key)

	if err != nil || data == nil {
		Logger.Debug("cache MISS.")
		c.Header("X-Potatofs-Cache", "MISS")
		data, err = EntityGet(key)
		if err == nil {
			if len(data) <= CACHE_MAX_SIZE {
				CacheSet(key, data)
			}
		}
	} else {
		Logger.Debug("cache HIT.")
		c.Header("X-Potatofs-Cache", "HIT")
	}

	var ettobj EntityObject
	erru := json.Unmarshal(data, &ettobj)

	var eo_name, eo_mime, eo_size string
	var eo_data []byte
	time_response := strings.Join([]string{strconv.Itoa(int(int64(time.Since(timer_start) / time.Millisecond))), "ms"}, " ")

	if erru == nil {
		eo_name = ettobj.Name
		eo_mime = ettobj.Mime
		eo_size = ettobj.Size
		eo_data = ettobj.Data

		Logger.Debug(eo_name)

		c.Header("X-Potatofs-Response-Time", time_response)
		c.Header("Content-Length", eo_size)
		c.Data(http.StatusOK, eo_mime, eo_data)

	} else {
		eo_name = "Error: 404"
		eo_mime = "text/html"
		eo_size = strconv.Itoa(len([]byte("Error:404")))
		eo_data = []byte("Error:404")

		c.Header("X-Potatofs-Response-Time", time_response)
		c.Header("Content-Length", eo_size)
		c.Data(http.StatusNotFound, eo_mime, eo_data)
	}

}

func HttpUpload(c *gin.Context) {
	form, _ := c.MultipartForm()
	files := form.File["uploads[]"]
	Logger.Debug("files:", len(files))
	var resp []*EntityResponse
	for _, file := range files {
		Logger.Debug("uploading: ", file.Filename)
		ftemp := strings.Join([]string{HTTP_TEMP_DIR, file.Filename}, "/")
		c.SaveUploadedFile(file, ftemp)

		fileData, err := ioutil.ReadFile(ftemp)
		if err == nil {
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
			sb_key := ByteMD5(fileData)
			sb_key = strings.ToLower(strings.Join([]string{sb_key, fileExt}, ""))

			if EntityExists(sb_key) == true {
				ett.URL = strings.Join([]string{HTTP_SITE_URL, "v1", "k", sb_key}, "/")
				ett.Name = sb_key
				ett.Size = fsize
				ett.Mime = fmime

				resp = append(resp, ett)
			} else {
				byteEntityObject, err := json.Marshal(sb)

				if err == nil {
					err := EntitySet(sb_key, byteEntityObject)
					if err != nil {
						Logger.Debug("Error while EntitySet: ", sb_key)
					} else {
						Logger.Debug("OK while EntitySet: ", sb_key)
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
	listHtml := MetaSyncList()
	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, listHtml)
}

func HttpMetaList(c *gin.Context) {
	listHtml := MetaList()
	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, listHtml)
}

func HttpList(c *gin.Context) {
	listHtml := EntityScan("fd")
	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, listHtml)
}

func HttpStats(c *gin.Context) {
	var stats = make(map[string]string)
	stats["Gets"] = CACHE_GROUP.Stats.Gets.String()
	stats["DBGetCounter"] = strconv.FormatUint(atomic.LoadUint64(&DBGetCounter), 10)
	stats["DBSetCounter"] = strconv.FormatUint(atomic.LoadUint64(&DBSetCounter), 10)
	stats["CacheHits"] = CACHE_GROUP.Stats.CacheHits.String()
	stats["PeerLoads"] = CACHE_GROUP.Stats.PeerLoads.String()
	stats["PeerErrors"] = CACHE_GROUP.Stats.PeerErrors.String()
	stats["Loads"] = CACHE_GROUP.Stats.Loads.String()
	stats["LoadsDeduped"] = CACHE_GROUP.Stats.LoadsDeduped.String()
	stats["LocalLoads"] = CACHE_GROUP.Stats.LocalLoads.String()
	stats["LocalLoadErrs"] = CACHE_GROUP.Stats.LocalLoadErrs.String()
	stats["ServerRequests"] = CACHE_GROUP.Stats.ServerRequests.String()
	c.Header("Content-Type", "text/html")
	c.HTML(http.StatusOK, "v1/stats.tmpl", gin.H{"Stats": stats})

}

func HttpGroupCache(c *gin.Context) {
	key := c.Param("key")

	Error404 := "404 NOT Found"

	var data []byte

	err := CACHE_GROUP.Get(c, key, groupcache.AllocatingByteSliceSink(&data))

	if err != nil {
		c.Data(http.StatusOK, "text/html", []byte(Error404))
		return
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
		c.Data(http.StatusOK, "text/html", []byte(Error404))
	}
}
