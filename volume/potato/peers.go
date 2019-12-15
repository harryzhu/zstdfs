package potato

// import (
// 	"strings"
// )

func PeersMark(cat, action, key, val string) error {
	var metakey string
	for _, peer := range volumePeers {
		metakey = metaKeyJoin(cat, peer, key)
		logger.Debug("PeersMark:", metakey, ":", val)
		MetaSet([]byte(metakey), []byte(val))
	}

	return nil
}
