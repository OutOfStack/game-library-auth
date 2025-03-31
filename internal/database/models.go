package database

import (
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
)

// Role - user role
type Role string

// User role names
const (
	UserRoleName      Role = "user"
	PublisherRoleName Role = "publisher"
)

var (
	// ErrNotFound is used when requested entity is not found
	ErrNotFound = errors.New("not found")
)

// User represents a user
type User struct {
	ID           uuid.UUID      `db:"id"`
	Username     string         `db:"username"`
	Name         string         `db:"name"`
	PasswordHash []byte         `db:"password_hash"`
	Role         Role           `db:"role"`
	AvatarURL    sql.NullString `db:"avatar_url"`
	DateCreated  time.Time      `db:"date_created"`
	DateUpdated  sql.NullTime   `db:"date_updated"`
}
