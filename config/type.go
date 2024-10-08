package config

type Config struct {
	Db DbConfig `mapstructure:"db"`
	Fr Firebase `mapstructure:"firebase"`
}

type DbConfig struct {
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Database string `mapstructure:"database"`
	Url		 string `mapstructure:"url"`
}

type Firebase struct {
	Credentials     string `mapstructure:"credentials"`
}
