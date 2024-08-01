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
	t.Logf("Testing generation and validation of JWT\n")
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Generating private key: %v\n", err)
	}

	a, err := auth.New("RS256", privateKey)
	if err != nil {
		t.Fatalf("Initializing token service instance: %v", err)
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
