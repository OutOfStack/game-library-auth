package model

// User represents a user
type User struct {
	ID            string
	Username      string
	DisplayName   string
	Email         string
	EmailVerified bool
	Role          string
	OAuthProvider string
	OAuthID       string
}

// UpdateProfileParams contains parameters for updating user profile
type UpdateProfileParams struct {
	Name        *string
	Password    *string
	NewPassword *string
}
