package resendapi_test

import (
	"testing"
	"time"

	"github.com/OutOfStack/game-library-auth/internal/client/resendapi"
	"github.com/stretchr/testify/require"
)

func TestNewClient_Success(t *testing.T) {
	client, err := resendapi.NewClient(resendapi.Config{
		APIToken:       "test-token",
		FromEmail:      "test@example.com",
		ContactEmail:   "contact@example.com",
		BaseURL:        "https://example.com",
		UnsubscribeURL: "https://example.com/unsubscribe",
		Timeout:        5 * time.Second,
	})

	require.NoError(t, err)
	require.NotNil(t, client)
}
