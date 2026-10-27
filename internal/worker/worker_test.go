package worker //nolint:testpackage // Tests control scheduling and individual attempts without exporting test hooks.

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/OutOfStack/game-library-auth/internal/database"
	"github.com/OutOfStack/game-library-auth/internal/notification"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

type storeStub struct {
	eligible                bool
	status, reason, message string
	delay                   time.Duration
	completeErr             error
	housekeepingErr         error
	items                   []database.Notification
	claims                  int
}

func (s *storeStub) ClaimNotification(context.Context, string) (database.Notification, error) {
	s.claims++
	if len(s.items) == 0 {
		return database.Notification{}, database.ErrNotFound
	}
	n := s.items[0]
	s.items = s.items[1:]
	return n, nil
}
func (s *storeStub) NotificationEligible(context.Context, database.Notification, string) (bool, error) {
	return s.eligible, nil
}
func (s *storeStub) CompleteNotification(_ context.Context, _ database.Notification, id string) error {
	s.message = id
	return s.completeErr
}
func (s *storeStub) RetryNotification(_ context.Context, _ database.Notification, d time.Duration, reason string) error {
	s.delay = d
	s.reason = reason
	return nil
}
func (s *storeStub) FinishNotification(_ context.Context, _ database.Notification, status, reason string) error {
	s.status = status
	s.reason = reason
	return nil
}
func (s *storeStub) ExpireNotifications(context.Context) error { return s.housekeepingErr }
func (*storeStub) PurgeNotifications(context.Context) error    { return nil }
func (*storeStub) NotificationStats(context.Context) (database.NotificationStats, error) {
	return database.NotificationStats{}, nil
}

type senderStub struct {
	emails []notification.Email
	keys   []string
	err    error
	onSend func(context.Context)
}

func (s *senderStub) Send(ctx context.Context, e notification.Email, key string) (string, error) {
	s.emails = append(s.emails, e)
	s.keys = append(s.keys, key)
	if s.onSend != nil {
		s.onSend(ctx)
	}
	return "provider-id", s.err
}

func testWorker(t *testing.T) (*Worker, *storeStub, *senderStub, database.Notification) {
	t.Helper()
	s := &storeStub{eligible: true}
	sender := &senderStub{}
	w := New(zap.NewNop(), s, sender, prometheus.NewRegistry())
	w.random = func() float64 { return 0.5 }
	n := database.Notification{ID: "id", ReferenceID: "verification", Kind: notification.VerificationKind, Attempts: 1,
		ExpiresAt: time.Now().Add(time.Hour), LockedAt: time.Now(), LockedBy: "owner"}
	var err error
	n.Payload, err = json.Marshal(notification.Email{From: "sender@example.com", To: []string{"user@example.com"}, Subject: "Verify", Text: "123456"})
	require.NoError(t, err)
	return w, s, sender, n
}

func TestAmbiguousAcknowledgmentReusesEmailAndKey(t *testing.T) {
	w, s, sender, n := testWorker(t)
	s.completeErr = errors.New("database unavailable")
	_, err := w.process(t.Context(), n)
	require.Error(t, err)
	require.Zero(t, s.delay, "do not release a claim after an ambiguous acknowledgment")
	s.completeErr = nil
	_, err = w.process(t.Context(), n)
	require.NoError(t, err)
	require.Equal(t, sender.emails[0], sender.emails[1])
	require.Equal(t, []string{"notification/id", "notification/id"}, sender.keys)
	require.Equal(t, "provider-id", s.message)
}

func TestWorkerRetriesSDKFailuresWithoutLoggingProviderMessages(t *testing.T) {
	for _, tc := range []struct {
		name     string
		err      error
		reason   string
		minDelay time.Duration
	}{
		{"network", errors.New("timeout"), "delivery_error", time.Second},
		{"invalid recipient", errors.New("[ERROR]: invalid sensitive recipient"), "delivery_error", time.Second},
		{"rate limit", &notification.RateLimitError{RetryAfter: time.Minute}, "rate_limit", time.Minute},
		{"wrapped rate limit", fmt.Errorf("send: %w", &notification.RateLimitError{RetryAfter: time.Minute}), "rate_limit", time.Minute},
		{"no retry hint", &notification.RateLimitError{}, "rate_limit", time.Second},
		{"short retry", &notification.RateLimitError{RetryAfter: time.Second}, "rate_limit", 3750 * time.Millisecond},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w, s, sender, n := testWorker(t)
			core, logs := observer.New(zap.WarnLevel)
			w.log = zap.New(core)
			sender.err = tc.err
			pause, err := w.process(t.Context(), n)
			require.NoError(t, err)
			require.Empty(t, s.status)
			require.GreaterOrEqual(t, s.delay, tc.minDelay)
			require.Equal(t, s.delay, pause)
			require.Equal(t, tc.reason, s.reason)
			require.Len(t, logs.All(), 1)
			require.NotContains(t, fmt.Sprint(logs.All()), "sensitive recipient")
		})
	}
}

