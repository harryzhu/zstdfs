package potato

func init() {
	loadConfigFromFile()
	openDatabase()
	openMetaCollection()
	openCacheCollection()
	smokeTest()
	getSlavesLength()
	IsDBValueLogGCNeeded = true
	setEntityMaxSize()
	setCacheMaxSize()
	setIsMaster()
}
