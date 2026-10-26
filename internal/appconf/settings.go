package appconf

import (
	"errors"
	"net/url"
	"strings"
	"time"

	"github.com/OutOfStack/game-library-auth/internal/notification"
)

// ServiceName - service name
const ServiceName = "game-library-auth"

// API - api service configuration
type API struct {
	Log         Log         `mapstructure:",squash"`
	DB          DB          `mapstructure:",squash"`
	Web         Web         `mapstructure:",squash"`
	Debug       Debug       `mapstructure:",squash"`
	Auth        Auth        `mapstructure:",squash"`
	OAuth       OAuth       `mapstructure:",squash"`
	Jaeger      Jaeger      `mapstructure:",squash"`
	Graylog     Graylog     `mapstructure:",squash"`
	EmailSender EmailSender `mapstructure:",squash"`
	InfoAPI     InfoAPI     `mapstructure:",squash"`
	Outbox      Outbox      `mapstructure:",squash"`
}

// Worker - notification worker configuration. Deliberately omits JWT keys, OAuth and downstream API credentials
type Worker struct {
	Log           Log           `mapstructure:",squash"`
	DB            DB            `mapstructure:",squash"`
	Debug         Debug         `mapstructure:",squash"`
	Jaeger        Jaeger        `mapstructure:",squash"`
	Graylog       Graylog       `mapstructure:",squash"`
	Outbox        Outbox        `mapstructure:",squash"`
	EmailDelivery EmailDelivery `mapstructure:",squash"`
}

// Manage - database management tool configuration
type Manage struct {
	DB DB `mapstructure:",squash"`
}

// DB represents settings for database
type DB struct {
	DSN string `mapstructure:"DB_DSN"`
}

// Web represents settings for web server
type Web struct {
	HTTPAddress           string        `mapstructure:"APP_HTTP_ADDRESS"`
	GRPCAddress           string        `mapstructure:"APP_GRPC_ADDRESS"`
	ReadTimeout           time.Duration `mapstructure:"APP_READTIMEOUT"`
	WriteTimeout          time.Duration `mapstructure:"APP_WRITETIMEOUT"`
	AllowedCORSOrigin     string        `mapstructure:"APP_ALLOWEDCORSORIGIN"`
	RefreshCookieSameSite string        `mapstructure:"APP_REFRESH_TOKEN_COOKIE_SAMESITE"`
	RefreshCookieSecure   bool          `mapstructure:"APP_REFRESH_TOKEN_COOKIE_SECURE"`
}

// Debug represents settings for debug server (pprof, metrics, health probes)
type Debug struct {
	Address string `mapstructure:"APP_DEBUG_ADDRESS"`
}

// Auth represents settings for authentication and authorization
type Auth struct {
	PrivateKeyFile   string        `mapstructure:"AUTH_PRIVATEKEYFILE"`
	SigningAlgorithm string        `mapstructure:"AUTH_SIGNINGALG"`
	Issuer           string        `mapstructure:"AUTH_ISSUER"`
	AccessTokenTTL   time.Duration `mapstructure:"AUTH_ACCESSTOKENTTL"`
	RefreshTokenTTL  time.Duration `mapstructure:"AUTH_REFRESHTOKENTTL"`
}

// OAuth represents settings for OAuth providers
type OAuth struct {
	Timeout            time.Duration `mapstructure:"OAUTH_TIMEOUT"`
	GoogleClientID     string        `mapstructure:"OAUTH_GOOGLECLIENTID"`
	GitHubClientID     string        `mapstructure:"OAUTH_GITHUBCLIENTID"`
	GitHubClientSecret string        `mapstructure:"OAUTH_GITHUBCLIENTSECRET"`
}

// Jaeger represents settings for Jaeger OTLP trace export
type Jaeger struct {
	OTLPEndpoint string `mapstructure:"JAEGER_OTLP_ENDPOINT"`
}

// Graylog represents settings related to Graylog integration
type Graylog struct {
	Address string `mapstructure:"GRAYLOG_ADDR"`
}

// Log represents settings for logging
type Log struct {
	Level string `mapstructure:"LOG_LEVEL"`
}

// EmailSender represents settings for composing emails
type EmailSender struct {
	EmailFrom         string `mapstructure:"EMAIL_SENDER_EMAIL_FROM"`
	ContactEmail      string `mapstructure:"EMAIL_SENDER_CONTACT_EMAIL"`
	BaseURL           string `mapstructure:"EMAIL_SENDER_BASE_URL"`
	UnsubscribeURL    string `mapstructure:"EMAIL_SENDER_UNSUBSCRIBE_URL"`
	UnsubscribeSecret string `mapstructure:"EMAIL_SENDER_UNSUBSCRIBE_SECRET"`
}

// EmailDelivery represents settings for email provider client
type EmailDelivery struct {
	APIToken   string        `mapstructure:"EMAIL_SENDER_API_TOKEN"`
	APITimeout time.Duration `mapstructure:"EMAIL_SENDER_API_TIMEOUT"`
}

// InfoAPI represents settings for InfoAPI gRPC client
type InfoAPI struct {
	Address string        `mapstructure:"INFOAPI_ADDRESS"`
	Timeout time.Duration `mapstructure:"INFOAPI_TIMEOUT"`
}

