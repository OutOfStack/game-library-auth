package appconf

import "time"

// ServiceName - service name
const ServiceName = "game-library-auth"

// Cfg - app configuration
type Cfg struct {
	DB          DB          `mapstructure:",squash"`
	Web         Web         `mapstructure:",squash"`
	Auth        Auth        `mapstructure:",squash"`
	Zipkin      Zipkin      `mapstructure:",squash"`
	Graylog     Graylog     `mapstructure:",squash"`
	Log         Log         `mapstructure:",squash"`
	EmailSender EmailSender `mapstructure:",squash"`
}

// DB represents settings related to database
type DB struct {
	DSN string `mapstructure:"DB_DSN"`
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
	GoogleClientID   string `mapstructure:"AUTH_GOOGLECLIENTID"`
}

// Zipkin represents settings related to zipkin trace storage
type Zipkin struct {
	ReporterURL string `mapstructure:"ZIPKIN_REPORTERURL"`
}

// Graylog represents settings related to Graylog integration
type Graylog struct {
	Address string `mapstructure:"GRAYLOG_ADDR"`
}

// Log represents settings for logging
type Log struct {
	Level string `mapstructure:"LOG_LEVEL"`
}

// EmailSender represents settings for email sending service
type EmailSender struct {
	APIToken          string        `mapstructure:"EMAIL_SENDER_API_TOKEN"`
	APITimeout        time.Duration `mapstructure:"EMAIL_SENDER_API_TIMEOUT"`
	EmailFrom         string        `mapstructure:"EMAIL_SENDER_EMAIL_FROM"`
	ContactEmail      string        `mapstructure:"EMAIL_SENDER_CONTACT_EMAIL"`
	BaseURL           string        `mapstructure:"EMAIL_SENDER_BASE_URL"`
	UnsubscribeURL    string        `mapstructure:"EMAIL_SENDER_UNSUBSCRIBE_URL"`
	UnsubscribeSecret string        `mapstructure:"EMAIL_SENDER_UNSUBSCRIBE_SECRET"`
}
