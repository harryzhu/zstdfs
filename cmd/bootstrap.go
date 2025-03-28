package cmd

import (
	//"os"
	"fmt"
	"path/filepath"

	//"fmt"
	"time"
)

func BeforeStart() error {
	DATA_DIR = GetEnv("zstdfs_data", "data/zstdfs")
	TEMP_DIR = ToUnixSlash(filepath.Join(DATA_DIR, "www/temp"))
	ASSET_DIR = ToUnixSlash(filepath.Join(DATA_DIR, "www/assets"))
	CACHE_DIR = ToUnixSlash(filepath.Join(DATA_DIR, "www/cache"))
	if UploadDir == "" {
		UploadDir = ToUnixSlash(filepath.Join(DATA_DIR, "www/uploads"))
	}

	if StaticDir == "" {
		STATIC_DIR = ToUnixSlash(filepath.Join(DATA_DIR, "www/static"))
	}
	MakeDirs(DATA_DIR)
	MakeDirs(UploadDir)
	MakeDirs(TEMP_DIR)
	MakeDirs(CACHE_DIR)
	MakeDirs(ASSET_DIR)
	MakeDirs(STATIC_DIR)
	//
	if CacheTimeout > 0 {
		FunctionCacheExpires = CacheTimeout
	}
	//
	DefaultAsset(ASSET_DIR+"/video-js.min.css", "template/video-js.min.css")
	DefaultAsset(ASSET_DIR+"/video.min.js", "template/video.min.js")
	DefaultAsset(ASSET_DIR+"/style.css", "template/style.css")
	DefaultAsset(ASSET_DIR+"/favicon.png", "template/favicon.png")
	DefaultAsset(STATIC_DIR+"/test.jpg", "template/bg-01.jpg")
	DefaultAsset(STATIC_DIR+"/example.jpg", "template/bg-02.jpg")
	//
	DebugInfo("BeforeStart:Debug", IsDebug)
	DebugInfo("BeforeStart:DATA_DIR", DATA_DIR)
	DebugInfo("BeforeStart:FunctionCacheExpires", FunctionCacheExpires)
	//
	sqldb = mysqlConnect()
	mgodb = mongoConnect()
	bgrdb = badgerConnect()

	//
	mysqlPing(sqldb)
	ts := time.Now().Unix()

	mongoAdminSetIfEmpty("system_info", "init_boot", Int64ToString(ts))
	mongoAdminUpsert("system_info", "last_boot", Int64ToString(ts))
	//
	EntitySaveSmoke()

	return nil
}

func EntitySaveSmoke() bool {
	testUser := "harry"
	mongoUserCollectionInit(testUser)
	mongoAdminCreateIndex(testUser)
	DebugInfo("Visit test.jpg", fmt.Sprintf("http://%s/f/%s/%s", SiteURL(), testUser, "test.jpg"))

	return true
}
