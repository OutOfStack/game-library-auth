package handlers

import (
	"context"
	"errors"

	"github.com/OutOfStack/game-library-auth/internal/appconf"
	"github.com/OutOfStack/game-library-auth/internal/auth"
	"github.com/OutOfStack/game-library-auth/internal/database"
	"github.com/golang-jwt/jwt/v4"
	"go.uber.org/zap"
	"google.golang.org/api/idtoken"
)

// UserRepo provides methods for working with user repo
type UserRepo interface {
	CreateUser(ctx context.Context, user database.User) error
	UpdateUser(ctx context.Context, user database.User) error
	GetUserByID(ctx context.Context, userID string) (database.User, error)
	GetUserByUsername(ctx context.Context, username string) (database.User, error)
	GetUserByOAuth(ctx context.Context, provider string, oauthID string) (database.User, error)
	CheckUserExists(ctx context.Context, name string, role database.Role) (bool, error)
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

// AuthAPI describes dependencies for auth endpoints
type AuthAPI struct {
	googleOAuthClientID  string
	log                  *zap.Logger
	auth                 Auth
	userRepo             UserRepo
	googleTokenValidator GoogleTokenValidator
}

// NewAuthAPI return new instance of auth api
func NewAuthAPI(log *zap.Logger, cfg *appconf.Cfg, auth Auth, userRepo UserRepo, googleTokenValidator GoogleTokenValidator) (*AuthAPI, error) {
	if cfg.Auth.GoogleClientID == "" {
		return nil, errors.New("google client id is empty")
	}

	return &AuthAPI{
		googleOAuthClientID:  cfg.Auth.GoogleClientID,
		auth:                 auth,
		userRepo:             userRepo,
		log:                  log,
		googleTokenValidator: googleTokenValidator,
	}, nil
}
