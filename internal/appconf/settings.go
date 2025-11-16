package appconf

import (
	"errors"
	"strings"
	"time"
)

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
	Address               string        `mapstructure:"APP_ADDRESS"`
	DebugAddress          string        `mapstructure:"DEBUG_ADDRESS"`
	ReadTimeout           time.Duration `mapstructure:"APP_READTIMEOUT"`
	WriteTimeout          time.Duration `mapstructure:"APP_WRITETIMEOUT"`
	AllowedCORSOrigin     string        `mapstructure:"APP_ALLOWEDCORSORIGIN"`
	RefreshCookieSameSite string        `mapstructure:"APP_REFRESH_TOKEN_COOKIE_SAMESITE"`
	RefreshCookieSecure   bool          `mapstructure:"APP_REFRESH_TOKEN_COOKIE_SECURE"`
}

// Auth represents settings related to authentication and authorization
type Auth struct {
	PrivateKeyFile   string        `mapstructure:"AUTH_PRIVATEKEYFILE"`
	SigningAlgorithm string        `mapstructure:"AUTH_SIGNINGALG"`
	Issuer           string        `mapstructure:"AUTH_ISSUER"`
	GoogleClientID   string        `mapstructure:"AUTH_GOOGLECLIENTID"`
	AccessTokenTTL   time.Duration `mapstructure:"AUTH_ACCESSTOKENTTL"`
	RefreshTokenTTL  time.Duration `mapstructure:"AUTH_REFRESHTOKENTTL"`
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

// Validate validates configuration
func (cfg *Cfg) Validate() error {
	if cfg == nil {
		return errors.New("cfg is nil")
	}

	// DB validation
	if cfg.DB.DSN == "" {
		return errors.New("DB_DSN is required")
	}

	// Web validation
	if cfg.Web.Address == "" {
		return errors.New("APP_ADDRESS is required")
	}
	if cfg.Web.DebugAddress == "" {
		return errors.New("DEBUG_ADDRESS is required")
	}
	if cfg.Web.ReadTimeout <= 0 {
		return errors.New("APP_READTIMEOUT must be greater than 0")
	}
	if cfg.Web.WriteTimeout <= 0 {
		return errors.New("APP_WRITETIMEOUT must be greater than 0")
	}
	if cfg.Web.AllowedCORSOrigin == "" {
		return errors.New("APP_ALLOWEDCORSORIGIN is required")
	}

	switch strings.ToLower(cfg.Web.RefreshCookieSameSite) {
	case "lax", "strict", "none":
	default:
		return errors.New("APP_REFRESH_TOKEN_COOKIE_SAMESITE must be one of lax, strict, none")
	}

	// Auth validation
	if cfg.Auth.PrivateKeyFile == "" {
		return errors.New("AUTH_PRIVATEKEYFILE is required")
	}
	if cfg.Auth.SigningAlgorithm == "" {
		return errors.New("AUTH_SIGNINGALG is required")
	}
	if cfg.Auth.Issuer == "" {
		return errors.New("AUTH_ISSUER is required")
	}
	if cfg.Auth.GoogleClientID == "" {
		return errors.New("AUTH_GOOGLECLIENTID is required")
	}
	if cfg.Auth.AccessTokenTTL <= 0 {
		return errors.New("AUTH_ACCESSTOKENTTL must be greater than 0")
	}
	if cfg.Auth.RefreshTokenTTL <= 0 {
		return errors.New("AUTH_REFRESHTOKENTTL must be greater than 0")
	}

	// Zipkin validation
	if cfg.Zipkin.ReporterURL == "" {
		return errors.New("ZIPKIN_REPORTERURL is required")
	}

	// Graylog validation
	if cfg.Graylog.Address == "" {
		return errors.New("GRAYLOG_ADDR is required")
	}

	// Log validation
	if cfg.Log.Level == "" {
		return errors.New("LOG_LEVEL is required")
	}

	// EmailSender validation
	if cfg.EmailSender.APIToken == "" {
		return errors.New("EMAIL_SENDER_API_TOKEN is required")
	}
	if cfg.EmailSender.APITimeout <= 0 {
		return errors.New("EMAIL_SENDER_API_TIMEOUT must be greater than 0")
	}
	if cfg.EmailSender.EmailFrom == "" {
		return errors.New("EMAIL_SENDER_EMAIL_FROM is required")
	}
	if cfg.EmailSender.ContactEmail == "" {
		return errors.New("EMAIL_SENDER_CONTACT_EMAIL is required")
	}
	if cfg.EmailSender.BaseURL == "" {
		return errors.New("EMAIL_SENDER_BASE_URL is required")
	}
	if cfg.EmailSender.UnsubscribeURL == "" {
		return errors.New("EMAIL_SENDER_UNSUBSCRIBE_URL is required")
	}
	if cfg.EmailSender.UnsubscribeSecret == "" {
		return errors.New("EMAIL_SENDER_UNSUBSCRIBE_SECRET is required")
	}

	return nil
}
