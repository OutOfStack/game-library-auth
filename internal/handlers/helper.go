package handlers

import (
	"errors"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// getUserIDFromJWT extracts and validates JWT from Authorization header and returns the user ID
func (a *AuthAPI) getUserIDFromJWT(c *fiber.Ctx) (string, error) {
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("authorization header required")
	}

	// Expected format: "Bearer <token>"
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", errors.New("invalid authorization header format")
	}

	tokenStr := parts[1]
	claims, err := a.auth.ValidateToken(tokenStr)
	if err != nil {
		return "", fmt.Errorf("invalid or expired token: %w", err)
	}

	return claims.UserID, nil
}
