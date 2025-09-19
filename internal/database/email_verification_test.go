package database_test

import (
	"context"
	"testing"
	"time"

	"github.com/OutOfStack/game-library-auth/internal/database"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

const (
	vrfCodeTTL = 24 * time.Hour
)

func TestCreateEmailVerification_Ok(t *testing.T) {
	s := setup(t)
	defer teardown(t)

	ctx := context.Background()

	user := database.NewUser("testuser", "Test User", []byte("hashedpassword"), database.UserRoleName)
	err := s.CreateUser(ctx, user)
	require.NoError(t, err)

	verification := database.NewEmailVerification(user.ID, "hashedcode123", time.Now().Add(vrfCodeTTL))

	err = s.CreateEmailVerification(ctx, verification)
	require.NoError(t, err)

	// Verify it was created
	createdVerification, err := s.GetEmailVerificationByUserID(ctx, user.ID)
	require.NoError(t, err)
	require.Equal(t, verification.ID, createdVerification.ID)
	require.Equal(t, verification.UserID, createdVerification.UserID)
	require.Equal(t, verification.CodeHash, createdVerification.CodeHash)
}

func TestGetEmailVerificationByUserID_Ok(t *testing.T) {
	s := setup(t)
	defer teardown(t)

	ctx := context.Background()

	user := database.NewUser("testuser", "Test User", []byte("hashedpassword"), database.UserRoleName)
	err := s.CreateUser(ctx, user)
	require.NoError(t, err)

	verification := database.NewEmailVerification(user.ID, "hashedcode123", time.Now().Add(vrfCodeTTL))
	err = s.CreateEmailVerification(ctx, verification)
	require.NoError(t, err)

	foundVerification, err := s.GetEmailVerificationByUserID(ctx, user.ID)
	require.NoError(t, err)
	require.Equal(t, verification.ID, foundVerification.ID)
	require.Equal(t, verification.UserID, foundVerification.UserID)
	require.Equal(t, verification.CodeHash, foundVerification.CodeHash)
}

func TestGetEmailVerificationByUserID_NotFound(t *testing.T) {
	s := setup(t)
	defer teardown(t)

	ctx := context.Background()

	_, err := s.GetEmailVerificationByUserID(ctx, uuid.New().String())
	require.Error(t, err)
	require.Equal(t, database.ErrNotFound, err)
}

func TestGetEmailVerificationByUserID_MostRecent(t *testing.T) {
	s := setup(t)
	defer teardown(t)

	ctx := context.Background()

	user := database.NewUser("testuser", "Test User", []byte("hashedpassword"), database.UserRoleName)
	err := s.CreateUser(ctx, user)
	require.NoError(t, err)

	// Create first verification
	verification1 := database.NewEmailVerification(user.ID, "hashedcode123", time.Now().Add(vrfCodeTTL))
	err = s.CreateEmailVerification(ctx, verification1)
	require.NoError(t, err)

	// Wait a bit and create second verification
	time.Sleep(10 * time.Millisecond)
	verification2 := database.NewEmailVerification(user.ID, "hashedcode456", time.Now().Add(vrfCodeTTL))
	err = s.CreateEmailVerification(ctx, verification2)
	require.NoError(t, err)

	// Should return the most recent one
	foundVerification, err := s.GetEmailVerificationByUserID(ctx, user.ID)
	require.NoError(t, err)
	require.Equal(t, verification2.ID, foundVerification.ID)
	require.Equal(t, verification2.CodeHash, foundVerification.CodeHash)
}

func TestSetEmailVerificationMessageID_Ok(t *testing.T) {
	s := setup(t)
	defer teardown(t)

	ctx := context.Background()

	user := database.NewUser("testuser", "Test User", []byte("hashedpassword"), database.UserRoleName)
	err := s.CreateUser(ctx, user)
	require.NoError(t, err)

	verification := database.NewEmailVerification(user.ID, "hashedcode123", time.Now().Add(vrfCodeTTL))
	err = s.CreateEmailVerification(ctx, verification)
	require.NoError(t, err)

	messageID := "msg_12345"
	err = s.SetEmailVerificationMessageID(ctx, verification.ID, messageID)
	require.NoError(t, err)

	// Verify message ID was set
	updatedVerification, err := s.GetEmailVerificationByUserID(ctx, user.ID)
	require.NoError(t, err)
	require.Equal(t, messageID, updatedVerification.MessageID.String)
}

func TestSetEmailVerificationUsed_Ok(t *testing.T) {
	s := setup(t)
	defer teardown(t)

	ctx := context.Background()

	user := database.NewUser("testuser", "Test User", []byte("hashedpassword"), database.UserRoleName)
	err := s.CreateUser(ctx, user)
	require.NoError(t, err)

	verification := database.NewEmailVerification(user.ID, "hashedcode123", time.Now().Add(vrfCodeTTL))
	err = s.CreateEmailVerification(ctx, verification)
	require.NoError(t, err)

	err = s.SetEmailVerificationUsed(ctx, verification.ID, true)
	require.NoError(t, err)

	// Verify verification is no longer returned by GetEmailVerificationByUserID
	_, err = s.GetEmailVerificationByUserID(ctx, user.ID)
	require.Error(t, err)
	require.Equal(t, database.ErrNotFound, err)
}

func TestEmailVerification_IsExpired_True(t *testing.T) {
	// Expired 1 hour ago
	verification := database.NewEmailVerification(uuid.New().String(), "hashedcode123", time.Now().Add(-1*time.Hour))

	require.True(t, verification.IsExpired())
}

func TestEmailVerification_IsExpired_False(t *testing.T) {
	// Expires in 1 hour
	verification := database.NewEmailVerification(uuid.New().String(), "hashedcode123", time.Now().Add(1*time.Hour))

	require.False(t, verification.IsExpired())
}
