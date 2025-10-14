package handlers

import (
	"context"
	"errors"

	"github.com/OutOfStack/game-library-auth/internal/appconf"
	"github.com/OutOfStack/game-library-auth/internal/auth"
	"github.com/OutOfStack/game-library-auth/internal/model"
	"github.com/golang-jwt/jwt/v4"
	"go.uber.org/zap"
	"google.golang.org/api/idtoken"
)

// Auth provides authentication methods
type Auth interface {
	GenerateToken(claims jwt.Claims) (string, error)
	ValidateToken(tokenStr string) (auth.Claims, error)
	CreateUserClaims(user model.User) jwt.Claims
}

// GoogleTokenValidator provides methods for validating Google ID tokens
type GoogleTokenValidator interface {
	Validate(ctx context.Context, idToken string, audience string) (*idtoken.Payload, error)
}

// UserFacade provides methods for working with user facade
type UserFacade interface {
	GoogleOAuth(ctx context.Context, oauthID, email string) (model.User, error)
	DeleteUser(ctx context.Context, userID string) error
	UpdateUserProfile(ctx context.Context, userID string, params model.UpdateProfileParams) (model.User, error)
	VerifyEmail(ctx context.Context, userID string, code string) (model.User, error)
	ResendVerificationEmail(ctx context.Context, userID string) error
	SignIn(ctx context.Context, username, password string) (model.User, error)
	SignUp(ctx context.Context, username, displayName, email, password string, isPublisher bool) (model.User, error)
}

// AuthAPI describes dependencies for auth endpoints
type AuthAPI struct {
	googleOAuthClientID  string
	log                  *zap.Logger
	auth                 Auth
	googleTokenValidator GoogleTokenValidator
	userFacade           UserFacade
}

// NewAuthAPI return new instance of auth api
func NewAuthAPI(log *zap.Logger, cfg *appconf.Cfg, auth Auth, googleTokenValidator GoogleTokenValidator, userFacade UserFacade) (*AuthAPI, error) {
	if cfg.Auth.GoogleClientID == "" {
		return nil, errors.New("google client id is empty")
	}

	return &AuthAPI{
		googleOAuthClientID:  cfg.Auth.GoogleClientID,
		auth:                 auth,
		log:                  log,
		googleTokenValidator: googleTokenValidator,
		userFacade:           userFacade,
	}, nil
}
