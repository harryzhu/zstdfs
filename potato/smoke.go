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

	testMetaKey := ""
	testMetaVal := []byte(testVal)
	for _, p := range volumePeers {
		if len(p) > 0 {
			testMetaKey = metaKeyJoin("test", "get", p, ByteSHA256([]byte(testKey)))
			logger.Info("smokeTest: MetaKeyJoin: ", testMetaKey)
			logger.Info("smokeTest: MetaKeySplit: ", strings.Join(metaKeySplit(testMetaKey), ";"))
			if len(metaKeySplit(testMetaKey)) != 4 {
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
