package potato

import (
	"errors"

	//"github.com/golang/groupcache"
	pb "github.com/golang/groupcache/groupcachepb"
)

func CacheGet(key string) ([]byte, error) {
	data, err := getFromPeer(CACHEGROUPNAME, key)
	if err != nil {
		Logger.Debug("CacheGet: ", err)
		return nil, err
	}
	return data, nil
}

func CacheSet(key string, data []byte) error {

	return nil
}

func getFromPeer(groupName, key string) ([]byte, error) {
	req := &pb.GetRequest{Group: &groupName, Key: &key}
	res := &pb.GetResponse{}

	peer, ok := CACHE_PEERS.PickPeer(key)
	if ok == false {
		Logger.Debug("getFromPeer: cannot PickPeer.")
		return nil, errors.New("getFromPeer: cannot PickPeer.")
	}

	err := peer.Get(nil, req, res)
	if err != nil {
		Logger.Debug("getFromPeer: cannot get by key:", key)
		return nil, err
	}
	return res.Value, nil
}
