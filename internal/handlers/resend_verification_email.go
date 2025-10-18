package handlers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/OutOfStack/game-library-auth/internal/facade"
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
			Error: invalidAuthTokenMsg,
		})
	}

	// resend verification email
	if err = a.userFacade.ResendVerificationEmail(ctx, claims.UserID); err != nil {
		switch {
		case errors.Is(err, facade.ErrVerifyEmailAlreadyVerified):
			return c.Status(http.StatusBadRequest).JSON(web.ErrResp{
				Error: "Email is already verified",
			})
		case errors.Is(err, facade.ErrResendVerificationNoEmail):
			return c.Status(http.StatusBadRequest).JSON(web.ErrResp{
				Error: "User does not have an email address",
			})
		case errors.Is(err, facade.ErrTooManyRequests):
			return c.Status(http.StatusTooManyRequests).JSON(web.ErrResp{
				Error: "Please wait before requesting another code",
			})
		case errors.Is(err, facade.ErrSendVerifyEmailUnsubscribed):
			return c.Status(http.StatusBadRequest).JSON(web.ErrResp{
				Error: fmt.Sprintf("User is unsubscribed. You may contact us at mailto:%s to resubscribe", a.contactEmail),
			})
		default:
			a.log.Error("resend verification", zap.Error(err))
			return c.Status(http.StatusInternalServerError).JSON(web.ErrResp{
				Error: internalErrorMsg,
			})
		}
	}

	return c.SendStatus(http.StatusNoContent)
}
