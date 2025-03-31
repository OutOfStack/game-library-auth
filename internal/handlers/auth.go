//go:generate mockgen -source=auth.go -destination=mocks/auth.go -package=handlers_mocks

package handlers

import (
	"context"

	"github.com/OutOfStack/game-library-auth/internal/auth"
	"github.com/OutOfStack/game-library-auth/internal/database"
	"github.com/golang-jwt/jwt/v4"
	"go.uber.org/zap"
)

// Storage provides methods for working with repo
type Storage interface {
	CreateUser(ctx context.Context, user database.User) error
	UpdateUser(ctx context.Context, user database.User) error
	GetUserByID(ctx context.Context, userID string) (database.User, error)
	GetUserByUsername(ctx context.Context, username string) (database.User, error)
	CheckUserExists(ctx context.Context, name string, role database.Role) (bool, error)
}

// Auth provides authentication methods
type Auth interface {
	GenerateToken(claims jwt.Claims) (string, error)
	ValidateToken(tokenStr string) (auth.Claims, error)
	CreateClaims(user database.User) jwt.Claims
}

// AuthAPI describes dependencies for auth endpoints
type AuthAPI struct {
	auth    Auth
	storage Storage
	log     *zap.Logger
}

// NewAuthAPI return new instance of auth api
func NewAuthAPI(log *zap.Logger, auth Auth, storage Storage) *AuthAPI {
	return &AuthAPI{
		auth:    auth,
		storage: storage,
		log:     log,
	}
}
