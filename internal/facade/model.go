package facade

import (
	"errors"
	"time"
)

const (
	maxUsernameLen = 32

	defaultVrfCodeLen = 6
)

// TooManyRequestsError - error with retry after
type TooManyRequestsError struct {
	RetryAfter time.Duration
}

// Error implements error interface
func (e TooManyRequestsError) Error() string {
	return "too many requests, retry after" + e.RetryAfter.String()
}

// NewTooManyRequestsError - creates a new ErrTooManyRequestsRetryAfter
func NewTooManyRequestsError(retryAfter time.Duration) TooManyRequestsError {
	return TooManyRequestsError{
		RetryAfter: retryAfter,
	}
}

// AsTooManyRequestsError - returns *TooManyRequestsError if err is of type TooManyRequestsError
func AsTooManyRequestsError(err error) *TooManyRequestsError {
	var tooManyRequestsErr *TooManyRequestsError
	if errors.As(err, &tooManyRequestsErr) {
		return tooManyRequestsErr
	}
	return nil
}

// emailVerificationResult - result of creating an email verification record
type emailVerificationResult struct {
	ID               string
	Code             string
	UnsubscribeToken string
}

// RefreshToken represents a refresh token
type RefreshToken struct {
	Token     string
	ExpiresAt time.Time
}

// TokenPair represents an access token and refresh token pair
type TokenPair struct {
	AccessToken  string
	RefreshToken RefreshToken
}
