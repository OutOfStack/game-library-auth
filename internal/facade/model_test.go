package facade_test

import (
	"errors"
	"testing"
	"time"

	"github.com/OutOfStack/game-library-auth/internal/facade"
)

func TestTooManyRequestsError(t *testing.T) {
	t.Run("error message", func(t *testing.T) {
		retryAfter := 5 * time.Minute
		err := facade.NewTooManyRequestsError(retryAfter)

		expectedMsg := "too many requests, retry after5m0s"
		if err.Error() != expectedMsg {
			t.Errorf("expected error message %q, got %q", expectedMsg, err.Error())
		}
	})

	t.Run("retry after duration", func(t *testing.T) {
		retryAfter := 10 * time.Minute
		err := facade.NewTooManyRequestsError(retryAfter)

		if err.RetryAfter != retryAfter {
			t.Errorf("expected retry after %v, got %v", retryAfter, err.RetryAfter)
		}
	})
}

func TestAsTooManyRequestsError(t *testing.T) {
	t.Run("with TooManyRequestsError", func(t *testing.T) {
		retryAfter := 5 * time.Minute
		err := facade.NewTooManyRequestsError(retryAfter)

		result := facade.AsTooManyRequestsError(&err)
		if result == nil {
			t.Fatal("expected non-nil result")
			return
		}
		if result.RetryAfter != retryAfter {
			t.Errorf("expected retry after %v, got %v", retryAfter, result.RetryAfter)
		}
	})

	t.Run("with different error", func(t *testing.T) {
		err := errors.New("some other error")

		result := facade.AsTooManyRequestsError(err)
		if result != nil {
			t.Errorf("expected nil result, got %v", result)
		}
	})

	t.Run("with wrapped TooManyRequestsError", func(t *testing.T) {
		retryAfter := 3 * time.Minute
		baseErr := facade.NewTooManyRequestsError(retryAfter)
		wrappedErr := errors.New("wrapped: " + baseErr.Error())

		result := facade.AsTooManyRequestsError(wrappedErr)
		if result != nil {
			t.Errorf("expected nil for non-wrapped error, got %v", result)
		}
	})
}
