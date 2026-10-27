package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/OutOfStack/game-library-auth/internal/notification"
)

// NotificationLease exceeds the worker attempt deadline to allow crash recovery
const NotificationLease = 60 * time.Second

// ErrLeaseLost means the claim has expired or another worker owns it
var ErrLeaseLost = errors.New("notification lease lost")

// Notification contains the JSON payload and the exact lease returned by a claim
type Notification struct {
	ID          string    `db:"id"`
	ReferenceID string    `db:"reference_id"`
	Kind        string    `db:"kind"`
	Payload     []byte    `db:"payload"`
	Attempts    int       `db:"attempts"`
	LockedBy    string    `db:"locked_by"`
	LockedAt    time.Time `db:"locked_at"`
	ExpiresAt   time.Time `db:"expires_at"`
	CreatedAt   time.Time `db:"created_at"`
	Reclaimed   bool      `db:"reclaimed"`
}

// leaseArgs returns $1-$4 for the lease check: a pending row still owned by the claim and not past the lease
func leaseArgs(n Notification, args ...any) []any {
	return append([]any{n.ID, n.LockedBy, n.LockedAt, NotificationLease.Seconds()}, args...)
}

// CreateNotification participates in the caller's transaction without querying the referenced entity
func (r *UserRepo) CreateNotification(ctx context.Context, n Notification) error {
	const q = `
INSERT INTO notification_outbox (id, reference_id, kind, payload, created_at, expires_at, next_attempt_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $5, $5)`
	// lib/pq encodes []byte as bytea, so jsonb needs the payload as text
	_, err := r.query().Exec(ctx, q, n.ID, n.ReferenceID, n.Kind, string(n.Payload), n.CreatedAt, n.ExpiresAt)
	if err != nil {
		return fmt.Errorf("enqueue notification: %w", err)
	}
	return nil
}

// ClaimNotification commits its short row lock before the provider is contacted
func (r *UserRepo) ClaimNotification(ctx context.Context, owner string) (Notification, error) {
	const q = `
WITH candidate AS (
	SELECT id, locked_at IS NOT NULL AS reclaimed
	FROM notification_outbox
	WHERE status = 'pending'
		AND next_attempt_at <= clock_timestamp()
		AND expires_at > clock_timestamp()
		AND (locked_at IS NULL OR locked_at <= clock_timestamp() - $2 * interval '1 second')
	ORDER BY next_attempt_at, created_at, id
	FOR UPDATE SKIP LOCKED
	LIMIT 1
)
UPDATE notification_outbox n
SET locked_by = $1, locked_at = clock_timestamp(), attempts = attempts + 1, updated_at = clock_timestamp()
FROM candidate c
WHERE n.id = c.id
RETURNING n.id, n.reference_id, n.kind, n.payload, n.attempts, n.locked_by, n.locked_at, n.expires_at, c.reclaimed`

	var n Notification
	err := r.db.GetContext(ctx, &n, q, owner, NotificationLease.Seconds())
	if errors.Is(err, sql.ErrNoRows) {
		return Notification{}, ErrNotFound
	}
	return n, err
}

// NotificationEligible reports whether a still-owned claim targets a pending code for an unverified, subscribed publisher at email.
// It runs just before sending, without holding locks over HTTP
func (r *UserRepo) NotificationEligible(ctx context.Context, n Notification, email string) (bool, error) {
	const q = `
SELECT EXISTS (
	SELECT 1
	FROM notification_outbox n
	JOIN email_verifications v ON v.id = CASE WHEN n.kind = $6 THEN n.reference_id::uuid END
	JOIN users u ON u.id = v.user_id
	WHERE n.id = $1 AND n.status = 'pending' AND n.locked_by = $2 AND n.locked_at = $3
		AND n.locked_at > clock_timestamp() - $4 * interval '1 second'
		AND n.expires_at > clock_timestamp()
		AND v.verified_at IS NULL
		AND v.verification_code IS NOT NULL
		AND NOT u.email_verified
		AND u.role = 'publisher'
		AND u.email = $5
		AND NOT EXISTS (
			SELECT 1
			FROM email_unsubscribes
			WHERE email = $5
		)
)`

	var eligible bool
	err := r.db.GetContext(ctx, &eligible, q, leaseArgs(n, email, notification.VerificationKind)...)
	return eligible, err
}

