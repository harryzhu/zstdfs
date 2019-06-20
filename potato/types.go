package potato

type Config struct {
	Welcome string
	Global  globalConfig `toml:"global"`
	Volume  volumeConfig `toml:"volume"`
}
type globalConfig struct {
	Log_level string
}

type volumeConfig struct {
	Ip           string
	Port         string
	Vmode        string
	Vmax         string
	Master       string
	Peers        string
	Db_data_dir  string
	Db_value_dir string
}
