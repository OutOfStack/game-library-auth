package config

import (
	"github.com/spf13/viper"
)

// Load reads config from provided file to specified config data structure
func Load(path, name, ext string, config interface{}) error {
	viper.AddConfigPath(path)
	viper.SetConfigName(name)
	viper.SetConfigType(ext)

	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		return err
	}

	return viper.Unmarshal(&config)
}
