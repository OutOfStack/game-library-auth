package database_test

import (
	"context"
	"testing"

	"github.com/OutOfStack/game-library-auth/internal/database"
	"github.com/OutOfStack/game-library-auth/internal/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestCreateUser_Ok(t *testing.T) {
	s := setup(t)
	defer teardown(t)

	ctx := context.Background()

	user := database.NewUser("testuser", "Test User", []byte("hashedpassword"), model.UserRoleName)
	user.SetEmail("test@example.com", false)

	err := s.CreateUser(ctx, user)
	require.NoError(t, err)

	// Verify user was created
	createdUser, err := s.GetUserByID(ctx, user.ID)
	require.NoError(t, err)
	require.Equal(t, user.Username, createdUser.Username)
	require.Equal(t, user.DisplayName, createdUser.DisplayName)
	require.Equal(t, user.Email, createdUser.Email)
	require.Equal(t, user.Role, createdUser.Role)
}

func TestCreateUser_DuplicateUsername(t *testing.T) {
	s := setup(t)
	defer teardown(t)

	ctx := context.Background()

	user1 := database.NewUser("testuser", "Test User 1", []byte("hashedpassword1"), model.UserRoleName)
	user2 := database.NewUser("testuser", "Test User 2", []byte("hashedpassword2"), model.UserRoleName)

	err := s.CreateUser(ctx, user1)
	require.NoError(t, err)

	err = s.CreateUser(ctx, user2)
	require.Error(t, err)
	require.Equal(t, database.ErrUserExists, err)
}

func TestGetUserByID_Ok(t *testing.T) {
	s := setup(t)
	defer teardown(t)

	ctx := context.Background()

	user := database.NewUser("testuser", "Test User", []byte("hashedpassword"), model.UserRoleName)
	user.SetEmail("test@example.com", true)
	err := s.CreateUser(ctx, user)
	require.NoError(t, err)

	foundUser, err := s.GetUserByID(ctx, user.ID)
	require.NoError(t, err)
	require.Equal(t, user.ID, foundUser.ID)
	require.Equal(t, user.Username, foundUser.Username)
	require.Equal(t, user.DisplayName, foundUser.DisplayName)
	require.Equal(t, user.Email, foundUser.Email)
	require.Equal(t, user.EmailVerified, foundUser.EmailVerified)
	require.Equal(t, user.Role, foundUser.Role)
}

func TestGetUserByID_NotFound(t *testing.T) {
	s := setup(t)
	defer teardown(t)

	ctx := context.Background()

	_, err := s.GetUserByID(ctx, uuid.New().String())
	require.Error(t, err)
	require.Equal(t, database.ErrNotFound, err)
}

func TestGetUserByUsername_Ok(t *testing.T) {
	s := setup(t)
	defer teardown(t)

	ctx := context.Background()

	user := database.NewUser("testuser", "Test User", []byte("hashedpassword"), model.PublisherRoleName)
	err := s.CreateUser(ctx, user)
	require.NoError(t, err)

	foundUser, err := s.GetUserByUsername(ctx, user.Username)
	require.NoError(t, err)
	require.Equal(t, user.ID, foundUser.ID)
	require.Equal(t, user.Username, foundUser.Username)
	require.Equal(t, user.Role, foundUser.Role)
}

func TestGetUserByUsername_NotFound(t *testing.T) {
	s := setup(t)
	defer teardown(t)

	ctx := context.Background()

	_, err := s.GetUserByUsername(ctx, "nonexistent")
	require.Error(t, err)
	require.Equal(t, database.ErrNotFound, err)
}

func TestCheckUserExists_True(t *testing.T) {
	s := setup(t)
	defer teardown(t)

	ctx := context.Background()

	user := database.NewUser("testuser", "Test User", []byte("hashedpassword"), model.UserRoleName)
	err := s.CreateUser(ctx, user)
	require.NoError(t, err)

	exists, err := s.CheckUserExists(ctx, user.DisplayName, user.Role)
	require.NoError(t, err)
	require.True(t, exists)
}

func TestCheckUserExists_False(t *testing.T) {
	s := setup(t)
	defer teardown(t)

	ctx := context.Background()

	exists, err := s.CheckUserExists(ctx, "Nonexistent User", model.UserRoleName)
	require.NoError(t, err)
	require.False(t, exists)
}

