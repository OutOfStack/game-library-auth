package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel"
)

var (
	tracer = otel.Tracer("")
)

// Repo manages API for user access
type Repo struct {
	db *sqlx.DB
}

// NewRepo constructs a Repo
func NewRepo(db *sqlx.DB) *Repo {
	return &Repo{
		db: db,
	}
}

// CreateUser inserts a new user into the database
func (r *Repo) CreateUser(ctx context.Context, user User) error {
	ctx, span := tracer.Start(ctx, "db.createUser")
	defer span.End()

	const q = `INSERT INTO users
		(id, username, name, password_hash, role, avatar_url, date_created)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())`

	_, err := r.db.ExecContext(ctx, q, user.ID, user.Username, user.Name, user.PasswordHash, user.Role, user.AvatarURL)
	if err != nil {
		return fmt.Errorf("insert user: %w", err)
	}

	return nil
}

// UpdateUser updates user
func (r *Repo) UpdateUser(ctx context.Context, user User) error {
	ctx, span := tracer.Start(ctx, "db.updateUser")
	defer span.End()

	const q = `UPDATE users
		SET name = COALESCE(NULLIF($2, ''), name),
		avatar_url = COALESCE(NULLIF($3, ''), avatar_url),
		password_hash = COALESCE(NULLIF($4, ''), password_hash),
		date_updated = NOW()
		WHERE id = $1`

	_, err := r.db.ExecContext(ctx, q, user.ID, user.Name, user.AvatarURL, user.PasswordHash)
	if err != nil {
		return fmt.Errorf("update user: %w", err)
	}

	return nil
}

// GetUserByID returns user by id
func (r *Repo) GetUserByID(ctx context.Context, userID string) (user User, err error) {
	ctx, span := tracer.Start(ctx, "db.getUserByID")
	defer span.End()

	const q = `SELECT id, username, name, password_hash, role, avatar_url, date_created, date_updated
		FROM users
		WHERE id = $1`

	if err = r.db.GetContext(ctx, &user, q, userID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, ErrNotFound
		}
		return User{}, fmt.Errorf("select user %v: %w", userID, err)
	}

	return user, nil
}

// GetUserByUsername returns user by username
func (r *Repo) GetUserByUsername(ctx context.Context, username string) (user User, err error) {
	ctx, span := tracer.Start(ctx, "db.getUserByUsername")
	defer span.End()

	const q = `SELECT id, username, name, password_hash, role, avatar_url, date_created, date_updated
		FROM users
		WHERE username = $1`

	if err = r.db.GetContext(ctx, &user, q, username); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, ErrNotFound
		}
		return User{}, fmt.Errorf("select user %s: %w", username, err)
	}

	return user, nil
}

// CheckUserExists checks whether user with provided name and role exists
func (r *Repo) CheckUserExists(ctx context.Context, name string, role Role) (bool, error) {
	ctx, span := tracer.Start(ctx, "db.checkUserExists")
	defer span.End()

	const q = `SELECT EXISTS(
		SELECT id
		FROM users
		WHERE name = $1 AND role = $2)`

	var exists bool
	if err := r.db.GetContext(ctx, &exists, q, name, role); err != nil {
		return false, fmt.Errorf("select publisher with name %s: %w", name, err)
	}

	return exists, nil
}