func TestWorkerSkipsObsoleteAndExpiredMessages(t *testing.T) {
	w, s, sender, n := testWorker(t)
	s.eligible = false
	_, err := w.process(t.Context(), n)
	require.NoError(t, err)
	require.Equal(t, "cancelled", s.status)
	require.Empty(t, sender.keys)
	n.ExpiresAt = time.Now().Add(-time.Second)
	_, err = w.process(t.Context(), n)
	require.NoError(t, err)
	require.Equal(t, "expired", s.status)
	require.Empty(t, sender.keys)
}

func TestMalformedPayloadFailsWithoutSending(t *testing.T) {
	for _, payload := range []string{`{`, `{"from":"sender@example.com","to":[],"subject":"Verify","text":"123456"}`} {
		w, s, sender, n := testWorker(t)
		n.Payload = []byte(payload)
		pause, err := w.process(t.Context(), n)
		require.NoError(t, err)
		require.Zero(t, pause)
		require.Equal(t, "failed", s.status)
		require.Equal(t, "invalid_payload", s.reason)
		require.Empty(t, sender.keys)
	}
}

func TestProviderRejectionFailsWithoutRetry(t *testing.T) {
	w, s, sender, n := testWorker(t)
	sender.err = fmt.Errorf("%w: status 422", notification.ErrRejected)
	pause, err := w.process(t.Context(), n)
	require.NoError(t, err)
	require.Zero(t, pause)
	require.Equal(t, "failed", s.status)
	require.Equal(t, "rejected", s.reason)
	require.Zero(t, s.delay)
}

func TestProviderMisconfigurationKeepsPayloadAndPausesWorker(t *testing.T) {
	w, s, sender, n := testWorker(t)
	core, logs := observer.New(zap.WarnLevel)
	w.log = zap.New(core)
	sender.err = fmt.Errorf("%w: status 401", notification.ErrMisconfigured)
	pause, err := w.process(t.Context(), n)
	require.NoError(t, err)
	require.Empty(t, s.status)
	require.Equal(t, "provider_configuration", s.reason)
	require.GreaterOrEqual(t, s.delay, MisconfigurationPause)
	require.Equal(t, s.delay, pause)
	require.Len(t, logs.FilterLevelExact(zap.ErrorLevel).All(), 1)
}

func TestProviderMisconfigurationPausesFurtherClaims(t *testing.T) {
	w, s, sender, n := testWorker(t)
	s.items = []database.Notification{n, n}
	sender.err = fmt.Errorf("%w: status 403", notification.ErrMisconfigured)
	started := make(chan struct{}, 2)
	sender.onSend = func(context.Context) { started <- struct{}{} }
	w.random = func() float64 { return 0 }
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	done := make(chan struct{})
	go func() { w.Run(ctx); close(done) }()
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("worker did not start")
	}
	time.AfterFunc(30*time.Millisecond, cancel)
	<-done
	require.Equal(t, 1, s.claims)
}

func TestHousekeepingFailureDoesNotBlockClaims(t *testing.T) {
	w, s, sender, n := testWorker(t)
	s.housekeepingErr = errors.New("purge timeout")
	s.items = []database.Notification{n}
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	sender.onSend = func(context.Context) { cancel() }
	w.random = func() float64 { return 0 }
	w.Run(ctx)
	require.Equal(t, 1, s.claims)
	require.Equal(t, "provider-id", s.message)
}

func TestBackoffIsBoundedAndJittered(t *testing.T) {
	w, _, _, _ := testWorker(t)
	for _, attempt := range []int{1, 2, 10, 1000000} {
		w.random = func() float64 { return 0 }
		low := w.retryDelay(attempt)
		w.random = func() float64 { return 0.999 }
		high := w.retryDelay(attempt)
		require.GreaterOrEqual(t, low, 2500*time.Millisecond)
		require.Greater(t, high, low)
		require.LessOrEqual(t, high, 5*time.Minute)
	}
}

func TestShutdownFinishesCurrentAttemptAndStopsClaims(t *testing.T) {
	w, s, sender, n := testWorker(t)
	s.items = []database.Notification{n, n}
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	sender.onSend = func(ctx context.Context) { cancel(); require.NoError(t, ctx.Err()) }
	w.random = func() float64 { return 0 }
	w.Run(ctx)
	require.Equal(t, 1, s.claims)
	require.Equal(t, "provider-id", s.message)
}

func TestProviderTransportOutagePausesFurtherClaims(t *testing.T) {
	w, s, sender, n := testWorker(t)
	s.items = []database.Notification{n, n}
	sender.err = errors.New("connection refused")
	started := make(chan struct{}, 2)
	sender.onSend = func(context.Context) { started <- struct{}{} }
	w.random = func() float64 { return 0 }
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	done := make(chan struct{})
	go func() { w.Run(ctx); close(done) }()
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("worker did not start")
	}
	// the retry delay is 2.5 seconds; cancellation should occur before another claim
	time.AfterFunc(30*time.Millisecond, cancel)
	<-done
	require.Equal(t, 1, s.claims)
}
