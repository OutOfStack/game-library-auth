package model

// Role - user role
type Role string

// User role names
const (
	UserRoleName      Role = "user"
	PublisherRoleName Role = "publisher"
)

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

// IsPublisher checks if user is a publisher
func (u *User) IsPublisher() bool {
	return u.Role == string(PublisherRoleName)
}

// UpdateProfileParams contains parameters for updating user profile
type UpdateProfileParams struct {
	Name        *string
	Password    *string
	NewPassword *string
}
