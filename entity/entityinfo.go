package entity

import (
	"encoding/json"

	"github.com/harryzhu/potatofs/util"
)

type EntityMeta struct {
	Name    string   `json:"name"`
	Tags    []string `json:"tags"`
	Comment string   `json:"comment"`
	Size    int64    `json:"size"`
	Width   int      `json:"width"`
	Height  int      `json:"height"`
	Mime    string   `json:"mime"`
	Ref     string   `json:"ref"`
}

func NewEntityMeta() EntityMeta {
	return EntityMeta{}
}

func (em EntityMeta) WithName(name string) EntityMeta {
	em.Name = name
	return em
}

func (em EntityMeta) WithTags(tags string) EntityMeta {
	em.Tags = util.FormatTags(tags)
	return em
}

func (em EntityMeta) WithComment(comment string) EntityMeta {
	em.Comment = comment
	return em
}

func (em EntityMeta) WithSize(size int64) EntityMeta {
	em.Size = size
	return em
}

func (em EntityMeta) WithWidth(width int) EntityMeta {
	em.Width = width
	return em
}

func (em EntityMeta) WithHeight(height int) EntityMeta {
	em.Height = height
	return em
}

func (em EntityMeta) WithMime(m string) EntityMeta {
	em.Mime = m
	return em
}

func (em EntityMeta) WithRef(ref string) EntityMeta {
	em.Ref = ref
	return em
}

func (em EntityMeta) Marshal() ([]byte, error) {
	jsonEm, err := json.Marshal(em)
	if err != nil {
		return nil, err
	}
	return jsonEm, nil
}
