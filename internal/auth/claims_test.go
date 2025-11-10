package auth_test

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	"github.com/OutOfStack/game-library-auth/internal/auth"
	"github.com/OutOfStack/game-library-auth/internal/model"
	"github.com/golang-jwt/jwt/v4"
)

func TestCreateUserClaims_RegularUser(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate private key: %v", err)
	}

	a, err := auth.New("RS256", privateKey, "test-issuer", 15*time.Minute, 7*24*time.Hour)
	if err != nil {
		t.Fatalf("failed to create auth: %v", err)
	}

	user := model.User{
		ID:            "user-123",
		Username:      "testuser",
		DisplayName:   "Test User",
		Email:         "test@example.com",
		EmailVerified: true,
		Role:          "user",
	}

	claims := a.CreateUserClaims(user)

	authClaims, ok := claims.(auth.Claims)
	if !ok {
		t.Fatal("expected claims to be of type auth.Claims")
	}

	if authClaims.UserID != user.ID {
		t.Errorf("expected UserID %s, got %s", user.ID, authClaims.UserID)
	}

	if authClaims.Username != user.Username {
		t.Errorf("expected Username %s, got %s", user.Username, authClaims.Username)
	}

	if authClaims.Name != user.DisplayName {
		t.Errorf("expected Name %s, got %s", user.DisplayName, authClaims.Name)
	}

	if authClaims.UserRole != user.Role {
		t.Errorf("expected UserRole %s, got %s", user.Role, authClaims.UserRole)
	}

	if authClaims.VerificationRequired {
		t.Error("expected VerificationRequired to be false for regular user")
	}

	if authClaims.Issuer != "test-issuer" {
		t.Errorf("expected Issuer test-issuer, got %s", authClaims.Issuer)
	}

	if authClaims.Subject != user.ID {
		t.Errorf("expected Subject %s, got %s", user.ID, authClaims.Subject)
	}

	if len(authClaims.Audience) != 1 || authClaims.Audience[0] != "game_lib_svc" {
		t.Errorf("expected Audience [game_lib_svc], got %v", authClaims.Audience)
	}
}

func TestCreateUserClaims_Publisher_EmailVerified(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate private key: %v", err)
	}

	a, err := auth.New("RS256", privateKey, "test-issuer", 15*time.Minute, 7*24*time.Hour)
	if err != nil {
		t.Fatalf("failed to create auth: %v", err)
	}

	user := model.User{
		ID:            "publisher-123",
		Username:      "testpublisher",
		DisplayName:   "Test Publisher",
		Email:         "publisher@example.com",
		EmailVerified: true,
		Role:          "publisher",
	}

	claims := a.CreateUserClaims(user)

	authClaims, ok := claims.(auth.Claims)
	if !ok {
		t.Fatal("expected claims to be of type auth.Claims")
	}

	if authClaims.VerificationRequired {
		t.Error("expected VerificationRequired to be false for verified publisher")
	}

	if authClaims.UserRole != "publisher" {
		t.Errorf("expected UserRole publisher, got %s", authClaims.UserRole)
	}
}

func TestCreateUserClaims_Publisher_EmailNotVerified(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate private key: %v", err)
	}

	a, err := auth.New("RS256", privateKey, "test-issuer", 15*time.Minute, 7*24*time.Hour)
	if err != nil {
		t.Fatalf("failed to create auth: %v", err)
	}

	user := model.User{
		ID:            "publisher-123",
		Username:      "testpublisher",
		DisplayName:   "Test Publisher",
		Email:         "publisher@example.com",
		EmailVerified: false,
		Role:          "publisher",
	}

	claims := a.CreateUserClaims(user)

	authClaims, ok := claims.(auth.Claims)
	if !ok {
		t.Fatal("expected claims to be of type auth.Claims")
	}

	if !authClaims.VerificationRequired {
		t.Error("expected VerificationRequired to be true for unverified publisher")
	}

	if authClaims.UserRole != "publisher" {
		t.Errorf("expected UserRole publisher, got %s", authClaims.UserRole)
	}
}

func TestCreateUserClaims_TokenExpiry(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate private key: %v", err)
	}

	a, err := auth.New("RS256", privateKey, "test-issuer", 15*time.Minute, 7*24*time.Hour)
	if err != nil {
		t.Fatalf("failed to create auth: %v", err)
	}

	user := model.User{
		ID:       "user-123",
		Username: "testuser",
		Role:     "user",
	}

	now := time.Now()
	claims := a.CreateUserClaims(user)

	authClaims, ok := claims.(auth.Claims)
	if !ok {
		t.Fatal("expected claims to be of type auth.Claims")
	}

	if authClaims.ExpiresAt == nil {
		t.Fatal("expected ExpiresAt to be set")
	}

	if authClaims.IssuedAt == nil {
		t.Fatal("expected IssuedAt to be set")
	}

	if authClaims.NotBefore == nil {
		t.Fatal("expected NotBefore to be set")
	}

	expectedExpiryMin := now.Add(15*time.Minute - 1*time.Minute)
	expectedExpiryMax := now.Add(15*time.Minute + 1*time.Minute)
	if authClaims.ExpiresAt.Before(expectedExpiryMin) || authClaims.ExpiresAt.After(expectedExpiryMax) {
		t.Errorf("expected ExpiresAt to be around %v in the future, got %v", 15*time.Minute, authClaims.ExpiresAt)
	}

	timeTolerance := 2 * time.Second
	if authClaims.IssuedAt.Before(now.Add(-timeTolerance)) || authClaims.IssuedAt.After(now.Add(timeTolerance)) {
		t.Errorf("expected IssuedAt to be around now, got %v", authClaims.IssuedAt)
	}

	if authClaims.NotBefore.Before(now.Add(-timeTolerance)) || authClaims.NotBefore.After(now.Add(timeTolerance)) {
		t.Errorf("expected NotBefore to be around now, got %v", authClaims.NotBefore)
	}
}

func TestCreateUserClaims_Subject(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate private key: %v", err)
	}

	a, err := auth.New("RS256", privateKey, "test-issuer", 15*time.Minute, 7*24*time.Hour)
	if err != nil {
		t.Fatalf("failed to create auth: %v", err)
	}

	user := model.User{
		ID:       "unique-user-id-456",
		Username: "testuser",
		Role:     "user",
	}

	claims := a.CreateUserClaims(user)

	authClaims, ok := claims.(auth.Claims)
	if !ok {
		t.Fatal("expected claims to be of type auth.Claims")
	}

	if authClaims.Subject != user.ID {
		t.Errorf("expected Subject to be %s, got %s", user.ID, authClaims.Subject)
	}
}

func TestClaims_ImplementsJWTClaims(t *testing.T) {
	claims := auth.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "test",
			Subject:   "user-123",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
		},
		UserID: "user-123",
	}

	var _ jwt.Claims = claims

	err := claims.Valid()
	if err != nil {
		t.Errorf("expected claims to be valid, got error: %v", err)
	}
}
