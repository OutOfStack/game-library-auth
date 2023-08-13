package user

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

// Create inserts a new user into the database
func (r *Repo) Create(ctx context.Context, usr *Info) (*Info, error) {
	ctx, span := tracer.Start(ctx, "sql.user.create")
	defer span.End()

	const q = `INSERT INTO users
		(id, username, name, password_hash, role_id, avatar_url, date_created)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())`

	_, err := r.db.ExecContext(ctx, q, usr.ID, usr.Username, usr.Name, usr.PasswordHash, usr.RoleID, usr.AvatarURL)
	if err != nil {
		return nil, fmt.Errorf("inserting user: %w", err)
	}

	return usr, nil
}

// Update updates user
func (r *Repo) Update(ctx context.Context, usr *Info) (*Info, error) {
	ctx, span := tracer.Start(ctx, "sql.user.update")
	defer span.End()

	const q = `UPDATE users
		SET name = COALESCE(NULLIF($2, ''), name),
		avatar_url = COALESCE(NULLIF($3, ''), avatar_url),
		password_hash = COALESCE(NULLIF($4, ''), password_hash),
		date_updated = NOW()
		WHERE id = $1`

	_, err := r.db.ExecContext(ctx, q, usr.ID, usr.Name, usr.AvatarURL, usr.PasswordHash)
	if err != nil {
		return nil, fmt.Errorf("update user: %w", err)
	}

	return usr, nil
}

// GetByID returns user by id
func (r *Repo) GetByID(ctx context.Context, userID string) (*Info, error) {
	ctx, span := tracer.Start(ctx, "sql.user.getbyid")
	defer span.End()

	const q = `SELECT id, username, name, password_hash, role_id, avatar_url, date_created, date_updated
		FROM users
		WHERE id = $1`

	var usr Info
	if err := r.db.GetContext(ctx, &usr, q, userID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("selecting user %v: %w", userID, err)
	}

	return &usr, nil
}

// GetByUsername returns user by username
func (r *Repo) GetByUsername(ctx context.Context, username string) (*Info, error) {
	ctx, span := tracer.Start(ctx, "sql.user.getbyusername")
	defer span.End()

	const q = `SELECT id, username, name, password_hash, role_id, avatar_url, date_created, date_updated
		FROM users
		WHERE username = $1`

	var usr Info
	if err := r.db.GetContext(ctx, &usr, q, username); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("selecting user %s: %w", username, err)
	}

	return &usr, nil
}

// CheckExistPublisherWithName returns true if publisher with such name already exists otherwise returns false
func (r *Repo) CheckExistPublisherWithName(ctx context.Context, name string) (bool, error) {
	ctx, span := tracer.Start(ctx, "sql.user.checkexistpublisher")
	defer span.End()

	const q = `SELECT u.id
		FROM users u
		INNER JOIN roles r ON r.id = u.role_id
		WHERE u.name = $1 AND r.name = $2`

	var usr Info
	if err := r.db.GetContext(ctx, &usr, q, name, PublisherRoleName); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("selecting publisher with name %s: %w", name, err)
	}

	return true, nil
}

// GetRoleByID returns role by id
func (r *Repo) GetRoleByID(ctx context.Context, roleID uuid.UUID) (*Role, error) {
	ctx, span := tracer.Start(ctx, "sql.role.getbyid")
	defer span.End()

	const q = `SELECT id, name, description, date_created, date_updated
		FROM roles
		WHERE id = $1`

	var role Role
	if err := r.db.GetContext(ctx, &role, q, roleID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("selecting role %v: %w", roleID, err)
	}

	return &role, nil
}

// GetRoleByName returns role by name
func (r *Repo) GetRoleByName(ctx context.Context, roleName string) (*Role, error) {
	ctx, span := tracer.Start(ctx, "sql.role.getbyname")
	defer span.End()

	const q = `SELECT id, name, description, date_created, date_updated
		FROM roles
		WHERE name = $1`

	var role Role
	if err := r.db.GetContext(ctx, &role, q, roleName); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("selecting role %s: %w", roleName, err)
	}

	return &role, nil
}
