package cmd

import (
	//"bytes"
	"context"
	"encoding/json"

	"fmt"
	"strings"
	"time"

	"github.com/allegro/bigcache/v3"
)

var bcache *bigcache.BigCache

func bigcacheInit() {
	var err error
	config := bigcache.Config{
		Shards:             1024,
		LifeWindow:         10 * time.Minute,
		CleanWindow:        5 * time.Minute,
		MaxEntriesInWindow: 1000 * 10 * 60,
		MaxEntrySize:       512,
		Verbose:            true,
		HardMaxCacheSize:   MaxCacheSizeMB,
		OnRemove:           nil,
		OnRemoveWithReason: nil,
	}

	if IsDebug {
		config.LifeWindow = 30 * time.Second
		config.CleanWindow = 15 * time.Second
	}

	DebugInfo("bigcacheInit:LifeWindow", config.LifeWindow)
	DebugInfo("bigcacheInit:CleanWindow", config.CleanWindow)
	DebugInfo("bigcacheInit:HardMaxCacheSize", config.HardMaxCacheSize)

	bcache, err = bigcache.New(context.Background(), config)
	PrintError("bigcacheInit", err)
}

func bcacheSet(k string, v []byte) error {
	err := bcache.Set(k, v)
	return err
}

func bcacheGet(k string) []byte {
	v, err := bcache.Get(k)
	if err != nil {
		return nil
	}
	DebugInfo("bcacheGet from cache", k)
	return v
}

func bcacheDelete(k string) error {
	return bcache.Delete(k)
}

func jsonEnc(data any) []byte {
	b, err := json.Marshal(data)
	if err != nil {
		PrintError("jsonEnc", err)
		return nil
	}
	return b
}

func jsonDec(data []byte, dataStruct any) error {
	err := json.Unmarshal(data, &dataStruct)
	PrintError("jsonDec", err)
	return err
}

func bcacheKeyJoin(args ...any) string {
	var info []string
	for _, arg := range args {
		info = append(info, fmt.Sprintf("%v", arg))
	}
	return strings.Join(info, "::")
}

func bcacheScan(uname string) (data map[string]string) {
	if uname == "" {
		return data
	}
	data = make(map[string]string, 100)
	iterator := bcache.Iterator()
	count := 0
	val_safe := ""
	prefix := strings.Join([]string{uname, "::"}, "")
	emptyJSON := make(map[string]any)
	for iterator.SetNext() {
		if count > 2000 {
			break
		}
		current, err := iterator.Value()
		k := current.Key()
		if strings.HasPrefix(k, prefix) {
			m := emptyJSON
			err := jsonDec(current.Value(), &m)
			if err != nil {
				DebugWarn("bcacheScan", err.Error())
				continue
			}
			DebugInfo("======", m)
			if _, ok := m["_fsum"]; ok {
				delete(m, "_fsum")
			}
			if _, ok := m["fsha256"]; ok {
				delete(m, "fsha256")
			}
			val_safe = string(jsonEnc(m))
			if len(val_safe) > 1024 {
				val_safe = strings.Join([]string{val_safe[0:1024], "..."}, " ")
			}
			data[k] = val_safe
			count++
		}
		PrintError("bcacheScan", err)
	}
	return data
}
