package auth_test

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	"github.com/OutOfStack/game-library-auth/internal/auth"
	"github.com/golang-jwt/jwt/v4"
)

// TestGenerateValidate tests GenerateToken and ValidateToken functions
func TestGenerateValidate(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Generating private key: %v", err)
	}

	a, err := auth.New("RS256", privateKey, "", 15*time.Minute, 7*24*time.Hour)
	if err != nil {
		t.Fatalf("Initializing auth service instance: %v", err)
	}

	claims := auth.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "test_runner",
			Subject:   "12345qwerty",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(720 * time.Hour)),
			NotBefore: jwt.NewNumericDate(time.Now()),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		UserRole: "super_admin",
	}

	tokenStr, err := a.GenerateToken(claims)
	if err != nil {
		t.Fatalf("Generating token: %v\n", err)
	}

	parsedClaims, err := a.ValidateToken(tokenStr)
	if err != nil {
		t.Fatalf("Validating token: %v\n", err)
	}

	if issuer := parsedClaims.Issuer; issuer != claims.Issuer {
		t.Fatalf("Expected Issuer claim to be %v, got %v", claims.Issuer, issuer)
	}
	if subject := parsedClaims.Subject; subject != claims.Subject {
		t.Fatalf("Expected Subject claim to be %v, got %v", claims.Subject, subject)
	}
	if userRole := parsedClaims.UserRole; userRole != claims.UserRole {
		t.Fatalf("Expected UserRole claim to be %v, got %v", claims.UserRole, userRole)
	}
}

func TestGenerateRefreshToken(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Generating private key: %v", err)
	}

	a, err := auth.New("RS256", privateKey, "", 15*time.Minute, 7*24*time.Hour)
	if err != nil {
		t.Fatalf("Initializing auth service instance: %v", err)
	}

	token1, _, err := a.GenerateRefreshToken()
	if err != nil {
		t.Fatalf("Generating refresh token: %v", err)
	}

	if token1 == "" {
		t.Fatal("Expected non-empty refresh token")
	}

	token2, _, err := a.GenerateRefreshToken()
	if err != nil {
		t.Fatalf("Generating second refresh token: %v", err)
	}

	if token1 == token2 {
		t.Fatal("Expected different tokens on subsequent calls")
	}

	if len(token1) < 40 {
		t.Fatalf("Expected token length to be at least 40 characters, got %d", len(token1))
	}
}
