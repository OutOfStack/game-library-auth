package user

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// Info represents an individual user
type Info struct {
	ID           uuid.UUID    `db:"id" json:"id"`
	Username     string       `db:"username" json:"username"`
	PasswordHash []byte       `db:"password_hash" json:"-"`
	RoleID       uuid.UUID    `db:"role_id" json:"role_id"`
	DateCreated  time.Time    `db:"date_created" json:"date_created"`
	DateUpdated  sql.NullTime `db:"date_updated" json:"date_updated"`
}

// GetUser represents user data to be returned to client
type GetUser struct {
	ID          uuid.UUID `json:"id"`
	Username    string    `json:"username"`
	RoleID      uuid.UUID `json:"role_id"`
	DateCreated string    `json:"date_created"`
	DateUpdated string    `json:"date_updated"`
}

// NewUser represents data for user creation
type NewUser struct {
	Username string
	Password string
	RoleID   uuid.UUID
}

// Role represents a user role
type Role struct {
	ID          uuid.UUID    `db:"id" json:"id"`
	Name        string       `db:"name" json:"name"`
	Description string       `db:"description" json:"description"`
	DateCreated time.Time    `db:"date_created" json:"date_created"`
	DateUpdated sql.NullTime `db:"date_updated" json:"date_updated"`
}
