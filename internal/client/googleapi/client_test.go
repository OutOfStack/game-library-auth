package googleapi_test

import (
	"testing"
	"time"

	"github.com/OutOfStack/game-library-auth/internal/client/googleapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient_Success(t *testing.T) {
	client, err := googleapi.NewClient(t.Context(), "test-client-id", 5*time.Second)

	require.NoError(t, err)
	require.NotNil(t, client)
}

func TestValidateIDToken_InvalidToken(t *testing.T) {
	client, err := googleapi.NewClient(t.Context(), "test-client-id", 5*time.Second)
	require.NoError(t, err)

	_, err = client.ValidateIDToken(t.Context(), "invalid-token")

	assert.Error(t, err)
}
