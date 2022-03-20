package cmd

import (
	"log"
	"sync"

	"github.com/BurntSushi/toml"
	"github.com/dgraph-io/badger/v3"
)

const (
	MaxMessageSize = 1024 << 20
)

var (
	wg sync.WaitGroup
)

var (
	Config *Configuration
	MetaDB *badger.DB
	DataDB *badger.DB
)

type Configuration struct {
	Welcome      string `toml:"welcome"`
	Version      string `toml:"version"`
	Debug        bool   `toml:"debug"`
	IP           string `toml:"ip"`
	SiteURL      string `toml:"site_url"`
	UploadDir    string `toml:"upload_dir"`
	VolumePort   string `toml:"volume_port"`
	MaxSizeMB    int    `toml:"max_size_mb"`
	HttpPort     string `toml:"http_port"`
	MetaDir      string `toml:"meta_dir"`
	DataDir      string `toml:"data_dir"`
	LogFile      string `toml:"log_file"`
	LogLevel     string `toml:"log_level"`
	IsMaster     bool   `toml:"is_master"`
	SyncSlave    string `toml:"sync_slave"`
	SyncPageSize int    `toml:"sync_pagesize"`
}

func LoadConfigurationFile() *Configuration {
	if _, err := toml.DecodeFile("conf.toml", &Config); err != nil {
		log.Fatal("cannot load the config file: conf.toml", err)
	}

	return Config
}
