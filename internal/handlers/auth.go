package handlers

import (
	"context"

	"github.com/OutOfStack/game-library-auth/internal/auth"
	"github.com/OutOfStack/game-library-auth/internal/data"
	"github.com/golang-jwt/jwt/v4"
	"go.uber.org/zap"
)

// Storage provides methods for working with repo
type Storage interface {
	CreateUser(ctx context.Context, user data.User) error
	UpdateUser(ctx context.Context, user data.User) error
	GetUserByID(ctx context.Context, userID string) (data.User, error)
	GetUserByUsername(ctx context.Context, username string) (data.User, error)
	CheckUserExists(ctx context.Context, name string, role data.Role) (bool, error)
}

// Auther provides authentication methods
type Auther interface {
	GenerateToken(claims jwt.Claims) (string, error)
	ValidateToken(tokenStr string) (auth.Claims, error)
	CreateClaims(user data.User) jwt.Claims
}

// AuthAPI describes dependencies for auth endpoints
type AuthAPI struct {
	auth    Auther
	storage Storage
	log     *zap.Logger
}

// NewAuthAPI return new instance of auth api
func NewAuthAPI(log *zap.Logger, auth Auther, storage Storage) *AuthAPI {
	return &AuthAPI{
		auth:    auth,
		storage: storage,
		log:     log,
	}
}