// CompleteNotification acknowledges delivery and records the provider message ID in one atomic statement
func (r *UserRepo) CompleteNotification(ctx context.Context, n Notification, messageID string) error {
	const q = `
WITH done AS (
	UPDATE notification_outbox n
	SET status = 'sent', payload = NULL, updated_at = clock_timestamp(),
		provider_message_id = $5::text, last_error = NULL, locked_by = NULL, locked_at = NULL
	WHERE n.id = $1 AND n.status = 'pending' AND n.locked_by = $2 AND n.locked_at = $3
		AND n.locked_at > clock_timestamp() - $4 * interval '1 second'
	RETURNING n.kind, n.reference_id
), verification AS (
	UPDATE email_verifications v
	SET message_id = $5::text
	FROM done
	WHERE v.id = CASE WHEN done.kind = $6 THEN done.reference_id::uuid END
)
SELECT count(*)
FROM done`
	var count int
	if err := r.query().Get(ctx, &count, q, leaseArgs(n, messageID, notification.VerificationKind)...); err != nil {
		return err
	}
	if count != 1 {
		return ErrLeaseLost
	}
	return nil
}

// FinishNotification clears the payload after a terminal delivery decision
func (r *UserRepo) FinishNotification(ctx context.Context, n Notification, status, reason string) error {
	if status != "cancelled" && status != "failed" && status != "expired" {
		return errors.New("invalid notification terminal status")
	}
	const q = `
UPDATE notification_outbox n
SET status = $5, payload = NULL, updated_at = clock_timestamp(), last_error = NULLIF($6, ''), locked_by = NULL, locked_at = NULL
WHERE n.id = $1 AND n.status = 'pending' AND n.locked_by = $2 AND n.locked_at = $3
	AND n.locked_at > clock_timestamp() - $4 * interval '1 second'`

	result, err := r.query().Exec(ctx, q, leaseArgs(n, status, reason)...)
	return checkLease(result, err)
}

// RetryNotification releases an owned lease and persists the next attempt time
func (r *UserRepo) RetryNotification(ctx context.Context, n Notification, delay time.Duration, reason string) error {
	const q = `
UPDATE notification_outbox n
SET next_attempt_at = clock_timestamp() + $5 * interval '1 second',
	last_error = $6, locked_by = NULL, locked_at = NULL, updated_at = clock_timestamp()
WHERE n.id = $1 AND n.status = 'pending' AND n.locked_by = $2 AND n.locked_at = $3
	AND n.locked_at > clock_timestamp() - $4 * interval '1 second'`

	result, err := r.query().Exec(ctx, q, leaseArgs(n, delay.Seconds(), reason)...)
	return checkLease(result, err)
}

func checkLease(result sql.Result, err error) error {
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if count != 1 {
		return ErrLeaseLost
	}
	return nil
}

// ExpireNotifications ignores retry eligibility, including long provider Retry-After delays
func (r *UserRepo) ExpireNotifications(ctx context.Context) error {
	const q = `
WITH expired AS (
	SELECT id
	FROM notification_outbox
	WHERE status = 'pending'
		AND expires_at <= clock_timestamp()
		AND (locked_at IS NULL OR locked_at <= clock_timestamp() - $1 * interval '1 second')
	ORDER BY expires_at
	FOR UPDATE SKIP LOCKED
	LIMIT 100
)
UPDATE notification_outbox
SET status = 'expired', payload = NULL, updated_at = clock_timestamp(),
	locked_by = NULL, locked_at = NULL, last_error = 'expired'
WHERE id IN (SELECT id FROM expired)`

	_, err := r.db.ExecContext(ctx, q, NotificationLease.Seconds())
	return err
}

// PurgeNotifications removes a bounded batch of old terminal metadata
func (r *UserRepo) PurgeNotifications(ctx context.Context) error {
	const q = `
DELETE FROM notification_outbox
WHERE id IN (
	SELECT id
	FROM notification_outbox
	WHERE status <> 'pending' AND updated_at < clock_timestamp() - interval '30 days'
	ORDER BY updated_at
	FOR UPDATE SKIP LOCKED
	LIMIT 1000
)`

	_, err := r.db.ExecContext(ctx, q)
	return err
}

// NotificationStats describes the shared queue, not one worker process
type NotificationStats struct {
	Pending   int64   `db:"pending"`
	OldestAge float64 `db:"oldest_age"`
}

// NotificationStats returns queue depth and age for monitoring
func (r *UserRepo) NotificationStats(ctx context.Context) (NotificationStats, error) {
	const q = `
SELECT count(*) AS pending,
	COALESCE(EXTRACT(EPOCH FROM clock_timestamp() - min(created_at)), 0)::double precision AS oldest_age
FROM notification_outbox
WHERE status = 'pending'`

	var stats NotificationStats
	err := r.db.GetContext(ctx, &stats, q)
	return stats, err
}
