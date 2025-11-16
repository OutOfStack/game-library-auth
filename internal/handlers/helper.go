package handlers

import (
	"errors"
	"fmt"
	"strings"

	"github.com/OutOfStack/game-library-auth/internal/auth"
	"github.com/OutOfStack/game-library-auth/internal/facade"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// getClaims extracts and validates JWT from Authorization header and returns the claims
func (a *AuthAPI) getClaims(c *fiber.Ctx) (auth.Claims, error) {
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return auth.Claims{}, errors.New("authorization header required")
	}

	// Expected format: "Bearer <token>"
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return auth.Claims{}, errors.New("invalid authorization header format")
	}

	tokenStr := parts[1]
	claims, err := a.userFacade.ValidateAccessToken(tokenStr)
	if err != nil {
		return auth.Claims{}, fmt.Errorf("invalid or expired token: %w", err)
	}

	if _, err = uuid.Parse(claims.UserID); err != nil {
		return auth.Claims{}, errors.New("invalid user ID")
	}

	return claims, nil
}

// getUserIDFromJWT extracts and validates JWT from Authorization header and returns the user ID
func (a *AuthAPI) getUserIDFromJWT(c *fiber.Ctx) (string, error) {
	claims, err := a.getClaims(c)
	if err != nil {
		return "", err
	}

	return claims.UserID, nil
}

// setRefreshTokenCookie sets the refresh token as an httpOnly cookie
func (a *AuthAPI) setRefreshTokenCookie(c *fiber.Ctx, refreshToken facade.RefreshToken) {
	c.Cookie(&fiber.Cookie{
		Name:     refreshTokenCookieName,
		Value:    refreshToken.Token,
		Path:     "/",
		HTTPOnly: true,
		Secure:   a.cfg.RefreshTokenCookieSecure,
		SameSite: a.cfg.RefreshTokenCookieSameSite,
		Expires:  refreshToken.ExpiresAt,
	})
}
