package database

import (
	"database/sql"
	"errors"
	"time"

	"github.com/OutOfStack/game-library-auth/internal/model"
	"github.com/google/uuid"
)

const (
	pgUniqueViolationCode = "23505"
)

var (
	// ErrNotFound is used when a record is not found
	ErrNotFound = errors.New("not found")
	// ErrUserExists is used when username/email already exists
	ErrUserExists = errors.New("user already exists")
)

// User represents a user
type User struct {
	ID            string         `db:"id"`
	Username      string         `db:"username"`
	DisplayName   string         `db:"name"`
	Email         sql.NullString `db:"email"`
	EmailVerified bool           `db:"email_verified"`
	PasswordHash  []byte         `db:"password_hash"`
	Role          model.Role     `db:"role"`
	OAuthProvider sql.NullString `db:"oauth_provider"`
	OAuthID       sql.NullString `db:"oauth_id"`
	DateCreated   time.Time      `db:"date_created"`
	DateUpdated   sql.NullTime   `db:"date_updated"`
}

// NewUser creates a new user
func NewUser(username, name string, passwordHash []byte, role model.Role) User {
	return User{
		ID:           uuid.New().String(),
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
	ID               string         `db:"id"`
	UserID           string         `db:"user_id"`
	CodeHash         sql.NullString `db:"verification_code"`
	MessageID        sql.NullString `db:"message_id"`
	UnsubscribeToken sql.NullString `db:"unsubscribe_token"`
	DateCreated      time.Time      `db:"date_created"`
}

// NewEmailVerification creates a new email verification record
func NewEmailVerification(userID, codeHash, unsubscribeToken string, createdAt time.Time) EmailVerification {
	return EmailVerification{
		ID:               uuid.New().String(),
		UserID:           userID,
		CodeHash:         sql.NullString{String: codeHash, Valid: codeHash != ""},
		UnsubscribeToken: sql.NullString{String: unsubscribeToken, Valid: unsubscribeToken != ""},
		DateCreated:      createdAt,
	}
}

// IsExpired checks if the verification code has expired
func (ev *EmailVerification) IsExpired() bool {
	return time.Now().After(ev.DateCreated.Add(model.VerificationCodeTTL))
}

// EmailUnsubscribe represents an email unsubscribe record
type EmailUnsubscribe struct {
	ID    string `db:"id"`
	Email string `db:"email"`
}

// NewEmailUnsubscribe creates a new email unsubscribe record
func NewEmailUnsubscribe(email string) EmailUnsubscribe {
	return EmailUnsubscribe{
		ID:    uuid.New().String(),
		Email: email,
	}
}

// RefreshToken represents a refresh token
type RefreshToken struct {
	ID          string    `db:"id"`
	UserID      string    `db:"user_id"`
	Token       string    `db:"token"`
	ExpiresAt   time.Time `db:"expires_at"`
	DateCreated time.Time `db:"date_created"`
}

// NewRefreshToken creates a new refresh token
func NewRefreshToken(userID, token string, expiresAt time.Time) RefreshToken {
	return RefreshToken{
		ID:        uuid.New().String(),
		UserID:    userID,
		Token:     token,
		ExpiresAt: expiresAt,
	}
}

// IsExpired checks if the refresh token has expired
func (rt *RefreshToken) IsExpired() bool {
	return !time.Now().Before(rt.ExpiresAt)
}