func TestGetUserByOAuth_Ok(t *testing.T) {
	s := setup(t)
	defer teardown(t)

	ctx := context.Background()

	user := database.NewUser("testuser", "Test User", []byte("hashedpassword"), model.UserRoleName)
	user.SetOAuthID("google", "google123456")
	err := s.CreateUser(ctx, user)
	require.NoError(t, err)

	foundUser, err := s.GetUserByOAuth(ctx, "google", "google123456")
	require.NoError(t, err)
	require.Equal(t, user.ID, foundUser.ID)
	require.Equal(t, user.Username, foundUser.Username)
	require.Equal(t, user.OAuthProvider, foundUser.OAuthProvider)
	require.Equal(t, user.OAuthID, foundUser.OAuthID)
}

func TestGetUserByOAuth_NotFound(t *testing.T) {
	s := setup(t)
	defer teardown(t)

	ctx := context.Background()

	_, err := s.GetUserByOAuth(ctx, "google", "nonexistent")
	require.Error(t, err)
	require.Equal(t, database.ErrNotFound, err)
}

func TestUpdateUser_Ok(t *testing.T) {
	s := setup(t)
	defer teardown(t)

	ctx := context.Background()

	user := database.NewUser("testuser", "Test User", []byte("hashedpassword"), model.UserRoleName)
	err := s.CreateUser(ctx, user)
	require.NoError(t, err)

	user.DisplayName = "Updated Name"
	user.PasswordHash = []byte("newhashedpassword")
	err = s.UpdateUser(ctx, user)
	require.NoError(t, err)

	updatedUser, err := s.GetUserByID(ctx, user.ID)
	require.NoError(t, err)
	require.Equal(t, "Updated Name", updatedUser.DisplayName)
	require.Equal(t, []byte("newhashedpassword"), updatedUser.PasswordHash)
	require.NotNil(t, updatedUser.DateUpdated)
}

func TestDeleteUser_Ok(t *testing.T) {
	s := setup(t)
	defer teardown(t)

	ctx := context.Background()

	user := database.NewUser("testuser", "Test User", []byte("hashedpassword"), model.UserRoleName)
	err := s.CreateUser(ctx, user)
	require.NoError(t, err)

	err = s.DeleteUser(ctx, user.ID)
	require.NoError(t, err)

	_, err = s.GetUserByID(ctx, user.ID)
	require.Error(t, err)
	require.Equal(t, database.ErrNotFound, err)
}

func TestGetUserByEmail_Ok(t *testing.T) {
	s := setup(t)
	defer teardown(t)

	ctx := context.Background()

	user := database.NewUser("testuser", "Test User", []byte("hashedpassword"), model.UserRoleName)
	user.SetEmail("test@example.com", true)
	err := s.CreateUser(ctx, user)
	require.NoError(t, err)

	foundUser, err := s.GetUserByEmail(ctx, "test@example.com")
	require.NoError(t, err)
	require.Equal(t, user.ID, foundUser.ID)
	require.Equal(t, user.Email, foundUser.Email)
	require.Equal(t, user.EmailVerified, foundUser.EmailVerified)
}

func TestGetUserByEmail_NotFound(t *testing.T) {
	s := setup(t)
	defer teardown(t)

	ctx := context.Background()

	_, err := s.GetUserByEmail(ctx, "nonexistent@example.com")
	require.Error(t, err)
	require.Equal(t, database.ErrNotFound, err)
}

func TestUpdateUserEmail_Ok(t *testing.T) {
	s := setup(t)
	defer teardown(t)

	ctx := context.Background()

	user := database.NewUser("testuser", "Test User", []byte("hashedpassword"), model.UserRoleName)
	err := s.CreateUser(ctx, user)
	require.NoError(t, err)

	err = s.SetUserEmailVerified(ctx, user.ID)
	require.NoError(t, err)

	updatedUser, err := s.GetUserByID(ctx, user.ID)
	require.NoError(t, err)
	require.True(t, updatedUser.EmailVerified)
	require.NotNil(t, updatedUser.DateUpdated)
}
