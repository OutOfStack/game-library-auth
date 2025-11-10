package appconf

import (
	"errors"
	"sync"

	"github.com/spf13/viper"
)

const (
	appEnvFilePath = "."
	appEnvFileName = "app"
	appEnvFileExt  = "env"
)

var (
	cfg  *Cfg
	once sync.Once
)

// Get returns application config
func Get() (*Cfg, error) {
	var err error

	once.Do(func() {
		cfg, err = readFile()
	})

	if err != nil {
		return nil, err
	}

	if cfg == nil {
		return nil, errors.New("invalid config")
	}

	return cfg, nil
}

func readFile() (c *Cfg, err error) {
	viper.AddConfigPath(appEnvFilePath)
	viper.SetConfigName(appEnvFileName)
	viper.SetConfigType(appEnvFileExt)

	if err = viper.ReadInConfig(); err != nil {
		// if file not found use env variables only
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFoundError) {
			return nil, err
		}
	}

	viper.AutomaticEnv()

	if err = viper.Unmarshal(&c); err != nil {
		return nil, err
	}

	if err = c.Validate(); err != nil {
		return nil, err
	}

	return c, nil
}
