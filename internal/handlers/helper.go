package handlers

import (
	"errors"
	"fmt"
	"net/mail"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
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

	if _, err = uuid.Parse(claims.UserID); err != nil {
		return "", errors.New("invalid user ID")
	}

	return claims.UserID, nil
}

// extractUsernameFromEmail extracts and sanitizes username from email for OAuth users
func extractUsernameFromEmail(email string) (string, error) {
	if email == "" {
		return "", errors.New("email cannot be empty")
	}

	if _, err := mail.ParseAddress(email); err != nil {
		return "", fmt.Errorf("invalid email format: %w", err)
	}

	// extract username part (before @)
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return "", errors.New("invalid email format")
	}

	username := parts[0]

	username = strings.ToLower(strings.TrimSpace(username))

	if len(username) > maxUsernameLen {
		username = username[:maxUsernameLen]
	}

	return username, nil
}
