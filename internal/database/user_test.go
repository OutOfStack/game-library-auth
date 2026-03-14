package database_test

import (
	"testing"

	"github.com/OutOfStack/game-library-auth/internal/database"
	"github.com/OutOfStack/game-library-auth/internal/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestCreateUser_Ok(t *testing.T) {
	s := setup(t)
	defer teardown(t)

	ctx := t.Context()

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

	ctx := t.Context()

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

	ctx := t.Context()

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

	ctx := t.Context()

	_, err := s.GetUserByID(ctx, uuid.New().String())
	require.Error(t, err)
	require.Equal(t, database.ErrNotFound, err)
}

func TestGetUserByUsername_Ok(t *testing.T) {
	s := setup(t)
	defer teardown(t)

	ctx := t.Context()

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

	ctx := t.Context()

	_, err := s.GetUserByUsername(ctx, "nonexistent")
	require.Error(t, err)
	require.Equal(t, database.ErrNotFound, err)
}

func TestCheckUserExists_True(t *testing.T) {
	s := setup(t)
	defer teardown(t)

	ctx := t.Context()

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

	ctx := t.Context()

	exists, err := s.CheckUserExists(ctx, "Nonexistent User", model.UserRoleName)
	require.NoError(t, err)
	require.False(t, exists)
}

func TestGetUserByOAuthLink_Ok(t *testing.T) {
	s := setup(t)
	defer teardown(t)

	ctx := t.Context()

	user := database.NewUser("testuser", "Test User", []byte("hashedpassword"), model.UserRoleName)
	err := s.CreateUser(ctx, user)
	require.NoError(t, err)

	link := database.NewUserOAuthLink(user.ID, "google", "google123456")
	err = s.CreateUserOAuthLink(ctx, link)
	require.NoError(t, err)

	foundUser, err := s.GetUserByOAuthLink(ctx, "google", "google123456")
	require.NoError(t, err)
	require.Equal(t, user.ID, foundUser.ID)
	require.Equal(t, user.Username, foundUser.Username)
}

func TestGetUserByOAuthLink_NotFound(t *testing.T) {
	s := setup(t)
	defer teardown(t)

	ctx := t.Context()

	_, err := s.GetUserByOAuthLink(ctx, "google", "nonexistent")
	require.Error(t, err)
	require.Equal(t, database.ErrNotFound, err)
}

func TestCreateUserOAuthLink_Ok(t *testing.T) {
	s := setup(t)
	defer teardown(t)

	ctx := t.Context()

	user := database.NewUser("testuser", "Test User", []byte("hashedpassword"), model.UserRoleName)
	err := s.CreateUser(ctx, user)
	require.NoError(t, err)

	link := database.NewUserOAuthLink(user.ID, "github", "gh123")
	err = s.CreateUserOAuthLink(ctx, link)
	require.NoError(t, err)

	foundUser, err := s.GetUserByOAuthLink(ctx, "github", "gh123")
	require.NoError(t, err)
	require.Equal(t, user.ID, foundUser.ID)
}

func TestCreateUserOAuthLink_Duplicate(t *testing.T) {
	s := setup(t)
	defer teardown(t)

	ctx := t.Context()

	user := database.NewUser("testuser", "Test User", []byte("hashedpassword"), model.UserRoleName)
	err := s.CreateUser(ctx, user)
	require.NoError(t, err)

	link := database.NewUserOAuthLink(user.ID, "google", "google123")
	err = s.CreateUserOAuthLink(ctx, link)
	require.NoError(t, err)

	// duplicate should be ignored (idempotent)
	link2 := database.NewUserOAuthLink(user.ID, "google", "google123")
	err = s.CreateUserOAuthLink(ctx, link2)
	require.NoError(t, err)
}

func TestHasOAuthLink_True(t *testing.T) {
	s := setup(t)
	defer teardown(t)

	ctx := t.Context()

	user := database.NewUser("testuser", "Test User", []byte("hashedpassword"), model.UserRoleName)
	err := s.CreateUser(ctx, user)
	require.NoError(t, err)

	link := database.NewUserOAuthLink(user.ID, "google", "google123")
	err = s.CreateUserOAuthLink(ctx, link)
	require.NoError(t, err)

	has, err := s.HasOAuthLink(ctx, user.ID)
	require.NoError(t, err)
	require.True(t, has)
}

func TestHasOAuthLink_False(t *testing.T) {
	s := setup(t)
	defer teardown(t)

	ctx := t.Context()

	user := database.NewUser("testuser", "Test User", []byte("hashedpassword"), model.UserRoleName)
	err := s.CreateUser(ctx, user)
	require.NoError(t, err)

	has, err := s.HasOAuthLink(ctx, user.ID)
	require.NoError(t, err)
	require.False(t, has)
}

func TestUpdateUser_Ok(t *testing.T) {
	s := setup(t)
	defer teardown(t)

	ctx := t.Context()

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

	ctx := t.Context()

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

	ctx := t.Context()

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

	ctx := t.Context()

	_, err := s.GetUserByEmail(ctx, "nonexistent@example.com")
	require.Error(t, err)
	require.Equal(t, database.ErrNotFound, err)
}

func TestUpdateUserEmail_Ok(t *testing.T) {
	s := setup(t)
	defer teardown(t)

	ctx := t.Context()

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
