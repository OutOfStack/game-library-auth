package handlers

import (
	"errors"
	"net/http"

	"github.com/OutOfStack/game-library-auth/internal/database"
	"github.com/OutOfStack/game-library-auth/internal/web"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// VerifyEmailHandler godoc
// @Summary      Verify email address using verification code
// @Description  Verifies user email address using the code sent via email
// @Tags         auth
// @Security     Bearer
// @Accept       json
// @Produce      json
// @Param        verification body VerifyEmailReq true "Email verification code"
// @Success      200 {object} TokenResp
// @Failure      400 {object} web.ErrResp "Invalid or expired verification code"
// @Failure      401 {object} web.ErrResp "Invalid or missing authorization token"
// @Failure      404 {object} web.ErrResp "Verification code is not found"
// @Failure      500 {object} web.ErrResp "Internal server error"
// @Router       /verify-email [post]
func (a *AuthAPI) VerifyEmailHandler(c *fiber.Ctx) error {
	ctx, span := tracer.Start(c.Context(), "verifyEmail")
	defer span.End()

	claims, err := a.getClaims(c)
	if err != nil {
		a.log.Error("get claims", zap.Error(err))
		return c.Status(http.StatusUnauthorized).JSON(web.ErrResp{
			Error: invalidAuthTokenMsg,
		})
	}

	var req VerifyEmailReq
	if err = c.BodyParser(&req); err != nil {
		a.log.Error("parsing data", zap.Error(err))
		return c.Status(http.StatusBadRequest).JSON(web.ErrResp{
			Error: "Cannot parse request",
		})
	}

	if fields, vErr := web.Validate(req); vErr != nil {
		a.log.Info("validating verify email data", zap.Error(vErr))
		return c.Status(http.StatusBadRequest).JSON(web.ErrResp{
			Error:  validationErrorMsg,
			Fields: fields,
		})
	}

	// get user by user id
	user, err := a.userRepo.GetUserByID(ctx, claims.UserID)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return c.Status(http.StatusNotFound).JSON(web.ErrResp{
				Error: "User not found",
			})
		}
		a.log.Error("get user by email", zap.Error(err))
		return c.Status(http.StatusInternalServerError).JSON(web.ErrResp{
			Error: internalErrorMsg,
		})
	}

	// check if email is already verified
	if user.EmailVerified {
		return c.Status(http.StatusBadRequest).JSON(web.ErrResp{
			Error: "Email is already verified",
		})
	}

	// get verification record by user ID
	verification, err := a.userRepo.GetEmailVerificationByUserID(ctx, user.ID)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return c.Status(http.StatusBadRequest).JSON(web.ErrResp{
				Error: invalidOrExpiredVrfCodeMsg,
			})
		}
		a.log.Error("get email verification", zap.Error(err))
		return c.Status(http.StatusInternalServerError).JSON(web.ErrResp{
			Error: internalErrorMsg,
		})
	}

	// check if verification code has expired
	if verification.IsExpired() {
		// clear expired verification without marking as verified
		if err = a.userRepo.SetEmailVerificationUsed(ctx, verification.ID, false); err != nil {
			a.log.Error("clear expired verification", zap.Error(err))
			return c.Status(http.StatusInternalServerError).JSON(web.ErrResp{
				Error: internalErrorMsg,
			})
		}
		return c.Status(http.StatusBadRequest).JSON(web.ErrResp{
			Error: invalidOrExpiredVrfCodeMsg,
		})
	}

	// validate code
	if err = bcrypt.CompareHashAndPassword([]byte(verification.CodeHash.String), []byte(req.Code)); err != nil {
		return c.Status(http.StatusBadRequest).JSON(web.ErrResp{
			Error: invalidOrExpiredVrfCodeMsg,
		})
	}

	// update user email and mark as verified
	if err = a.userRepo.UpdateUserEmail(ctx, verification.UserID, verification.Email, true); err != nil {
		a.log.Error("update user email", zap.String("userID", verification.UserID.String()), zap.Error(err))
		return c.Status(http.StatusInternalServerError).JSON(web.ErrResp{
			Error: internalErrorMsg,
		})
	}

	// mark verification record as used and verified
	if err = a.userRepo.SetEmailVerificationUsed(ctx, verification.ID, true); err != nil {
		a.log.Error("mark email verification as used", zap.Error(err))
		return c.Status(http.StatusInternalServerError).JSON(web.ErrResp{
			Error: internalErrorMsg,
		})
	}

	// generate JWT token
	user.EmailVerified = true
	jwtClaims := a.auth.CreateClaims(user)
	tokenStr, err := a.auth.GenerateToken(jwtClaims)
	if err != nil {
		a.log.Error("generating user token", zap.Error(err))
		return c.Status(http.StatusInternalServerError).JSON(web.ErrResp{
			Error: internalErrorMsg,
		})
	}

	return c.JSON(TokenResp{
		AccessToken: tokenStr,
	})
}
