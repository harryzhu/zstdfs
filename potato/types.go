package potato

type EntityObject struct {
	Name string
	Mime string
	Size string
	Data []byte
}

type EntityResponse struct {
	URL  string
	Name string
	Mime string
	Size string
}

type Config struct {
	Welcome string
	Global  globalConfig `toml:"global"`
	Volume  volumeConfig `toml:"volume"`
	Http    httpConfig   `toml:"http"`
}
type globalConfig struct {
	Log_level string
}

type volumeConfig struct {
	Ip           string
	Port         string
	Db_data_dir  string
	Db_value_dir string
}

type httpConfig struct {
	Ip       string
	Port     string
	Site_url string
	Temp_dir string
}
