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
	// ErrUsernameExists is used when username already exists
	ErrUsernameExists = errors.New("username already exists")
)

// User represents a user
type User struct {
	ID            uuid.UUID      `db:"id"`
	Username      string         `db:"username"`
	DisplayName   string         `db:"name"`
	PasswordHash  []byte         `db:"password_hash"`
	Role          Role           `db:"role"`
	OAuthProvider sql.NullString `db:"oauth_provider"`
	OAuthID       sql.NullString `db:"oauth_id"`
	DateCreated   time.Time      `db:"date_created"`
	DateUpdated   sql.NullTime   `db:"date_updated"`
}

// NewUser creates a new user
func NewUser(username, name string, passwordHash []byte, role Role) User {
	return User{
		ID:           uuid.New(),
		Username:     username,
		DisplayName:  name,
		PasswordHash: passwordHash,
		Role:         role,
	}
}

// SetOAuthID sets oauth provider and oauth id
func (u *User) SetOAuthID(provider string, oauthID string) {
	u.OAuthProvider = sql.NullString{String: provider, Valid: true}
	u.OAuthID = sql.NullString{String: oauthID, Valid: true}
}
