package config

type Config struct {
	BindPort        string                `mapstructure:"bindport"`
	Logs            Log                   `mapstructure:"logs"`
	ManagerWorker   int                   `mapstructure:"manager_worker"`
	Mysql           *MysqlSetting         `mapstructure:"mysql"`
	Kakfa           *KafkaSetting         `mapstructure:"kafka"`
	ElasticSearch   *ElasticSearchSetting `mapstructure:"elasticsearch"`
	ElasticSearchV7 *ElasticSearchSetting `mapstructure:"elasticsearchv7"`
}

type Agent struct {
	Sources []Source `mapstructure:"source"`
	Matches []Match  `mapstructure:"match"`
}

type Log struct {
	LogLevel string `mapstructure:"log_level"`
	Addr     string `mapstructure:"addr"`
}

type Source struct {
	SourceType    string      `mapstructure:"source_type"`
	Tag           string      `mapstructure:"tag"`
	CollectWorker int         `mapstructure:"collect_worker"`
	Setting       interface{} `mapstructure:"setting"`
}

type Match struct {
	MatchType  string      `mapstructure:"match_type"`
	Tag        string      `mapstructure:"tag"`
	SendWorker int         `mapstructure:"send_worker"`
	Setting    interface{} `mapstructure:"setting"`
}

type KafkaSetting struct {
	Brokers []string `mapstructure:"brokers"`
	Topic   []string `mapstructure:"topic"`
	Group   string   `mapstructure:"group"`
}

type MysqlSetting struct {
	Debug    string `mapstructure:"debug"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Address  string `mapstructure:"address"`
	Type     string `mapstructure:"type"`
	DBName   string `mapstructure:"dbname"`
}

type ElasticSearchSetting struct {
	Sniff   bool     `mapstructure:"sniff"`
	Address []string `mapstructure:"address"`
}
