package facade_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/OutOfStack/game-library-auth/internal/auth"
	"go.uber.org/mock/gomock"
)

func TestProvider_UnsubscribeEmail_Success(t *testing.T) {
	provider, mockUserRepo, _, _, ctrl := setupTest(t)
	defer ctrl.Finish()

	ctx := context.Background()

	email := "test@example.com"
	expiresAt := time.Now().Add(24 * time.Hour)
	tokenGen := auth.NewUnsubscribeTokenGenerator([]byte("test-secret-key"))
	token := tokenGen.GenerateToken(email, expiresAt)

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
	provider, _, _, _, ctrl := setupTest(t) //nolint
	defer ctrl.Finish()

	ctx := context.Background()

	invalidToken := "invalid-token"

	_, err := provider.UnsubscribeEmail(ctx, invalidToken)
	if err == nil {
		t.Error("expected error for invalid token")
	}
}

func TestProvider_UnsubscribeEmail_ExpiredToken(t *testing.T) {
	provider, _, _, _, ctrl := setupTest(t) //nolint
	defer ctrl.Finish()

	ctx := context.Background()

	email := "test@example.com"
	expiresAt := time.Now().Add(-1 * time.Hour) // Expired token
	tokenGen := auth.NewUnsubscribeTokenGenerator([]byte("test-secret-key"))
	token := tokenGen.GenerateToken(email, expiresAt)

	_, err := provider.UnsubscribeEmail(ctx, token)
	if err == nil {
		t.Error("expected error for expired token")
	}
}

func TestProvider_UnsubscribeEmail_CreateUnsubscribeFails(t *testing.T) {
	provider, mockUserRepo, _, _, ctrl := setupTest(t)
	defer ctrl.Finish()

	ctx := context.Background()

	email := "test@example.com"
	expiresAt := time.Now().Add(24 * time.Hour)
	tokenGen := auth.NewUnsubscribeTokenGenerator([]byte("test-secret-key"))
	token := tokenGen.GenerateToken(email, expiresAt)

	createError := errors.New("failed to create unsubscribe record")
	mockUserRepo.EXPECT().
		CreateEmailUnsubscribe(ctx, gomock.Any()).
		Return(createError)

	_, err := provider.UnsubscribeEmail(ctx, token)
	if err == nil {
		t.Error("expected error when CreateEmailUnsubscribe fails")
	}
}
