package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// User role names
const (
	DefaultRoleName   string = "user"
	PublisherRoleName string = "publisher"
)

var (
	// ErrNotFound is used when requested entity is not found
	ErrNotFound = errors.New("not found")
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
	const q = `INSERT INTO users
		(id, username, name, password_hash, role_id, date_created, date_updated)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := r.db.ExecContext(ctx, q, usr.ID, usr.Username, usr.Name, usr.PasswordHash, usr.RoleID, usr.DateCreated, usr.DateUpdated)
	if err != nil {
		return nil, fmt.Errorf("inserting user: %w", err)
	}

	return usr, nil
}

// GetByID returns user by id
func (r *Repo) GetByID(ctx context.Context, userID uuid.UUID) (*Info, error) {
	const q = `SELECT id, username, name, role_id, date_created, date_updated FROM users
		WHERE id = $1`

	var usr Info
	if err := r.db.GetContext(ctx, &usr, q, userID); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("selecting user %v: %w", userID, err)
	}

	return &usr, nil
}

// GetByUsername returns user by username
func (r *Repo) GetByUsername(ctx context.Context, username string) (*Info, error) {
	const q = `SELECT id, username, name, password_hash, role_id, date_created, date_updated FROM users
		WHERE username = $1`

	var usr Info
	if err := r.db.GetContext(ctx, &usr, q, username); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("selecting user %s: %w", username, err)
	}

	return &usr, nil
}

// CheckExistPublisherWithName returns true if publisher with such name already exists otherwise returns false
func (r *Repo) CheckExistPublisherWithName(ctx context.Context, name string) (bool, error) {
	const q = `SELECT u.id 
		FROM users u
		INNER JOIN roles r ON r.id = u.role_id
		WHERE u.name = $1 AND r.name = $2`

	var usr Info
	if err := r.db.GetContext(ctx, &usr, q, name, PublisherRoleName); err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, fmt.Errorf("selecting publisher with name %s: %w", name, err)
	}

	return true, nil
}

// GetRoleByID returns role by id
func (r *Repo) GetRoleByID(ctx context.Context, roleID uuid.UUID) (*Role, error) {
	const q = `SELECT id, name, description, date_created, date_updated FROM roles
		WHERE id = $1`

	var role Role
	if err := r.db.GetContext(ctx, &role, q, roleID); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("selecting role %v: %w", roleID, err)
	}

	return &role, nil
}

// GetRoleByName returns role by name
func (r *Repo) GetRoleByName(ctx context.Context, roleName string) (*Role, error) {
	const q = `SELECT id, name, description, date_created, date_updated FROM roles
		WHERE name = $1`

	var role Role
	if err := r.db.GetContext(ctx, &role, q, roleName); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("selecting role %s: %w", roleName, err)
	}

	return &role, nil
}
