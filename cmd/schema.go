package cmd

type UploadEntitySchema struct {
	ID     string                 `json:"fid"`
	User   string                 `json:"fuser"`
	ApiKey string                 `json:"fapikey"`
	Meta   UploadEntityMetaSchema `json:"fmeta"`
	Data   []byte                 `json:"file"`
}

type UploadEntityMetaSchema struct {
	Size               string `json:"size"`
	Mime               string `json:"mime"`
	Mtime              string `json:"mtime"`
	IsPublic           string `json:"is_public"`
	IsBan              string `json:"is_ban"`
	Tags               string `json:"ftags"`
	URI                string `json:"uri"`
	StatsDiggCount     string `json:"stats_digg_count"`
	StatsCollectCount  string `json:"stats_collect_count"`
	StatsShareCount    string `json:"stats_share_count"`
	StatsCommentCount  string `json:"stats_comment_count"`
	StatsDownloadCount string `json:"stats_download_count"`
	DotColor           string `json:"dot_color"`
}
