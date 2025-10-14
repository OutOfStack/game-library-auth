package facade_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/OutOfStack/game-library-auth/internal/database"
	"github.com/OutOfStack/game-library-auth/internal/facade"
	"github.com/OutOfStack/game-library-auth/internal/model"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
)

func TestProvider_GoogleOAuth(t *testing.T) {
	ctx := context.Background()

	t.Run("existing user with oauth", func(t *testing.T) {
		provider, mockUserRepo, _, ctrl := setupTest(t)
		defer ctrl.Finish()

		expectedUser := database.User{
			ID:       "user-123",
			Username: "testuser",
			Email:    sql.NullString{String: "test@example.com", Valid: true},
			Role:     model.UserRoleName,
		}

		mockUserRepo.EXPECT().
			GetUserByOAuth(ctx, model.GoogleAuthTokenProvider, "oauth-123").
			Return(expectedUser, nil)

		result, err := provider.GoogleOAuth(ctx, "oauth-123", "test@example.com")

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if result.ID != expectedUser.ID {
			t.Errorf("expected user ID %s, got %s", expectedUser.ID, result.ID)
		}
	})

	t.Run("new user creation", func(t *testing.T) {
		provider, mockUserRepo, _, ctrl := setupTest(t)
		defer ctrl.Finish()

		mockUserRepo.EXPECT().
			GetUserByOAuth(ctx, model.GoogleAuthTokenProvider, "oauth-123").
			Return(database.User{}, database.ErrNotFound)

		mockUserRepo.EXPECT().
			CreateUser(ctx, gomock.Any()).
			Return(nil)

		result, err := provider.GoogleOAuth(ctx, "oauth-123", "newuser@example.com")

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if result.Username != "newuser" {
			t.Errorf("expected username newuser, got %s", result.Username)
		}
	})

	t.Run("invalid email", func(t *testing.T) {
		provider, mockUserRepo, _, ctrl := setupTest(t)
		defer ctrl.Finish()

		mockUserRepo.EXPECT().
			GetUserByOAuth(ctx, model.GoogleAuthTokenProvider, "oauth-123").
			Return(database.User{}, database.ErrNotFound)

		_, err := provider.GoogleOAuth(ctx, "oauth-123", "invalid-email")

		if !errors.Is(err, facade.ErrInvalidEmail) {
			t.Errorf("expected ErrInvalidEmail, got %v", err)
		}
	})

	t.Run("username conflict on creation", func(t *testing.T) {
		provider, mockUserRepo, _, ctrl := setupTest(t)
		defer ctrl.Finish()

		mockUserRepo.EXPECT().
			GetUserByOAuth(ctx, model.GoogleAuthTokenProvider, "oauth-123").
			Return(database.User{}, database.ErrNotFound)

		mockUserRepo.EXPECT().
			CreateUser(ctx, gomock.Any()).
			Return(database.ErrUserExists)

		_, err := provider.GoogleOAuth(ctx, "oauth-123", "existing@example.com")

		if !errors.Is(err, facade.ErrOAuthSignInConflict) {
			t.Errorf("expected ErrOAuthSignInConflict, got %v", err)
		}
	})
}

