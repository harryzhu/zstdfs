package cmd

type Config struct {
	Welcome string
	Global  globalConfig `toml:"global"`
	Master  masterConfig `toml:"master"`
	Volume  volumeConfig `toml:"volume"`
	Http    httpConfig   `toml:"http"`
}

type globalConfig struct {
	Mq_address string
}

type masterConfig struct {
	Ip       string
	Port     string
	Dir_meta string
}

type volumeConfig struct {
	Dc               string
	Rack             string
	Ip               string
	Port             string
	Vdir             string
	Vgroup           string
	Vmode            string
	Vmax             string
	Master           string
	Dir_data_default string
	Dir_meta         string
	Dirs_data        []string
}

type httpConfig struct {
	Ip             string
	Port           string
	Beip           string
	Beport         string
	Dir_http       string
	Expire_seconds string
}
