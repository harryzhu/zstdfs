package potato

import (
	"encoding/json"
	//"io"
	"io/ioutil"
	"mime"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang/groupcache"
)

func StartHttpServer() {
	addressHttp := strings.Join([]string{CFG.Http.Ip, CFG.Http.Port}, ":")
	if MODE == "PRODUCTION" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}
	r := gin.Default()
	r.Use(gin.Recovery())
	v1 := r.Group("/v1")
	{
		v1.GET("/ping", HttpPing)
		v1.GET("/k/:key", HttpGet)
		v1.GET("/form-files.html", HttpFormFiles)
		v1.GET("/meta-sync-list.html", HttpMetaSyncList)
		v1.GET("/meta-list.html", HttpMetaList)
		v1.GET("/list", HttpList)
		v1.GET("/_groupcache/:key", HttpGroupCache)
		v1.POST("/uploads", HttpUpload)
	}

	r.GET("/favicon.ico", HttpFavicon)
	r.Run(addressHttp)
}

func HttpFavicon(c *gin.Context) {
	c.Data(http.StatusOK, "image/x-icon", []byte("-"))
}

func HttpPing(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "pong",
	})
}

func HttpGet(c *gin.Context) {
	timer_start := time.Now()
	key := c.Param("key")

	data, err := CacheGet(key)

	if err != nil || data == nil {
		Logger.Debug("cache miss.")
		c.Header("X-Potatofs-Cache", "MISS")
		data, err = EntityGet(key)
		if err == nil {
			if len(data) <= CACHE_MAX_SIZE {
				CacheSet(key, data)
			}
		}
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
	Logger.Debug("fffiles:", len(files))
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

			sb_key := ByteMD5(fileData)
			if EntityExists(sb_key) == true {
				ett.URL = strings.Join([]string{HTTP_SITE_URL, "v1", "s", sb_key}, "/")
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
						ett.URL = strings.Join([]string{HTTP_SITE_URL, "v1", "s", sb_key}, "/")
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
	f1 := `<form class="form-files" method="POST" action="`
	f2 := strings.Join([]string{HTTP_SITE_URL, "v1", "uploads"}, "/")
	f3 := `" enctype="multipart/form-data">
		<input type="file" class="frm-file" name="uploads[]" multiple />
		<input type="submit" class="frm-submit" value="Upload">
</form>
	`
	f := strings.Join([]string{f1, f2, f3}, "")
	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, f)
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

func HttpGroupCache(c *gin.Context) {
	key := c.Param("key")

	Logger.Info("adfasdadsa:", key)

	var data []byte

	CACHE_GROUP.Get(nil, key, groupcache.AllocatingByteSliceSink(&data))
	Logger.Info("rrrrr:", len(data))

	var ettobj EntityObject
	//erru := json.Unmarshal(data, &ettobj)
	erru := json.Unmarshal(data, &ettobj)

	var eo_name, eo_mime, eo_size string
	var eo_data []byte

	if erru == nil {
		eo_name = ettobj.Name
		eo_mime = ettobj.Mime
		eo_size = ettobj.Size
		eo_data = ettobj.Data

		Logger.Debug(eo_name)

		c.Header("Content-Length", eo_size)
		c.Data(http.StatusOK, eo_mime, eo_data)

	} else {
		m404 := "404 NOT Found"

		c.Header("Content-Length", strconv.Itoa(len([]byte(m404))))
		c.Data(http.StatusOK, "text/html", []byte(m404))
	}
}
