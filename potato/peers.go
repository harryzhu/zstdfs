package potato

// import (
// 	"strings"
// )

func PeersMark(cat, action, key string) error {
	if isMaster == true && volumePeersLength > 0 {
		var key string
		for _, peer := range volumePeers {
			if key = metaKeyJoin(cat, action, peer, key); len(key) > 0 {
				MetaSet([]byte(key), []byte("1"))
			}
		}
	}

	return nil
}
