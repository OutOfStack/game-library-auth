package database_test

import (
	"testing"
	"time"
	"uuid"

	"github.com/OutOfStack/game-library-auth/internal/database"
	"github.com/OutOfStack/game-library-auth/internal/model"
	"github.com/stretchr/testify/require"
)

func TestCreateEmailVerification_Ok(t *testing.T) {
	s := setup(t)
	defer teardown(t)

	ctx := t.Context()

	user := database.NewUser("testuser", "Test User", []byte("hashedpassword"), model.UserRoleName)
	err := s.CreateUser(ctx, user)
	require.NoError(t, err)

	verification := database.NewEmailVerification(user.ID, "hashedcode123", "unsubscribe_token", time.Now())

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

	ctx := t.Context()

	user := database.NewUser("testuser", "Test User", []byte("hashedpassword"), model.UserRoleName)
	err := s.CreateUser(ctx, user)
	require.NoError(t, err)

	verification := database.NewEmailVerification(user.ID, "hashedcode123", "unsubscribe_token", time.Now())
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

	ctx := t.Context()

	_, err := s.GetEmailVerificationByUserID(ctx, uuid.New().String())
	require.Error(t, err)
	require.Equal(t, database.ErrNotFound, err)
}

func TestGetEmailVerificationByUserID_MostRecent(t *testing.T) {
	s := setup(t)
	defer teardown(t)

	ctx := t.Context()

	user := database.NewUser("testuser", "Test User", []byte("hashedpassword"), model.UserRoleName)
	err := s.CreateUser(ctx, user)
	require.NoError(t, err)

	// Create first verification
	verification1 := database.NewEmailVerification(user.ID, "hashedcode123", "unsubscribe_token1", time.Now())
	err = s.CreateEmailVerification(ctx, verification1)
	require.NoError(t, err)

	// Wait a bit and create second verification
	time.Sleep(10 * time.Millisecond)
	verification2 := database.NewEmailVerification(user.ID, "hashedcode456", "unsubscribe_token2", time.Now())
	err = s.CreateEmailVerification(ctx, verification2)
	require.NoError(t, err)

	// Should return the most recent one
	foundVerification, err := s.GetEmailVerificationByUserID(ctx, user.ID)
	require.NoError(t, err)
	require.Equal(t, verification2.ID, foundVerification.ID)
	require.Equal(t, verification2.CodeHash, foundVerification.CodeHash)
}

func TestGetVerificationState(t *testing.T) {
	s := setup(t)
	defer teardown(t)

	ctx := t.Context()
	email := "state@example.com"
	user := database.NewUser("testuser", "Test User", []byte("hashedpassword"), model.PublisherRoleName)
	require.NoError(t, s.CreateUser(ctx, user))

	state, err := s.GetVerificationState(ctx, user.ID, email)
	require.NoError(t, err)
	require.False(t, state.LatestID.Valid)
	require.False(t, state.Unsubscribed)
	require.WithinDuration(t, time.Now(), state.Now, time.Minute)

	older := database.NewEmailVerification(user.ID, "hash1", "token1", time.Now().Add(-time.Hour))
	latest := database.NewEmailVerification(user.ID, "hash2", "token2", time.Now())
	require.NoError(t, s.CreateEmailVerification(ctx, older))
	require.NoError(t, s.CreateEmailVerification(ctx, latest))
	require.NoError(t, s.CreateEmailUnsubscribe(ctx, database.NewEmailUnsubscribe(email)))

	state, err = s.GetVerificationState(ctx, user.ID, email)
	require.NoError(t, err)
	require.Equal(t, latest.ID, state.LatestID.String)
	require.WithinDuration(t, latest.DateCreated, state.LatestCreated.Time, time.Millisecond)
	require.True(t, state.Unsubscribed)

	require.NoError(t, s.SetEmailVerificationUsed(ctx, latest.ID, true))
	state, err = s.GetVerificationState(ctx, user.ID, email)
	require.NoError(t, err)
	require.Equal(t, older.ID, state.LatestID.String, "used codes are skipped")
}

func TestSetEmailVerificationUsed_Ok(t *testing.T) {
	s := setup(t)
	defer teardown(t)

	ctx := t.Context()

	user := database.NewUser("testuser", "Test User", []byte("hashedpassword"), model.UserRoleName)
	err := s.CreateUser(ctx, user)
	require.NoError(t, err)

	verification := database.NewEmailVerification(user.ID, "hashedcode123", "unsubscribe_token", time.Now())
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
	// Create verification with old date (expired)
	verification := database.NewEmailVerification(uuid.New().String(), "hashedcode123", "unsubscribe_token", time.Now().Add(-25*time.Hour))

	require.True(t, verification.IsExpired())
}

func TestEmailVerification_IsExpired_False(t *testing.T) {
	// Create verification with current time (not expired)
	verification := database.NewEmailVerification(uuid.New().String(), "hashedcode123", "unsubscribe_token", time.Now())

	require.False(t, verification.IsExpired())
}
