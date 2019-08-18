package potato

import (
	"sync/atomic"
	//"github.com/coocood/freecache"
)

func cache_set(key, val []byte) (err error) {
	err = cacheFree.Set(key, val, cacheExpiration)
	atomic.AddUint64(&cacheSetCounter, 1)
	return err
}

func cache_get(key []byte) (val []byte, err error) {
	val, err = cacheFree.Get(key)
	if err != nil {
		return nil, err
	}
	return val, nil
}

func cache_del(key []byte) (tf bool) {
	tf = cacheFree.Del(key)
	return tf
}

func CacheGet(key []byte) (val []byte, err error) {
	val, err = cache_get(key)
	if err == nil {
		return val, nil
	}

	val, err = EntityGet([]byte(key))
	if err == nil {
		cache_set(key, val)
		return val, nil
	}

	val, err = EntityGetRoundRobin([]byte(key))
	if err == nil {
		cache_set(key, val)
		return val, nil
	}

	return nil, err
}
