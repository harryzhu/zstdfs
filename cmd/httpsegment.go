package cmd

import (
	"fmt"
	"sort"
	"strings"
)

func genMetaStatistics(ss map[string]string) (map[string]string, error) {
	counterSet := make(map[string]string)

	diggCount, ok := ss["stats_digg_count"]
	if ok && Str2Int(diggCount) > minDiggCount {
		counterSet["点赞"] = ToKWM(Str2Int(diggCount))
	}
	commentCount, ok := ss["stats_comment_count"]
	if ok && Str2Int(commentCount) > minCommentCount {
		counterSet["评论"] = ToKWM(Str2Int(commentCount))
	}
	collectCount, ok := ss["stats_collect_count"]
	if ok && Str2Int(collectCount) > minCollectCount {
		counterSet["收藏"] = ToKWM(Str2Int(collectCount))
	}
	shareCount, ok := ss["stats_share_count"]
	if ok && Str2Int(shareCount) > minShareCount {
		counterSet["转发"] = ToKWM(Str2Int(shareCount))
	}
	downloadCount, ok := ss["stats_download_count"]
	if ok && Str2Int(downloadCount) > minDownloadCount {
		counterSet["下载"] = ToKWM(Str2Int(downloadCount))
	}
	return counterSet, nil
}

func genNavFileList(files []string, fkey, uname string) []map[string]map[string]any {
	var navFileKVS []map[string]map[string]any
	var aText string
	var lineMeta map[string]string

	for _, line := range files {
		navFileList := make(map[string]map[string]any)
		if line != "" && uname != "" {
			if fkey == "" {
				lineMeta = mongoGet(uname, line)
				aText = line

				navFileList[aText] = map[string]any{
					"uid":   fmt.Sprintf("%s/%s", uname, line),
					"uri":   fmt.Sprintf("%s/%s", uname, lineMeta["uri"]),
					"size":  ToKMGTB(Str2Int(lineMeta["size"])),
					"mtime": UnixFormat(Int2Int64(Str2Int(lineMeta["mtime"])), "06-01-02 15:04"),
				}
			} else {
				lineMeta = mongoGet(uname, strings.Join([]string{fkey, line}, "/"))
				aText = line
				navFileList[aText] = map[string]any{
					"uid":   fmt.Sprintf("%s/%s/%s", uname, fkey, line),
					"uri":   fmt.Sprintf("%s/%s", uname, lineMeta["uri"]),
					"size":  ToKMGTB(Str2Int(lineMeta["size"])),
					"mtime": UnixFormat(Int2Int64(Str2Int(lineMeta["mtime"])), "06-01-02 15:04"),
				}
			}

			navFileList[aText]["site_url"] = GetSiteURL()

			if lineMeta["dot_color"] != "" {
				navFileList[aText]["dot_color"] = lineMeta["dot_color"]
				navFileList[aText]["is_having_dot_color"] = "1"
			}

			if strings.Index(lineMeta["mime"], "video") > -1 || strings.Index(lineMeta["mime"], "mpeg") > -1 {
				navFileList[aText]["is_video"] = "1"
			}
			if lineMeta["tags"] != "" {
				metaTags := strings.Split(lineMeta["tags"], ",")
				navFileList[aText]["tags"] = metaTags
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

			counterSet, err := genMetaStatistics(ss)
			if err == nil {
				navFileList[aText]["statistics"] = counterSet
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
				aText := strings.TrimPrefix(line, fkey)
				navDirList[aText] = map[string]string{
					"uid": fmt.Sprintf("%s/%s/%s", uname, fkey, line),
				}
			}
		}
		navDirKVS = append(navDirKVS, navDirList)
	}

	return navDirKVS
}
