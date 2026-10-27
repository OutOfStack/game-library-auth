// Package notification contains the immutable email payload stored in the outbox
package notification

import (
	"errors"
	"time"
)

// VerificationKind identifies the supported notification payload type
const VerificationKind = "email_verification"

// ErrRejected marks a delivery the provider refused permanently, so retrying cannot help
var ErrRejected = errors.New("email rejected by provider")

// ErrMisconfigured marks a delivery refused because of sender credentials or configuration, so it can succeed once fixed
var ErrMisconfigured = errors.New("email provider rejected sender credentials or configuration")

// RateLimitError marks a delivery the provider throttled. RetryAfter is zero when the provider gave no usable hint
type RateLimitError struct {
	RetryAfter time.Duration
}

// Error implements error without provider details, which may contain recipient data
func (e *RateLimitError) Error() string {
	return "email provider rate limit exceeded, retry after " + e.RetryAfter.String()
}

// Email is rendered once, before enqueueing, so retries have identical provider payloads
type Email struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	HTML    string   `json:"html"`
	Text    string   `json:"text"`
}
