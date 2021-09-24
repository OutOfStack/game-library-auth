package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

var (
	// ErrNotFound is used when requested entity is not found
	ErrNotFound = errors.New("not found")
)

// Repo manages API for user access
type Repo struct {
	log *log.Logger
	db  *sqlx.DB
}

// New constructs a Repo
func New(log *log.Logger, db *sqlx.DB) Repo {
	return Repo{
		log: log,
		db:  db,
	}
}

// Create inserts a new user into a the database
func (r Repo) Create(ctx context.Context, nu NewUser, now time.Time) (Info, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(nu.Password), bcrypt.DefaultCost)
	if err != nil {
		return Info{}, fmt.Errorf("generation password hash: %w", err)
	}

	usr := Info{
		ID:           uuid.New(),
		Username:     nu.Username,
		PasswordHash: hash,
		RoleID:       nu.RoleID,
		DateCreated:  now.UTC(),
		DateUpdated:  now.UTC(),
	}

	const q = `INSERT INTO users
		(id, username, password_hash, role_id, date_created, date_updated)
		VALUES ($1, $2, $3, $4, $5, $6)`

	_, err = r.db.ExecContext(ctx, q, usr.ID, usr.Username, usr.PasswordHash, usr.RoleID, usr.DateCreated, usr.DateUpdated)
	if err != nil {
		return Info{}, fmt.Errorf("inserting user: %w", err)
	}

	return usr, nil
}

func (r Repo) GetByID(ctx context.Context, userID uuid.UUID) (Info, error) {
	const q = `SELECT id, username, password_hash, role_id, date_created, date_updated FROM users
		WHERE id = $1`

	var usr Info
	if err := r.db.GetContext(ctx, &usr, q, userID); err != nil {
		if err == sql.ErrNoRows {
			return Info{}, ErrNotFound
		}
		return Info{}, fmt.Errorf("selecting user %v: %w", userID, err)
	}

	return usr, nil
}
