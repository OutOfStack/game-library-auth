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

	const q = `INSERT INTO email_verifications
        (id, user_id, verification_code, unsubscribe_token, date_created)
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

	const q = `SELECT id, user_id, verification_code, message_id, date_created
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

// SetEmailVerificationMessageID sets the message_id for an email verification record
func (r *UserRepo) SetEmailVerificationMessageID(ctx context.Context, id string, messageID string) error {
	ctx, span := tracer.Start(ctx, "setEmailVerificationMessageID")
	defer span.End()

	const q = `UPDATE email_verifications SET message_id = $1 WHERE id = $2`

	_, err := r.query().Exec(ctx, q, messageID, id)
	if err != nil {
		return fmt.Errorf("set email verification message_id: %w", err)
	}

	return nil
}

// SetEmailVerificationUsed sets email verification as used by clearing code hash and optionally setting verified_at
func (r *UserRepo) SetEmailVerificationUsed(ctx context.Context, id string, verified bool) error {
	ctx, span := tracer.Start(ctx, "setEmailVerificationUsed")
	defer span.End()

	verifiedAt := sql.NullTime{
		Time:  time.Now(),
		Valid: verified,
	}

	const q = `UPDATE email_verifications 
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
