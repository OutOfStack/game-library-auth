package resendapi //nolint:testpackage

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/OutOfStack/game-library-auth/internal/notification"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSendReusesPreparedPayloadAndIdempotencyKey(t *testing.T) {
	var bodies []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "notification/id", r.Header.Get("Idempotency-Key"))
		body, err := io.ReadAll(r.Body)
		assert.NoError(t, err)
		bodies = append(bodies, string(body))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"message"}`))
	}))
	defer server.Close()
	c := NewClient("test", time.Second)
	var err error
	c.client.BaseURL, err = url.Parse(server.URL + "/")
	require.NoError(t, err)
	email := notification.Email{From: "from@example.com", To: []string{"to@example.com"}, Subject: "Verify", HTML: "123456", Text: "123456"}
	for range 2 {
		id, err := c.Send(t.Context(), email, "notification/id")
		require.NoError(t, err)
		require.Equal(t, "message", id)
	}
	require.Len(t, bodies, 2)
	require.Equal(t, bodies[0], bodies[1])
}

func TestSendReturnsRateLimitDelay(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Retry-After", "60")
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"name":"rate_limit_exceeded","message":"sensitive recipient"}`))
	}))
	defer server.Close()
	c := NewClient("", time.Second)
	var err error
	c.client.BaseURL, err = url.Parse(server.URL + "/")
	require.NoError(t, err)
	_, err = c.Send(t.Context(), notification.Email{}, "notification/id")
	var rateLimit *notification.RateLimitError
	require.ErrorAs(t, err, &rateLimit)
	require.Equal(t, time.Minute, rateLimit.RetryAfter)
	require.NotContains(t, err.Error(), "sensitive recipient")
}

func TestRetryAfter(t *testing.T) {
	require.Equal(t, time.Minute, retryAfter("60"))
	require.InDelta(t, time.Minute, retryAfter(time.Now().Add(time.Minute).UTC().Format(http.TimeFormat)), float64(2*time.Second))
	for _, value := range []string{"", "invalid", "-1", "0", time.Now().Add(-time.Minute).UTC().Format(http.TimeFormat)} {
		require.Zero(t, retryAfter(value), value)
	}
}

func TestSendClassifiesProviderFailures(t *testing.T) {
	for status, want := range map[int]error{
		http.StatusUnprocessableEntity: notification.ErrRejected,
		http.StatusBadRequest:          notification.ErrRejected,
		http.StatusUnauthorized:        notification.ErrMisconfigured,
		http.StatusForbidden:           notification.ErrMisconfigured,
		http.StatusNotFound:            notification.ErrMisconfigured,
		http.StatusConflict:            nil,
		http.StatusInternalServerError: nil,
	} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(status)
				_, _ = w.Write([]byte(`{"message":"sensitive recipient"}`))
			}))
			defer server.Close()
			c := NewClient("", time.Second)
			var err error
			c.client.BaseURL, err = url.Parse(server.URL + "/")
			require.NoError(t, err)
			_, err = c.Send(t.Context(), notification.Email{}, "notification/id")
			require.Error(t, err)
			if want == nil {
				require.NotErrorIs(t, err, notification.ErrRejected)
				require.NotErrorIs(t, err, notification.ErrMisconfigured)
				return
			}
			require.ErrorIs(t, err, want)
			require.NotContains(t, err.Error(), "sensitive recipient")
		})
	}
}

func TestRendererPreparesEmailWithoutProviderCredentials(t *testing.T) {
	r, err := NewRenderer(Config{FromEmail: "from@example.com", BaseURL: "https://example.com", UnsubscribeURL: "https://example.com/unsubscribe"})
	require.NoError(t, err)
	email, err := r.PrepareVerification(SendEmailVerificationRequest{Email: "to@example.com", Username: "user", VerificationCode: "123456", UnsubscribeToken: "token"})
	require.NoError(t, err)
	require.Contains(t, email.HTML, "123456")
	require.Contains(t, email.Text, "123456")
	require.Contains(t, email.HTML, "token")
	require.Equal(t, []string{"to@example.com"}, email.To)
}
