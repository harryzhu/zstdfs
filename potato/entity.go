package potato

import (
	"encoding/json"
	//"context"
	//"errors"
	//"io"
	"strings"
	//"time"
	//pbv "./pb/volume"
	//"golang.org/x/net/context"
	//"google.golang.org/grpc"
)

type Entity struct {
	Key  []byte
	Data []byte
}

// swagger:operation EntitySet
func EntitySet(key []byte, data []byte) error {
	if err := bdb_set(key, data); err != nil {
		return err
	}
	return nil
}

func EntityGet(key []byte) ([]byte, error) {
	data, err := bdb_get(key)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func EntityDelete(key []byte) error {
	if EntityExists(key) == true {
		if err := bdb_delete(key); err != nil {
			return err
		}
	}

	return nil
}

func EntityBan(key []byte) error {
	if EntityExists(key) == false {
		return nil
	}

	data, err := EntityGet(key)

	if err != nil {
		return err
	}

	var ettobj EntityObject
	err = json.Unmarshal(data, &ettobj)

	if err != nil {
		return err
	}

	sb := &EntityObject{
		Name: ettobj.Name,
		Size: ettobj.Size,
		Mime: ettobj.Mime,
		Stat: -1,
		Data: ettobj.Data,
	}

	byteEntityObject, err := json.Marshal(sb)

	if err != nil {
		return err
	}

	err = EntitySet(key, byteEntityObject)

	if err != nil {
		return err
	}

	return nil
}

func EntityExists(key []byte) bool {
	_, err := bdb_get(key)
	if err != nil {
		return false
	}
	return true
}

func EntityScan(prefix string) string {
	keys := bdb_scan()
	listHtml := ""
	href := strings.Join([]string{"<a href=\"", cfg.Http.Site_url, "/v1/k"}, "")
	if len(keys) > 0 {
		for _, v := range keys {
			listHtml = strings.Join([]string{href, "/", v, "\">", v, "</a><br/>", listHtml}, "")
		}
	}

	return listHtml
}

func EntityScanHtmlChecker(prefix string) string {
	keys := bdb_scan()
	listHtml := ""
	href := strings.Join([]string{"<a href=\"", cfg.Http.Site_url, "/v1/checker"}, "")
	if len(keys) > 0 {
		for _, v := range keys {
			listHtml = strings.Join([]string{href, "/", v, "\">/v1/checker/", v, "</a><br/>", listHtml}, "")
		}
	}

	return listHtml
}
