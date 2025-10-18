package auth_test

import (
	"encoding/base64"
	"testing"
	"time"

	"github.com/OutOfStack/game-library-auth/internal/auth"
)

func TestGenerateToken(t *testing.T) {
	secretKey := []byte("test-secret-key-32-bytes-long!")
	generator := auth.NewUnsubscribeTokenGenerator(secretKey)

	email := "test@example.com"
	expiresAt := time.Now().Add(24 * time.Hour)

	token := generator.GenerateToken(email, expiresAt)

	if token == "" {
		t.Error("expected non-empty token")
	}

	// verify token can be base64 decoded
	_, err := base64.URLEncoding.DecodeString(token)
	if err != nil {
		t.Errorf("token should be valid base64: %v", err)
	}
}

func TestValidateToken_Success(t *testing.T) {
	secretKey := []byte("test-secret-key-32-bytes-long!")
	generator := auth.NewUnsubscribeTokenGenerator(secretKey)

	email := "test@example.com"
	expiresAt := time.Now().Add(24 * time.Hour)

	token := generator.GenerateToken(email, expiresAt)

	validatedEmail, err := generator.ValidateToken(token)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if validatedEmail != email {
		t.Errorf("expected email %s, got %s", email, validatedEmail)
	}
}

func TestValidateToken_ExpiredToken(t *testing.T) {
	secretKey := []byte("test-secret-key-32-bytes-long!")
	generator := auth.NewUnsubscribeTokenGenerator(secretKey)

	email := "test@example.com"
	expiresAt := time.Now().Add(-1 * time.Hour)

	token := generator.GenerateToken(email, expiresAt)

	_, err := generator.ValidateToken(token)
	if err == nil {
		t.Error("expected error for expired token")
	}

	if err.Error() != "token expired" {
		t.Errorf("expected 'token expired' error, got %v", err)
	}
}

func TestValidateToken_InvalidFormat(t *testing.T) {
	secretKey := []byte("test-secret-key-32-bytes-long!")
	generator := auth.NewUnsubscribeTokenGenerator(secretKey)

	invalidToken := "not-a-valid-base64-token!!!"

	_, err := generator.ValidateToken(invalidToken)
	if err == nil {
		t.Error("expected error for invalid token format")
	}

	if err.Error() != "invalid token format" {
		t.Errorf("expected 'invalid token format' error, got %v", err)
	}
}

func TestValidateToken_TokenTooShort(t *testing.T) {
	secretKey := []byte("test-secret-key-32-bytes-long!")
	generator := auth.NewUnsubscribeTokenGenerator(secretKey)

	shortToken := base64.URLEncoding.EncodeToString([]byte("short"))

	_, err := generator.ValidateToken(shortToken)
	if err == nil {
		t.Error("expected error for short token")
	}

	if err.Error() != "token too short" {
		t.Errorf("expected 'token too short' error, got %v", err)
	}
}

func TestValidateToken_InvalidSignature(t *testing.T) {
	secretKey1 := []byte("test-secret-key-32-bytes-long!")
	secretKey2 := []byte("different-secret-key-32-bytes!")

	generator1 := auth.NewUnsubscribeTokenGenerator(secretKey1)
	generator2 := auth.NewUnsubscribeTokenGenerator(secretKey2)

	email := "test@example.com"
	expiresAt := time.Now().Add(24 * time.Hour)

	token := generator1.GenerateToken(email, expiresAt)

	_, err := generator2.ValidateToken(token)
	if err == nil {
		t.Error("expected error for invalid signature")
	}

	if err.Error() != "invalid token signature" {
		t.Errorf("expected 'invalid token signature' error, got %v", err)
	}
}

func TestValidateToken_InvalidPayloadFormat(t *testing.T) {
	secretKey := []byte("test-secret-key-32-bytes-long!")
	generator := auth.NewUnsubscribeTokenGenerator(secretKey)

	// manually create a token with invalid payload (no colon separator)
	invalidPayload := "invalid-payload-format"
	tokenData := make([]byte, 32+len(invalidPayload))
	copy(tokenData[32:], []byte(invalidPayload))
	invalidToken := base64.URLEncoding.EncodeToString(tokenData)

	_, err := generator.ValidateToken(invalidToken)
	if err == nil {
		t.Error("expected error for invalid payload format")
	}

	if err.Error() != "invalid token signature" {
		t.Errorf("expected 'invalid token signature' error, got %v", err)
	}
}

func TestValidateToken_InvalidExpiryTime(t *testing.T) {
	secretKey := []byte("test-secret-key-32-bytes-long!")
	generator := auth.NewUnsubscribeTokenGenerator(secretKey)

	// manually create a token with invalid expiry time
	invalidPayload := "test@example.com:not-a-number"
	tokenData := make([]byte, 32+len(invalidPayload))
	copy(tokenData[32:], []byte(invalidPayload))
	invalidToken := base64.URLEncoding.EncodeToString(tokenData)

	_, err := generator.ValidateToken(invalidToken)
	if err == nil {
		t.Error("expected error for invalid expiry time")
	}

	if err.Error() != "invalid token signature" {
		t.Errorf("expected 'invalid token signature' error, got %v", err)
	}
}

func TestGenerateToken_Deterministic(t *testing.T) {
	secretKey := []byte("test-secret-key-32-bytes-long!")
	generator := auth.NewUnsubscribeTokenGenerator(secretKey)

	email := "test@example.com"
	expiresAt := time.Unix(1704067200, 0)

	token1 := generator.GenerateToken(email, expiresAt)
	token2 := generator.GenerateToken(email, expiresAt)

	if token1 != token2 {
		t.Error("expected same token for same inputs")
	}
}

func TestGenerateToken_DifferentEmails(t *testing.T) {
	secretKey := []byte("test-secret-key-32-bytes-long!")
	generator := auth.NewUnsubscribeTokenGenerator(secretKey)

	expiresAt := time.Now().Add(24 * time.Hour)

	token1 := generator.GenerateToken("user1@example.com", expiresAt)
	token2 := generator.GenerateToken("user2@example.com", expiresAt)

	if token1 == token2 {
		t.Error("expected different tokens for different emails")
	}

	email1, err := generator.ValidateToken(token1)
	if err != nil {
		t.Fatalf("expected no error for token1, got %v", err)
	}
	if email1 != "user1@example.com" {
		t.Errorf("expected email user1@example.com, got %s", email1)
	}

	email2, err := generator.ValidateToken(token2)
	if err != nil {
		t.Fatalf("expected no error for token2, got %v", err)
	}
	if email2 != "user2@example.com" {
		t.Errorf("expected email user2@example.com, got %s", email2)
	}
}

func TestValidateToken_EmailWithColon(t *testing.T) {
	secretKey := []byte("test-secret-key-32-bytes-long!")
	generator := auth.NewUnsubscribeTokenGenerator(secretKey)

	// emails with colons will cause parsing issues because colon is used as separator
	// payload format: "email:timestamp", but "test:user@example.com:timestamp" splits into 3 parts
	email := "test:user@example.com"
	expiresAt := time.Now().Add(24 * time.Hour)

	token := generator.GenerateToken(email, expiresAt)

	_, err := generator.ValidateToken(token)
	if err == nil {
		t.Error("expected error for email with colon")
	}

	if err.Error() != "invalid token payload" {
		t.Errorf("expected 'invalid token payload' error, got %v", err)
	}
}
