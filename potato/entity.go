package potato

import (
	//"context"
	"errors"
	//"io"
	"strings"
	"time"

	pbv "./pb/volume"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type Entity struct {
	Key  string
	Data []byte
}

// swagger:operation EntitySet
func EntitySet(key string, data []byte) error {
	err := db_set(key, data)
	if err != nil {
		return err
	}
	if IsMaster == true && SLAVES_LENGTH > 0 {
		for _, slave := range SLAVES {
			prefix := strings.Join([]string{"sync", slave}, ":")
			MetaSet(prefix, key, []byte("0"))

		}
	}

	return nil
}

func EntityGet(key string) ([]byte, error) {
	v, err := db_get(key)
	if err != nil {
		return nil, err
	}
	return v, nil
}

func EntityGetRoundRobin(key string) ([]byte, error) {
	Logger.Info("EntityGetRoundRobin:", key)
	var data []byte
	var err_return error = errors.New("not exist")

	for ip_port, is_live := range VOLUME_PEERS_LIVE {
		if err_return == nil {
			break
		}

		if is_live == false {
			continue
		}
		Logger.Info("EntityGetRoundRobin:ip_port:", ip_port)
		conn, err := grpc.Dial(ip_port, grpc.WithInsecure(),
			grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(GRPCMAXMSGSIZE), grpc.MaxCallRecvMsgSize(GRPCMAXMSGSIZE)))
		if err != nil {
			continue
		}
		defer conn.Close()

		client := pbv.NewVolumeServiceClient(conn)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		fileRead, err := client.ReadFile(ctx, &pbv.File{Key: key})
		if err != nil {
			continue
		}
		data = fileRead.Data
		err_return = nil
		break

	}
	return data, err_return
}

func EntityDelete(key string) error {
	err := db_delete(key)
	if err != nil {
		return err
	}
	return nil
}

func EntityExists(key string) bool {
	_, err := db_get(key)
	if err != nil {
		return false
	}
	return true
}

func EntityCompaction() error {
	Logger.Debug("DB Compaction is starting...")
	db_compact()
	return nil
}

func EntityScan(prefix string) string {
	keys := db_scan()
	listHtml := ""
	href := strings.Join([]string{"<a href=\"", CFG.Http.Site_url, "/v1/k"}, "")
	if len(keys) > 0 {
		for _, v := range keys {
			listHtml = strings.Join([]string{href, "/", v, "\">", v, "</a><br/>", listHtml}, "")
		}
	}

	return listHtml
}

func EntityScanHtmlChecker(prefix string) string {
	keys := db_scan()
	listHtml := ""
	href := strings.Join([]string{"<a href=\"", CFG.Http.Site_url, "/v1/checker"}, "")
	if len(keys) > 0 {
		for _, v := range keys {
			listHtml = strings.Join([]string{href, "/", v, "\">/v1/checker/", v, "</a><br/>", listHtml}, "")
		}
	}

	return listHtml
}
