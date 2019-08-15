package potato

// import (
// 	"github.com/coocood/freecache"
// )

func CacheSet(key, val []byte) (err error) {
	err = cacheFree.Set(key, val, cacheExpiration)
	return err
}
