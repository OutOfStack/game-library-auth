package facade_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/OutOfStack/game-library-auth/internal/auth"
	"github.com/OutOfStack/game-library-auth/internal/database"
	"go.uber.org/mock/gomock"
)

func TestProvider_UnsubscribeEmail_Success(t *testing.T) {
	provider, mockUserRepo, _, ctrl := setupTest(t)
	defer ctrl.Finish()

	ctx := context.Background()

	email := "test@example.com"
	expiresAt := time.Now().Add(24 * time.Hour)
	tokenGen := auth.NewUnsubscribeTokenGenerator([]byte("test-secret-key"))
	token := tokenGen.GenerateToken(email, expiresAt)

	verification := database.EmailVerification{
		ID:          "verification-123",
		UserID:      "user-123",
		DateCreated: time.Now().Add(-1 * time.Hour),
	}

	mockUserRepo.EXPECT().
		GetEmailVerificationByUnsubscribeToken(ctx, token).
		Return(verification, nil)

	mockUserRepo.EXPECT().
		CreateEmailUnsubscribe(ctx, gomock.Any()).
		Return(nil)

	resultEmail, err := provider.UnsubscribeEmail(ctx, token)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if resultEmail != email {
		t.Errorf("expected email %s, got %s", email, resultEmail)
	}
}

func TestProvider_UnsubscribeEmail_InvalidToken(t *testing.T) {
	provider, _, _, ctrl := setupTest(t)
	defer ctrl.Finish()

	ctx := context.Background()

	invalidToken := "invalid-token"

	_, err := provider.UnsubscribeEmail(ctx, invalidToken)
	if err == nil {
		t.Error("expected error for invalid token")
	}
}

func TestProvider_UnsubscribeEmail_ExpiredToken(t *testing.T) {
	provider, mockUserRepo, _, ctrl := setupTest(t)
	defer ctrl.Finish()

	ctx := context.Background()

	email := "test@example.com"
	expiresAt := time.Now().Add(24 * time.Hour)
	tokenGen := auth.NewUnsubscribeTokenGenerator([]byte("test-secret-key"))
	token := tokenGen.GenerateToken(email, expiresAt)

	verification := database.EmailVerification{
		ID:          "verification-123",
		UserID:      "user-123",
		DateCreated: time.Now().Add(-31 * 24 * time.Hour),
	}

	mockUserRepo.EXPECT().
		GetEmailVerificationByUnsubscribeToken(ctx, token).
		Return(verification, nil)

	_, err := provider.UnsubscribeEmail(ctx, token)
	if err == nil {
		t.Error("expected error for expired unsubscribe link")
	}

	if err.Error() != "unsubscribe link has expired" {
		t.Errorf("expected 'unsubscribe link has expired' error, got %v", err)
	}
}

func TestProvider_UnsubscribeEmail_VerificationNotFound(t *testing.T) {
	provider, mockUserRepo, _, ctrl := setupTest(t)
	defer ctrl.Finish()

	ctx := context.Background()

	email := "test@example.com"
	expiresAt := time.Now().Add(24 * time.Hour)
	tokenGen := auth.NewUnsubscribeTokenGenerator([]byte("test-secret-key"))
	token := tokenGen.GenerateToken(email, expiresAt)

	mockUserRepo.EXPECT().
		GetEmailVerificationByUnsubscribeToken(ctx, token).
		Return(database.EmailVerification{}, database.ErrNotFound)

	_, err := provider.UnsubscribeEmail(ctx, token)
	if !errors.Is(err, database.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestProvider_UnsubscribeEmail_DatabaseError(t *testing.T) {
	provider, mockUserRepo, _, ctrl := setupTest(t)
	defer ctrl.Finish()

	ctx := context.Background()

	email := "test@example.com"
	expiresAt := time.Now().Add(24 * time.Hour)
	tokenGen := auth.NewUnsubscribeTokenGenerator([]byte("test-secret-key"))
	token := tokenGen.GenerateToken(email, expiresAt)

	dbError := errors.New("database connection error")

	mockUserRepo.EXPECT().
		GetEmailVerificationByUnsubscribeToken(ctx, token).
		Return(database.EmailVerification{}, dbError)

	_, err := provider.UnsubscribeEmail(ctx, token)
	if err == nil {
		t.Error("expected error for database error")
	}
}

func TestProvider_UnsubscribeEmail_CreateUnsubscribeFails(t *testing.T) {
	provider, mockUserRepo, _, ctrl := setupTest(t)
	defer ctrl.Finish()

	ctx := context.Background()

	email := "test@example.com"
	expiresAt := time.Now().Add(24 * time.Hour)
	tokenGen := auth.NewUnsubscribeTokenGenerator([]byte("test-secret-key"))
	token := tokenGen.GenerateToken(email, expiresAt)

	verification := database.EmailVerification{
		ID:          "verification-123",
		UserID:      "user-123",
		DateCreated: time.Now().Add(-1 * time.Hour),
	}

	mockUserRepo.EXPECT().
		GetEmailVerificationByUnsubscribeToken(ctx, token).
		Return(verification, nil)

	createError := errors.New("failed to create unsubscribe record")
	mockUserRepo.EXPECT().
		CreateEmailUnsubscribe(ctx, gomock.Any()).
		Return(createError)

	_, err := provider.UnsubscribeEmail(ctx, token)
	if err == nil {
		t.Error("expected error when CreateEmailUnsubscribe fails")
	}
}

func TestProvider_UnsubscribeEmail_RecentToken(t *testing.T) {
	provider, mockUserRepo, _, ctrl := setupTest(t)
	defer ctrl.Finish()

	ctx := context.Background()

	email := "recent@example.com"
	expiresAt := time.Now().Add(24 * time.Hour)
	tokenGen := auth.NewUnsubscribeTokenGenerator([]byte("test-secret-key"))
	token := tokenGen.GenerateToken(email, expiresAt)

	verification := database.EmailVerification{
		ID:          "verification-123",
		UserID:      "user-123",
		DateCreated: time.Now().Add(-30 * time.Minute),
	}

	mockUserRepo.EXPECT().
		GetEmailVerificationByUnsubscribeToken(ctx, token).
		Return(verification, nil)

	mockUserRepo.EXPECT().
		CreateEmailUnsubscribe(ctx, gomock.Any()).
		Return(nil)

	resultEmail, err := provider.UnsubscribeEmail(ctx, token)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if resultEmail != email {
		t.Errorf("expected email %s, got %s", email, resultEmail)
	}
}
