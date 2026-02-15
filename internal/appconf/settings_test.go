package appconf_test

import (
	"testing"
	"time"

	"github.com/OutOfStack/game-library-auth/internal/appconf"
	"github.com/stretchr/testify/require"
)

func TestCfgValidate_ValidConfig(t *testing.T) {
	cfg := validCfg()

	err := cfg.Validate()
	require.NoError(t, err)
}

func TestCfgValidate_AllowsCaseInsensitiveSameSite(t *testing.T) {
	cfg := validCfg()
	cfg.Web.RefreshCookieSameSite = "StRiCt"

	err := cfg.Validate()
	require.NoError(t, err)
}

func TestCfgValidate_NilConfig(t *testing.T) {
	var cfg *appconf.Cfg

	err := cfg.Validate()
	require.EqualError(t, err, "cfg is nil")
}

func TestCfgValidate_ErrorCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		mutate  func(*appconf.Cfg)
		wantErr string
	}{
		{
			name: "missing db dsn",
			mutate: func(cfg *appconf.Cfg) {
				cfg.DB.DSN = ""
			},
			wantErr: "DB_DSN is required",
		},
		{
			name: "missing app http address",
			mutate: func(cfg *appconf.Cfg) {
				cfg.Web.HTTPAddress = ""
			},
			wantErr: "APP_HTTP_ADDRESS is required",
		},
		{
			name: "missing app grpc address",
			mutate: func(cfg *appconf.Cfg) {
				cfg.Web.GRPCAddress = ""
			},
			wantErr: "APP_GRPC_ADDRESS is required",
		},
		{
			name: "missing app debug address",
			mutate: func(cfg *appconf.Cfg) {
				cfg.Web.DebugAddress = ""
			},
			wantErr: "APP_DEBUG_ADDRESS is required",
		},
		{
			name: "invalid app read timeout",
			mutate: func(cfg *appconf.Cfg) {
				cfg.Web.ReadTimeout = 0
			},
			wantErr: "APP_READTIMEOUT must be greater than 0",
		},
		{
			name: "invalid app write timeout",
			mutate: func(cfg *appconf.Cfg) {
				cfg.Web.WriteTimeout = 0
			},
			wantErr: "APP_WRITETIMEOUT must be greater than 0",
		},
		{
			name: "missing cors origin",
			mutate: func(cfg *appconf.Cfg) {
				cfg.Web.AllowedCORSOrigin = ""
			},
			wantErr: "APP_ALLOWEDCORSORIGIN is required",
		},
		{
			name: "invalid cookie samesite",
			mutate: func(cfg *appconf.Cfg) {
				cfg.Web.RefreshCookieSameSite = "invalid"
			},
			wantErr: "APP_REFRESH_TOKEN_COOKIE_SAMESITE must be one of lax, strict, none",
		},
		{
			name: "missing auth private key file",
			mutate: func(cfg *appconf.Cfg) {
				cfg.Auth.PrivateKeyFile = ""
			},
			wantErr: "AUTH_PRIVATEKEYFILE is required",
		},
		{
			name: "missing auth signing algorithm",
			mutate: func(cfg *appconf.Cfg) {
				cfg.Auth.SigningAlgorithm = ""
			},
			wantErr: "AUTH_SIGNINGALG is required",
		},
		{
			name: "missing auth issuer",
			mutate: func(cfg *appconf.Cfg) {
				cfg.Auth.Issuer = ""
			},
			wantErr: "AUTH_ISSUER is required",
		},
		{
			name: "missing auth google client id",
			mutate: func(cfg *appconf.Cfg) {
				cfg.Auth.GoogleClientID = ""
			},
			wantErr: "AUTH_GOOGLECLIENTID is required",
		},
		{
			name: "invalid auth access token ttl",
			mutate: func(cfg *appconf.Cfg) {
				cfg.Auth.AccessTokenTTL = 0
			},
			wantErr: "AUTH_ACCESSTOKENTTL must be greater than 0",
		},
		{
			name: "invalid auth refresh token ttl",
			mutate: func(cfg *appconf.Cfg) {
				cfg.Auth.RefreshTokenTTL = 0
			},
			wantErr: "AUTH_REFRESHTOKENTTL must be greater than 0",
		},
		{
			name: "missing jaeger endpoint",
			mutate: func(cfg *appconf.Cfg) {
				cfg.Jaeger.OTLPEndpoint = ""
			},
			wantErr: "JAEGER_OTLP_ENDPOINT is required",
		},
		{
			name: "invalid jaeger endpoint format",
			mutate: func(cfg *appconf.Cfg) {
				cfg.Jaeger.OTLPEndpoint = "%"
			},
			wantErr: "JAEGER_OTLP_ENDPOINT must be host:port without scheme",
		},
		{
			name: "missing graylog addr",
			mutate: func(cfg *appconf.Cfg) {
				cfg.Graylog.Address = ""
			},
			wantErr: "GRAYLOG_ADDR is required",
		},
		{
			name: "missing log level",
			mutate: func(cfg *appconf.Cfg) {
				cfg.Log.Level = ""
			},
			wantErr: "LOG_LEVEL is required",
		},
		{
			name: "missing email sender api token",
			mutate: func(cfg *appconf.Cfg) {
				cfg.EmailSender.APIToken = ""
			},
			wantErr: "EMAIL_SENDER_API_TOKEN is required",
		},
		{
			name: "invalid email sender api timeout",
			mutate: func(cfg *appconf.Cfg) {
				cfg.EmailSender.APITimeout = 0
			},
			wantErr: "EMAIL_SENDER_API_TIMEOUT must be greater than 0",
		},
		{
			name: "missing email sender from",
			mutate: func(cfg *appconf.Cfg) {
				cfg.EmailSender.EmailFrom = ""
			},
			wantErr: "EMAIL_SENDER_EMAIL_FROM is required",
		},
		{
			name: "missing email sender contact email",
			mutate: func(cfg *appconf.Cfg) {
				cfg.EmailSender.ContactEmail = ""
			},
			wantErr: "EMAIL_SENDER_CONTACT_EMAIL is required",
		},
		{
			name: "missing email sender base url",
			mutate: func(cfg *appconf.Cfg) {
				cfg.EmailSender.BaseURL = ""
			},
			wantErr: "EMAIL_SENDER_BASE_URL is required",
		},
		{
			name: "missing email sender unsubscribe url",
			mutate: func(cfg *appconf.Cfg) {
				cfg.EmailSender.UnsubscribeURL = ""
			},
			wantErr: "EMAIL_SENDER_UNSUBSCRIBE_URL is required",
		},
		{
			name: "missing email sender unsubscribe secret",
			mutate: func(cfg *appconf.Cfg) {
				cfg.EmailSender.UnsubscribeSecret = ""
			},
			wantErr: "EMAIL_SENDER_UNSUBSCRIBE_SECRET is required",
		},
		{
			name: "missing infoapi address",
			mutate: func(cfg *appconf.Cfg) {
				cfg.InfoAPI.Address = ""
			},
			wantErr: "INFOAPI_ADDRESS is required",
		},
		{
			name: "invalid infoapi timeout",
			mutate: func(cfg *appconf.Cfg) {
				cfg.InfoAPI.Timeout = 0
			},
			wantErr: "INFOAPI_TIMEOUT must be greater than 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := validCfg()
			tt.mutate(cfg)

			err := cfg.Validate()
			require.EqualError(t, err, tt.wantErr)
		})
	}
}

