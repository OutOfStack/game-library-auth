package facade_test

import (
	"errors"
	"testing"

	"github.com/OutOfStack/game-library-auth/internal/auth"
)

func TestProvider_ValidateAccessToken(t *testing.T) {
	t.Run("valid token", func(t *testing.T) {
		provider, _, _, mockAuth, ctrl := setupTest(t)
		defer ctrl.Finish()

		expected := auth.Claims{UserID: "user-123", Username: "testuser"}

		mockAuth.EXPECT().
			ValidateToken("valid.jwt.token").
			Return(expected, nil)

		got, err := provider.ValidateAccessToken("valid.jwt.token")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if got.UserID != expected.UserID || got.Username != expected.Username {
			t.Errorf("unexpected claims: got=%+v expected=%+v", got, expected)
		}
	})

	t.Run("invalid token", func(t *testing.T) {
		provider, _, _, mockAuth, ctrl := setupTest(t)
		defer ctrl.Finish()

		mockAuth.EXPECT().
			ValidateToken("invalid.jwt.token").
			Return(auth.Claims{}, errors.New("invalid token"))

		_, err := provider.ValidateAccessToken("invalid.jwt.token")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}
