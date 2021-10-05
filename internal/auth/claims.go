package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

type Claims struct {
	jwt.RegisteredClaims
	UserRole string `json:"user_role,omitempty"`
}

// CreateClaims creates claims for user
func CreateClaims(issuer string, userID uuid.UUID, role string) jwt.Claims {
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    issuer,
			Subject:   userID.String(),
			Audience:  jwt.ClaimStrings{"game_lib_svc"},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(720 * time.Hour)),
			NotBefore: jwt.NewNumericDate(time.Now()),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		UserRole: role,
	}

	return claims
}