func validCfg() *appconf.Cfg {
	return &appconf.Cfg{
		DB: appconf.DB{
			DSN: "postgres://localhost:5432/game_library?sslmode=disable",
		},
		Web: appconf.Web{
			HTTPAddress:           ":8080",
			GRPCAddress:           ":9090",
			DebugAddress:          ":6060",
			ReadTimeout:           time.Second,
			WriteTimeout:          time.Second,
			AllowedCORSOrigin:     "*",
			RefreshCookieSameSite: "lax",
			RefreshCookieSecure:   true,
		},
		Auth: appconf.Auth{
			PrivateKeyFile:   "private.pem",
			SigningAlgorithm: "RS256",
			Issuer:           "game-library-auth",
			GoogleClientID:   "google-client-id",
			AccessTokenTTL:   15 * time.Minute,
			RefreshTokenTTL:  24 * time.Hour,
		},
		Jaeger: appconf.Jaeger{
			OTLPEndpoint: "localhost:4317",
		},
		Graylog: appconf.Graylog{
			Address: "localhost:12201",
		},
		Log: appconf.Log{
			Level: "info",
		},
		EmailSender: appconf.EmailSender{
			APIToken:          "api-token",
			APITimeout:        5 * time.Second,
			EmailFrom:         "noreply@example.com",
			ContactEmail:      "support@example.com",
			BaseURL:           "https://example.com",
			UnsubscribeURL:    "https://example.com/unsubscribe",
			UnsubscribeSecret: "unsubscribe-secret",
		},
		InfoAPI: appconf.InfoAPI{
			Address: "localhost:50051",
			Timeout: 2 * time.Second,
		},
	}
}
