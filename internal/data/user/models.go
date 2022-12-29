package user

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// Info represents a user
type Info struct {
	ID           uuid.UUID      `db:"id"`
	Username     string         `db:"username"`
	Name         string         `db:"name"`
	PasswordHash []byte         `db:"password_hash"`
	RoleID       uuid.UUID      `db:"role_id"`
	AvatarURL    sql.NullString `db:"avatar_url"`
	DateCreated  time.Time      `db:"date_created"`
	DateUpdated  sql.NullTime   `db:"date_updated"`
}

// Role represents a user role
type Role struct {
	ID          uuid.UUID    `db:"id"`
	Name        string       `db:"name"`
	Description string       `db:"description"`
	DateCreated time.Time    `db:"date_created"`
	DateUpdated sql.NullTime `db:"date_updated"`
}
