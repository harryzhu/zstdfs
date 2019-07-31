package potato

import (
	"errors"

	//"github.com/golang/groupcache"
	pb "github.com/golang/groupcache/groupcachepb"
)

func CacheGet(key string) ([]byte, error) {
	_,ok:=CACHE_PEERS.PickPeer(key)
	if ok == false{
		Logger.Debug("CACHE_PEERS.PickPeer: Error: ",key)
	}

	Logger.Debug("CACHE_GROUP.Name: ",CACHE_GROUP.Name)

	data, err := getByteFromPeer(CACHEGROUPNAME, key)
	if err != nil {
		Logger.Debug("Error: CacheGet: ",CACHEGROUPNAME,": key: ",key, ": ", err)
		return nil, err
	}
	return data, nil
}

func CacheSet(key string, data []byte) error {

	return nil
}

func getByteFromPeer(groupName, key string) ([]byte, error) {
	req := &pb.GetRequest{Group: &groupName, Key: &key}
	res := &pb.GetResponse{}

	peer, ok := CACHE_PEERS.PickPeer(key)
	if ok == false {
		Logger.Debug("getByteFromPeer: cannot PickPeer.")
		return nil, errors.New("getByteFromPeer: cannot PickPeer.")
	}

	err := peer.Get(nil, req, res)
	if err != nil {
		Logger.Debug("getByteFromPeer: cannot get by key:", key)
		return nil, err
	}
	return res.Value, nil
}
