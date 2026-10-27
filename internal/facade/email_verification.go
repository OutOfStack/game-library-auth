package facade

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/big"
	"strconv"
	"time"
	"uuid"

	"github.com/OutOfStack/game-library-auth/internal/client/resendapi"
	"github.com/OutOfStack/game-library-auth/internal/database"
	"github.com/OutOfStack/game-library-auth/internal/model"
	"github.com/OutOfStack/game-library-auth/internal/notification"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// errors
var (
	ErrResendVerificationNoEmail   = errors.New("resend verification: user has no email")
	ErrVerifyEmailUserNotFound     = errors.New("verify email: user not found")
	ErrVerifyEmailAlreadyVerified  = errors.New("verify email: already verified")
	ErrVerifyEmailInvalidOrExpired = errors.New("verify email: invalid or expired code")
	ErrSendVerifyEmailUnsubscribed = errors.New("send verify email: user is unsubscribed")
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
	_, err := p.enqueueVerificationEmail(ctx, userID)
	return err
}

// locks the user and queues a verification email. Returns the user as read under lock, so callers see concurrent verification
func (p *Provider) enqueueVerificationEmail(ctx context.Context, userID string) (database.User, error) {
	var user database.User
	err := p.userRepo.RunWithTx(ctx, func(ctx context.Context) error {
		// lock the user before reading the current code, including when no code exists yet
		var err error
		if user, err = p.userRepo.GetUserByID(ctx, userID); err != nil {
			return err
		}
		return p.queueVerificationEmail(ctx, user)
	})
	return user, err
}

// queues verification email for a user whose row the caller's transaction holds locked.
// Returns ErrTooManyRequests if a code was queued recently
func (p *Provider) queueVerificationEmail(ctx context.Context, user database.User) error {
	if user.EmailVerified || user.Role != model.PublisherRoleName {
		return ErrVerifyEmailAlreadyVerified
	}
	if !user.Email.Valid {
		return ErrResendVerificationNoEmail
	}
	state, err := p.userRepo.GetVerificationState(ctx, user.ID, user.Email.String)
	if err != nil {
		return fmt.Errorf("get verification state: %w", err)
	}
	if state.Unsubscribed {
		return ErrSendVerifyEmailUnsubscribed
	}
	if state.LatestID.Valid {
		if elapsed := state.Now.Sub(state.LatestCreated.Time); elapsed < model.ResendVerificationCodeCooldown {
			return NewTooManyRequestsError(model.ResendVerificationCodeCooldown - elapsed)
		}
		if err = p.userRepo.SetEmailVerificationUsed(ctx, state.LatestID.String, false); err != nil {
			return err
		}
	}
	result, err := p.createEmailVerificationRecord(ctx, user.ID, user.Email.String, state.Now)
	if err != nil {
		return fmt.Errorf("create verification record: %w", err)
	}
	email, err := p.emailPreparer.PrepareVerification(resendapi.SendEmailVerificationRequest{
		Email: user.Email.String, Username: user.Username, VerificationCode: result.Code, UnsubscribeToken: result.UnsubscribeToken,
	})
	if err != nil {
		return fmt.Errorf("prepare verification email: %w", err)
	}
	n := database.Notification{ID: uuid.New().String(), ReferenceID: result.ID, Kind: notification.VerificationKind,
		CreatedAt: state.Now, ExpiresAt: state.Now.Add(model.VerificationCodeTTL)}
	n.Payload, err = json.Marshal(email)
	if err != nil {
		return fmt.Errorf("marshal verification email: %w", err)
	}
	return p.userRepo.CreateNotification(ctx, n)
}

// creates a new email verification record and returns the result
func (p *Provider) createEmailVerificationRecord(ctx context.Context, userID, email string, now time.Time) (emailVerificationResult, error) {
	code := generate6DigitCode()

	// hash the code
	codeHash, err := bcrypt.GenerateFromPassword([]byte(code), bcrypt.DefaultCost)
	if err != nil {
		return emailVerificationResult{}, fmt.Errorf("hash verification code: %w", err)
	}

	// generate unsubscribe token
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

// generates a secure random 6-digit verification code
func generate6DigitCode() string {
	var minVal = int64(math.Pow10(defaultVrfCodeLen - 1))  // min codeLen len value
	var maxValPlus1 = int64(math.Pow10(defaultVrfCodeLen)) // max codeLen len value + 1, but it is exclusive

	// generate random number with codeLen
	n, _ := rand.Int(rand.Reader, big.NewInt(maxValPlus1-minVal)) // for codeLen = 2: range [0, 90)
	return strconv.FormatInt(n.Int64()+minVal, 10)                // for codeLen = 2: [0, 90] + 10 = [10; 100)
}
