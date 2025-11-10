package auth

import (
	"time"

	"github.com/OutOfStack/game-library-auth/internal/model"
	"github.com/golang-jwt/jwt/v4"
)

// Claims represent jwt claims
type Claims struct {
	jwt.RegisteredClaims
	UserID   string `json:"user_id"`
	UserRole string `json:"user_role,omitempty"`
	Username string `json:"username,omitempty"`
	Name     string `json:"name,omitempty"`
	// VerificationRequired - represents requirement to verify email (true = not verified, false = verified or does not require verification)
	VerificationRequired bool `json:"vrf_required"`
}

// CreateUserClaims creates claims for user
func (a *Auth) CreateUserClaims(user model.User) jwt.Claims {
	now := time.Now()
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    a.claimsIssuer,
			Subject:   user.ID,
			Audience:  jwt.ClaimStrings{"game_lib_svc"},
			ExpiresAt: jwt.NewNumericDate(now.Add(a.accessTokenTTL)),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
		},
		UserID:               user.ID,
		UserRole:             user.Role,
		Username:             user.Username,
		Name:                 user.DisplayName,
		VerificationRequired: user.IsPublisher() && !user.EmailVerified,
	}

	return claims
}