// Outbox represents settings for notification outbox shared by api and worker
type Outbox struct {
	Key string `mapstructure:"NOTIFICATION_OUTBOX_KEY"`
}

// Validate validates db settings
func (c DB) Validate() error {
	if c.DSN == "" {
		return errors.New("DB_DSN is required")
	}
	return nil
}

// Validate validates web server settings
func (c Web) Validate() error {
	if c.HTTPAddress == "" {
		return errors.New("APP_HTTP_ADDRESS is required")
	}
	if c.GRPCAddress == "" {
		return errors.New("APP_GRPC_ADDRESS is required")
	}
	if c.ReadTimeout <= 0 {
		return errors.New("APP_READTIMEOUT must be greater than 0")
	}
	if c.WriteTimeout <= 0 {
		return errors.New("APP_WRITETIMEOUT must be greater than 0")
	}
	if c.AllowedCORSOrigin == "" {
		return errors.New("APP_ALLOWEDCORSORIGIN is required")
	}

	switch strings.ToLower(c.RefreshCookieSameSite) {
	case "lax", "strict", "none":
	default:
		return errors.New("APP_REFRESH_TOKEN_COOKIE_SAMESITE must be one of lax, strict, none")
	}
	return nil
}

// Validate validates debug server settings
func (c Debug) Validate() error {
	if c.Address == "" {
		return errors.New("APP_DEBUG_ADDRESS is required")
	}
	return nil
}

// Validate validates auth settings
func (c Auth) Validate() error {
	if c.PrivateKeyFile == "" {
		return errors.New("AUTH_PRIVATEKEYFILE is required")
	}
	if c.SigningAlgorithm == "" {
		return errors.New("AUTH_SIGNINGALG is required")
	}
	if c.Issuer == "" {
		return errors.New("AUTH_ISSUER is required")
	}
	if c.AccessTokenTTL <= 0 {
		return errors.New("AUTH_ACCESSTOKENTTL must be greater than 0")
	}
	if c.RefreshTokenTTL <= 0 {
		return errors.New("AUTH_REFRESHTOKENTTL must be greater than 0")
	}
	return nil
}

// Validate validates OAuth settings
func (c OAuth) Validate() error {
	if c.GoogleClientID == "" {
		return errors.New("OAUTH_GOOGLECLIENTID is required")
	}
	if c.GitHubClientID == "" {
		return errors.New("OAUTH_GITHUBCLIENTID is required")
	}
	if c.GitHubClientSecret == "" {
		return errors.New("OAUTH_GITHUBCLIENTSECRET is required")
	}
	if c.Timeout <= 0 {
		return errors.New("OAUTH_TIMEOUT must be greater than 0")
	}
	return nil
}

// Validate validates Jaeger settings
func (c Jaeger) Validate() error {
	if c.OTLPEndpoint == "" {
		return errors.New("JAEGER_OTLP_ENDPOINT is required")
	}
	if strings.Contains(c.OTLPEndpoint, "://") {
		return errors.New("JAEGER_OTLP_ENDPOINT must be host:port without scheme")
	}
	u, err := url.Parse("http://" + c.OTLPEndpoint)
	if err != nil || u.Host != c.OTLPEndpoint {
		return errors.New("JAEGER_OTLP_ENDPOINT must be host:port without scheme")
	}
	return nil
}

// Validate validates Graylog settings
func (c Graylog) Validate() error {
	if c.Address == "" {
		return errors.New("GRAYLOG_ADDR is required")
	}
	return nil
}

// Validate validates log settings
func (c Log) Validate() error {
	if c.Level == "" {
		return errors.New("LOG_LEVEL is required")
	}
	return nil
}

// Validate validates email sender settings
func (c EmailSender) Validate() error {
	if c.EmailFrom == "" {
		return errors.New("EMAIL_SENDER_EMAIL_FROM is required")
	}
	if c.ContactEmail == "" {
		return errors.New("EMAIL_SENDER_CONTACT_EMAIL is required")
	}
	if c.BaseURL == "" {
		return errors.New("EMAIL_SENDER_BASE_URL is required")
	}
	if c.UnsubscribeURL == "" {
		return errors.New("EMAIL_SENDER_UNSUBSCRIBE_URL is required")
	}
	if c.UnsubscribeSecret == "" {
		return errors.New("EMAIL_SENDER_UNSUBSCRIBE_SECRET is required")
	}
	return nil
}

// Validate validates email delivery settings. Timeout must not outlive a worker attempt
func (c EmailDelivery) Validate() error {
	if c.APIToken == "" {
		return errors.New("EMAIL_SENDER_API_TOKEN is required")
	}
	if c.APITimeout <= 0 || c.APITimeout >= 30*time.Second {
		return errors.New("EMAIL_SENDER_API_TIMEOUT must be greater than 0 and less than 30s")
	}
	return nil
}

// Validate validates InfoAPI settings
func (c InfoAPI) Validate() error {
	if c.Address == "" {
		return errors.New("INFOAPI_ADDRESS is required")
	}
	if c.Timeout <= 0 {
		return errors.New("INFOAPI_TIMEOUT must be greater than 0")
	}
	return nil
}

// Validate validates outbox encryption key
func (c Outbox) Validate() error {
	_, err := notification.NewCipher(c.Key)
	return err
}
