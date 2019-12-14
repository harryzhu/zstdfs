package potato

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

func EntityExists(key []byte) bool {
	if true == bdb_exists(key) {
		return true
	}
	return false
}

// func EntityKeyScan(prefix []byte, pageNum int) string {
// 	keys := bdb_key_scan(prefix, pageNum)
// 	listHtml := ""
// 	href := strings.Join([]string{"<a href=\"", cfg.Http.Site_url, "/v1/k"}, "")
// 	if len(keys) > 0 {
// 		btns := ""

// 		for _, v := range keys {
// 			btns = strings.Join([]string{
// 				"<td><a class=\"ett-btn\" href=\"", HTTP_SITE_URL, "/v1/del/k/", v, "\">Del</a></td>",
// 				"<td><a class=\"ett-btn\" href=\"", HTTP_SITE_URL, "/v1/ban/k/", v, "\">Ban</a></td>",
// 				"<td><a class=\"ett-btn\" href=\"", HTTP_SITE_URL, "/v1/pub/k/", v, "\">Pub</a></td>",
// 			}, "")
// 			listHtml = strings.Join([]string{"<tr><td>", href, "/", v, "\">", v, "</a></td>", btns, "</tr>", listHtml}, "")
// 		}
// 	}
// 	listHtml = strings.Join([]string{"<table class=\"ett-list\">", listHtml, "</table>"}, "")
// 	return listHtml
// }

// func EntityScanHtmlChecker(prefix string) string {
// 	keys := bdb_scan()
// 	listHtml := ""
// 	href := strings.Join([]string{"<a href=\"", cfg.Http.Site_url, "/v1/checker"}, "")
// 	if len(keys) > 0 {
// 		for _, v := range keys {
// 			listHtml = strings.Join([]string{href, "/", v, "\">/v1/checker/", v, "</a><br/>", listHtml}, "")
// 		}
// 	}

// 	return listHtml
// }
