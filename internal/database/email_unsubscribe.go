package database

import (
	"context"
	"fmt"
)

// CreateEmailUnsubscribe creates a new email unsubscribe record
func (r *UserRepo) CreateEmailUnsubscribe(ctx context.Context, unsubscribe EmailUnsubscribe) error {
	ctx, span := tracer.Start(ctx, "createEmailUnsubscribe")
	defer span.End()

	const q = `INSERT INTO email_unsubscribes (id, email, date_created)
        VALUES ($1, $2, NOW())
        ON CONFLICT (email) DO NOTHING`

	_, err := r.query().Exec(ctx, q, unsubscribe.ID, unsubscribe.Email)
	if err != nil {
		return fmt.Errorf("insert email unsubscribe: %w", err)
	}

	return nil
}

// IsEmailUnsubscribed checks if an email address is unsubscribed
func (r *UserRepo) IsEmailUnsubscribed(ctx context.Context, email string) (bool, error) {
	ctx, span := tracer.Start(ctx, "isEmailUnsubscribed")
	defer span.End()

	const q = `SELECT EXISTS(SELECT 1 FROM email_unsubscribes WHERE email = $1)`

	var exists bool
	if err := r.query().Get(ctx, &exists, q, email); err != nil {
		return false, fmt.Errorf("check email unsubscribed: %w", err)
	}

	return exists, nil
}

// SetUnsubscribeToken sets the unsubscribe token for an email verification record
func (r *UserRepo) SetUnsubscribeToken(ctx context.Context, id string, token string) error {
	ctx, span := tracer.Start(ctx, "setUnsubscribeToken")
	defer span.End()

	const q = `UPDATE email_verifications SET unsubscribe_token = $1 WHERE id = $2`

	_, err := r.query().Exec(ctx, q, token, id)
	if err != nil {
		return fmt.Errorf("set unsubscribe token: %w", err)
	}

	return nil
}
