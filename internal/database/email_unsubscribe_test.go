package database_test

import (
	"context"
	"testing"
	"time"

	"github.com/OutOfStack/game-library-auth/internal/database"
	"github.com/OutOfStack/game-library-auth/internal/model"
	"github.com/stretchr/testify/require"
)

func TestCreateEmailUnsubscribe_Ok(t *testing.T) {
	s := setup(t)
	defer teardown(t)

	ctx := context.Background()

	email := "test@example.com"
	unsubscribe := database.NewEmailUnsubscribe(email)

	err := s.CreateEmailUnsubscribe(ctx, unsubscribe)
	require.NoError(t, err)

	isUnsubscribed, err := s.IsEmailUnsubscribed(ctx, email)
	require.NoError(t, err)
	require.True(t, isUnsubscribed)
}

func TestCreateEmailUnsubscribe_Duplicate(t *testing.T) {
	s := setup(t)
	defer teardown(t)

	ctx := context.Background()

	email := "duplicate@example.com"
	unsubscribe1 := database.NewEmailUnsubscribe(email)
	unsubscribe2 := database.NewEmailUnsubscribe(email)

	err := s.CreateEmailUnsubscribe(ctx, unsubscribe1)
	require.NoError(t, err)

	err = s.CreateEmailUnsubscribe(ctx, unsubscribe2)
	require.NoError(t, err)

	isUnsubscribed, err := s.IsEmailUnsubscribed(ctx, email)
	require.NoError(t, err)
	require.True(t, isUnsubscribed)
}

func TestIsEmailUnsubscribed_NotUnsubscribed(t *testing.T) {
	s := setup(t)
	defer teardown(t)

	ctx := context.Background()

	isUnsubscribed, err := s.IsEmailUnsubscribed(ctx, "notunsubscribed@example.com")
	require.NoError(t, err)
	require.False(t, isUnsubscribed)
}

func TestIsEmailUnsubscribed_Unsubscribed(t *testing.T) {
	s := setup(t)
	defer teardown(t)

	ctx := context.Background()

	email := "unsubscribed@example.com"
	unsubscribe := database.NewEmailUnsubscribe(email)

	err := s.CreateEmailUnsubscribe(ctx, unsubscribe)
	require.NoError(t, err)

	isUnsubscribed, err := s.IsEmailUnsubscribed(ctx, email)
	require.NoError(t, err)
	require.True(t, isUnsubscribed)
}

func TestSetUnsubscribeToken_Ok(t *testing.T) {
	s := setup(t)
	defer teardown(t)

	ctx := context.Background()

	user := database.NewUser("testuser", "Test User", []byte("hashedpassword"), model.UserRoleName)
	user.SetEmail("test@example.com", false)
	err := s.CreateUser(ctx, user)
	require.NoError(t, err)

	verification := database.NewEmailVerification(user.ID, "hashedcode123", "initial-token", time.Now())
	err = s.CreateEmailVerification(ctx, verification)
	require.NoError(t, err)

	newToken := "new-unsubscribe-token"
	err = s.SetUnsubscribeToken(ctx, verification.ID, newToken)
	require.NoError(t, err)
}

func TestGetEmailVerificationByUnsubscribeToken_Ok(t *testing.T) {
	s := setup(t)
	defer teardown(t)

	ctx := context.Background()

	user := database.NewUser("testuser", "Test User", []byte("hashedpassword"), model.UserRoleName)
	user.SetEmail("test@example.com", false)
	err := s.CreateUser(ctx, user)
	require.NoError(t, err)

	token := "test-unsubscribe-token"
	verification := database.NewEmailVerification(user.ID, "hashedcode123", token, time.Now())
	err = s.CreateEmailVerification(ctx, verification)
	require.NoError(t, err)

	foundVerification, err := s.GetEmailVerificationByUnsubscribeToken(ctx, token)
	require.NoError(t, err)
	require.Equal(t, verification.ID, foundVerification.ID)
	require.Equal(t, verification.UserID, foundVerification.UserID)
}

func TestGetEmailVerificationByUnsubscribeToken_NotFound(t *testing.T) {
	s := setup(t)
	defer teardown(t)

	ctx := context.Background()

	_, err := s.GetEmailVerificationByUnsubscribeToken(ctx, "nonexistent-token")
	require.Error(t, err)
	require.Equal(t, database.ErrNotFound, err)
}

func TestGetEmailVerificationByUnsubscribeToken_UserWithoutEmail(t *testing.T) {
	s := setup(t)
	defer teardown(t)

	ctx := context.Background()

	user := database.NewUser("testuser", "Test User", []byte("hashedpassword"), model.UserRoleName)
	err := s.CreateUser(ctx, user)
	require.NoError(t, err)

	token := "test-unsubscribe-token"
	verification := database.NewEmailVerification(user.ID, "hashedcode123", token, time.Now())
	err = s.CreateEmailVerification(ctx, verification)
	require.NoError(t, err)

	_, err = s.GetEmailVerificationByUnsubscribeToken(ctx, token)
	require.Error(t, err)
	require.Equal(t, database.ErrNotFound, err)
}
