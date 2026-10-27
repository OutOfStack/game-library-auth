package tools_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/OutOfStack/game-library-auth/internal/api/tools"
	"github.com/stretchr/testify/require"
)

type databasePing struct{ err error }

func (p databasePing) PingContext(context.Context) error { return p.err }

func TestHealthCheck_Liveness(t *testing.T) {
	t.Setenv("KUBERNETES_PODNAME", "auth-1")

	response := httptest.NewRecorder()
	tools.NewHealthCheck(databasePing{errors.New("db down")}).Liveness(response, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/liveness", nil))

	require.Equal(t, http.StatusOK, response.Code)
	body := checkHealthResponse(t, response, "OK")
	require.Equal(t, "auth-1", body["pod"])
}

func TestHealthCheck_Readiness(t *testing.T) {
	tests := []struct {
		name       string
		ready      bool
		pingErr    error
		wantCode   int
		wantStatus string
	}{
		{name: "ready", ready: true, wantCode: http.StatusOK, wantStatus: "OK"},
		{name: "database unavailable", ready: true, pingErr: errors.New("db down"), wantCode: http.StatusServiceUnavailable, wantStatus: "database not ready"},
		{name: "not ready", wantCode: http.StatusServiceUnavailable, wantStatus: "not ready"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hc := tools.NewHealthCheck(databasePing{tt.pingErr})
			hc.SetReady(tt.ready)

			response := httptest.NewRecorder()
			hc.Readiness(response, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/readiness", nil))

			require.Equal(t, tt.wantCode, response.Code)
			checkHealthResponse(t, response, tt.wantStatus)
		})
	}
}

func checkHealthResponse(t *testing.T, response *httptest.ResponseRecorder, status string) map[string]string {
	t.Helper()
	require.Contains(t, response.Header().Get("Content-Type"), "application/json")
	var body map[string]string
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &body))
	require.Equal(t, status, body["status"])
	return body
}
