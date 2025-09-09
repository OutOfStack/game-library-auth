package handlers

import (
	"errors"
	"net/http"

	"github.com/OutOfStack/game-library-auth/internal/database"
	"github.com/OutOfStack/game-library-auth/internal/web"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// ResendVerificationEmailHandler godoc
// @Summary      Resend email verification code
// @Description  Resends email verification code to the user's email address
// @Tags         auth
// @Security     Bearer
// @Produce      json
// @Success      204 "Successfully resent verification email"
// @Failure      400 {object} web.ErrResp "User email is already verified"
// @Failure      401 {object} web.ErrResp "Invalid or missing authorization token"
// @Failure      429 {object} web.ErrResp "Too many resend requests"
// @Failure      500 {object} web.ErrResp "Internal server error"
// @Router       /resend-verification [post]
func (a *AuthAPI) ResendVerificationEmailHandler(c *fiber.Ctx) error {
	ctx, span := tracer.Start(c.Context(), "resendVerificationEmail")
	defer span.End()

	claims, err := a.getClaims(c)
	if err != nil {
		a.log.Error("get claims", zap.Error(err))
		return c.Status(http.StatusUnauthorized).JSON(web.ErrResp{
			Error: invalidAuthToken,
		})
	}

	// get user by ID
	user, err := a.userRepo.GetUserByID(ctx, claims.UserID)
	if err != nil {
		a.log.Error("get user by id", zap.String("user_id", claims.UserID), zap.Error(err))
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

	// check if user has an email address
	if !user.Email.Valid {
		return c.Status(http.StatusBadRequest).JSON(web.ErrResp{
			Error: "User does not have an email address",
		})
	}

	// get verification record by user ID
	verification, err := a.userRepo.GetEmailVerificationByUserID(ctx, user.ID)
	if err != nil && !errors.Is(err, database.ErrNotFound) {
		a.log.Error("get email verification", zap.Error(err))
		return c.Status(http.StatusInternalServerError).JSON(web.ErrResp{
			Error: internalErrorMsg,
		})
	}
	// mark existing verification as used
	if err = a.userRepo.SetEmailVerificationUsed(ctx, verification.ID, false); err != nil {
		a.log.Error("clear previous verification", zap.Error(err))
	}

	// resend verification email
	if err = a.sendVerificationEmail(ctx, user.ID, user.Email.String, user.Username); err != nil {
		if errors.Is(err, errTooManyRequests) {
			return c.Status(http.StatusTooManyRequests).JSON(web.ErrResp{
				Error: "Please wait before requesting another code",
			})
		}
		a.log.Error("sending verification email", zap.Error(err))
		return c.Status(http.StatusInternalServerError).JSON(web.ErrResp{
			Error: internalErrorMsg,
		})
	}

	return c.SendStatus(http.StatusNoContent)
}
