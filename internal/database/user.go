package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/OutOfStack/game-library-auth/pkg/database"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
)

var (
	tracer = otel.Tracer("db")
)

// UserRepo manages API for user access
type UserRepo struct {
	db  *sqlx.DB
	log *zap.Logger
}

// NewUserRepo constructs a user Repo
func NewUserRepo(db *sqlx.DB, log *zap.Logger) *UserRepo {
	return &UserRepo{
		db:  db,
		log: log,
	}
}

// RunWithTx runs a function in a transaction
func (r *UserRepo) RunWithTx(ctx context.Context, f func(context.Context) error) error {
	// check if we're already in a transaction
	if _, ok := database.TxFromContext(ctx); ok {
		// just exec func
		return f(ctx)
	}

	// begin new tx
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}

	txCtx := database.WithTx(ctx, tx)
	err = f(txCtx)
	if err != nil {
		txErr := tx.Rollback()
		if txErr != nil {
			r.log.Error("rolling back tx", zap.Error(txErr))
			return err
		}
		return err
	}

	return tx.Commit()
}

// query returns the appropriate executor - transaction if present in context, otherwise the database
func (r *UserRepo) query() database.Querier {
	return database.NewQuerier(r.db)
}

// CreateUser inserts a new user into the database
func (r *UserRepo) CreateUser(ctx context.Context, user User) error {
	ctx, span := tracer.Start(ctx, "createUser")
	defer span.End()

	const q = `INSERT INTO users
        (id, username, name, email, email_verified, password_hash, role, oauth_provider, oauth_id, date_created)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW())`

	_, err := r.query().Exec(ctx, q, user.ID, user.Username, user.DisplayName, user.Email, user.EmailVerified, user.PasswordHash, user.Role, user.OAuthProvider, user.OAuthID)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == pgUniqueViolationCode {
			return ErrUserExists
		}
		return fmt.Errorf("insert user: %w", err)
	}

	return nil
}

// UpdateUser updates user
func (r *UserRepo) UpdateUser(ctx context.Context, user User) error {
	ctx, span := tracer.Start(ctx, "updateUser")
	defer span.End()

	const q = `UPDATE users
		SET name = COALESCE(NULLIF($2, ''), name),
		password_hash = COALESCE(NULLIF($3, ''), password_hash),
		date_updated = NOW()
		WHERE id = $1`

	_, err := r.query().Exec(ctx, q, user.ID, user.DisplayName, user.PasswordHash)
	if err != nil {
		return fmt.Errorf("update user: %w", err)
	}

	return nil
}

// GetUserByID returns user by id
func (r *UserRepo) GetUserByID(ctx context.Context, userID string) (user User, err error) {
	ctx, span := tracer.Start(ctx, "getUserByID")
	defer span.End()

	const q = `SELECT id, username, name, email, email_verified, password_hash, role, oauth_provider, oauth_id, date_created, date_updated
		FROM users
		WHERE id = $1
		FOR NO KEY UPDATE`

	if err = r.query().Get(ctx, &user, q, userID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, ErrNotFound
		}
		return User{}, fmt.Errorf("select user by id: %w", err)
	}

	return user, nil
}

// GetUserByUsername returns user by username
func (r *UserRepo) GetUserByUsername(ctx context.Context, username string) (user User, err error) {
	ctx, span := tracer.Start(ctx, "getUserByUsername")
	defer span.End()

	const q = `SELECT id, username, name, email, email_verified, password_hash, role, oauth_provider, oauth_id, date_created, date_updated
		FROM users
		WHERE username = $1`

	if err = r.query().Get(ctx, &user, q, username); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, ErrNotFound
		}
		return User{}, fmt.Errorf("select user by username: %w", err)
	}

	return user, nil
}

// CheckUserExists checks whether user with provided name and role exists
func (r *UserRepo) CheckUserExists(ctx context.Context, name string, role Role) (bool, error) {
	ctx, span := tracer.Start(ctx, "checkUserExists")
	defer span.End()

	const q = `SELECT EXISTS(
		SELECT id
		FROM users
		WHERE name = $1 AND role = $2)`

	var exists bool
	if err := r.query().Get(ctx, &exists, q, name, role); err != nil {
		return false, fmt.Errorf("check user exists: %w", err)
	}

	return exists, nil
}

// GetUserByOAuth returns a user by oauth provider and oauth_id
func (r *UserRepo) GetUserByOAuth(ctx context.Context, provider string, oauthID string) (User, error) {
	ctx, span := tracer.Start(ctx, "getUserByOAuth")
	defer span.End()

	const q = `SELECT id, username, name, email, email_verified, role, oauth_provider, oauth_id, date_created, date_updated
        FROM users
        WHERE oauth_provider = $1 AND oauth_id = $2`

	var user User
	if err := r.query().Get(ctx, &user, q, provider, oauthID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, ErrNotFound
		}
		return User{}, fmt.Errorf("select user (oauth): %w", err)
	}
	return user, nil
}

// DeleteUser deletes a user by user id
func (r *UserRepo) DeleteUser(ctx context.Context, userID string) error {
	ctx, span := tracer.Start(ctx, "deleteUser")
	defer span.End()

	const q = `DELETE FROM users WHERE id = $1`

	_, err := r.query().Exec(ctx, q, userID)
	if err != nil {
		return fmt.Errorf("delete user: %w", err)
	}

	return nil
}

// GetUserByEmail gets user by email address
func (r *UserRepo) GetUserByEmail(ctx context.Context, email string) (User, error) {
	ctx, span := tracer.Start(ctx, "getUserByEmail")
	defer span.End()

	const q = `SELECT id, username, name, email, email_verified, password_hash, role, oauth_provider, oauth_id, date_created, date_updated
		FROM users
		WHERE email = $1`

	var user User
	if err := r.query().Get(ctx, &user, q, email); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, ErrNotFound
		}
		return User{}, fmt.Errorf("select user by email: %w", err)
	}
	return user, nil
}

// SetUserEmailVerified sets user email as verified
func (r *UserRepo) SetUserEmailVerified(ctx context.Context, userID string) error {
	ctx, span := tracer.Start(ctx, "setUserVerified")
	defer span.End()

	const q = `UPDATE users
        SET email_verified = TRUE, date_updated = NOW()
        WHERE id = $1`

	_, err := r.query().Exec(ctx, q, userID)
	if err != nil {
		return fmt.Errorf("set user email verified: %w", err)
	}

	return nil
}
