package facade

import (
	"errors"
	"time"
)

const (
	maxUsernameLen = 32

	defaultVrfCodeLen = 6

	verificationCodeTTL            = 24 * time.Hour
	resendVerificationCodeCooldown = 60 * time.Second
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
	ErrSignUpPublisherNameExists    = errors.New("sign up: publisher name already exists")
)
