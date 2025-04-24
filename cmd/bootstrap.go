package cmd

import (
	"fmt"
	"path/filepath"
	"time"
)

func BeforeStart() error {
	DataDir = GetEnv("zstdfs_data", "data/zstdfs")
	TempDir = ToUnixSlash(filepath.Join(DataDir, "www/temp"))
	AssetDir = ToUnixSlash(filepath.Join(DataDir, "www/assets"))
	CacheDir = ToUnixSlash(filepath.Join(DataDir, "www/cache"))
	if UploadDir == "" {
		UploadDir = ToUnixSlash(filepath.Join(DataDir, "www/uploads"))
	}

	if StaticDir == "" {
		StaticDir = ToUnixSlash(filepath.Join(DataDir, "www/static"))
	}
	MakeDirs(DataDir)
	MakeDirs(UploadDir)
	MakeDirs(TempDir)
	MakeDirs(CacheDir)
	MakeDirs(AssetDir)
	MakeDirs(StaticDir)
	//
	if IsDebug {
		FunctionCacheExpires = 0
	}
	//
	minDiggCount = Str2Int(GetEnv("zstdfs_min_digg_count", Int2Str(minDiggCount)))
	minCommentCount = Str2Int(GetEnv("zstdfs_min_comment_count", Int2Str(minCommentCount)))
	minCollectCount = Str2Int(GetEnv("zstdfs_min_collect_count", Int2Str(minCollectCount)))
	minShareCount = Str2Int(GetEnv("zstdfs_min_share_count", Int2Str(minShareCount)))
	minDownloadCount = Str2Int(GetEnv("zstdfs_min_download_count", Int2Str(minDownloadCount)))
	//
	DefaultAsset(AssetDir+"/video-js.min.css", "template/video-js.min.css")
	DefaultAsset(AssetDir+"/video.min.js", "template/video.min.js")
	DefaultAsset(AssetDir+"/style.css", "template/style.css")
	DefaultAsset(AssetDir+"/favicon.png", "template/favicon.png")
	DefaultAsset(StaticDir+"/test.jpg", "template/bg-01.jpg")
	DefaultAsset(StaticDir+"/example.jpg", "template/bg-02.jpg")
	//
	DebugInfo("BeforeStart:Debug", IsDebug)
	DebugInfo("BeforeStart:DataDir", DataDir)
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

	if BulkLoadDir != "" && BulkLoadExt != "" && BulkLoadUser != "" {
		badgerBulkLoad(BulkLoadDir, BulkLoadExt)
	}
	return nil
}

func EntitySaveSmoke() bool {
	mongoUserCollectionInit(testUser)
	mongoAdminCreateIndex(testUser)
	DebugInfo("Visit test.jpg", fmt.Sprintf("%s/f/%s/%s", GetSiteURL(), testUser, testKey))

	return true
}
