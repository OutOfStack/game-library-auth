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

const (
	pgUniqueViolationCode = "23505"
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
	Email         sql.NullString `db:"email"`
	EmailVerified bool           `db:"email_verified"`
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

// SetEmail sets user email and verification status
func (u *User) SetEmail(email string, verified bool) {
	u.Email = sql.NullString{String: email, Valid: email != ""}
	u.EmailVerified = u.Email.Valid && verified
}

// EmailVerification represents an email verification record
type EmailVerification struct {
	ID          uuid.UUID      `db:"id"`
	UserID      uuid.UUID      `db:"user_id"`
	Email       string         `db:"email"`
	CodeHash    sql.NullString `db:"verification_code"`
	ExpiresAt   time.Time      `db:"expires_at"`
	MessageID   sql.NullString `db:"message_id"`
	DateCreated time.Time      `db:"date_created"`
}

// NewEmailVerification creates a new email verification record
func NewEmailVerification(userID uuid.UUID, email, codeHash string, expiresAt time.Time) EmailVerification {
	return EmailVerification{
		ID:        uuid.New(),
		UserID:    userID,
		Email:     email,
		CodeHash:  sql.NullString{String: codeHash, Valid: codeHash != ""},
		ExpiresAt: expiresAt,
	}
}

// IsExpired checks if the verification code has expired
func (ev *EmailVerification) IsExpired() bool {
	return time.Now().After(ev.ExpiresAt)
}
