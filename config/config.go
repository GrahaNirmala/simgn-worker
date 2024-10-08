package config

import (
	"os"

	"github.com/fsnotify/fsnotify"
	"github.com/Roofiif/sim-graha-nirmala-worker/logger"
	"github.com/spf13/viper"
)

func load() *Config {
	v := viper.New()

	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	configLocation, available := os.LookupEnv("CONFIG_LOCATION")
	if available {
		v.AddConfigPath(configLocation)
	}

	v.AutomaticEnv()
	err := v.ReadInConfig()
	if err != nil {
		logger.Log().Fatal("failed to read config")
	}

	c := new(Config)
	err = v.Unmarshal(c)
	if err != nil {
		logger.Log().Fatal("failed to unmarshal config")
	}

	v.WatchConfig()
	v.OnConfigChange(func(fsnotify.Event) {
		err := v.Unmarshal(c)
		if err != nil {
			logger.Log().Error("failed to reload config")
			return
		}
		logger.Log().Info("config reloaded")
	})

	return c
}

var config = load()

func Cfg() *Config { return config }
