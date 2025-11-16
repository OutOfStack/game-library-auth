package database_test

import (
	"context"
	"testing"
	"time"

	"github.com/OutOfStack/game-library-auth/internal/database"
	"github.com/OutOfStack/game-library-auth/internal/model"
	"github.com/stretchr/testify/require"
)

func TestCreateRefreshToken_Ok(t *testing.T) {
	s := setup(t)
	defer teardown(t)

	ctx := context.Background()

	user := database.NewUser("testuser", "Test User", []byte("hashedpassword"), model.UserRoleName)
	err := s.CreateUser(ctx, user)
	require.NoError(t, err)

	refreshToken := database.NewRefreshToken(user.ID, "test-refresh-token-abc123", time.Now().Add(24*time.Hour))

	err = s.CreateRefreshToken(ctx, refreshToken)
	require.NoError(t, err)

	foundToken, err := s.GetRefreshTokenByHash(ctx, refreshToken.TokenHash)
	require.NoError(t, err)
	require.Equal(t, refreshToken.ID, foundToken.ID)
	require.Equal(t, refreshToken.UserID, foundToken.UserID)
	require.Equal(t, refreshToken.TokenHash, foundToken.TokenHash)
}

func TestGetRefreshTokenByToken_Ok(t *testing.T) {
	s := setup(t)
	defer teardown(t)

	ctx := context.Background()

	user := database.NewUser("testuser", "Test User", []byte("hashedpassword"), model.UserRoleName)
	err := s.CreateUser(ctx, user)
	require.NoError(t, err)

	refreshToken := database.NewRefreshToken(user.ID, "test-refresh-token-xyz789", time.Now().Add(24*time.Hour))
	err = s.CreateRefreshToken(ctx, refreshToken)
	require.NoError(t, err)

	foundToken, err := s.GetRefreshTokenByHash(ctx, refreshToken.TokenHash)
	require.NoError(t, err)
	require.Equal(t, refreshToken.ID, foundToken.ID)
	require.Equal(t, refreshToken.UserID, foundToken.UserID)
	require.Equal(t, refreshToken.TokenHash, foundToken.TokenHash)
	require.False(t, foundToken.IsExpired())
}

func TestGetRefreshTokenByToken_NotFound(t *testing.T) {
	s := setup(t)
	defer teardown(t)

	ctx := context.Background()

	_, err := s.GetRefreshTokenByHash(ctx, "non-existent-token")
	require.Error(t, err)
	require.Equal(t, database.ErrNotFound, err)
}

func TestDeleteRefreshToken_Ok(t *testing.T) {
	s := setup(t)
	defer teardown(t)

	ctx := context.Background()

	user := database.NewUser("testuser", "Test User", []byte("hashedpassword"), model.UserRoleName)
	err := s.CreateUser(ctx, user)
	require.NoError(t, err)

	refreshToken := database.NewRefreshToken(user.ID, "test-refresh-token-delete", time.Now().Add(24*time.Hour))
	err = s.CreateRefreshToken(ctx, refreshToken)
	require.NoError(t, err)

	err = s.DeleteRefreshToken(ctx, refreshToken.TokenHash)
	require.NoError(t, err)

	_, err = s.GetRefreshTokenByHash(ctx, refreshToken.TokenHash)
	require.Error(t, err)
	require.Equal(t, database.ErrNotFound, err)
}

func TestDeleteRefreshTokensByUserID_Ok(t *testing.T) {
	s := setup(t)
	defer teardown(t)

	ctx := context.Background()

	user := database.NewUser("testuser", "Test User", []byte("hashedpassword"), model.UserRoleName)
	err := s.CreateUser(ctx, user)
	require.NoError(t, err)

	token1 := database.NewRefreshToken(user.ID, "test-refresh-token-1", time.Now().Add(24*time.Hour))
	token2 := database.NewRefreshToken(user.ID, "test-refresh-token-2", time.Now().Add(24*time.Hour))

	err = s.CreateRefreshToken(ctx, token1)
	require.NoError(t, err)
	err = s.CreateRefreshToken(ctx, token2)
	require.NoError(t, err)

	err = s.DeleteRefreshTokensByUserID(ctx, user.ID)
	require.NoError(t, err)

	_, err = s.GetRefreshTokenByHash(ctx, token1.TokenHash)
	require.Error(t, err)
	require.Equal(t, database.ErrNotFound, err)

	_, err = s.GetRefreshTokenByHash(ctx, token2.TokenHash)
	require.Error(t, err)
	require.Equal(t, database.ErrNotFound, err)
}

func TestDeleteExpiredRefreshTokens_Ok(t *testing.T) {
	s := setup(t)
	defer teardown(t)

	ctx := context.Background()

	user := database.NewUser("testuser", "Test User", []byte("hashedpassword"), model.UserRoleName)
	err := s.CreateUser(ctx, user)
	require.NoError(t, err)

	expiredToken := database.NewRefreshToken(user.ID, "test-expired-token", time.Now().Add(-1*time.Hour))
	validToken := database.NewRefreshToken(user.ID, "test-valid-token", time.Now().Add(24*time.Hour))

	err = s.CreateRefreshToken(ctx, expiredToken)
	require.NoError(t, err)
	err = s.CreateRefreshToken(ctx, validToken)
	require.NoError(t, err)

	err = s.DeleteExpiredRefreshTokens(ctx)
	require.NoError(t, err)

	_, err = s.GetRefreshTokenByHash(ctx, expiredToken.TokenHash)
	require.Error(t, err)
	require.Equal(t, database.ErrNotFound, err)

	foundToken, err := s.GetRefreshTokenByHash(ctx, validToken.TokenHash)
	require.NoError(t, err)
	require.Equal(t, validToken.TokenHash, foundToken.TokenHash)
}

func TestRefreshToken_IsExpired(t *testing.T) {
	tests := []struct {
		name      string
		expiresAt time.Time
		expected  bool
	}{
		{
			name:      "expired token",
			expiresAt: time.Now().Add(-1 * time.Hour),
			expected:  true,
		},
		{
			name:      "valid token",
			expiresAt: time.Now().Add(24 * time.Hour),
			expected:  false,
		},
		{
			name:      "token expiring now",
			expiresAt: time.Now(),
			expected:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := database.NewRefreshToken("user-id", "token", tt.expiresAt)
			require.Equal(t, tt.expected, token.IsExpired())
		})
	}
}
