package facade

import (
	"errors"
	"time"
)

const (
	maxUsernameLen = 32

	verificationCodeTTL            = 24 * time.Hour
	resendVerificationCodeCooldown = 60 * time.Second
)

// errors
var (
	InvalidEmailErr                 = errors.New("invalid email")
	OAutSignInConflictErr           = errors.New("oauth sign in name conflict")
	UpdateProfileUserNotFoundErr    = errors.New("update profile: user not found")
	UpdateProfileInvalidPasswordErr = errors.New("update profile: invalid current password")
	UpdateProfileNotAllowedErr      = errors.New("update profile: password change not allowed for oauth users")
	ErrTooManyRequests              = errors.New("too many requests")
	ResendVerificationNoEmailErr    = errors.New("resend verification: user has no email")
	VerifyEmailUserNotFoundErr      = errors.New("verify email: user not found")
	VerifyEmailAlreadyVerifiedErr   = errors.New("verify email: already verified")
	VerifyEmailInvalidOrExpiredErr  = errors.New("verify email: invalid or expired code")
	SignInInvalidCredentialsErr     = errors.New("sign in: invalid credentials")
	SignUpUsernameExistsErr         = errors.New("sign up: username already exists")
	SignUpPublisherNameExistsErr    = errors.New("sign up: publisher name already exists")
)
