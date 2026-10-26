package appconf_test

import (
	"encoding/base64"
	"testing"
	"time"

	"github.com/OutOfStack/game-library-auth/internal/appconf"
	"github.com/stretchr/testify/require"
)

func TestCfgValidateValidConfig(t *testing.T) {
	for _, cfg := range []any{api(nil), worker(nil), &appconf.Manage{DB: validTestAPI().DB}} {
		require.NoError(t, appconf.Validate(cfg))
	}
}

func TestCfgValidateValidSpecialCases(t *testing.T) {
	tests := []struct {
		name string
		cfg  any
	}{
		{
			name: "case insensitive cookie samesite",
			cfg: api(func(cfg *appconf.API) {
				cfg.Web.RefreshCookieSameSite = "StRiCt"
			}),
		},
		{
			name: "ipv6 jaeger otlp endpoint",
			cfg: api(func(cfg *appconf.API) {
				cfg.Jaeger.OTLPEndpoint = "[::1]:4318"
			}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NoError(t, appconf.Validate(tt.cfg))
		})
	}
}

func TestCfgValidateErrorCases(t *testing.T) {
	tests := []struct {
		name      string
		cfg       any
		wantError string
	}{
		{
			name: "missing db dsn",
			cfg: api(func(cfg *appconf.API) {
				cfg.DB.DSN = ""
			}),
			wantError: "DB_DSN is required",
		},
		{
			name: "missing app http address",
			cfg: api(func(cfg *appconf.API) {
				cfg.Web.HTTPAddress = ""
			}),
			wantError: "APP_HTTP_ADDRESS is required",
		},
		{
			name: "missing app grpc address",
			cfg: api(func(cfg *appconf.API) {
				cfg.Web.GRPCAddress = ""
			}),
			wantError: "APP_GRPC_ADDRESS is required",
		},
		{
			name: "invalid app read timeout",
			cfg: api(func(cfg *appconf.API) {
				cfg.Web.ReadTimeout = 0
			}),
			wantError: "APP_READTIMEOUT must be greater than 0",
		},
		{
			name: "invalid app write timeout",
			cfg: api(func(cfg *appconf.API) {
				cfg.Web.WriteTimeout = 0
			}),
			wantError: "APP_WRITETIMEOUT must be greater than 0",
		},
		{
			name: "missing cors origin",
			cfg: api(func(cfg *appconf.API) {
				cfg.Web.AllowedCORSOrigin = ""
			}),
			wantError: "APP_ALLOWEDCORSORIGIN is required",
		},
		{
			name: "invalid cookie samesite",
			cfg: api(func(cfg *appconf.API) {
				cfg.Web.RefreshCookieSameSite = "invalid"
			}),
			wantError: "APP_REFRESH_TOKEN_COOKIE_SAMESITE must be one of lax, strict, none",
		},
		{
			name: "missing app debug address",
			cfg: api(func(cfg *appconf.API) {
				cfg.Debug.Address = ""
			}),
			wantError: "APP_DEBUG_ADDRESS is required",
		},
		{
			name: "missing worker debug address",
			cfg: worker(func(cfg *appconf.Worker) {
				cfg.Debug.Address = ""
			}),
			wantError: "APP_DEBUG_ADDRESS is required",
		},
		{
			name: "missing auth private key file",
			cfg: api(func(cfg *appconf.API) {
				cfg.Auth.PrivateKeyFile = ""
			}),
			wantError: "AUTH_PRIVATEKEYFILE is required",
		},
		{
			name: "missing auth signing algorithm",
			cfg: api(func(cfg *appconf.API) {
				cfg.Auth.SigningAlgorithm = ""
			}),
			wantError: "AUTH_SIGNINGALG is required",
		},
		{
			name: "missing auth issuer",
			cfg: api(func(cfg *appconf.API) {
				cfg.Auth.Issuer = ""
			}),
			wantError: "AUTH_ISSUER is required",
		},
		{
			name: "invalid auth access token ttl",
			cfg: api(func(cfg *appconf.API) {
				cfg.Auth.AccessTokenTTL = 0
			}),
			wantError: "AUTH_ACCESSTOKENTTL must be greater than 0",
		},
		{
			name: "invalid auth refresh token ttl",
			cfg: api(func(cfg *appconf.API) {
				cfg.Auth.RefreshTokenTTL = 0
			}),
			wantError: "AUTH_REFRESHTOKENTTL must be greater than 0",
		},
		{
			name: "missing oauth google client id",
			cfg: api(func(cfg *appconf.API) {
				cfg.OAuth.GoogleClientID = ""
			}),
			wantError: "OAUTH_GOOGLECLIENTID is required",
		},
		{
			name: "missing oauth github client id",
			cfg: api(func(cfg *appconf.API) {
				cfg.OAuth.GitHubClientID = ""
			}),
			wantError: "OAUTH_GITHUBCLIENTID is required",
		},
		{
			name: "missing oauth github client secret",
			cfg: api(func(cfg *appconf.API) {
				cfg.OAuth.GitHubClientSecret = ""
			}),
			wantError: "OAUTH_GITHUBCLIENTSECRET is required",
		},
		{
			name: "invalid oauth timeout",
			cfg: api(func(cfg *appconf.API) {
				cfg.OAuth.Timeout = 0
			}),
			wantError: "OAUTH_TIMEOUT must be greater than 0",
		},
		{
			name: "missing jaeger otlp endpoint",
			cfg: api(func(cfg *appconf.API) {
				cfg.Jaeger.OTLPEndpoint = ""
			}),
			wantError: "JAEGER_OTLP_ENDPOINT is required",
		},
		{
			name: "jaeger otlp endpoint with scheme",
			cfg: api(func(cfg *appconf.API) {
				cfg.Jaeger.OTLPEndpoint = "http://jaeger:4318"
			}),
			wantError: "JAEGER_OTLP_ENDPOINT must be host:port without scheme",
		},
		{
			name: "jaeger otlp endpoint invalid format",
			cfg: api(func(cfg *appconf.API) {
				cfg.Jaeger.OTLPEndpoint = "not a valid endpoint"
			}),
			wantError: "JAEGER_OTLP_ENDPOINT must be host:port without scheme",
		},
		{
			name: "missing graylog addr",
			cfg: api(func(cfg *appconf.API) {
				cfg.Graylog.Address = ""
			}),
			wantError: "GRAYLOG_ADDR is required",
		},
		{
			name: "missing log level",
			cfg: api(func(cfg *appconf.API) {
				cfg.Log.Level = ""
			}),
			wantError: "LOG_LEVEL is required",
		},
		{
			name: "missing email sender from",
			cfg: api(func(cfg *appconf.API) {
				cfg.EmailSender.EmailFrom = ""
			}),
			wantError: "EMAIL_SENDER_EMAIL_FROM is required",
		},
		{
			name: "missing email sender contact email",
			cfg: api(func(cfg *appconf.API) {
				cfg.EmailSender.ContactEmail = ""
			}),
			wantError: "EMAIL_SENDER_CONTACT_EMAIL is required",
		},
		{
			name: "missing email sender base url",
			cfg: api(func(cfg *appconf.API) {
				cfg.EmailSender.BaseURL = ""
			}),
			wantError: "EMAIL_SENDER_BASE_URL is required",
		},
		{
			name: "missing email sender unsubscribe url",
			cfg: api(func(cfg *appconf.API) {
				cfg.EmailSender.UnsubscribeURL = ""
			}),
			wantError: "EMAIL_SENDER_UNSUBSCRIBE_URL is required",
		},
		{
			name: "missing email sender unsubscribe secret",
			cfg: api(func(cfg *appconf.API) {
				cfg.EmailSender.UnsubscribeSecret = ""
			}),
			wantError: "EMAIL_SENDER_UNSUBSCRIBE_SECRET is required",
		},
		{
			name: "missing email sender api token",
			cfg: worker(func(cfg *appconf.Worker) {
				cfg.EmailDelivery.APIToken = ""
			}),
			wantError: "EMAIL_SENDER_API_TOKEN is required",
		},
		{
			name: "zero email sender api timeout",
			cfg: worker(func(cfg *appconf.Worker) {
				cfg.EmailDelivery.APITimeout = 0
			}),
			wantError: "EMAIL_SENDER_API_TIMEOUT must be greater than 0 and less than 30s",
		},
		{
			name: "too long email sender api timeout",
			cfg: worker(func(cfg *appconf.Worker) {
				cfg.EmailDelivery.APITimeout = 30 * time.Second
			}),
			wantError: "EMAIL_SENDER_API_TIMEOUT must be greater than 0 and less than 30s",
		},
		{
			name: "missing infoapi address",
			cfg: api(func(cfg *appconf.API) {
				cfg.InfoAPI.Address = ""
			}),
			wantError: "INFOAPI_ADDRESS is required",
		},
		{
			name: "invalid infoapi timeout",
			cfg: api(func(cfg *appconf.API) {
				cfg.InfoAPI.Timeout = 0
			}),
			wantError: "INFOAPI_TIMEOUT must be greater than 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := appconf.Validate(tt.cfg)

			require.EqualError(t, err, tt.wantError)
		})
	}
}

