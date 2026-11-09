package auth

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/OutOfStack/game-library-auth/internal/facade"
	"github.com/OutOfStack/game-library-auth/internal/web"
	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

// ResendVerificationEmail queues email verification code for asynchronous delivery
func (a *API) ResendVerificationEmail(c fiber.Ctx) error {
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
		var tooManyRequestErr *facade.TooManyRequestsError
		switch {
		case errors.Is(err, facade.ErrVerifyEmailAlreadyVerified):
			return c.Status(http.StatusBadRequest).JSON(web.ErrResp{
				Error: emailAlreadyVerifiedMsg,
			})
		case errors.Is(err, facade.ErrResendVerificationNoEmail):
			return c.Status(http.StatusBadRequest).JSON(web.ErrResp{
				Error: "User does not have an email address",
			})
		case errors.As(err, &tooManyRequestErr):
			c.Set("Retry-After", fmt.Sprintf("%.0f", tooManyRequestErr.RetryAfter.Seconds()))
			return c.Status(http.StatusTooManyRequests).JSON(web.ErrResp{
				Error: "Please wait before requesting another code",
			})
		case errors.Is(err, facade.ErrSendVerifyEmailUnsubscribed):
			return c.Status(http.StatusBadRequest).JSON(web.ErrResp{
				Error: fmt.Sprintf("User is unsubscribed. You may contact us at mailto:%s to resubscribe", a.cfg.ContactEmail),
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
