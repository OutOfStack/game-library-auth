package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// CreateEmailVerification creates a new email verification record
func (r *UserRepo) CreateEmailVerification(ctx context.Context, vrf EmailVerification) error {
	ctx, span := tracer.Start(ctx, "createEmailVerification")
	defer span.End()

	const q = `
INSERT INTO email_verifications (id, user_id, verification_code, unsubscribe_token, date_created)
VALUES ($1, $2, $3, $4, $5)`

	_, err := r.query().Exec(ctx, q, vrf.ID, vrf.UserID, vrf.CodeHash, vrf.UnsubscribeToken, vrf.DateCreated)
	if err != nil {
		return fmt.Errorf("insert email verification: %w", err)
	}

	return nil
}

// GetEmailVerificationByUserID gets email verification by user ID (most recent unused)
func (r *UserRepo) GetEmailVerificationByUserID(ctx context.Context, userID string) (EmailVerification, error) {
	ctx, span := tracer.Start(ctx, "getEmailVerificationByUserID")
	defer span.End()

	const q = `
SELECT id, user_id, verification_code, message_id, date_created
FROM email_verifications
WHERE user_id = $1 AND verified_at IS NULL AND verification_code IS NOT NULL
ORDER BY date_created DESC
LIMIT 1
FOR NO KEY UPDATE`

	var verification EmailVerification
	if err := r.query().Get(ctx, &verification, q, userID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return EmailVerification{}, ErrNotFound
		}
		return EmailVerification{}, fmt.Errorf("select email verification: %w", err)
	}
	return verification, nil
}

// VerificationState holds what queueing a verification code depends on
type VerificationState struct {
	Now           time.Time      `db:"now"`
	Unsubscribed  bool           `db:"unsubscribed"`
	LatestID      sql.NullString `db:"latest_id"`
	LatestCreated sql.NullTime   `db:"latest_created"`
}

// GetVerificationState returns the database time, whether email is unsubscribed and the user's latest unused code, locking that code.
// Call it after locking the user row so the latest code is read from a fresh snapshot
func (r *UserRepo) GetVerificationState(ctx context.Context, userID, email string) (VerificationState, error) {
	ctx, span := tracer.Start(ctx, "getVerificationState")
	defer span.End()

	const q = `
SELECT clock_timestamp() AS now,
	EXISTS (
		SELECT 1
		FROM email_unsubscribes
		WHERE email = $2
	) AS unsubscribed,
	v.id AS latest_id,
	v.date_created AS latest_created
FROM (SELECT 1) AS one
LEFT JOIN (
	SELECT id, date_created
	FROM email_verifications
	WHERE user_id = $1 AND verified_at IS NULL AND verification_code IS NOT NULL
	ORDER BY date_created DESC
	LIMIT 1
	FOR NO KEY UPDATE
) v ON true`

	var state VerificationState
	if err := r.query().Get(ctx, &state, q, userID, email); err != nil {
		return VerificationState{}, fmt.Errorf("select verification state: %w", err)
	}
	return state, nil
}

// SetEmailVerificationUsed sets email verification as used by clearing code hash and optionally setting verified_at
func (r *UserRepo) SetEmailVerificationUsed(ctx context.Context, id string, verified bool) error {
	ctx, span := tracer.Start(ctx, "setEmailVerificationUsed")
	defer span.End()

	verifiedAt := sql.NullTime{
		Time:  time.Now(),
		Valid: verified,
	}

	const q = `
UPDATE email_verifications
SET verification_code = NULL,
	unsubscribe_token = NULL,
	verified_at = $2
WHERE id = $1`

	_, err := r.query().Exec(ctx, q, id, verifiedAt)
	if err != nil {
		return fmt.Errorf("set email verification as used: %w", err)
	}

	return nil
}
