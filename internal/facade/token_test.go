package facade_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/OutOfStack/game-library-auth/internal/auth"
	"github.com/OutOfStack/game-library-auth/internal/database"
	"github.com/OutOfStack/game-library-auth/internal/facade"
	"github.com/OutOfStack/game-library-auth/internal/model"
	"go.uber.org/mock/gomock"
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

func TestProvider_RefreshTokens(t *testing.T) {
	ctx := context.Background()

	t.Run("successful refresh", func(t *testing.T) {
		provider, mockUserRepo, _, mockAuth, ctrl := setupTest(t)
		defer ctrl.Finish()

		refreshToken := database.RefreshToken{
			UserID:      "user-123",
			Token:       "old-refresh-token",
			ExpiresAt:   time.Now().Add(24 * time.Hour),
			DateCreated: time.Now(),
		}

		user := database.User{
			ID:       "user-123",
			Username: "testuser",
			Email:    sql.NullString{String: "test@example.com", Valid: true},
		}

		mockUserRepo.EXPECT().
			GetRefreshTokenByToken(gomock.Any(), "old-refresh-token").
			Return(refreshToken, nil)

		mockUserRepo.EXPECT().
			GetUserByID(gomock.Any(), "user-123").
			Return(user, nil)

		mockAuth.EXPECT().
			CreateUserClaims(gomock.Any()).
			Return(auth.Claims{UserID: "user-123", Username: "testuser"})

		mockAuth.EXPECT().
			GenerateToken(gomock.Any()).
			Return("new-access-token", nil)

		mockAuth.EXPECT().
			GenerateRefreshToken().
			Return("new-refresh-token", time.Now().Add(7*24*time.Hour), nil)

		mockUserRepo.EXPECT().
			RunWithTx(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
				return fn(ctx)
			})

		mockUserRepo.EXPECT().
			DeleteRefreshToken(gomock.Any(), "old-refresh-token").
			Return(nil)

		mockUserRepo.EXPECT().
			CreateRefreshToken(gomock.Any(), gomock.Any()).
			Return(nil)

		tokens, err := provider.RefreshTokens(ctx, "old-refresh-token")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if tokens.AccessToken != "new-access-token" {
			t.Errorf("expected access token 'new-access-token', got '%s'", tokens.AccessToken)
		}
		if tokens.RefreshToken.Token != "new-refresh-token" {
			t.Errorf("expected refresh token 'new-refresh-token', got '%s'", tokens.RefreshToken.Token)
		}
	})

	t.Run("token not found", func(t *testing.T) {
		provider, mockUserRepo, _, _, ctrl := setupTest(t)
		defer ctrl.Finish()

		mockUserRepo.EXPECT().
			GetRefreshTokenByToken(gomock.Any(), "invalid-token").
			Return(database.RefreshToken{}, database.ErrNotFound)

		_, err := provider.RefreshTokens(ctx, "invalid-token")
		if !errors.Is(err, facade.ErrRefreshTokenNotFound) {
			t.Errorf("expected ErrRefreshTokenNotFound, got %v", err)
		}
	})

	t.Run("expired token", func(t *testing.T) {
		provider, mockUserRepo, _, _, ctrl := setupTest(t)
		defer ctrl.Finish()

		expiredToken := database.RefreshToken{
			UserID:      "user-123",
			Token:       "expired-token",
			ExpiresAt:   time.Now().Add(-1 * time.Hour),
			DateCreated: time.Now().Add(-25 * time.Hour),
		}

		mockUserRepo.EXPECT().
			GetRefreshTokenByToken(gomock.Any(), "expired-token").
			Return(expiredToken, nil)

		mockUserRepo.EXPECT().
			DeleteRefreshToken(gomock.Any(), "expired-token").
			Return(nil).
			AnyTimes()

		_, err := provider.RefreshTokens(ctx, "expired-token")
		if !errors.Is(err, facade.ErrRefreshTokenExpired) {
			t.Errorf("expected ErrRefreshTokenExpired, got %v", err)
		}
	})

	t.Run("user not found", func(t *testing.T) {
		provider, mockUserRepo, _, _, ctrl := setupTest(t)
		defer ctrl.Finish()

		refreshToken := database.RefreshToken{
			UserID:      "user-123",
			Token:       "valid-token",
			ExpiresAt:   time.Now().Add(24 * time.Hour),
			DateCreated: time.Now(),
		}

		mockUserRepo.EXPECT().
			GetRefreshTokenByToken(gomock.Any(), "valid-token").
			Return(refreshToken, nil)

		mockUserRepo.EXPECT().
			GetUserByID(gomock.Any(), "user-123").
			Return(database.User{}, database.ErrNotFound)

		mockUserRepo.EXPECT().
			DeleteRefreshToken(gomock.Any(), "valid-token").
			Return(nil).
			AnyTimes()

		_, err := provider.RefreshTokens(ctx, "valid-token")
		if !errors.Is(err, facade.ErrRefreshTokenNotFound) {
			t.Errorf("expected ErrRefreshTokenNotFound, got %v", err)
		}
	})

	t.Run("error deleting old token", func(t *testing.T) {
		provider, mockUserRepo, _, mockAuth, ctrl := setupTest(t)
		defer ctrl.Finish()

		refreshToken := database.RefreshToken{
			UserID:      "user-123",
			Token:       "valid-token",
			ExpiresAt:   time.Now().Add(24 * time.Hour),
			DateCreated: time.Now(),
		}

		user := database.User{
			ID:       "user-123",
			Username: "testuser",
			Email:    sql.NullString{String: "test@example.com", Valid: true},
		}

		mockUserRepo.EXPECT().
			GetRefreshTokenByToken(gomock.Any(), "valid-token").
			Return(refreshToken, nil)

		mockUserRepo.EXPECT().
			GetUserByID(gomock.Any(), "user-123").
			Return(user, nil)

		mockAuth.EXPECT().
			CreateUserClaims(gomock.Any()).
			Return(auth.Claims{UserID: "user-123", Username: "testuser"})

		mockAuth.EXPECT().
			GenerateToken(gomock.Any()).
			Return("new-access-token", nil)

		mockAuth.EXPECT().
			GenerateRefreshToken().
			Return("new-refresh-token", time.Now().Add(7*24*time.Hour), nil)

		mockUserRepo.EXPECT().
			RunWithTx(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
				return fn(ctx)
			})

		mockUserRepo.EXPECT().
			DeleteRefreshToken(gomock.Any(), "valid-token").
			Return(errors.New("database error"))

		_, err := provider.RefreshTokens(ctx, "valid-token")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("error creating new access token", func(t *testing.T) {
		provider, mockUserRepo, _, mockAuth, ctrl := setupTest(t)
		defer ctrl.Finish()

		refreshToken := database.RefreshToken{
			UserID:      "user-123",
			Token:       "valid-token",
			ExpiresAt:   time.Now().Add(24 * time.Hour),
			DateCreated: time.Now(),
		}

		user := database.User{
			ID:       "user-123",
			Username: "testuser",
			Email:    sql.NullString{String: "test@example.com", Valid: true},
		}

		mockUserRepo.EXPECT().
			GetRefreshTokenByToken(gomock.Any(), "valid-token").
			Return(refreshToken, nil)

		mockUserRepo.EXPECT().
			GetUserByID(gomock.Any(), "user-123").
			Return(user, nil)

		mockAuth.EXPECT().
			CreateUserClaims(gomock.Any()).
			Return(auth.Claims{UserID: "user-123", Username: "testuser"})

		mockAuth.EXPECT().
			GenerateToken(gomock.Any()).
			Return("", errors.New("token generation error"))

		_, err := provider.RefreshTokens(ctx, "valid-token")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("error creating new refresh token", func(t *testing.T) {
		provider, mockUserRepo, _, mockAuth, ctrl := setupTest(t)
		defer ctrl.Finish()

		refreshToken := database.RefreshToken{
			UserID:      "user-123",
			Token:       "valid-token",
			ExpiresAt:   time.Now().Add(24 * time.Hour),
			DateCreated: time.Now(),
		}

		user := database.User{
			ID:       "user-123",
			Username: "testuser",
			Email:    sql.NullString{String: "test@example.com", Valid: true},
		}

		mockUserRepo.EXPECT().
			GetRefreshTokenByToken(gomock.Any(), "valid-token").
			Return(refreshToken, nil)

		mockUserRepo.EXPECT().
			GetUserByID(gomock.Any(), "user-123").
			Return(user, nil)

		mockAuth.EXPECT().
			CreateUserClaims(gomock.Any()).
			Return(auth.Claims{UserID: "user-123", Username: "testuser"})

		mockAuth.EXPECT().
			GenerateToken(gomock.Any()).
			Return("new-access-token", nil)

		mockAuth.EXPECT().
			GenerateRefreshToken().
			Return("", time.Time{}, errors.New("refresh token generation error"))

		_, err := provider.RefreshTokens(ctx, "valid-token")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestProvider_CreateTokens(t *testing.T) {
	ctx := context.Background()

	t.Run("successful creation", func(t *testing.T) {
		provider, mockUserRepo, _, mockAuth, ctrl := setupTest(t)
		defer ctrl.Finish()

		user := model.User{
			ID:       "user-123",
			Username: "testuser",
			Email:    "test@example.com",
		}

		mockAuth.EXPECT().
			CreateUserClaims(user).
			Return(auth.Claims{UserID: "user-123", Username: "testuser"})

		mockAuth.EXPECT().
			GenerateToken(gomock.Any()).
			Return("access-token", nil)

		mockAuth.EXPECT().
			GenerateRefreshToken().
			Return("refresh-token", time.Now().Add(7*24*time.Hour), nil)

		mockUserRepo.EXPECT().
			CreateRefreshToken(gomock.Any(), gomock.Any()).
			Return(nil)

		tokens, err := provider.CreateTokens(ctx, user)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if tokens.AccessToken != "access-token" {
			t.Errorf("expected access token 'access-token', got '%s'", tokens.AccessToken)
		}
		if tokens.RefreshToken.Token != "refresh-token" {
			t.Errorf("expected refresh token 'refresh-token', got '%s'", tokens.RefreshToken.Token)
		}
	})

	t.Run("error generating access token", func(t *testing.T) {
		provider, _, _, mockAuth, ctrl := setupTest(t)
		defer ctrl.Finish()

		user := model.User{
			ID:       "user-123",
			Username: "testuser",
		}

		mockAuth.EXPECT().
			CreateUserClaims(user).
			Return(auth.Claims{UserID: "user-123"})

		mockAuth.EXPECT().
			GenerateToken(gomock.Any()).
			Return("", errors.New("token generation error"))

		_, err := provider.CreateTokens(ctx, user)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("error creating refresh token", func(t *testing.T) {
		provider, _, _, mockAuth, ctrl := setupTest(t)
		defer ctrl.Finish()

		user := model.User{
			ID:       "user-123",
			Username: "testuser",
		}

		mockAuth.EXPECT().
			CreateUserClaims(user).
			Return(auth.Claims{UserID: "user-123"})

		mockAuth.EXPECT().
			GenerateToken(gomock.Any()).
			Return("access-token", nil)

		mockAuth.EXPECT().
			GenerateRefreshToken().
			Return("", time.Time{}, errors.New("refresh token generation error"))

		_, err := provider.CreateTokens(ctx, user)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestProvider_RevokeRefreshToken(t *testing.T) {
	ctx := context.Background()

	t.Run("successful revoke", func(t *testing.T) {
		provider, mockUserRepo, _, _, ctrl := setupTest(t)
		defer ctrl.Finish()

		mockUserRepo.EXPECT().
			DeleteRefreshToken(gomock.Any(), "valid-token").
			Return(nil)

		err := provider.RevokeRefreshToken(ctx, "valid-token")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("empty token string", func(t *testing.T) {
		provider, _, _, _, ctrl := setupTest(t)
		defer ctrl.Finish()

		err := provider.RevokeRefreshToken(ctx, "")
		if err != nil {
			t.Fatalf("expected no error for empty token, got %v", err)
		}
	})

	t.Run("token not found", func(t *testing.T) {
		provider, mockUserRepo, _, _, ctrl := setupTest(t)
		defer ctrl.Finish()

		mockUserRepo.EXPECT().
			DeleteRefreshToken(gomock.Any(), "non-existent-token").
			Return(database.ErrNotFound)

		err := provider.RevokeRefreshToken(ctx, "non-existent-token")
		if err != nil {
			t.Fatalf("expected no error when token not found, got %v", err)
		}
	})

	t.Run("database error", func(t *testing.T) {
		provider, mockUserRepo, _, _, ctrl := setupTest(t)
		defer ctrl.Finish()

		dbError := errors.New("database connection error")
		mockUserRepo.EXPECT().
			DeleteRefreshToken(gomock.Any(), "some-token").
			Return(dbError)

		err := provider.RevokeRefreshToken(ctx, "some-token")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, dbError) {
			t.Errorf("expected database error, got %v", err)
		}
	})
}