func TestCfgValidateInvalidOutboxKey(t *testing.T) {
	cfg := worker(func(cfg *appconf.Worker) {
		cfg.Outbox.Key = "bad"
	})

	err := appconf.Validate(cfg)

	require.Error(t, err)
}

func api(mutate func(*appconf.API)) *appconf.API {
	cfg := validTestAPI()
	if mutate != nil {
		mutate(cfg)
	}
	return cfg
}

func worker(mutate func(*appconf.Worker)) *appconf.Worker {
	a := validTestAPI()
	cfg := &appconf.Worker{
		Log:     a.Log,
		DB:      a.DB,
		Debug:   a.Debug,
		Jaeger:  a.Jaeger,
		Graylog: a.Graylog,
		Outbox:  a.Outbox,
		EmailDelivery: appconf.EmailDelivery{
			APIToken:   "token",
			APITimeout: 5 * time.Second,
		},
	}
	if mutate != nil {
		mutate(cfg)
	}
	return cfg
}

func validTestAPI() *appconf.API {
	return &appconf.API{
		Log: appconf.Log{
			Level: "INFO",
		},
		DB: appconf.DB{
			DSN: "postgres://localhost:5432/auth?sslmode=disable",
		},
		Web: appconf.Web{
			HTTPAddress:           "0.0.0.0:8000",
			GRPCAddress:           "0.0.0.0:9000",
			ReadTimeout:           time.Second,
			WriteTimeout:          time.Second,
			AllowedCORSOrigin:     "http://localhost:3000",
			RefreshCookieSameSite: "lax",
			RefreshCookieSecure:   true,
		},
		Debug: appconf.Debug{
			Address: "0.0.0.0:6060",
		},
		Auth: appconf.Auth{
			PrivateKeyFile:   "private.pem",
			SigningAlgorithm: "RS256",
			Issuer:           "game-library-auth",
			AccessTokenTTL:   15 * time.Minute,
			RefreshTokenTTL:  24 * time.Hour,
		},
		OAuth: appconf.OAuth{
			GoogleClientID:     "google-client-id",
			GitHubClientID:     "github-client-id",
			GitHubClientSecret: "github-client-secret",
			Timeout:            5 * time.Second,
		},
		Jaeger: appconf.Jaeger{
			OTLPEndpoint: "jaeger-service:4318",
		},
		Graylog: appconf.Graylog{
			Address: "localhost:12201",
		},
		EmailSender: appconf.EmailSender{
			EmailFrom:         "noreply@example.com",
			ContactEmail:      "support@example.com",
			BaseURL:           "https://example.com",
			UnsubscribeURL:    "https://example.com/unsubscribe",
			UnsubscribeSecret: "unsubscribe-secret",
		},
		InfoAPI: appconf.InfoAPI{
			Address: "localhost:9001",
			Timeout: 2 * time.Second,
		},
		Outbox: appconf.Outbox{
			Key: base64.StdEncoding.EncodeToString(make([]byte, 32)),
		},
	}
}
