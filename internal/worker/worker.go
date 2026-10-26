// Package worker delivers encrypted notifications claimed from the PostgreSQL outbox
package worker

import (
	"context"
	"errors"
	"math/rand/v2"
	"net/http"
	"strconv"
	"time"
	"uuid"

	"github.com/OutOfStack/game-library-auth/internal/database"
	"github.com/OutOfStack/game-library-auth/internal/notification"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/resend/resend-go/v2"
	"go.uber.org/zap"
)

// AttemptTimeout bounds work below the database lease and shutdown grace period
const AttemptTimeout = 30 * time.Second

// MisconfigurationPause is the minimal worker-wide pause after the provider refuses sender credentials or configuration
const MisconfigurationPause = time.Minute

// Store persists claims and ownership-fenced delivery transitions
type Store interface {
	ClaimNotification(ctx context.Context, owner string) (database.Notification, error)
	NotificationEligible(ctx context.Context, n database.Notification, email string) (bool, error)
	CompleteNotification(ctx context.Context, n database.Notification, messageID string) error
	RetryNotification(ctx context.Context, n database.Notification, delay time.Duration, reason string) error
	FinishNotification(ctx context.Context, n database.Notification, status, reason string) error
	ExpireNotifications(ctx context.Context) error
	PurgeNotifications(ctx context.Context) error
	NotificationStats(ctx context.Context) (database.NotificationStats, error)
}

// Sender submits an immutable email using a stable idempotency key
type Sender interface {
	Send(ctx context.Context, email notification.Email, idempotencyKey string) (string, error)
}

// Worker drains due notifications with bounded attempts and crash-recoverable leases
type Worker struct {
	log     *zap.Logger
	store   Store
	sender  Sender
	cipher  *notification.Cipher
	owner   string
	random  func() float64
	metrics *metrics
}

// New creates one independently identified worker and registers its metrics
func New(log *zap.Logger, store Store, sender Sender, cipher *notification.Cipher, registry prometheus.Registerer) (*Worker, error) {
	m, err := newMetrics(registry)
	if err != nil {
		return nil, err
	}
	return &Worker{log: log, store: store, sender: sender, cipher: cipher, owner: uuid.New().String(), random: rand.Float64, metrics: m}, nil
}

// Run stops acquiring work on cancellation, while an already claimed attempt may finish
func (w *Worker) Run(ctx context.Context) error {
	if !sleep(ctx, time.Duration(w.random()*float64(5*time.Second))) {
		return nil
	}
	var lastSweep, lastPurge, pausedUntil time.Time
	dbFailures := 0
	for ctx.Err() == nil {
		if time.Since(lastSweep) >= time.Minute {
			// housekeeping failures must not block delivery; claim failures back off on their own
			purge := time.Since(lastPurge) >= time.Hour
			if err := w.housekeeping(ctx, purge); err != nil {
				w.log.Error("notification housekeeping failed", zap.Error(err))
			} else if purge {
				lastPurge = time.Now()
			}
			lastSweep = time.Now()
		}
		if delay := time.Until(pausedUntil); delay > 0 {
			if !sleep(ctx, min(delay, w.pollDelay())) {
				return nil
			}
			continue
		}
		claimCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		n, err := w.store.ClaimNotification(claimCtx, w.owner)
		cancel()
		if errors.Is(err, database.ErrNotFound) {
			dbFailures = 0
			if !sleep(ctx, w.pollDelay()) {
				return nil
			}
			continue
		}
		if err != nil {
			w.log.Error("claim notification failed", zap.Error(err))
			dbFailures++
			if !sleep(ctx, w.retryDelay(dbFailures)) {
				return nil
			}
			continue
		}
		dbFailures = 0
		if n.Reclaimed {
			w.metrics.reclaimed.Inc()
		}
		pause, err := w.process(ctx, n)
		if err != nil {
			w.log.Error("notification processing failed; lease will recover", zap.String("notification_id", n.ID), zap.Error(err))
			if !sleep(ctx, w.retryDelay(1)) {
				return nil
			}
		}
		if pause > 0 {
			pausedUntil = time.Now().Add(pause)
		}
	}
	return nil
}

