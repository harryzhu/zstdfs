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
	Welcome     string
	Global      globalConfig      `toml:"global"`
	Volume      volumeConfig      `toml:"volume"`
	Http        httpConfig        `toml:"http"`
	Replication replicationConfig `toml:"replication"`
}
type globalConfig struct {
	Is_debug  bool
	Log_level string
}

type volumeConfig struct {
	Ip             string
	Port           string
	Db_data_dir    string
	Db_value_dir   string
	Meta_dir       string
	Cache_self     string
	Cache_peers    []string
	Cache_basepath string
	Max_size       int
	Max_cache_size int
}

type httpConfig struct {
	Ip       string
	Port     string
	Site_url string
	Temp_dir string
}

type replicationConfig struct {
	Is_master bool
	Slaves    []string
}
