package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/OutOfStack/game-library-auth/internal/client/mailersend"
	"github.com/OutOfStack/game-library-auth/internal/database"
	"github.com/OutOfStack/game-library-auth/internal/web"
	"github.com/cenkalti/backoff/v4"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// SignUpHandler godoc
// @Summary		Register a new user
// @Description Create a new user account with the provided information
// @Tags  		auth
// @Accept 		json
// @Produce 	json
// @Param 		signup body SignUpReq true "User signup information"
// @Success		200 {object} TokenResp 	 "User credentials"
// @Failure 	400 {object} web.ErrResp "Invalid input data"
// @Failure 	409 {object} web.ErrResp "Username or publisher name already exists"
// @Failure 	500 {object} web.ErrResp "Internal server error"
// @Router		/signup [post]
func (a *AuthAPI) SignUpHandler(c *fiber.Ctx) error {
	ctx, span := tracer.Start(c.Context(), "signUp")
	defer span.End()

	var signUp SignUpReq
	if err := c.BodyParser(&signUp); err != nil {
		a.log.Error("parsing data", zap.Error(err))
		return c.Status(http.StatusBadRequest).JSON(web.ErrResp{
			Error: "Error parsing data",
		})
	}

	log := a.log.With(zap.String("username", signUp.Username))
	// validate
	if fields, err := web.Validate(signUp); err != nil {
		log.Info("validating sign up data", zap.Error(err))
		return c.Status(http.StatusBadRequest).JSON(web.ErrResp{
			Error:  validationErrorMsg,
			Fields: fields,
		})
	}

	// check if such username already exists
	_, err := a.userRepo.GetUserByUsername(ctx, signUp.Username)
	// if err is ErrNotFound then continue
	if err != nil && !errors.Is(err, database.ErrNotFound) {
		log.Error("checking existence of user", zap.Error(err))
		return c.Status(http.StatusInternalServerError).JSON(web.ErrResp{
			Error: internalErrorMsg,
		})
	}
	if err == nil {
		log.Info("username already exists")
		return c.Status(http.StatusConflict).JSON(web.ErrResp{
			Error: "This username is already taken",
		})
	}

	// get role
	var userRole = database.UserRoleName
	if signUp.IsPublisher {
		// check uniqueness of publisher name
		exists, cErr := a.userRepo.CheckUserExists(ctx, signUp.DisplayName, database.PublisherRoleName)
		if cErr != nil {
			log.Error("checking existence of publisher with name", zap.Error(cErr))
			return c.Status(http.StatusInternalServerError).JSON(web.ErrResp{
				Error: internalErrorMsg,
			})
		}
		if exists {
			log.Info("publisher already exists")
			return c.Status(http.StatusConflict).JSON(web.ErrResp{
				Error: "Publisher with this name already exists",
			})
		}
		userRole = database.PublisherRoleName
	}

	// hash password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(signUp.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Error("generating password hash", zap.Error(err))
		return c.Status(http.StatusInternalServerError).JSON(web.ErrResp{
			Error: internalErrorMsg,
		})
	}

	usr := database.NewUser(signUp.Username, signUp.DisplayName, passwordHash, userRole)
	usr.SetEmail(signUp.Email, false)

	// create new user
	if err = a.userRepo.CreateUser(ctx, usr); err != nil {
		log.Error("creating new user", zap.Error(err))
		return c.Status(http.StatusInternalServerError).JSON(web.ErrResp{
			Error: internalErrorMsg,
		})
	}

	// send verification email
	if usr.Email.Valid {
		if err = a.sendVerificationEmail(ctx, usr.ID, signUp.Email, signUp.Username); err != nil {
			// if error occurs, don't return it, user  will see "resend email verification code" button on UI
			log.Error("sending verification email", zap.Error(err))
		}
	}

	// generate JWT token
	jwtClaims := a.auth.CreateClaims(usr)
	tokenStr, err := a.auth.GenerateToken(jwtClaims)
	if err != nil {
		log.Error("generating token", zap.Error(err))
		return c.Status(http.StatusInternalServerError).JSON(web.ErrResp{
			Error: internalErrorMsg,
		})
	}

	return c.JSON(TokenResp{
		AccessToken: tokenStr,
	})
}

// sendVerificationEmail sends verification email. Returns errTooManyRequests if email was sent recently
func (a *AuthAPI) sendVerificationEmail(ctx context.Context, userID uuid.UUID, email, username string) error {
	if a.disableEmailSender {
		return nil
	}

	// check if code was already sent
	record, err := a.userRepo.GetEmailVerificationByUserID(ctx, userID)
	if err != nil && !errors.Is(err, database.ErrNotFound) {
		return fmt.Errorf("get verification record: %w", err)
	}
	if err == nil {
		// if sent before resend cooldown, don't resend
		// if sent after resend cooldown, resend
		if time.Since(record.DateCreated) < resendVerificationCodeCooldown {
			return errTooManyRequests
		}

		// mark verification as used
		if err = a.userRepo.SetEmailVerificationUsed(ctx, record.ID, false); err != nil {
			a.log.Error("clear verification", zap.Error(err))
		}
	}

	// create new verification record
	recordID, code, err := a.createEmailVerificationRecord(ctx, userID, email)
	if err != nil {
		return fmt.Errorf("create verification record: %w", err)
	}

	// send verification email
	messageID, err := a.sendVerificationEmailWithRetry(ctx, email, username, code)
	if err != nil {
		return fmt.Errorf("send verification email: %w", err)
	}

	// set message id
	err = a.userRepo.SetEmailVerificationMessageID(ctx, recordID, messageID)
	if err != nil {
		return fmt.Errorf("set email verification message_id: %w", err)
	}

	return nil
}

// createEmailVerificationRecord creates a new email verification record and returns verification record id and code
func (a *AuthAPI) createEmailVerificationRecord(ctx context.Context, userID uuid.UUID, email string) (uuid.UUID, string, error) {
	code, err := generate6DigitCode()
	if err != nil {
		return uuid.Nil, "", fmt.Errorf("generate verification code: %w", err)
	}

	// hash the code
	codeHash, err := bcrypt.GenerateFromPassword([]byte(code), bcrypt.DefaultCost)
	if err != nil {
		return uuid.Nil, "", fmt.Errorf("hash verification code: %w", err)
	}

	expiresAt := time.Now().Add(verificationCodeTTL)
	verification := database.NewEmailVerification(userID, email, string(codeHash), expiresAt)

	if err = a.userRepo.CreateEmailVerification(ctx, verification); err != nil {
		return uuid.Nil, "", err
	}

	return verification.ID, code, nil
}

// sendVerificationEmailWithRetry sends verification email with retry logic and returns message id
func (a *AuthAPI) sendVerificationEmailWithRetry(ctx context.Context, email, username, code string) (messageID string, err error) {
	op := func() error {
		messageID, err = a.emailSender.SendEmailVerification(ctx, mailersend.SendEmailVerificationRequest{
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
