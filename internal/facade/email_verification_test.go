package facade_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/OutOfStack/game-library-auth/internal/database"
	"github.com/OutOfStack/game-library-auth/internal/facade"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
)

func TestProvider_VerifyEmail(t *testing.T) {
	ctx := context.Background()

	t.Run("successful email verification", func(t *testing.T) {
		provider, mockUserRepo, _, ctrl := setupTest(t)
		defer ctrl.Finish()

		user := database.User{
			ID:            "user-123",
			Username:      "testuser",
			Email:         sql.NullString{String: "test@example.com", Valid: true},
			EmailVerified: false,
			Role:          database.UserRoleName,
		}

		code := "123456"
		codeHash, _ := bcrypt.GenerateFromPassword([]byte(code), bcrypt.DefaultCost)
		verification := database.EmailVerification{
			ID:          "verification-123",
			UserID:      "user-123",
			CodeHash:    sql.NullString{String: string(codeHash), Valid: true},
			ExpiresAt:   time.Now().Add(24 * time.Hour),
			DateCreated: time.Now(),
		}

		mockUserRepo.EXPECT().
			GetUserByID(ctx, "user-123").
			Return(user, nil)

		mockUserRepo.EXPECT().
			GetEmailVerificationByUserID(ctx, "user-123").
			Return(verification, nil)

		mockUserRepo.EXPECT().
			SetUserEmailVerified(ctx, "user-123").
			Return(nil)

		mockUserRepo.EXPECT().
			SetEmailVerificationUsed(ctx, "verification-123", true).
			Return(nil)

		result, err := provider.VerifyEmail(ctx, "user-123", code)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !result.EmailVerified {
			t.Error("expected email to be verified")
		}
	})

	t.Run("user not found", func(t *testing.T) {
		provider, mockUserRepo, _, ctrl := setupTest(t)
		defer ctrl.Finish()

		mockUserRepo.EXPECT().
			GetUserByID(ctx, "nonexistent").
			Return(database.User{}, database.ErrNotFound)

		_, err := provider.VerifyEmail(ctx, "nonexistent", "123456")

		if !errors.Is(err, facade.ErrVerifyEmailUserNotFound) {
			t.Errorf("expected ErrVerifyEmailUserNotFound, got %v", err)
		}
	})

	t.Run("email already verified", func(t *testing.T) {
		provider, mockUserRepo, _, ctrl := setupTest(t)
		defer ctrl.Finish()

		user := database.User{
			ID:            "user-123",
			Username:      "testuser",
			EmailVerified: true,
			Role:          database.UserRoleName,
		}

		mockUserRepo.EXPECT().
			GetUserByID(ctx, "user-123").
			Return(user, nil)

		_, err := provider.VerifyEmail(ctx, "user-123", "123456")

		if !errors.Is(err, facade.ErrVerifyEmailAlreadyVerified) {
			t.Errorf("expected ErrVerifyEmailAlreadyVerified, got %v", err)
		}
	})

	t.Run("verification not found", func(t *testing.T) {
		provider, mockUserRepo, _, ctrl := setupTest(t)
		defer ctrl.Finish()

		user := database.User{
			ID:            "user-123",
			Username:      "testuser",
			EmailVerified: false,
			Role:          database.UserRoleName,
		}

		mockUserRepo.EXPECT().
			GetUserByID(ctx, "user-123").
			Return(user, nil)

		mockUserRepo.EXPECT().
			GetEmailVerificationByUserID(ctx, "user-123").
			Return(database.EmailVerification{}, database.ErrNotFound)

		_, err := provider.VerifyEmail(ctx, "user-123", "123456")

		if !errors.Is(err, facade.ErrVerifyEmailInvalidOrExpired) {
			t.Errorf("expected ErrVerifyEmailInvalidOrExpired, got %v", err)
		}
	})

	t.Run("verification expired", func(t *testing.T) {
		provider, mockUserRepo, _, ctrl := setupTest(t)
		defer ctrl.Finish()

		user := database.User{
			ID:            "user-123",
			Username:      "testuser",
			EmailVerified: false,
			Role:          database.UserRoleName,
		}

		code := "123456"
		codeHash, _ := bcrypt.GenerateFromPassword([]byte(code), bcrypt.DefaultCost)
		verification := database.EmailVerification{
			ID:          "verification-123",
			UserID:      "user-123",
			CodeHash:    sql.NullString{String: string(codeHash), Valid: true},
			ExpiresAt:   time.Now().Add(-1 * time.Hour), // expired
			DateCreated: time.Now().Add(-25 * time.Hour),
		}

		mockUserRepo.EXPECT().
			GetUserByID(ctx, "user-123").
			Return(user, nil)

		mockUserRepo.EXPECT().
			GetEmailVerificationByUserID(ctx, "user-123").
			Return(verification, nil)

		mockUserRepo.EXPECT().
			SetEmailVerificationUsed(ctx, "verification-123", false).
			Return(nil)

		_, err := provider.VerifyEmail(ctx, "user-123", code)

		if !errors.Is(err, facade.ErrVerifyEmailInvalidOrExpired) {
			t.Errorf("expected ErrVerifyEmailInvalidOrExpired, got %v", err)
		}
	})

	t.Run("invalid verification code", func(t *testing.T) {
		provider, mockUserRepo, _, ctrl := setupTest(t)
		defer ctrl.Finish()

		user := database.User{
			ID:            "user-123",
			Username:      "testuser",
			EmailVerified: false,
			Role:          database.UserRoleName,
		}

		correctCode := "123456"
		codeHash, _ := bcrypt.GenerateFromPassword([]byte(correctCode), bcrypt.DefaultCost)
		verification := database.EmailVerification{
			ID:          "verification-123",
			UserID:      "user-123",
			CodeHash:    sql.NullString{String: string(codeHash), Valid: true},
			ExpiresAt:   time.Now().Add(24 * time.Hour),
			DateCreated: time.Now(),
		}

		mockUserRepo.EXPECT().
			GetUserByID(ctx, "user-123").
			Return(user, nil)

		mockUserRepo.EXPECT().
			GetEmailVerificationByUserID(ctx, "user-123").
			Return(verification, nil)

		_, err := provider.VerifyEmail(ctx, "user-123", "wrongcode")

		if !errors.Is(err, facade.ErrVerifyEmailInvalidOrExpired) {
			t.Errorf("expected ErrVerifyEmailInvalidOrExpired, got %v", err)
		}
	})
}

