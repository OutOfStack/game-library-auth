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
	Name         string       `db:"name" json:"name"`
	PasswordHash []byte       `db:"password_hash" json:"-"`
	RoleID       uuid.UUID    `db:"role_id" json:"roleId"`
	DateCreated  time.Time    `db:"date_created" json:"dateCreated"`
	DateUpdated  sql.NullTime `db:"date_updated" json:"dateUpdated"`
}

// GetUser represents user data to be returned to client
type GetUser struct {
	ID          uuid.UUID `json:"id"`
	Username    string    `json:"username"`
	Name        string    `json:"name"`
	RoleID      uuid.UUID `json:"roleId"`
	DateCreated string    `json:"dateCreated"`
	DateUpdated string    `json:"dateUpdated"`
}

// Role represents a user role
type Role struct {
	ID          uuid.UUID    `db:"id" json:"id"`
	Name        string       `db:"name" json:"name"`
	Description string       `db:"description" json:"description"`
	DateCreated time.Time    `db:"date_created" json:"dateCreated"`
	DateUpdated sql.NullTime `db:"date_updated" json:"dateUpdated"`
}
