package user

import (
	"time"

	"github.com/google/uuid"
)

// Info represents an individual user
type Info struct {
	ID           uuid.UUID `db:"id" json:"id"`
	Username     string    `db:"username" json:"username"`
	PasswordHash []byte    `db:"password_hash" json:"-"`
	RoleID       uuid.UUID `db:"role_id" json:"role_id"`
	DateCreated  time.Time `db:"date_created" json:"date_created"`
	DateUpdated  time.Time `db:"date_updated" json:"date_updated"`
}

// SignUp represets data for user sign up
type SignUp struct {
	Username        string `json:"id" validate:"required"`
	Password        string `json:"password" validate:"required"`
	ConfirmPassword string `json:"confirm_password" validate:"eqfield=Password"`
}

// NewUser represents data for user creation
type NewUser struct {
	Username string
	Password string
	RoleID   uuid.UUID
}

// SignIn represets data for user sign in
type SignIn struct {
	Username string `json:"id" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// Role represents a user role
type Role struct {
	ID          uuid.UUID `db:"id" json:"id"`
	Name        string    `db:"name" json:"name"`
	Description string    `db:"description" json:"description"`
	DateCreated time.Time `db:"date_created" json:"date_created"`
	DateUpdated time.Time `db:"date_updated" json:"date_updated"`
}
