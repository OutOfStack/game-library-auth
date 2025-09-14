package auth

import (
	"time"

	"github.com/OutOfStack/game-library-auth/internal/model"
	"github.com/golang-jwt/jwt/v4"
)

const (
	// GoogleAuthTokenProvider is google auth token provider
	GoogleAuthTokenProvider = "google"
)

// Claims represent jwt claims
type Claims struct {
	jwt.RegisteredClaims
	UserID        string `json:"user_id"`
	UserRole      string `json:"user_role,omitempty"`
	Username      string `json:"username,omitempty"`
	Name          string `json:"name,omitempty"`
	EmailVerified bool   `json:"email_verified"`
}

// CreateUserClaims creates claims for user
func (a *Auth) CreateUserClaims(user model.User) jwt.Claims {
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    a.claimsIssuer,
			Subject:   user.ID,
			Audience:  jwt.ClaimStrings{"game_lib_svc"},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(360 * time.Hour)),
			NotBefore: jwt.NewNumericDate(time.Now()),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		UserID:        user.ID,
		UserRole:      user.Role,
		Username:      user.Username,
		Name:          user.DisplayName,
		EmailVerified: user.EmailVerified,
	}

	return claims
}
