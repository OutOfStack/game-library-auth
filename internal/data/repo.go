package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel"
)

// User role names
const (
	DefaultRoleName   string = "user"
	PublisherRoleName string = "publisher"
)

var (
	// ErrNotFound is used when requested entity is not found
	ErrNotFound = errors.New("not found")

	tracer = otel.Tracer("")
)

// Repo manages API for user access
type Repo struct {
	db *sqlx.DB
}

// NewRepo constructs a Repo
func NewRepo(db *sqlx.DB) Repo {
	return Repo{
		db: db,
	}
}

// CreateUser inserts a new user into the database
func (r *Repo) CreateUser(ctx context.Context, user User) error {
	ctx, span := tracer.Start(ctx, "db.createUser")
	defer span.End()

	const q = `INSERT INTO users
		(id, username, name, password_hash, role_id, avatar_url, date_created)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())`

	_, err := r.db.ExecContext(ctx, q, user.ID, user.Username, user.Name, user.PasswordHash, user.RoleID, user.AvatarURL)
	if err != nil {
		return fmt.Errorf("inserting user: %w", err)
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

	const q = `SELECT id, username, name, password_hash, role_id, avatar_url, date_created, date_updated
		FROM users
		WHERE id = $1`

	if err = r.db.GetContext(ctx, &user, q, userID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, ErrNotFound
		}
		return User{}, fmt.Errorf("selecting user %v: %w", userID, err)
	}

	return user, nil
}

// GetUserByUsername returns user by username
func (r *Repo) GetUserByUsername(ctx context.Context, username string) (user User, err error) {
	ctx, span := tracer.Start(ctx, "db.getUserByUsername")
	defer span.End()

	const q = `SELECT id, username, name, password_hash, role_id, avatar_url, date_created, date_updated
		FROM users
		WHERE username = $1`

	if err = r.db.GetContext(ctx, &user, q, username); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, ErrNotFound
		}
		return User{}, fmt.Errorf("selecting user %s: %w", username, err)
	}

	return user, nil
}

// CheckExistsPublisherWithName returns true if publisher with such name already exists otherwise returns false
func (r *Repo) CheckExistsPublisherWithName(ctx context.Context, name string) (bool, error) {
	ctx, span := tracer.Start(ctx, "db.checkExistsPublisherWithName")
	defer span.End()

	const q = `SELECT u.id
		FROM users u
		INNER JOIN roles r ON r.id = u.role_id
		WHERE u.name = $1 AND r.name = $2`

	var usr User
	if err := r.db.GetContext(ctx, &usr, q, name, PublisherRoleName); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("selecting publisher with name %s: %w", name, err)
	}

	return true, nil
}

// GetRoleByID returns role by id
func (r *Repo) GetRoleByID(ctx context.Context, roleID uuid.UUID) (role Role, err error) {
	ctx, span := tracer.Start(ctx, "db.getRoleByID")
	defer span.End()

	const q = `SELECT id, name, description, date_created, date_updated
		FROM roles
		WHERE id = $1`

	if err = r.db.GetContext(ctx, &role, q, roleID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Role{}, ErrNotFound
		}
		return Role{}, fmt.Errorf("selecting role %v: %w", roleID, err)
	}

	return role, nil
}

// GetRoleByName returns role by name
func (r *Repo) GetRoleByName(ctx context.Context, roleName string) (role Role, err error) {
	ctx, span := tracer.Start(ctx, "db.getRoleByName")
	defer span.End()

	const q = `SELECT id, name, description, date_created, date_updated
		FROM roles
		WHERE name = $1`

	if err = r.db.GetContext(ctx, &role, q, roleName); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Role{}, ErrNotFound
		}
		return Role{}, fmt.Errorf("selecting role %s: %w", roleName, err)
	}

	return role, nil
}
