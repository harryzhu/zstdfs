package cmd

import (
	"fmt"
	"io"
	"mime"
	"os"
	"path/filepath"
	"strings"
)

type Entity struct {
	ID   string
	User string
	Meta map[string]string
	Data []byte
}

func metaDefault() map[string]string {
	meta := make(map[string]string)
	meta["mime"] = ""
	meta["size"] = ""
	meta["_fsum"] = ""
	meta["fsha256"] = ""
	meta["mtime"] = ""
	meta["tags"] = ""
	meta["is_public"] = "1"
	meta["is_ban"] = "0"
	meta["dot_color"] = ""
	meta["stats_digg_count"] = "0"
	meta["stats_comment_count"] = "0"
	meta["stats_collect_count"] = "0"
	meta["stats_share_count"] = "0"
	meta["stats_download_count"] = "0"
	meta["caption.en"] = ""
	meta["caption.cn"] = ""

	return meta
}

func NewEntity(user, id string) Entity {
	if IsAnyEmpty(user, id) {
		return Entity{}
	}

	meta := metaDefault()

	ett := Entity{
		ID:   id,
		User: user,
		Meta: meta,
		Data: nil,
	}
	return ett
}

func (ett Entity) WithID(n string) Entity {
	if n != "" {
		ett.ID = n
	}
	return ett
}

func (ett Entity) WithData(b []byte) Entity {
	ett.Data = b
	return ett
}

func (ett Entity) WithFile(fpath string) Entity {
	if IsAnyEmpty(fpath) {
		return Entity{}
	}

	finfo, err := os.Stat(fpath)
	if err != nil {
		PrintError("NewEntity", err)
		return Entity{}
	}

	fp, err := os.Open(fpath)
	if err != nil {
		PrintError("NewEntity", err)
		return Entity{}
	}

	data, err := io.ReadAll(fp)
	if err != nil {
		PrintError("NewEntity", err)
		return Entity{}
	}
	fp.Close()

	mimeType := "text/plain"
	if filepath.Ext(fpath) != "" {
		mimeType = mime.TypeByExtension(filepath.Ext(fpath))
	}

	ett.Meta["mtime"] = Int64ToString(finfo.ModTime().UTC().Unix())
	ett.Meta["size"] = Int64ToString(finfo.Size())
	ett.Meta["mime"] = mimeType
	//
	ett.Data = data

	if data == nil {
		ett.Data = []byte("")
	}

	return ett
}

func (ett Entity) WithMeta(k, v string) Entity {
	if k != "" {
		ett.Meta[k] = fmt.Sprintf("%v", v)
	}
	return ett
}

func (ett Entity) WithTags(tagline string) Entity {
	if tagline != "" {
		tagline = TagLineFormat(tagline)
		//
		taglist := strings.Split(tagline, ",")
		var tags []string
		for _, v := range taglist {
			if Contains(tags, v) == false {
				v = strings.TrimSpace(v)
				tags = append(tags, v)
			}
		}
		ett.Meta["tags"] = strings.Join(tags, ",")
	}

	return ett
}

func (ett Entity) Save() bool {
	if IsAnyEmpty(ett.ID, ett.User) || IsAnyNil(ett.Data) {
		return false
	}

	if len(ett.Data) != Str2Int(ett.Meta["size"]) {
		DebugWarn("ERROR: entity.Save:", "data length != meta[size], will SKIP save")
		return false
	}

	if SHA256Bytes(ett.Data) != ett.Meta["fsha256"] {
		DebugWarn("ERROR: entity.Save:", "data sha256 != meta[fsha256], will SKIP save")
		return false
	}

	bkey := badgerSave(ett.Data)
	DebugInfo("Entity Save", string(bkey))
	if bkey == nil {
		return false
	}
	ett.Meta["author"] = strings.ToLower(ett.User)
	ett.Meta["_fsum"] = string(bkey)
	ett.Meta["fsha256"] = SHA256Bytes(ett.Data)
	ett.Meta["uri"] = GetURI(ett.ID)

	for k, v := range ett.Meta {
		mongoSave(ett.User, ett.ID, k, v)
	}

	return true
}

func (ett Entity) SaveWithoutData() bool {
	if IsAnyEmpty(ett.ID, ett.User, ett.Meta["_fsum"]) {
		return false
	}

	if badgerExists([]byte(ett.Meta["_fsum"])) == false {
		DebugWarn("badgerExists", "_fsum does not exist, cannot SaveWithoutData", ", _fsum=", ett.Meta["_fsum"])
		return false
	}

	ett.Meta["author"] = strings.ToLower(ett.User)
	ett.Meta["uri"] = GetURI(ett.ID)
	dataInBadger := badgerGet([]byte(ett.Meta["_fsum"]))
	if dataInBadger != nil {
		ett.Meta["fsha256"] = SHA256Bytes(dataInBadger)
	}

	for k, v := range ett.Meta {
		mongoSave(ett.User, ett.ID, k, v)
	}

	return true
}

func (ett Entity) Get() Entity {
	if IsAnyEmpty(ett.ID, ett.User) {
		return Entity{}
	}

	meta := mongoGet(ett.User, ett.ID)

	_, ok := meta["_id"]
	if !ok {
		return Entity{}
	}

	ett.Meta = meta

	fsum, ok := ett.Meta["_fsum"]
	if ok && fsum != "" {
		ett.Data = badgerGet([]byte(fsum))
	} else {
		return Entity{}
	}
	return ett
}

func (ett Entity) Head() Entity {
	if IsAnyEmpty(ett.ID, ett.User) {
		return Entity{}
	}

	meta := mongoGet(ett.User, ett.ID)
	ett.Meta = meta

	return ett
}
