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
)

func StartHttpServer() {
	addressHttp := strings.Join([]string{CFG.Http.Ip, CFG.Http.Port}, ":")
	r := gin.Default()
	v1 := r.Group("/v1")
	{
		v1.GET("/ping", HttpPing)
		v1.GET("/s/:key", HttpGet)
		v1.GET("/form-files.html", HttpFormFiles)
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
	var data []byte
	var err error
	data, err = CacheGet(key)
	if err != nil {
		Logger.Debug("cache miss.")
	} else {
		Logger.Debug("cache hit.")
		data, err = EntityGet(key)
		if err != nil {
			Logger.Debug("db miss.")
			CacheSet(key, data)
		}
	}

	if err == nil {
		var ettobj EntityObject
		erru := json.Unmarshal(data, &ettobj)

		var eo_name, eo_mime, eo_size string
		var eo_data []byte

		if erru == nil {
			eo_name = ettobj.Name
			eo_mime = ettobj.Mime
			eo_size = ettobj.Size
			eo_data = ettobj.Data
		}
		Logger.Debug(eo_name)
		// c.Header("Content-Type", "image/jpeg")
		time_response := strings.Join([]string{strconv.Itoa(int(int64(time.Since(timer_start) / time.Millisecond))), "ms"}, " ")
		c.Header("X-Potatofs-Response-Time", time_response)
		c.Header("Content-Length", eo_size)
		c.Data(http.StatusOK, eo_mime, eo_data)
	} else {
		c.String(http.StatusNotFound, "Error 404")
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

			byteEntityObject, err := json.Marshal(sb)

			if err == nil {
				sb_key := ByteMD5(fileData)
				sb_data := byteEntityObject

				err := EntitySet(sb_key, sb_data)
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
		<input type="file" name="uploads[]" multiple />
		<input type="submit" value="Upload">
</form>
	`
	f := strings.Join([]string{f1, f2, f3}, "")
	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, f)
}
