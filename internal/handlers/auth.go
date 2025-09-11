package handlers

import (
	"context"
	"errors"

	"github.com/OutOfStack/game-library-auth/internal/appconf"
	"github.com/OutOfStack/game-library-auth/internal/auth"
	"github.com/OutOfStack/game-library-auth/internal/client/mailersend"
	"github.com/OutOfStack/game-library-auth/internal/database"
	"github.com/golang-jwt/jwt/v4"
	"go.uber.org/zap"
	"google.golang.org/api/idtoken"
)

// UserRepo provides methods for working with user repo
type UserRepo interface {
	CreateUser(ctx context.Context, user database.User) error
	UpdateUser(ctx context.Context, user database.User) error
	DeleteUser(ctx context.Context, userID string) error
	GetUserByID(ctx context.Context, userID string) (database.User, error)
	GetUserByUsername(ctx context.Context, username string) (database.User, error)
	GetUserByEmail(ctx context.Context, email string) (database.User, error)
	GetUserByOAuth(ctx context.Context, provider string, oauthID string) (database.User, error)
	CheckUserExists(ctx context.Context, name string, role database.Role) (bool, error)
	SetUserEmailVerified(ctx context.Context, userID string) error

	CreateEmailVerification(ctx context.Context, verification database.EmailVerification) error
	GetEmailVerificationByUserID(ctx context.Context, userID string) (database.EmailVerification, error)
	SetEmailVerificationMessageID(ctx context.Context, verificationID string, messageID string) error
	SetEmailVerificationUsed(ctx context.Context, id string, verified bool) error
}

// Auth provides authentication methods
type Auth interface {
	GenerateToken(claims jwt.Claims) (string, error)
	ValidateToken(tokenStr string) (auth.Claims, error)
	CreateClaims(user database.User) jwt.Claims
}

// GoogleTokenValidator provides methods for validating Google ID tokens
type GoogleTokenValidator interface {
	Validate(ctx context.Context, idToken string, audience string) (*idtoken.Payload, error)
}

// EmailSender provides methods for sending emails
type EmailSender interface {
	SendEmailVerification(ctx context.Context, req mailersend.SendEmailVerificationRequest) (string, error)
}

// AuthAPI describes dependencies for auth endpoints
type AuthAPI struct {
	googleOAuthClientID  string
	baseURL              string
	log                  *zap.Logger
	auth                 Auth
	userRepo             UserRepo
	googleTokenValidator GoogleTokenValidator
	emailSender          EmailSender
	disableEmailSender   bool
}

// NewAuthAPI return new instance of auth api
func NewAuthAPI(log *zap.Logger, cfg *appconf.Cfg, auth Auth, userRepo UserRepo, googleTokenValidator GoogleTokenValidator, emailSender EmailSender) (*AuthAPI, error) {
	if cfg.Auth.GoogleClientID == "" {
		return nil, errors.New("google client id is empty")
	}

	if !cfg.EmailSender.EmailVerificationEnabled {
		log.Warn("email verification is disabled")
	}

	return &AuthAPI{
		googleOAuthClientID:  cfg.Auth.GoogleClientID,
		baseURL:              cfg.Auth.Issuer,
		auth:                 auth,
		userRepo:             userRepo,
		log:                  log,
		googleTokenValidator: googleTokenValidator,
		emailSender:          emailSender,
		disableEmailSender:   !cfg.EmailSender.EmailVerificationEnabled,
	}, nil
}
