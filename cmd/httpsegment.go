package cmd

import (
	"sort"
	//"encoding/json"
	"fmt"
	"strings"
	//"go.mongodb.org/mongo-driver/v2/bson"
)

func genMetaStatistics(ss map[string]string) (map[string]string, error) {
	counter_set := make(map[string]string)

	digg_count, ok := ss["stats_digg_count"]
	if ok && Str2Int(digg_count) > 100000 {
		//counter_set["点赞"] = fmt.Sprintf("%v", ToKWM(digg_count))
		counter_set["点赞"] = ToKWM(Str2Int(digg_count))
	}
	comment_count, ok := ss["stats_comment_count"]
	if ok && Str2Int(comment_count) > 20000 {
		counter_set["评论"] = ToKWM(Str2Int(comment_count))
	}
	collect_count, ok := ss["stats_collect_count"]
	if ok && Str2Int(collect_count) > 10000 {
		counter_set["收藏"] = ToKWM(Str2Int(collect_count))
	}
	share_count, ok := ss["stats_share_count"]
	if ok && Str2Int(share_count) > 10000 {
		counter_set["转发"] = ToKWM(Str2Int(share_count))
	}
	download_count, ok := ss["stats_download_count"]
	if ok && Str2Int(download_count) > 10000 {
		counter_set["下载"] = ToKWM(Str2Int(download_count))
	}
	return counter_set, nil
}

func genNavFileList(files []string, fkey, uname string) []map[string]map[string]any {
	var navFileKVS []map[string]map[string]any
	var a_text string
	var lineMeta map[string]string

	for _, line := range files {
		navFileList := make(map[string]map[string]any)
		if line != "" && uname != "" {
			if fkey == "" {
				lineMeta = mongoGet(uname, line)
				a_text = line

				navFileList[a_text] = map[string]any{
					"uid":   fmt.Sprintf("%s/%s", uname, line),
					"uri":   fmt.Sprintf("%s/%s", uname, lineMeta["uri"]),
					"size":  ToKMGTB(Str2Int(lineMeta["size"])),
					"mtime": UnixFormat(Int2Int64(Str2Int(lineMeta["mtime"])), "06-01-02 15:04"),
				}
			} else {
				lineMeta = mongoGet(uname, strings.Join([]string{fkey, line}, "/"))
				a_text = strings.TrimPrefix(line, fkey)

				navFileList[a_text] = map[string]any{
					"uid":   fmt.Sprintf("%s/%s/%s", uname, fkey, line),
					"uri":   fmt.Sprintf("%s/%s", uname, lineMeta["uri"]),
					"size":  ToKMGTB(Str2Int(lineMeta["size"])),
					"mtime": UnixFormat(Int2Int64(Str2Int(lineMeta["mtime"])), "06-01-02 15:04"),
				}
			}

			navFileList[a_text]["site_url"] = GetSiteURL()

			if lineMeta["dot_color"] != "" {
				navFileList[a_text]["dot_color"] = lineMeta["dot_color"]
				navFileList[a_text]["is_having_dot_color"] = "1"
			}

			if strings.Index(lineMeta["mime"], "video") > -1 || strings.Index(lineMeta["mime"], "mpeg") > -1 {
				navFileList[a_text]["is_video"] = "1"
			}
			if lineMeta["tags"] != "" {
				metaTags := strings.Split(lineMeta["tags"], ",")
				navFileList[a_text]["tags"] = metaTags
			}

			ss := make(map[string]string)
			if lineMeta["stats_digg_count"] != "" {
				ss["stats_digg_count"] = lineMeta["stats_digg_count"]
			}
			if lineMeta["stats_comment_count"] != "" {
				ss["stats_comment_count"] = lineMeta["stats_comment_count"]
			}
			if lineMeta["stats_collect_count"] != "" {
				ss["stats_collect_count"] = lineMeta["stats_collect_count"]
			}
			if lineMeta["stats_share_count"] != "" {
				ss["stats_share_count"] = lineMeta["stats_share_count"]
			}
			if lineMeta["stats_download_count"] != "" {
				ss["stats_download_count"] = lineMeta["stats_download_count"]
			}

			counter_set, err := genMetaStatistics(ss)
			if err == nil {
				navFileList[a_text]["statistics"] = counter_set
			}

		}
		navFileKVS = append(navFileKVS, navFileList)
	}

	return navFileKVS
}

func genNavDirList(dirs []string, fkey, uname string) []map[string]map[string]string {
	navDirList := make(map[string]map[string]string)
	var navDirKVS []map[string]map[string]string
	sort.Strings(dirs)
	if fkey == "" {
		for _, line := range dirs {
			line = strings.TrimPrefix(line, "/")
			if line != "" && uname != "" {
				navDirList[line] = map[string]string{
					"uid": fmt.Sprintf("%s/%s", uname, line),
				}
			}
		}
		navDirKVS = append(navDirKVS, navDirList)
	} else {
		for _, line := range dirs {
			if line != "" && uname != "" {
				//DebugInfo("====:line=", line, " :fkey=", fkey)
				a_text := strings.TrimPrefix(line, fkey)
				navDirList[a_text] = map[string]string{
					"uid": fmt.Sprintf("%s/%s/%s", uname, fkey, line),
				}
			}
		}
		navDirKVS = append(navDirKVS, navDirList)
	}

	return navDirKVS
}
