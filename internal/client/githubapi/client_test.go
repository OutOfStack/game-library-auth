package githubapi_test

import (
	"testing"
	"time"

	"github.com/OutOfStack/game-library-auth/internal/client/githubapi"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	client := githubapi.NewClient(githubapi.Config{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		Timeout:      5 * time.Second,
	})
	require.NotNil(t, client)
}
