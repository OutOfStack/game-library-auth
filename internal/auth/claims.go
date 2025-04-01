package auth

import (
	"time"

	"github.com/OutOfStack/game-library-auth/internal/database"
	"github.com/golang-jwt/jwt/v4"
)

// Claims represent jwt claims
type Claims struct {
	jwt.RegisteredClaims
	UserID    string `json:"user_id"`
	UserRole  string `json:"user_role,omitempty"`
	Username  string `json:"username,omitempty"`
	Name      string `json:"name,omitempty"`
	AvatarURL string `json:"avatar,omitempty"`
}

// CreateClaims creates claims for user
func (a *Auth) CreateClaims(user database.User) jwt.Claims {
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    a.claimsIssuer,
			Subject:   user.ID.String(),
			Audience:  jwt.ClaimStrings{"game_lib_svc"},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(360 * time.Hour)),
			NotBefore: jwt.NewNumericDate(time.Now()),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		UserID:    user.ID.String(),
		UserRole:  string(user.Role),
		Username:  user.Username,
		Name:      user.Name,
		AvatarURL: user.AvatarURL.String,
	}

	return claims
}
