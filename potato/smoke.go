package potato

import (
	"strings"
)

func smokeTest() {
	testKey := "testKey"
	testVal := "testVal"

	err := EntitySet([]byte(testKey), []byte(testVal))
	if err != nil {
		logger.Fatal("smokeTest: EntitySet Failed.")
	}

	if EntityExists([]byte(testKey)) == false {
		logger.Fatal("smokeTest: EntityExists Failed.")
	}

	data, err := EntityGet([]byte(testKey))
	if err != nil {
		logger.Fatal("smokeTest: EntityGet Failed.")
	}

	if string(data) != testVal {
		logger.Fatal("smokeTest: EntityGet data unzip failed: ", string(data))
	}

	err = EntityDelete([]byte(testKey))
	if err != nil {
		logger.Fatal("smokeTest: EntityDelete Failed.")
	}

	cacheFree.Set([]byte(testKey), []byte(testVal), 60)
	cget, err := cacheFree.Get([]byte(testKey))
	if err != nil {
		logger.Fatal("smokeTest: Cache Error: ", err)
	} else {
		logger.Info("smokeTest: Cache Get OK.", string(cget))
	}
	cache_affected := cacheFree.Del([]byte(testKey))
	logger.Info("smokeTest: Cache Delete: ", cache_affected)

	testMetaKey := ""
	testMetaVal := []byte(testVal)
	for _, p := range volumePeers {
		if len(p) > 0 {
			testMetaKey = metaKeyJoin("test", "get", p, ByteSHA256([]byte(testKey)))
			logger.Info("smokeTest: MetaKeyJoin: ", testMetaKey)
			logger.Info("smokeTest: MetaKeySplit: ", strings.Join(metaKeySplit(testMetaKey), ";"))
			if len(metaKeySplit(testMetaKey)) != 5 {
				logger.Fatal("smokeTest: MetaKeySplit: Error.")
			}

			if nil != MetaSet([]byte(testMetaKey), testMetaVal) {
				logger.Fatal("smokeTest: MetaSet failed: ")
			}
			val, err := MetaGet([]byte(testMetaKey))
			if err != nil {
				logger.Fatal("smokeTest: MetaGet failed: ", err)
			} else {
				if string(val) != testVal {
					logger.Fatal("smokeTest: MetaGet value failed.")
				}
			}

			if nil != MetaDelete([]byte(testMetaKey)) {
				logger.Fatal("smokeTest: MetaDelete failed.")
			}
		}
	}
}
