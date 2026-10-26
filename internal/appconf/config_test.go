package appconf_test

import (
	"encoding/base64"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/OutOfStack/game-library-auth/internal/appconf"
	"github.com/stretchr/testify/require"
)

type pointerSection struct{}

func (*pointerSection) Validate() error { return nil }

func Test_Load_DotEnv(t *testing.T) {
	t.Chdir(t.TempDir())
	config := `LOG_LEVEL=INFO
DB_DSN=postgres://test:test@localhost/auth
APP_HTTP_ADDRESS=localhost:8000
APP_GRPC_ADDRESS=localhost:9000
APP_DEBUG_ADDRESS=localhost:6060
APP_READTIMEOUT=5s
APP_WRITETIMEOUT=10s
APP_ALLOWEDCORSORIGIN=http://localhost:3000
APP_REFRESH_TOKEN_COOKIE_SAMESITE=lax
APP_REFRESH_TOKEN_COOKIE_SECURE=true
AUTH_PRIVATEKEYFILE=private.pem
AUTH_SIGNINGALG=RS256
AUTH_ISSUER=http://localhost:8000
AUTH_ACCESSTOKENTTL=15m
AUTH_REFRESHTOKENTTL=168h
OAUTH_TIMEOUT=5s
OAUTH_GOOGLECLIENTID=test
OAUTH_GITHUBCLIENTID=test
OAUTH_GITHUBCLIENTSECRET=test
JAEGER_OTLP_ENDPOINT=localhost:4318
GRAYLOG_ADDR=localhost:12201
EMAIL_SENDER_EMAIL_FROM=info@example.com
EMAIL_SENDER_CONTACT_EMAIL=contact@example.com
EMAIL_SENDER_BASE_URL=https://example.com
EMAIL_SENDER_UNSUBSCRIBE_URL=https://example.com/unsubscribe
EMAIL_SENDER_UNSUBSCRIBE_SECRET=test
INFOAPI_ADDRESS=localhost:9001
INFOAPI_TIMEOUT=5s
NOTIFICATION_OUTBOX_KEY=` + base64.StdEncoding.EncodeToString(make([]byte, 32)) + `
`
	for line := range strings.SplitSeq(strings.TrimSpace(config), "\n") {
		key, _, _ := strings.Cut(line, "=")
		t.Setenv(key, "")
	}
	t.Setenv("APP_HTTP_ADDRESS", "localhost:8001")
	require.NoError(t, os.WriteFile(".env", []byte(config), 0o600))

	cfg, err := appconf.Load[appconf.API]()

	require.NoError(t, err)
	require.Equal(t, "postgres://test:test@localhost/auth", cfg.DB.DSN)
	require.Equal(t, "localhost:8001", cfg.Web.HTTPAddress)
	require.Equal(t, 5*time.Second, cfg.Web.ReadTimeout)
	require.Equal(t, "localhost:6060", cfg.Debug.Address)
	require.True(t, cfg.Web.RefreshCookieSecure)
}

func Test_Load_EnvOnly(t *testing.T) {
	t.Chdir(t.TempDir())
	t.Setenv("DB_DSN", "postgres://env:env@localhost/auth")

	cfg, err := appconf.Load[appconf.Manage]()

	require.NoError(t, err)
	require.Equal(t, "postgres://env:env@localhost/auth", cfg.DB.DSN)
}

func Test_Load_MalformedFile(t *testing.T) {
	t.Chdir(t.TempDir())
	require.NoError(t, os.WriteFile(".env", []byte("DB_DSN=\"unfinished\n"), 0o600))

	_, err := appconf.Load[appconf.Manage]()

	require.Error(t, err)
}

func Test_Load_ReportsAllMissingKeys(t *testing.T) {
	t.Chdir(t.TempDir())
	t.Setenv("DB_DSN", "")
	t.Setenv("APP_DEBUG_ADDRESS", "")

	_, err := appconf.Load[struct {
		DB    appconf.DB    `mapstructure:",squash"`
		Debug appconf.Debug `mapstructure:",squash"`
	}]()

	require.EqualError(t, err, "DB_DSN is required\nAPP_DEBUG_ADDRESS is required")
}

func Test_Validate_NotStruct(t *testing.T) {
	for _, cfg := range []any{nil, (*appconf.API)(nil), "invalid"} {
		require.ErrorContains(t, appconf.Validate(cfg), "config must be a struct")
	}
}

func Test_Validate_SectionWithoutValidate(t *testing.T) {
	tests := []struct {
		name      string
		cfg       any
		wantError string
	}{
		{
			name: "no validate method",
			cfg: struct {
				DB  appconf.DB
				Bad struct{ X string }
			}{},
			wantError: "config section Bad has no Validate method",
		},
		{
			name:      "pointer receiver",
			cfg:       struct{ PointerOnly pointerSection }{},
			wantError: "config section PointerOnly has no Validate method",
		},
		{
			name:      "unexported field",
			cfg:       struct{ hidden appconf.DB }{},
			wantError: "config section hidden has no Validate method",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := appconf.Validate(tt.cfg)

			require.EqualError(t, err, tt.wantError)
		})
	}
}

func Test_Validate_AllConfigsHaveValidatableSections(t *testing.T) {
	for _, cfg := range []any{appconf.API{}, appconf.Worker{}, appconf.Manage{}} {
		err := appconf.Validate(cfg)

		require.Error(t, err)
		require.NotContains(t, err.Error(), "has no Validate method")
	}
}
