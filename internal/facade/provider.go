package facade

import (
	"context"
	"time"

	"github.com/OutOfStack/game-library-auth/internal/auth"
	"github.com/OutOfStack/game-library-auth/internal/client/resendapi"
	"github.com/OutOfStack/game-library-auth/internal/database"
	"github.com/OutOfStack/game-library-auth/internal/model"
	"github.com/OutOfStack/game-library-auth/internal/notification"
	"github.com/golang-jwt/jwt/v4"
	"go.uber.org/zap"
)

// Provider represents dependencies for facade layer
type Provider struct {
	log                       *zap.Logger
	userRepo                  UserRepo
	emailPreparer             EmailPreparer
	auth                      Auth
	unsubscribeTokenGenerator *auth.UnsubscribeTokenGenerator
	infoAPIClient             InfoAPIClient
}

// New creates a new facade provider
func New(log *zap.Logger, userRepo UserRepo, emailPreparer EmailPreparer, authService Auth, unsubscribeTokenGenerator *auth.UnsubscribeTokenGenerator, infoAPIClient InfoAPIClient) *Provider {
	return &Provider{
		log:                       log,
		userRepo:                  userRepo,
		emailPreparer:             emailPreparer,
		auth:                      authService,
		unsubscribeTokenGenerator: unsubscribeTokenGenerator,
		infoAPIClient:             infoAPIClient,
	}
}

// Auth provides authentication methods
type Auth interface {
	GenerateToken(claims jwt.Claims) (string, error)
	GenerateRefreshToken() (string, time.Time, error)
	CreateUserClaims(user model.User) jwt.Claims
	GetClaimsFromToken(tokenStr string) (auth.Claims, error)
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
	GetUserByOAuthLink(ctx context.Context, provider string, oauthID string) (database.User, error)
	CreateUserOAuthLink(ctx context.Context, link database.UserOAuthLink) error
	HasOAuthLink(ctx context.Context, userID string) (bool, error)
	CheckUserExists(ctx context.Context, name string, role model.Role) (bool, error)
	SetUserEmailVerified(ctx context.Context, userID string) error

	CreateEmailVerification(ctx context.Context, verification database.EmailVerification) error
	GetEmailVerificationByUserID(ctx context.Context, userID string) (database.EmailVerification, error)
	GetVerificationState(ctx context.Context, userID, email string) (database.VerificationState, error)
	CreateNotification(ctx context.Context, notification database.Notification) error
	SetEmailVerificationUsed(ctx context.Context, id string, verified bool) error
	SetUnsubscribeToken(ctx context.Context, id string, token string) error

	CreateEmailUnsubscribe(ctx context.Context, unsubscribe database.EmailUnsubscribe) error
	IsEmailUnsubscribed(ctx context.Context, email string) (bool, error)

	CreateRefreshToken(ctx context.Context, refreshToken database.RefreshToken) error
	GetRefreshTokenByHash(ctx context.Context, tokenHash string) (database.RefreshToken, error)
	DeleteRefreshToken(ctx context.Context, token string) error
	DeleteRefreshTokensByUserID(ctx context.Context, userID string) error
}

// EmailPreparer renders an immutable email without contacting the provider
type EmailPreparer interface {
	PrepareVerification(req resendapi.SendEmailVerificationRequest) (notification.Email, error)
}

// InfoAPIClient provides methods for interacting with game library info API
type InfoAPIClient interface {
	CompanyExists(ctx context.Context, companyName string) (bool, error)
}