func (w *Worker) process(parent context.Context, n database.Notification) (time.Duration, error) {
	// bound every attempt below the lease, including a graceful-shutdown attempt;
	// local time is used because locked_at comes from the database clock
	ctx, cancel := context.WithTimeout(context.WithoutCancel(parent), AttemptTimeout)
	defer cancel()
	if !time.Now().Before(n.ExpiresAt) {
		return 0, w.finish(ctx, n, "expired", "expired")
	}
	if n.Kind != notification.VerificationKind {
		return 0, w.finish(ctx, n, "failed", "unsupported_kind")
	}
	email, err := w.cipher.Decrypt(n.ID, n.Kind, n.ReferenceID, n.Payload)
	if err != nil {
		// keep the payload for recovery after a key fix and keep draining the rest of the queue
		w.log.Error("notification payload authentication failed; check NOTIFICATION_OUTBOX_KEY", zap.String("notification_id", n.ID))
		return 0, w.retry(ctx, n, w.retryDelay(n.Attempts), "payload_authentication")
	}
	if len(email.To) != 1 || email.To[0] == "" || email.From == "" || email.Subject == "" || (email.Text == "" && email.HTML == "") {
		return 0, w.finish(ctx, n, "failed", "invalid_payload")
	}
	eligible, err := w.store.NotificationEligible(ctx, n, email.To[0])
	if err != nil {
		return 0, err
	}
	if !eligible {
		return 0, w.finish(ctx, n, "cancelled", "obsolete_or_unsubscribed")
	}
	sendCtx, sendCancel := context.WithDeadline(ctx, n.ExpiresAt)
	defer sendCancel()
	if err = sendCtx.Err(); err != nil {
		return 0, err
	}
	messageID, err := w.sender.Send(sendCtx, email, "notification/"+n.ID)
	if err == nil {
		if err = w.store.CompleteNotification(ctx, n, messageID); err != nil {
			return 0, err
		}
		w.metrics.outcomes.WithLabelValues("sent").Inc()
		return 0, nil
	}
	if !time.Now().Before(n.ExpiresAt) {
		return 0, w.finish(ctx, n, "expired", "expired")
	}
	if errors.Is(err, notification.ErrRejected) {
		w.log.Warn("notification rejected by provider", zap.String("notification_id", n.ID), zap.Error(err))
		return 0, w.finish(ctx, n, "failed", "rejected")
	}
	delay := w.retryDelay(n.Attempts)
	reason := "delivery_error"
	if rateLimit, ok := errors.AsType[*resend.RateLimitError](err); ok {
		delay = max(delay, retryAfter(rateLimit.RetryAfter))
		reason = "rate_limit"
	}
	if errors.Is(err, notification.ErrMisconfigured) {
		// keep the payload so delivery resumes once credentials or sender configuration are fixed
		w.log.Error("email provider refused sender credentials or configuration; check EMAIL_SENDER_API_TOKEN and sender domain",
			zap.String("notification_id", n.ID), zap.Error(err))
		delay = max(delay, MisconfigurationPause)
		reason = "provider_configuration"
	}
	if err = w.retry(ctx, n, delay, reason); err != nil {
		return 0, err
	}
	// SDK error messages may contain recipient details; log only our retry reason
	w.log.Warn("notification delivery deferred", zap.String("notification_id", n.ID), zap.String("reason", reason), zap.Duration("retry_after", delay))
	return delay, nil
}

func (w *Worker) retry(ctx context.Context, n database.Notification, delay time.Duration, reason string) error {
	if err := w.store.RetryNotification(ctx, n, delay, reason); err != nil {
		return err
	}
	w.metrics.outcomes.WithLabelValues("retry").Inc()
	return nil
}

func (w *Worker) finish(ctx context.Context, n database.Notification, status, reason string) error {
	if err := w.store.FinishNotification(ctx, n, status, reason); err != nil {
		return err
	}
	w.metrics.outcomes.WithLabelValues(status).Inc()
	return nil
}

func (w *Worker) housekeeping(ctx context.Context, purge bool) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := w.store.ExpireNotifications(ctx); err != nil {
		return err
	}
	if purge {
		if err := w.store.PurgeNotifications(ctx); err != nil {
			return err
		}
	}
	stats, err := w.store.NotificationStats(ctx)
	if err != nil {
		return err
	}
	w.metrics.pending.Set(float64(stats.Pending))
	w.metrics.oldest.Set(stats.OldestAge)
	return nil
}

func (w *Worker) retryDelay(attempt int) time.Duration {
	base := min(5*time.Second<<min(max(attempt-1, 0), 6), 5*time.Minute)
	return base/2 + time.Duration(w.random()*float64(base/2))
}
func (w *Worker) pollDelay() time.Duration {
	return 4*time.Second + time.Duration(w.random()*float64(2*time.Second))
}

func sleep(ctx context.Context, d time.Duration) bool {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return ctx.Err() == nil
	}
}

func retryAfter(value string) time.Duration {
	if seconds, err := strconv.ParseInt(value, 10, 32); err == nil && seconds > 0 {
		return time.Duration(seconds) * time.Second
	}
	if deadline, err := http.ParseTime(value); err == nil {
		return max(0, time.Until(deadline))
	}
	return 0
}
