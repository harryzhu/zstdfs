package potato

import (
	"errors"
	//"net/http"

	//"github.com/golang/groupcache"
	//pb "github.com/golang/groupcache/groupcachepb"
)

func CacheGet(key string) ([]byte, error) {

	//data, err := getByteFromPeer(CACHEGROUPNAME, key, CACHE_PEERS)
	data := []byte("")
	err := errors.New("error")
	if err != nil {
		Logger.Debug("Error: CacheGet: ", CACHEGROUPNAME, ": key: ", key, ": ", err)
		return nil, err
	}
	return data, nil
}

func CacheSet(key string, data []byte) error {

	return nil
}
