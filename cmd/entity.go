package cmd

import (
	//"encoding/json"
	"fmt"
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

func NewEntity(user, id string) Entity {
	if IsAnyEmpty(user, id) {
		return Entity{}
	}

	meta := make(map[string]string)
	meta["mime"] = ""
	meta["size"] = ""
	meta["fsum"] = ""
	meta["mtime"] = ""
	meta["sha256"] = ""
	meta["tags"] = ""
	meta["is_public"] = "1"
	meta["is_ban"] = "0"
	meta["dot_color"] = ""
	meta["author"] = strings.ToLower(user)
	meta["stats_digg_count"] = "0"
	meta["stats_comment_count"] = "0"
	meta["stats_collect_count"] = "0"
	meta["stats_share_count"] = "0"
	meta["stats_download_count"] = "0"

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
	ett.Meta["sha256"] = SHA256Bytes(b)
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

	data, err := os.ReadFile(fpath)
	if err != nil {
		PrintError("NewEntity", err)
		return Entity{}
	}

	mimeType := "text/plain"
	if filepath.Ext(fpath) != "" {
		mimeType = mime.TypeByExtension(filepath.Ext(fpath))
	}

	ett.Meta["mtime"] = Int64ToString(finfo.ModTime().UTC().Unix())
	ett.Meta["size"] = Int64ToString(finfo.Size())
	ett.Meta["mime"] = mimeType
	ett.Meta["sha256"] = SHA256File(fpath)
	//
	ett.Data = data

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

	bkey := badgerSave(ett.Data)
	DebugInfo("Entity Save", string(bkey))
	if bkey == nil {
		return false
	}
	ett.Meta["author"] = strings.ToLower(ett.User)
	ett.Meta["fsum"] = string(bkey)

	for k, v := range ett.Meta {
		mongoSave(ett.User, ett.ID, k, v)
	}

	return true
}

func (ett Entity) SaveWithoutData() bool {
	if IsAnyEmpty(ett.ID, ett.User, ett.Meta["fsum"]) {
		return false
	}

	if badgerExists([]byte(ett.Meta["fsum"])) == false {
		DebugWarn("SaveWithoutData:badgerExists", "fsum does not exist, cannot SaveWithoutData", ", fsum=", ett.Meta["fsum"])
		//return false
	}

	ett.Meta["author"] = strings.ToLower(ett.User)

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
	ett.Meta = meta

	fsum := meta["fsum"]
	if fsum != "" {
		//DebugInfo("Entity::Get:fsum", fsum)
		ett.Data = badgerGet([]byte(fsum))
	} else {
		DebugInfo("Entity::Get:meta", meta)
	}
	return ett
}

func (ett Entity) GetMeta() Entity {
	if IsAnyEmpty(ett.ID, ett.User) {
		return Entity{}
	}

	meta := mongoGet(ett.User, ett.ID)
	ett.Meta = meta

	return ett
}
