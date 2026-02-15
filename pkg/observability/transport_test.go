package observability_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/OutOfStack/game-library-auth/pkg/observability"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTransport(t *testing.T) {
	transport := observability.NewTransport("test")
	require.NotNil(t, transport)
}

func TestNewTransportReturnsDistinctInstances(t *testing.T) {
	transport1 := observability.NewTransport("test")
	transport2 := observability.NewTransport("test")

	assert.NotSame(t, transport1, transport2, "should return distinct instances per call")
}

func TestNewTransportDifferentInstances(t *testing.T) {
	transport1 := observability.NewTransport("diff1")
	transport2 := observability.NewTransport("diff2")

	assert.NotSame(t, transport1, transport2, "should return different instances for different client names")
}

func TestNewTransportMakesRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	transport := observability.NewTransport("request")
	client := &http.Client{Transport: transport}

	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, server.URL, nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
