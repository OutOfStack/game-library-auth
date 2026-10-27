package resendapi

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/OutOfStack/game-library-auth/internal/notification"
	"github.com/OutOfStack/game-library-auth/pkg/observability"
	"github.com/resend/resend-go/v2"
	"go.opentelemetry.io/otel"
)

var tracer = otel.Tracer("resendapi")

// Client delivers prepared emails
type Client struct {
	client *resend.Client
}

// NewClient creates a delivery client with bounded HTTP requests
func NewClient(apiToken string, timeout time.Duration) *Client {
	httpClient := &http.Client{
		Transport: statusTransport{next: observability.NewTransport("resend_api_client", observability.WithOtel())},
		Timeout:   timeout,
	}
	return &Client{client: resend.NewCustomClient(httpClient, apiToken)}
}

// Send submits payload with its stable idempotency key
func (c *Client) Send(ctx context.Context, email notification.Email, idempotencyKey string) (string, error) {
	ctx, span := tracer.Start(ctx, "sendNotification")
	defer span.End()

	params := &resend.SendEmailRequest{
		From:    email.From,
		To:      email.To,
		Subject: email.Subject,
		Html:    email.HTML,
		Text:    email.Text,
	}
	// the SDK drops HTTP status codes, so capture them to tell rejections from transient failures
	var status int
	ctx = context.WithValue(ctx, statusKey{}, &status)
	sent, err := c.client.Emails.SendWithOptions(ctx, params, &resend.SendEmailOptions{IdempotencyKey: idempotencyKey})
	if err != nil {
		if rateLimit, ok := errors.AsType[*resend.RateLimitError](err); ok {
			return "", &notification.RateLimitError{RetryAfter: retryAfter(rateLimit.RetryAfter)}
		}
		// the SDK message may contain recipient details, so keep only the status
		switch status {
		case http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound:
			// invalid API key, unverified domain or wrong endpoint affect every message until fixed
			return "", fmt.Errorf("%w: status %d", notification.ErrMisconfigured, status)
		case http.StatusBadRequest, http.StatusUnprocessableEntity:
			return "", fmt.Errorf("%w: status %d", notification.ErrRejected, status)
		}
		return "", err
	}
	if sent == nil || sent.Id == "" {
		return "", errors.New("email provider returned no message ID")
	}
	return sent.Id, nil
}

// parses a Retry-After header given in seconds or as an HTTP date; zero when absent or invalid
func retryAfter(value string) time.Duration {
	if seconds, err := strconv.ParseInt(value, 10, 32); err == nil && seconds > 0 {
		return time.Duration(seconds) * time.Second
	}
	if deadline, err := http.ParseTime(value); err == nil {
		return max(0, time.Until(deadline))
	}
	return 0
}

type statusKey struct{}

// statusTransport records the response status into the *int stored in the request context under statusKey
type statusTransport struct{ next http.RoundTripper }

// RoundTrip implements http.RoundTripper
func (t statusTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := t.next.RoundTrip(req)
	if status, ok := req.Context().Value(statusKey{}).(*int); ok && resp != nil {
		*status = resp.StatusCode
	}
	return resp, err
}