func TestProvider_UpdateUserProfile(t *testing.T) {
	ctx := context.Background()

	t.Run("update name only", func(t *testing.T) {
		provider, mockUserRepo, _, ctrl := setupTest(t)
		defer ctrl.Finish()

		existingUser := database.User{
			ID:          "user-123",
			Username:    "testuser",
			DisplayName: "Old Name",
			Email:       sql.NullString{String: "test@example.com", Valid: true},
			Role:        model.UserRoleName,
		}

		newName := "New Name"
		params := model.UpdateProfileParams{
			Name: &newName,
		}

		mockUserRepo.EXPECT().
			RunWithTx(ctx, gomock.Any()).
			DoAndReturn(func(ctx context.Context, f func(context.Context) error) error {
				return f(ctx)
			})

		mockUserRepo.EXPECT().
			GetUserByID(ctx, "user-123").
			Return(existingUser, nil)

		mockUserRepo.EXPECT().
			UpdateUser(ctx, gomock.Any()).
			Return(nil)

		result, err := provider.UpdateUserProfile(ctx, "user-123", params)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if result.DisplayName != "New Name" {
			t.Errorf("expected display name 'New Name', got %s", result.DisplayName)
		}
	})

	t.Run("update password", func(t *testing.T) {
		provider, mockUserRepo, _, ctrl := setupTest(t)
		defer ctrl.Finish()

		oldPassword := "oldpass"
		oldPasswordHash, _ := bcrypt.GenerateFromPassword([]byte(oldPassword), bcrypt.DefaultCost)

		existingUser := database.User{
			ID:           "user-123",
			Username:     "testuser",
			PasswordHash: oldPasswordHash,
			Role:         model.UserRoleName,
		}

		newPassword := "newpass"
		params := model.UpdateProfileParams{
			Password:    &oldPassword,
			NewPassword: &newPassword,
		}

		mockUserRepo.EXPECT().
			RunWithTx(ctx, gomock.Any()).
			DoAndReturn(func(ctx context.Context, f func(context.Context) error) error {
				return f(ctx)
			})

		mockUserRepo.EXPECT().
			GetUserByID(ctx, "user-123").
			Return(existingUser, nil)

		mockUserRepo.EXPECT().
			UpdateUser(ctx, gomock.Any()).
			Return(nil)

		_, err := provider.UpdateUserProfile(ctx, "user-123", params)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("user not found", func(t *testing.T) {
		provider, mockUserRepo, _, ctrl := setupTest(t)
		defer ctrl.Finish()

		mockUserRepo.EXPECT().
			RunWithTx(ctx, gomock.Any()).
			DoAndReturn(func(ctx context.Context, f func(context.Context) error) error {
				return f(ctx)
			})

		mockUserRepo.EXPECT().
			GetUserByID(ctx, "nonexistent").
			Return(database.User{}, database.ErrNotFound)

		params := model.UpdateProfileParams{}
		_, err := provider.UpdateUserProfile(ctx, "nonexistent", params)

		if !errors.Is(err, facade.ErrUpdateProfileUserNotFound) {
			t.Errorf("expected ErrUpdateProfileUserNotFound, got %v", err)
		}
	})

	t.Run("password change not allowed for oauth users", func(t *testing.T) {
		provider, mockUserRepo, _, ctrl := setupTest(t)
		defer ctrl.Finish()

		existingUser := database.User{
			ID:            "user-123",
			Username:      "testuser",
			OAuthProvider: sql.NullString{String: "google", Valid: true},
			Role:          model.UserRoleName,
		}

		oldPassword := "oldpass"
		newPassword := "newpass"
		params := model.UpdateProfileParams{
			Password:    &oldPassword,
			NewPassword: &newPassword,
		}

		mockUserRepo.EXPECT().
			RunWithTx(ctx, gomock.Any()).
			DoAndReturn(func(ctx context.Context, f func(context.Context) error) error {
				return f(ctx)
			})

		mockUserRepo.EXPECT().
			GetUserByID(ctx, "user-123").
			Return(existingUser, nil)

		_, err := provider.UpdateUserProfile(ctx, "user-123", params)

		if !errors.Is(err, facade.ErrUpdateProfileNotAllowed) {
			t.Errorf("expected ErrUpdateProfileNotAllowed, got %v", err)
		}
	})

	t.Run("invalid current password", func(t *testing.T) {
		provider, mockUserRepo, _, ctrl := setupTest(t)
		defer ctrl.Finish()

		wrongPassword := "wrongpass"
		correctPasswordHash, _ := bcrypt.GenerateFromPassword([]byte("correctpass"), bcrypt.DefaultCost)

		existingUser := database.User{
			ID:           "user-123",
			Username:     "testuser",
			PasswordHash: correctPasswordHash,
			Role:         model.UserRoleName,
		}

		newPassword := "newpass"
		params := model.UpdateProfileParams{
			Password:    &wrongPassword,
			NewPassword: &newPassword,
		}

		mockUserRepo.EXPECT().
			RunWithTx(ctx, gomock.Any()).
			DoAndReturn(func(ctx context.Context, f func(context.Context) error) error {
				return f(ctx)
			})

		mockUserRepo.EXPECT().
			GetUserByID(ctx, "user-123").
			Return(existingUser, nil)

		_, err := provider.UpdateUserProfile(ctx, "user-123", params)

		if !errors.Is(err, facade.ErrUpdateProfileInvalidPassword) {
			t.Errorf("expected ErrUpdateProfileInvalidPassword, got %v", err)
		}
	})
}

func TestProvider_DeleteUser(t *testing.T) {
	ctx := context.Background()

	t.Run("successful deletion", func(t *testing.T) {
		provider, mockUserRepo, _, ctrl := setupTest(t)
		defer ctrl.Finish()

		mockUserRepo.EXPECT().
			DeleteUser(ctx, "user-123").
			Return(nil)

		err := provider.DeleteUser(ctx, "user-123")

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("deletion failure", func(t *testing.T) {
		provider, mockUserRepo, _, ctrl := setupTest(t)
		defer ctrl.Finish()

		expectedErr := database.ErrNotFound

		mockUserRepo.EXPECT().
			DeleteUser(ctx, "nonexistent").
			Return(expectedErr)

		err := provider.DeleteUser(ctx, "nonexistent")

		if err != expectedErr {
			t.Fatalf("expected %v, got %v", expectedErr, err)
		}
	})
}

func TestProvider_SignIn(t *testing.T) {
	ctx := context.Background()

	t.Run("successful sign in", func(t *testing.T) {
		provider, mockUserRepo, _, ctrl := setupTest(t)
		defer ctrl.Finish()

		password := "testpass"
		passwordHash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

		existingUser := database.User{
			ID:           "user-123",
			Username:     "testuser",
			PasswordHash: passwordHash,
			Role:         model.UserRoleName,
		}

		mockUserRepo.EXPECT().
			GetUserByUsername(ctx, "testuser").
			Return(existingUser, nil)

		result, err := provider.SignIn(ctx, "testuser", password)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if result.Username != "testuser" {
			t.Errorf("expected username testuser, got %s", result.Username)
		}
	})

	t.Run("user not found", func(t *testing.T) {
		provider, mockUserRepo, _, ctrl := setupTest(t)
		defer ctrl.Finish()

		mockUserRepo.EXPECT().
			GetUserByUsername(ctx, "nonexistent").
			Return(database.User{}, database.ErrNotFound)

		_, err := provider.SignIn(ctx, "nonexistent", "password")

		if !errors.Is(err, facade.ErrSignInInvalidCredentials) {
			t.Errorf("expected ErrSignInInvalidCredentials, got %v", err)
		}
	})

	t.Run("invalid password", func(t *testing.T) {
		provider, mockUserRepo, _, ctrl := setupTest(t)
		defer ctrl.Finish()

		correctPassword := "correctpass"
		passwordHash, _ := bcrypt.GenerateFromPassword([]byte(correctPassword), bcrypt.DefaultCost)

		existingUser := database.User{
			ID:           "user-123",
			Username:     "testuser",
			PasswordHash: passwordHash,
			Role:         model.UserRoleName,
		}

		mockUserRepo.EXPECT().
			GetUserByUsername(ctx, "testuser").
			Return(existingUser, nil)

		_, err := provider.SignIn(ctx, "testuser", "wrongpass")

		if !errors.Is(err, facade.ErrSignInInvalidCredentials) {
			t.Errorf("expected ErrSignInInvalidCredentials, got %v", err)
		}
	})
}

func TestProvider_SignUp(t *testing.T) {
	ctx := context.Background()

	t.Run("successful user signup", func(t *testing.T) {
		provider, mockUserRepo, _, ctrl := setupTest(t)
		defer ctrl.Finish()

		mockUserRepo.EXPECT().
			GetUserByUsername(ctx, "newuser").
			Return(database.User{}, database.ErrNotFound)

		mockUserRepo.EXPECT().
			RunWithTx(ctx, gomock.Any()).
			DoAndReturn(func(ctx context.Context, f func(context.Context) error) error {
				return f(ctx)
			})

		mockUserRepo.EXPECT().
			CreateUser(ctx, gomock.Any()).
			Return(nil)

		// regular users do not provide email

		result, err := provider.SignUp(ctx, "newuser", "New User", "", "password", false)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if result.Username != "newuser" {
			t.Errorf("expected username newuser, got %s", result.Username)
		}
		if result.Role != string(model.UserRoleName) {
			t.Errorf("expected role %s, got %s", model.UserRoleName, result.Role)
		}
		if result.Email != "" {
			t.Errorf("expected no email for regular user, got %s", result.Email)
		}
	})

	t.Run("successful publisher signup", func(t *testing.T) {
		provider, mockUserRepo, mockEmailSender, ctrl := setupTest(t)
		defer ctrl.Finish()

		mockUserRepo.EXPECT().
			GetUserByUsername(ctx, "newpublisher").
			Return(database.User{}, database.ErrNotFound)

		mockUserRepo.EXPECT().
			CheckUserExists(ctx, "Publisher Name", model.PublisherRoleName).
			Return(false, nil)

		mockUserRepo.EXPECT().
			RunWithTx(ctx, gomock.Any()).
			DoAndReturn(func(ctx context.Context, f func(context.Context) error) error {
				return f(ctx)
			}).AnyTimes()

		mockUserRepo.EXPECT().
			CreateUser(ctx, gomock.Any()).
			Return(nil)

		// mock unsubscribe check
		mockUserRepo.EXPECT().
			IsEmailUnsubscribed(ctx, "pub@example.com").
			Return(false, nil)

		// mock email verification calls (with email sender enabled)
		mockUserRepo.EXPECT().
			GetEmailVerificationByUserID(ctx, gomock.Any()).
			Return(database.EmailVerification{}, database.ErrNotFound).
			AnyTimes()

		mockUserRepo.EXPECT().
			CreateEmailVerification(ctx, gomock.Any()).
			Return(nil).
			AnyTimes()

		mockEmailSender.EXPECT().
			SendEmailVerification(ctx, gomock.Any()).
			Return("message-id-123", nil).
			AnyTimes()

		mockUserRepo.EXPECT().
			SetEmailVerificationMessageID(ctx, gomock.Any(), "message-id-123").
			Return(nil).
			AnyTimes()

		result, err := provider.SignUp(ctx, "newpublisher", "Publisher Name", "pub@example.com", "password", true)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if result.Role != string(model.PublisherRoleName) {
			t.Errorf("expected role %s, got %s", model.PublisherRoleName, result.Role)
		}
		if result.EmailVerified {
			t.Error("expected publisher email to not be auto-verified")
		}
	})

	t.Run("username already exists", func(t *testing.T) {
		provider, mockUserRepo, _, ctrl := setupTest(t)
		defer ctrl.Finish()

		existingUser := database.User{ID: "existing-123", Username: "existinguser"}

		mockUserRepo.EXPECT().
			GetUserByUsername(ctx, "existinguser").
			Return(existingUser, nil)

		_, err := provider.SignUp(ctx, "existinguser", "Display Name", "email@example.com", "password", false)

		if !errors.Is(err, facade.ErrSignUpUsernameExists) {
			t.Errorf("expected ErrSignUpUsernameExists, got %v", err)
		}
	})

	t.Run("publisher name already exists", func(t *testing.T) {
		provider, mockUserRepo, _, ctrl := setupTest(t)
		defer ctrl.Finish()

		mockUserRepo.EXPECT().
			GetUserByUsername(ctx, "newpublisher").
			Return(database.User{}, database.ErrNotFound)

		mockUserRepo.EXPECT().
			CheckUserExists(ctx, "Existing Publisher", model.PublisherRoleName).
			Return(true, nil)

		_, err := provider.SignUp(ctx, "newpublisher", "Existing Publisher", "pub@example.com", "password", true)

		if !errors.Is(err, facade.ErrSignUpPublisherNameExists) {
			t.Errorf("expected ErrSignUpPublisherNameExists, got %v", err)
		}
	})

	t.Run("signup without email", func(t *testing.T) {
		provider, mockUserRepo, _, ctrl := setupTest(t)
		defer ctrl.Finish()

		mockUserRepo.EXPECT().
			GetUserByUsername(ctx, "newuser").
			Return(database.User{}, database.ErrNotFound)

		mockUserRepo.EXPECT().
			RunWithTx(ctx, gomock.Any()).
			DoAndReturn(func(ctx context.Context, f func(context.Context) error) error {
				return f(ctx)
			})

		mockUserRepo.EXPECT().
			CreateUser(ctx, gomock.Any()).
			Return(nil)

		result, err := provider.SignUp(ctx, "newuser", "New User", "", "password", false)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if result.Email != "" {
			t.Errorf("expected empty email, got %s", result.Email)
		}
	})
}
