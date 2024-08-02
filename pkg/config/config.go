package config

import (
	"github.com/OutOfStack/game-library-auth/internal/appconf"
	"github.com/spf13/viper"
)

const (
	configPath = "./app.env"
)

// Init reads config from provided file to specified config data structure
func Init() (appconf.Cfg, error) {
	var cfg appconf.Cfg
	viper.SetConfigFile(configPath)

	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		return appconf.Cfg{}, err
	}

	err = viper.Unmarshal(&cfg)
	if err != nil {
		return appconf.Cfg{}, err
	}

	return cfg, nil
}
