package facade

import (
	"errors"
)

const (
	maxUsernameLen = 32

	defaultVrfCodeLen = 6
)

// errors
var (
	ErrInvalidEmail                 = errors.New("invalid email")
	ErrOAuthSignInConflict          = errors.New("oauth sign in name conflict")
	ErrUpdateProfileUserNotFound    = errors.New("update profile: user not found")
	ErrUpdateProfileInvalidPassword = errors.New("update profile: invalid current password")
	ErrUpdateProfileNotAllowed      = errors.New("update profile: password change not allowed for oauth users")
	ErrTooManyRequests              = errors.New("too many requests")
	ErrResendVerificationNoEmail    = errors.New("resend verification: user has no email")
	ErrVerifyEmailUserNotFound      = errors.New("verify email: user not found")
	ErrVerifyEmailAlreadyVerified   = errors.New("verify email: already verified")
	ErrVerifyEmailInvalidOrExpired  = errors.New("verify email: invalid or expired code")
	ErrSignInInvalidCredentials     = errors.New("sign in: invalid credentials")
	ErrSignUpUsernameExists         = errors.New("sign up: username already exists")
	ErrSignUpEmailExists            = errors.New("sign up: email already exists")
	ErrSignUpEmailRequired          = errors.New("sign up: email is required")
	ErrSignUpPublisherNameExists    = errors.New("sign up: publisher name already exists")
)

// emailVerificationResult - result of creating an email verification record
type emailVerificationResult struct {
	ID               string
	Code             string
	UnsubscribeToken string
}
