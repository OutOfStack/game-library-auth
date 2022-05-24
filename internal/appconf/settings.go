package appconf

import "time"

// DB represents settings related to database
type DB struct {
	Host       string `mapstructure:"DB_HOST"`
	Name       string `mapstructure:"DB_NAME"`
	User       string `mapstructure:"DB_USER"`
	Password   string `mapstructure:"DB_PASSWORD"`
	RequireSSL bool   `mapstructure:"DB_REQUIRESSL"`
}

// Web represents settings related to web server
type Web struct {
	Address           string        `mapstructure:"APP_ADDRESS"`
	DebugAddress      string        `mapstructure:"DEBUG_ADDRESS"`
	ReadTimeout       time.Duration `mapstructure:"APP_READTIMEOUT"`
	WriteTimeout      time.Duration `mapstructure:"APP_WRITETIMEOUT"`
	AllowedCORSOrigin string        `mapstructure:"APP_ALLOWEDCORSORIGIN"`
}

// Auth represents settings related to authentication and authorization
type Auth struct {
	PrivateKeyFile   string `mapstructure:"AUTH_PRIVATEKEYFILE"`
	SigningAlgorithm string `mapstructure:"AUTH_SIGNINGALG"`
	Issuer           string `mapstructure:"AUTH_ISSUER"`
}
