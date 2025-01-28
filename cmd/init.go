package cmd

func init() {

	//defer pprof.StopCPUProfile()
	MakeDirs("data")
	MakeDirs("data/fbin")
	MakeDirs("data/_sync")
	MakeDirs("data/logs")
	//
	MakeDirs("www")
	MakeDirs("www/uploads")
	MakeDirs("www/export")
	MakeDirs("www/temp")
	MakeDirs("www/assets")
	MakeDirs("www/static")
	//
	DefaultAsset("www/assets/video-js.min.css", "template/video-js.min.css")
	DefaultAsset("www/assets/video.min.js", "template/video.min.js")
	DefaultAsset("www/assets/style.css", "template/style.css")
	DefaultAsset("www/assets/favicon.png", "template/favicon.png")
	DefaultAsset("www/assets/video-bg.png", "template/video-bg.png")
	//
	MaxUploadSize = Int2Int64(MaxUploadSizeMB * MB)
	//
	if AdminUser != "" && AdminPassword != "" {
		userpass[AdminUser] = AdminPassword
	}
	//
	aesKey = []byte(SHA256String(GetEnv("zstdfsaeskey", aesKeyDefault))[10:42])
	aesPlaintext := "hello zstdfs!"
	if aesPlaintext != string(DecryptAES(EncryptAES([]byte(aesPlaintext)))) {
		FatalError("AES enc/dec:", ErrAESInvalid)
	}
}
