package observability_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/OutOfStack/game-library-auth/pkg/observability"
	"github.com/stretchr/testify/require"
)

func TestNewDebugServer_ServesEndpoints(t *testing.T) {
	probe := func(status int) http.HandlerFunc {
		return func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(status) }
	}
	server := observability.NewDebugServer(":6060", probe(http.StatusOK), probe(http.StatusServiceUnavailable))

	for path, want := range map[string]int{
		"/liveness":     http.StatusOK,
		"/readiness":    http.StatusServiceUnavailable,
		"/metrics":      http.StatusOK,
		"/debug/pprof/": http.StatusOK,
		"/debug/vars":   http.StatusOK,
	} {
		recorder := httptest.NewRecorder()
		server.Handler.ServeHTTP(recorder, httptest.NewRequestWithContext(t.Context(), http.MethodGet, path, nil))
		require.Equal(t, want, recorder.Code, path)
	}
}
