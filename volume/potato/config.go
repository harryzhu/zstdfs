package potato

import (
	"os"
	"strings"

	//"time"

	"github.com/BurntSushi/toml"
	log "github.com/sirupsen/logrus"
)

type Config struct {
	Welcome string
	Global  globalConfig `toml:"global"`
	Volume  volumeConfig `toml:"volume"`
}

type globalConfig struct {
	Is_debug     bool
	Log_level    string
	Profile_path string
}

type volumeConfig struct {
	Self          string
	Peers         []string
	Data_dir      string
	Meta_dir      string
	Is_syncwrites bool
	Max_size_mb   int
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

	_, err = os.Stat(cfg.Global.Profile_path)
	if err != nil {
		logger.Fatal("the Profile_path(cfg.Global.Profile_path) does not exist: ", cfg.Global.Profile_path)
	} else {
		profilePath = cfg.Global.Profile_path
		logger.Info("Profiling path: ", profilePath)
	}

	return nil
}
