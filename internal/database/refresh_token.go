package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// CreateRefreshToken inserts a new refresh token into the database
func (r *UserRepo) CreateRefreshToken(ctx context.Context, refreshToken RefreshToken) error {
	ctx, span := tracer.Start(ctx, "createRefreshToken")
	defer span.End()

	const q = `INSERT INTO refresh_tokens
		(id, user_id, token_hash, expires_at, date_created)
		VALUES ($1, $2, $3, $4, NOW())`

	_, err := r.query().Exec(ctx, q, refreshToken.ID, refreshToken.UserID, refreshToken.TokenHash, refreshToken.ExpiresAt)
	if err != nil {
		return fmt.Errorf("insert refresh token: %w", err)
	}

	return nil
}

// GetRefreshTokenByHash retrieves a refresh token by token hash string with row-level lock
func (r *UserRepo) GetRefreshTokenByHash(ctx context.Context, tokenHash string) (RefreshToken, error) {
	ctx, span := tracer.Start(ctx, "getRefreshTokenByHash")
	defer span.End()

	const q = `SELECT id, user_id, token_hash, expires_at, date_created
		FROM refresh_tokens
		WHERE token_hash = $1
		FOR UPDATE`

	var refreshToken RefreshToken
	if err := r.query().Get(ctx, &refreshToken, q, tokenHash); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return RefreshToken{}, ErrNotFound
		}
		return RefreshToken{}, fmt.Errorf("get refresh token by token: %w", err)
	}

	return refreshToken, nil
}

// DeleteRefreshToken deletes a refresh token by token hash string
func (r *UserRepo) DeleteRefreshToken(ctx context.Context, tokenHash string) error {
	ctx, span := tracer.Start(ctx, "deleteRefreshToken")
	defer span.End()

	const q = `DELETE FROM refresh_tokens WHERE token_hash = $1`

	_, err := r.query().Exec(ctx, q, tokenHash)
	if err != nil {
		return fmt.Errorf("delete refresh token: %w", err)
	}

	return nil
}

// DeleteRefreshTokensByUserID deletes all refresh tokens for a user
func (r *UserRepo) DeleteRefreshTokensByUserID(ctx context.Context, userID string) error {
	ctx, span := tracer.Start(ctx, "deleteRefreshTokensByUserID")
	defer span.End()

	const q = `DELETE FROM refresh_tokens WHERE user_id = $1`

	_, err := r.query().Exec(ctx, q, userID)
	if err != nil {
		return fmt.Errorf("delete refresh tokens by user id: %w", err)
	}

	return nil
}

// DeleteExpiredRefreshTokens deletes all expired refresh tokens
func (r *UserRepo) DeleteExpiredRefreshTokens(ctx context.Context) error {
	ctx, span := tracer.Start(ctx, "deleteExpiredRefreshTokens")
	defer span.End()

	const q = `DELETE FROM refresh_tokens WHERE expires_at < $1`

	_, err := r.query().Exec(ctx, q, time.Now())
	if err != nil {
		return fmt.Errorf("delete expired refresh tokens: %w", err)
	}

	return nil
}