func TestProvider_ResendVerificationEmail(t *testing.T) {
	ctx := context.Background()

	t.Run("successful resend with email sender enabled", func(t *testing.T) {
		provider, mockUserRepo, mockEmailSender, ctrl := setupTest(t)
		defer ctrl.Finish()

		user := database.User{
			ID:            "user-123",
			Username:      "testuser",
			Email:         sql.NullString{String: "test@example.com", Valid: true},
			EmailVerified: false,
			Role:          database.UserRoleName,
		}

		mockUserRepo.EXPECT().
			GetUserByID(ctx, "user-123").
			Return(user, nil)

		// mock the call to check existing verification record (returns not found)
		mockUserRepo.EXPECT().
			GetEmailVerificationByUserID(ctx, "user-123").
			Return(database.EmailVerification{}, database.ErrNotFound)

		// mock creating new verification record
		mockUserRepo.EXPECT().
			CreateEmailVerification(ctx, gomock.Any()).
			Return(nil)

		// mock email sending
		mockEmailSender.EXPECT().
			SendEmailVerification(ctx, gomock.Any()).
			Return("message-id-123", nil)

		// mock setting message ID
		mockUserRepo.EXPECT().
			SetEmailVerificationMessageID(ctx, gomock.Any(), "message-id-123").
			Return(nil)

		err := provider.ResendVerificationEmail(ctx, "user-123")

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("email already verified", func(t *testing.T) {
		provider, mockUserRepo, _, ctrl := setupTest(t)
		defer ctrl.Finish()

		user := database.User{
			ID:            "user-123",
			Username:      "testuser",
			Email:         sql.NullString{String: "test@example.com", Valid: true},
			EmailVerified: true,
			Role:          database.UserRoleName,
		}

		mockUserRepo.EXPECT().
			GetUserByID(ctx, "user-123").
			Return(user, nil)

		err := provider.ResendVerificationEmail(ctx, "user-123")

		if !errors.Is(err, facade.ErrVerifyEmailAlreadyVerified) {
			t.Errorf("expected ErrVerifyEmailAlreadyVerified, got %v", err)
		}
	})

	t.Run("user has no email", func(t *testing.T) {
		provider, mockUserRepo, _, ctrl := setupTest(t)
		defer ctrl.Finish()

		user := database.User{
			ID:            "user-123",
			Username:      "testuser",
			Email:         sql.NullString{Valid: false}, // no email
			EmailVerified: false,
			Role:          database.UserRoleName,
		}

		mockUserRepo.EXPECT().
			GetUserByID(ctx, "user-123").
			Return(user, nil)

		err := provider.ResendVerificationEmail(ctx, "user-123")

		if !errors.Is(err, facade.ErrResendVerificationNoEmail) {
			t.Errorf("expected ErrResendVerificationNoEmail, got %v", err)
		}
	})

	t.Run("user not found", func(t *testing.T) {
		provider, mockUserRepo, _, ctrl := setupTest(t)
		defer ctrl.Finish()

		mockUserRepo.EXPECT().
			GetUserByID(ctx, "nonexistent").
			Return(database.User{}, database.ErrNotFound)

		err := provider.ResendVerificationEmail(ctx, "nonexistent")

		if !errors.Is(err, database.ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})
}
