package facade

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math"
	"math/big"
	"strconv"
	"time"

	"github.com/OutOfStack/game-library-auth/internal/client/resendapi"
	"github.com/OutOfStack/game-library-auth/internal/database"
	"github.com/OutOfStack/game-library-auth/internal/model"
	"github.com/cenkalti/backoff/v4"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// VerifyEmail verifies user email by provided code
func (p *Provider) VerifyEmail(ctx context.Context, userID string, code string) (model.User, error) {
	var user database.User

	txErr := p.userRepo.RunWithTx(ctx, func(ctx context.Context) error {
		var err error

		// get user
		user, err = p.userRepo.GetUserByID(ctx, userID)
		if err != nil {
			if errors.Is(err, database.ErrNotFound) {
				return ErrVerifyEmailUserNotFound
			}
			p.log.Error("get user by id", zap.String("userID", userID), zap.Error(err))
			return err
		}

		if user.EmailVerified {
			return ErrVerifyEmailAlreadyVerified
		}

		// get verification by user id
		verification, err := p.userRepo.GetEmailVerificationByUserID(ctx, user.ID)
		if err != nil {
			if errors.Is(err, database.ErrNotFound) {
				return ErrVerifyEmailInvalidOrExpired
			}
			p.log.Error("get email verification", zap.String("userID", userID), zap.Error(err))
			return err
		}

		// check expiration
		if verification.IsExpired() {
			if err = p.userRepo.SetEmailVerificationUsed(ctx, verification.ID, false); err != nil {
				p.log.Error("clear expired verification", zap.String("verificationID", verification.ID), zap.Error(err))
				return err
			}
			return ErrVerifyEmailInvalidOrExpired
		}

		// compare codes
		if err = bcrypt.CompareHashAndPassword([]byte(verification.CodeHash.String), []byte(code)); err != nil {
			return ErrVerifyEmailInvalidOrExpired
		}

		// set verified
		if err = p.userRepo.SetUserEmailVerified(ctx, userID); err != nil {
			p.log.Error("set user email verified", zap.String("userID", verification.UserID), zap.Error(err))
			return err
		}
		user.EmailVerified = true

		// mark verification as used
		if err = p.userRepo.SetEmailVerificationUsed(ctx, verification.ID, true); err != nil {
			p.log.Error("mark email verification used", zap.String("verificationID", verification.ID), zap.Error(err))
			return err
		}

		return nil
	})
	if txErr != nil {
		return model.User{}, txErr
	}

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

	// check if already verified.
	// only publishers require email verification
	if user.EmailVerified || user.Role != model.PublisherRoleName {
		return ErrVerifyEmailAlreadyVerified
	}

	// check if user has an email address
	if !user.Email.Valid {
		return ErrResendVerificationNoEmail
	}

	// send verification email
	if err = p.sendVerificationEmail(ctx, user.ID, user.Email.String, user.Username); err != nil {
		return err
	}

	return nil
}

// sends verification email. Returns ErrTooManyRequests if email was sent recently
func (p *Provider) sendVerificationEmail(ctx context.Context, userID string, email, username string) error {
	// check if email is unsubscribed
	isUnsubscribed, uErr := p.userRepo.IsEmailUnsubscribed(ctx, email)
	if uErr != nil {
		p.log.Error("check email unsubscribe status", zap.String("email", email), zap.Error(uErr))
		return fmt.Errorf("check email unsubscribe status: %w", uErr)
	}
	if isUnsubscribed {
		p.log.Info("email is unsubscribed, skipping verification email", zap.String("email", email))
		return nil
	}

	txErr := p.userRepo.RunWithTx(ctx, func(ctx context.Context) error {
		// check if code was already sent
		vrfRecord, err := p.userRepo.GetEmailVerificationByUserID(ctx, userID)
		if err != nil && !errors.Is(err, database.ErrNotFound) {
			return fmt.Errorf("get verification record: %w", err)
		}
		if err == nil {
			// if sent before resend cooldown, don't resend
			// if sent after resend cooldown, resend
			if time.Since(vrfRecord.DateCreated) < model.ResendVerificationCodeCooldown {
				return ErrTooManyRequests
			}

			// mark verification as used
			if err = p.userRepo.SetEmailVerificationUsed(ctx, vrfRecord.ID, false); err != nil {
				return fmt.Errorf("clear verification: %w", err)
			}
		}

		// create new verification record
		result, err := p.createEmailVerificationRecord(ctx, userID, email)
		if err != nil {
			return fmt.Errorf("create verification record: %w", err)
		}

		// send verification email
		messageID, err := p.sendVerificationEmailWithRetry(ctx, email, username, result.Code, result.UnsubscribeToken)
		if err != nil {
			return fmt.Errorf("send verification email: %w", err)
		}

		// set message id
		err = p.userRepo.SetEmailVerificationMessageID(ctx, result.ID, messageID)
		if err != nil {
			return fmt.Errorf("set email verification message_id: %w", err)
		}

		return nil
	})
	if txErr != nil {
		return txErr
	}

	return nil
}

// creates a new email verification record and returns the result
func (p *Provider) createEmailVerificationRecord(ctx context.Context, userID, email string) (emailVerificationResult, error) {
	code := generate6DigitCode()

	// hash the code
	codeHash, err := bcrypt.GenerateFromPassword([]byte(code), bcrypt.DefaultCost)
	if err != nil {
		return emailVerificationResult{}, fmt.Errorf("hash verification code: %w", err)
	}

	// generate unsubscribe token
	now := time.Now()
	expiresAt := now.Add(model.UnsubscribeTokenTTL)
	unsubscribeToken := p.unsubscribeTokenGenerator.GenerateToken(email, expiresAt)

	verification := database.NewEmailVerification(userID, string(codeHash), unsubscribeToken, now)

	if err = p.userRepo.CreateEmailVerification(ctx, verification); err != nil {
		return emailVerificationResult{}, err
	}

	return emailVerificationResult{
		ID:               verification.ID,
		Code:             code,
		UnsubscribeToken: unsubscribeToken,
	}, nil
}

// sends verification email with retry logic and returns message id
func (p *Provider) sendVerificationEmailWithRetry(ctx context.Context, email, username, code, unsubscribeToken string) (messageID string, err error) {
	op := func() error {
		messageID, err = p.emailSender.SendEmailVerification(ctx, resendapi.SendEmailVerificationRequest{
			Email:            email,
			Username:         username,
			VerificationCode: code,
			UnsubscribeToken: unsubscribeToken,
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
	var minVal = int64(math.Pow10(defaultVrfCodeLen - 1))  // min codeLen len value
	var maxValPlus1 = int64(math.Pow10(defaultVrfCodeLen)) // max codeLen len value + 1, but it is exclusive

	// generate random number with codeLen
	n, _ := rand.Int(rand.Reader, big.NewInt(maxValPlus1-minVal)) // for codeLen = 2: range [0, 90)
	return strconv.FormatInt(n.Int64()+minVal, 10)                // for codeLen = 2: [0, 90] + 10 = [10; 100)
}
