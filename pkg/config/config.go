package config

import "github.com/spf13/viper"

// LoadConfig reads config from provided file to specified config data structure
func LoadConfig(path, name, ext string, config interface{}) (err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName(name)
	viper.SetConfigType(ext)

	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	err = viper.Unmarshal(&config)
	return
}
