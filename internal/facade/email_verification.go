package facade

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/OutOfStack/game-library-auth/internal/client/mailersend"
	"github.com/OutOfStack/game-library-auth/internal/database"
	"github.com/OutOfStack/game-library-auth/internal/model"
	"github.com/cenkalti/backoff/v4"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// VerifyEmail verifies user email by provided code
func (p *Provider) VerifyEmail(ctx context.Context, userID string, code string) (model.User, error) {
	// get user
	user, err := p.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return model.User{}, VerifyEmailUserNotFoundErr
		}
		p.log.Error("get user by id", zap.String("userID", userID), zap.Error(err))
		return model.User{}, err
	}

	if user.EmailVerified {
		return model.User{}, VerifyEmailAlreadyVerifiedErr
	}

	// get verification by user id
	verification, err := p.userRepo.GetEmailVerificationByUserID(ctx, user.ID)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return model.User{}, VerifyEmailInvalidOrExpiredErr
		}
		p.log.Error("get email verification", zap.String("userID", userID), zap.Error(err))
		return model.User{}, err
	}

	// check expiration
	if verification.IsExpired() {
		if err = p.userRepo.SetEmailVerificationUsed(ctx, verification.ID, false); err != nil {
			p.log.Error("clear expired verification", zap.String("verificationID", verification.ID), zap.Error(err))
			return model.User{}, err
		}
		return model.User{}, VerifyEmailInvalidOrExpiredErr
	}

	// compare codes
	if err = bcrypt.CompareHashAndPassword([]byte(verification.CodeHash.String), []byte(code)); err != nil {
		return model.User{}, VerifyEmailInvalidOrExpiredErr
	}

	// set verified
	if err = p.userRepo.SetUserEmailVerified(ctx, verification.UserID); err != nil {
		p.log.Error("set user email verified", zap.String("userID", verification.UserID), zap.Error(err))
		return model.User{}, err
	}
	if err = p.userRepo.SetEmailVerificationUsed(ctx, verification.ID, true); err != nil {
		p.log.Error("mark email verification used", zap.String("verificationID", verification.ID), zap.Error(err))
		return model.User{}, err
	}

	user.EmailVerified = true
	return mapDBUserToUser(user), nil
}

// ResendVerificationEmail resends email verification code to a user
func (p *Provider) ResendVerificationEmail(ctx context.Context, userID string) error {
	// get user
	user, err := p.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		p.log.Error("get user by id", zap.String("userID", userID), zap.Error(err))
		return err
	}

	// check if email is already verified
	if user.EmailVerified {
		return VerifyEmailAlreadyVerifiedErr
	}

	// check if user has an email address
	if !user.Email.Valid {
		return ResendVerificationNoEmailErr
	}

	// send verification email (includes cooldown and record management)
	if err = p.sendVerificationEmail(ctx, user.ID, user.Email.String, user.Username); err != nil {
		return err
	}

	return nil
}

// sends verification email. Returns ErrTooManyRequests if email was sent recently
func (p *Provider) sendVerificationEmail(ctx context.Context, userID string, email, username string) error {
	if p.disableEmailSender {
		return nil
	}

	// check if code was already sent
	record, err := p.userRepo.GetEmailVerificationByUserID(ctx, userID)
	if err != nil && !errors.Is(err, database.ErrNotFound) {
		return fmt.Errorf("get verification record: %w", err)
	}
	if err == nil {
		// if sent before resend cooldown, don't resend
		// if sent after resend cooldown, resend
		if time.Since(record.DateCreated) < resendVerificationCodeCooldown {
			return ErrTooManyRequests
		}

		// mark verification as used
		if err = p.userRepo.SetEmailVerificationUsed(ctx, record.ID, false); err != nil {
			p.log.Error("clear verification", zap.Error(err))
		}
	}

	// create new verification record
	recordID, code, err := p.createEmailVerificationRecord(ctx, userID)
	if err != nil {
		return fmt.Errorf("create verification record: %w", err)
	}

	// send verification email
	messageID, err := p.sendVerificationEmailWithRetry(ctx, email, username, code)
	if err != nil {
		return fmt.Errorf("send verification email: %w", err)
	}

	// set message id
	err = p.userRepo.SetEmailVerificationMessageID(ctx, recordID, messageID)
	if err != nil {
		return fmt.Errorf("set email verification message_id: %w", err)
	}

	return nil
}

// creates a new email verification record and returns verification record id and code
func (p *Provider) createEmailVerificationRecord(ctx context.Context, userID string) (string, string, error) {
	code := generate6DigitCode()

	// hash the code
	codeHash, err := bcrypt.GenerateFromPassword([]byte(code), bcrypt.DefaultCost)
	if err != nil {
		return "", "", fmt.Errorf("hash verification code: %w", err)
	}

	expiresAt := time.Now().Add(verificationCodeTTL)
	verification := database.NewEmailVerification(userID, string(codeHash), expiresAt)

	if err = p.userRepo.CreateEmailVerification(ctx, verification); err != nil {
		return "", "", err
	}

	return verification.ID, code, nil
}

// sends verification email with retry logic and returns message id
func (p *Provider) sendVerificationEmailWithRetry(ctx context.Context, email, username, code string) (messageID string, err error) {
	op := func() error {
		messageID, err = p.emailSender.SendEmailVerification(ctx, mailersend.SendEmailVerificationRequest{
			Email:            email,
			Username:         username,
			VerificationCode: code,
		})
		return err
	}

	bo := backoff.NewExponentialBackOff([]backoff.ExponentialBackOffOpts{
		backoff.WithInitialInterval(30 * time.Millisecond),
		backoff.WithMaxInterval(500 * time.Millisecond),
		backoff.WithMaxElapsedTime(3 * time.Second),
	}...)

	err = backoff.Retry(op, backoff.WithContext(bo, ctx))
	return messageID, err
}

// generates a secure random 6-digit verification code
func generate6DigitCode() string {
	const codeLength = 6
	const maxVal = 10
	var code string
	for range codeLength {
		n, _ := rand.Int(rand.Reader, big.NewInt(maxVal))
		code += n.String()
	}
	return code
}
