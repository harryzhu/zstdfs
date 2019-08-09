package potato

// import (
// 	"strings"
// )

func PeersMark(cat, action, key, val string) error {
	if isMaster == true && volumePeersLength > 0 {
		var metakey string
		for _, peer := range volumePeers {
			metakey = metaKeyJoin(cat, action, peer, key)
			MetaSet([]byte(metakey), []byte(val))
		}
	}

	return nil
}
