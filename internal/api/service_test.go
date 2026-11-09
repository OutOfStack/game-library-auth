package api_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	specs "github.com/OutOfStack/game-library-auth/api"
	"github.com/OutOfStack/game-library-auth/internal/api"
	"github.com/OutOfStack/game-library-auth/internal/api/auth"
	"github.com/OutOfStack/game-library-auth/internal/api/unsubscribe"
	"github.com/OutOfStack/game-library-auth/internal/appconf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestService(t *testing.T) {
	log := zap.NewNop()
	authAPI, err := auth.NewAPI(log, nil, nil, nil, auth.APICfg{RefreshTokenCookieSameSite: "strict"})
	require.NoError(t, err)
	unsubscribeAPI := unsubscribe.NewAPI(log, nil, nil, "")
	cfg := &appconf.API{
		Web:    appconf.Web{AllowedCORSOrigin: "http://localhost:3000"},
		Jaeger: appconf.Jaeger{OTLPEndpoint: "localhost:4318"},
	}

	app, tp, err := api.Service(log, authAPI, unsubscribeAPI, cfg)
	require.NoError(t, err)
	t.Cleanup(func() { _ = tp.Shutdown(t.Context()) })

	t.Run("loads embedded views", func(t *testing.T) {
		require.NoError(t, app.Config().Views.Load())
	})

	t.Run("serves embedded spec", func(t *testing.T) {
		resp, err := app.Test(httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/swagger/openapi.yaml", nil))
		require.NoError(t, err)
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "application/yaml", resp.Header.Get("Content-Type"))
		assert.Equal(t, specs.OpenAPISpec, body)
	})

	t.Run("registers spec routes", func(t *testing.T) {
		registered := make(map[[2]string]bool)
		for _, r := range app.GetRoutes(true) {
			registered[[2]string{r.Method, r.Path}] = true
		}
		for _, route := range [][2]string{
			{http.MethodPost, "/signin"},
			{http.MethodPost, "/signup"},
			{http.MethodPost, "/oauth/google"},
			{http.MethodPost, "/oauth/github"},
			{http.MethodPost, "/refresh"},
			{http.MethodPost, "/logout"},
			{http.MethodPatch, "/account"},
			{http.MethodDelete, "/account"},
			{http.MethodPost, "/verify-email"},
			{http.MethodPost, "/resend-verification"},
			{http.MethodGet, "/unsubscribe"},
			{http.MethodPost, "/unsubscribe"},
		} {
			assert.True(t, registered[route], route)
		}
	})
}
