package cmd

import (
	"fmt"
	"path/filepath"
)

func BeforeHTTPStart() error {
	DataDir = GetEnv("zstdfs_data", "data/zstdfs")
	DebugInfo("BeforeStart: DataDir", DataDir)

	//
	TempDir = ToUnixSlash(filepath.Join(DataDir, "www/temp"))
	AssetDir = ToUnixSlash(filepath.Join(DataDir, "www/assets"))
	CacheDir = ToUnixSlash(filepath.Join(DataDir, "www/cache"))
	if UploadDir == "" {
		UploadDir = ToUnixSlash(filepath.Join(DataDir, "www/uploads"))
	}

	if StaticDir == "" {
		StaticDir = ToUnixSlash(filepath.Join(DataDir, "www/static"))
	}

	if ThumbDir == "" {
		ThumbDir = ToUnixSlash(filepath.Join(DataDir, "www/thumb"))
	}
	MakeDirs(DataDir)
	MakeDirs(UploadDir)
	MakeDirs(TempDir)
	MakeDirs(CacheDir)
	MakeDirs(AssetDir)
	MakeDirs(StaticDir)
	MakeDirs(ThumbDir)
	//
	ChModDir(UploadDir, 0777)
	ChModDir(TempDir, 0777)
	ChModDir(CacheDir, 0777)
	ChModDir(StaticDir, 0777)
	ChModDir(ThumbDir, 0777)
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
	DefaultAsset(AssetDir+"/zstdfs.css", "template/zstdfs.css")
	DefaultAsset(AssetDir+"/style.css", "template/style.css")
	DefaultAsset(AssetDir+"/favicon.png", "template/favicon.png")
	DefaultAsset(AssetDir+"/403-logo.png", "template/403-logo.png")
	DefaultAsset(AssetDir+"/404-logo.png", "template/404-logo.png")
	DefaultAsset(AssetDir+"/500-logo.png", "template/500-logo.png")
	DefaultAsset(AssetDir+"/thumb_logo_empty.png", "template/thumb_logo_empty.png")
	DefaultAsset(AssetDir+"/banned-logo.png", "template/banned-logo.png")
	DefaultAsset(AssetDir+"/document-logo.png", "template/document-logo.png")
	DefaultAsset(AssetDir+"/thumb_logo_zip.png", "template/thumb_logo_zip.png")
	DefaultAsset(AssetDir+"/thumb_logo_txt.png", "template/thumb_logo_txt.png")
	DefaultAsset(AssetDir+"/thumb_logo_pdf.png", "template/thumb_logo_pdf.png")
	DefaultAsset(StaticDir+"/test.jpg", "template/bg-01.jpg")
	//
	DefaultAsset(StaticDir+"/example.jpg", "template/bg-02.jpg")
	//
	binFileDocumentLogo = LoadFileBytes(AssetDir + "/document-logo.png")
	bin403Logo = LoadFileBytes(AssetDir + "/403-logo.png")
	bin404Logo = LoadFileBytes(AssetDir + "/404-logo.png")
	bin500Logo = LoadFileBytes(AssetDir + "/500-logo.png")
	binBannedLogo = LoadFileBytes(AssetDir + "/banned-logo.png")
	binEmptyLogo = LoadFileBytes(AssetDir + "/thumb_logo_empty.png")
	//
	if MinTopCaption < 1 {
		MinTopCaption = 1
	}
	//
	DebugInfo("BeforeStart:Debug", IsDebug)
	DebugInfo("BeforeStart:DataDir", DataDir)
	DebugInfo("BeforeStart:FunctionCacheExpires", FunctionCacheExpires)
	//
	bigcacheInit()
	SetGrpcClient()

	sqldb = mysqlConnect()
	mgodb = mongoConnect()

	mysqlPing(sqldb)
	ts := GetNowUnix()

	mongoAdminSetIfEmpty(AdminBucket, "system_info", "init_boot", Int64ToString(ts))
	mongoAdminUpsert(AdminBucket, "system_info", "last_boot", Int64ToString(ts), true)
	//
	EntitySaveSmoke()
	return nil
}

func EntitySaveSmoke() bool {
	mongoUserCollectionInit(testUser)
	mongoAdminCreateIndex(testUser)
	DebugInfo("Visit test.jpg", fmt.Sprintf("%s/f/%s/%s", GetSiteURL(), testUser, testKey))

	return true
}
