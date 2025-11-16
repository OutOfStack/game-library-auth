package facade

import (
	"context"
	"time"

	"github.com/OutOfStack/game-library-auth/internal/auth"
	"github.com/OutOfStack/game-library-auth/internal/client/resendapi"
	"github.com/OutOfStack/game-library-auth/internal/database"
	"github.com/OutOfStack/game-library-auth/internal/model"
	"github.com/golang-jwt/jwt/v4"
	"go.uber.org/zap"
)

// Provider represents dependencies for facade layer
type Provider struct {
	log                       *zap.Logger
	userRepo                  UserRepo
	emailSender               EmailSender
	auth                      Auth
	unsubscribeTokenGenerator *auth.UnsubscribeTokenGenerator
}

// New creates a new facade provider
func New(log *zap.Logger, userRepo UserRepo, emailSender EmailSender, authService Auth, unsubscribeTokenGenerator *auth.UnsubscribeTokenGenerator) *Provider {
	return &Provider{
		log:                       log,
		userRepo:                  userRepo,
		emailSender:               emailSender,
		auth:                      authService,
		unsubscribeTokenGenerator: unsubscribeTokenGenerator,
	}
}

// Auth provides authentication methods
type Auth interface {
	GenerateToken(claims jwt.Claims) (string, error)
	GenerateRefreshToken() (string, time.Time, error)
	CreateUserClaims(user model.User) jwt.Claims
	ValidateToken(tokenStr string) (auth.Claims, error)
}

// UserRepo provides methods for working with user repo
type UserRepo interface {
	RunWithTx(ctx context.Context, f func(context.Context) error) error

	CreateUser(ctx context.Context, user database.User) error
	UpdateUser(ctx context.Context, user database.User) error
	DeleteUser(ctx context.Context, userID string) error
	GetUserByID(ctx context.Context, userID string) (database.User, error)
	GetUserByUsername(ctx context.Context, username string) (database.User, error)
	GetUserByEmail(ctx context.Context, email string) (database.User, error)
	GetUserByOAuth(ctx context.Context, provider string, oauthID string) (database.User, error)
	CheckUserExists(ctx context.Context, name string, role model.Role) (bool, error)
	SetUserEmailVerified(ctx context.Context, userID string) error

	CreateEmailVerification(ctx context.Context, verification database.EmailVerification) error
	GetEmailVerificationByUserID(ctx context.Context, userID string) (database.EmailVerification, error)
	SetEmailVerificationMessageID(ctx context.Context, verificationID string, messageID string) error
	SetEmailVerificationUsed(ctx context.Context, id string, verified bool) error
	SetUnsubscribeToken(ctx context.Context, id string, token string) error

	CreateEmailUnsubscribe(ctx context.Context, unsubscribe database.EmailUnsubscribe) error
	IsEmailUnsubscribed(ctx context.Context, email string) (bool, error)

	CreateRefreshToken(ctx context.Context, refreshToken database.RefreshToken) error
	GetRefreshTokenByHash(ctx context.Context, tokenHash string) (database.RefreshToken, error)
	DeleteRefreshToken(ctx context.Context, token string) error
	DeleteRefreshTokensByUserID(ctx context.Context, userID string) error
}

// EmailSender provides methods for sending emails
type EmailSender interface {
	SendEmailVerification(ctx context.Context, req resendapi.SendEmailVerificationRequest) (string, error)
}
