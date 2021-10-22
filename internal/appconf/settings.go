package appconf

import "time"

type DB struct {
	Host       string `mapstructure:"DB_HOST"`
	Name       string `mapstructure:"DB_NAME"`
	User       string `mapstructure:"DB_USER"`
	Password   string `mapstructure:"DB_PASSWORD"`
	RequireSSL bool   `mapstructure:"DB_REQUIRESSL"`
}

type Web struct {
	Address           string        `mapstructure:"APP_ADDRESS"`
	DebugAddress      string        `mapstructure:"DEBUG_ADDRESS"`
	ReadTimeout       time.Duration `mapstructure:"APP_READTIMEOUT"`
	WriteTimeout      time.Duration `mapstructure:"APP_WRITETIMEOUT"`
	AllowedCORSOrigin string        `mapstructure:"APP_ALLOWEDCORSORIGIN"`
}

type Auth struct {
	PrivateKeyFile   string `mapstructure:"AUTH_PRIVATEKEYFILE"`
	SigningAlgorithm string `mapstructure:"AUTH_SIGNINGALG"`
	Issuer           string `mapstructure:"AUTH_ISSUER"`
}
