package facade_test

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/OutOfStack/game-library-auth/internal/facade"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTooManyRequestsError(t *testing.T) {
	t.Run("error message", func(t *testing.T) {
		retryAfter := 5 * time.Minute
		err := facade.NewTooManyRequestsError(retryAfter)

		expectedMsg := "too many requests, retry after5m0s"
		assert.Equal(t, expectedMsg, err.Error())
	})

	t.Run("retry after duration", func(t *testing.T) {
		retryAfter := 10 * time.Minute
		err := facade.NewTooManyRequestsError(retryAfter)

		assert.Equal(t, retryAfter, err.RetryAfter)
	})
}

func TestAsTooManyRequestsError(t *testing.T) {
	t.Run("with constructor pointer and wrapped pointer", func(t *testing.T) {
		err := facade.NewTooManyRequestsError(time.Minute)
		require.IsType(t, (*facade.TooManyRequestsError)(nil), err)
		for _, input := range []error{err, fmt.Errorf("resend: %w", err)} {
			result := facade.AsTooManyRequestsError(input)
			require.NotNil(t, result)
			assert.Equal(t, time.Minute, result.RetryAfter)
		}
	})

	t.Run("with TooManyRequestsError", func(t *testing.T) {
		retryAfter := 5 * time.Minute
		result := facade.AsTooManyRequestsError(facade.NewTooManyRequestsError(retryAfter))
		require.NotNil(t, result)
		assert.Equal(t, retryAfter, result.RetryAfter)
	})

	t.Run("with different error", func(t *testing.T) {
		err := errors.New("some other error")

		result := facade.AsTooManyRequestsError(err)
		assert.Nil(t, result)
	})

	t.Run("with error message copied without wrapping", func(t *testing.T) {
		retryAfter := 3 * time.Minute
		baseErr := facade.NewTooManyRequestsError(retryAfter)
		wrappedErr := errors.New("wrapped: " + baseErr.Error())

		result := facade.AsTooManyRequestsError(wrappedErr)
		assert.Nil(t, result)
	})
}
