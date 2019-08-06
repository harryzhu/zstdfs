package potato

import (
	"os"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	log "github.com/sirupsen/logrus"
)

type Config struct {
	Welcome string
	Global  globalConfig `toml:"global"`
	Volume  volumeConfig `toml:"volume"`
	Http    httpConfig   `toml:"http"`
}

type globalConfig struct {
	Is_debug  bool
	Log_level string
}

type volumeConfig struct {
	Ip            string
	Port          string
	Peers         []string
	Is_master     bool
	Db_data_dir   string
	Db_value_dir  string
	Meta_dir      string
	Max_size_mb   int
	Cache_size_mb int
}

type httpConfig struct {
	Ip                     string
	Port                   string
	Site_url               string
	Log_file               string
	Favicon_file           string
	Temp_upload_dir        string
	Cors_enabled           bool
	Cors_allow_credentials bool
	Cors_allow_origins     []string
	Cors_allow_methods     []string
	Cors_allow_headers     []string
	Cors_expose_headers    []string
	Cors_maxage_hours      time.Duration
}

func loadConfigFromFile() error {
	cfgFileName := "conf.toml"

	ENV_POTATOFS_CONF_SUFFIX := strings.Trim(os.Getenv("POTATOFS_CONF_SUFFIX"), " ")
	if len(ENV_POTATOFS_CONF_SUFFIX) > 0 {
		cfgFileNameOverwrite := strings.ToLower(strings.Join([]string{"conf", ENV_POTATOFS_CONF_SUFFIX, "toml"}, "."))

		_, err := os.Stat(cfgFileNameOverwrite)
		if err == nil {
			logger.Info(cfgFileNameOverwrite, " will overwrite ", cfgFileName)
			cfgFileName = cfgFileNameOverwrite
		} else {
			logger.Info(cfgFileNameOverwrite, " does not exist, will try to load ", cfgFileName)
		}

	}
	_, err := os.Stat(cfgFileName)
	if err != nil {
		logger.Fatal("cannot find the configuration file: conf.toml")
	}

	if _, err := toml.DecodeFile(cfgFileName, &cfg); err != nil {
		logger.Fatal(err)
	} else {
		logger.Info(cfgFileName, " was loaded.")
		logger.Info(cfg.Welcome)
		switch strings.ToUpper(cfg.Global.Log_level) {
		case "DEBUG":
			logger.SetLevel(log.DebugLevel)
		case "INFO":
			logger.SetLevel(log.InfoLevel)
		case "WARN":
			logger.SetLevel(log.WarnLevel)
		case "ERROR":
			logger.SetLevel(log.ErrorLevel)
		default:
			logger.SetLevel(log.ErrorLevel)
		}

		logger.SetLevel(log.DebugLevel)
	}
	return nil
}
