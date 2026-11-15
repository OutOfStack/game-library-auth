package auth

import (
	"errors"
	"net/http"

	"github.com/OutOfStack/game-library-auth/internal/api/openapi"
	"github.com/OutOfStack/game-library-auth/internal/facade"
	"github.com/OutOfStack/game-library-auth/internal/web"
	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

// VerifyEmail verifies user email with code sent via email, returns new access token and sets refresh token cookie
func (a *API) VerifyEmail(c fiber.Ctx) error {
	ctx, span := tracer.Start(c.Context(), "verifyEmail")
	defer span.End()

	claims, err := a.getClaims(c)
	if err != nil {
		a.log.Error("get claims", zap.Error(err))
		return c.Status(http.StatusUnauthorized).JSON(web.ErrResp{
			Error: invalidAuthTokenMsg,
		})
	}

	var req openapi.VerifyEmailReq
	if err = c.Bind().Body(&req); err != nil {
		a.log.Error("parsing data", zap.Error(err))
		return c.Status(http.StatusBadRequest).JSON(web.ErrResp{
			Error: invalidRequestMsg,
		})
	}

	if fields, vErr := web.Validate(req); vErr != nil {
		a.log.Info("validating verify email data", zap.Error(vErr))
		return c.Status(http.StatusBadRequest).JSON(web.ErrResp{
			Error:  validationErrorMsg,
			Fields: fields,
		})
	}

	// verify user email
	verifiedUser, err := a.userFacade.VerifyEmail(ctx, claims.UserID, req.Code)
	if err != nil {
		switch {
		case errors.Is(err, facade.ErrVerifyEmailUserNotFound):
			return c.Status(http.StatusNotFound).JSON(web.ErrResp{
				Error: "User not found",
			})
		case errors.Is(err, facade.ErrVerifyEmailAlreadyVerified):
			return c.Status(http.StatusBadRequest).JSON(web.ErrResp{
				Error: emailAlreadyVerifiedMsg,
			})
		case errors.Is(err, facade.ErrVerifyEmailInvalidOrExpired):
			return c.Status(http.StatusBadRequest).JSON(web.ErrResp{
				Error: invalidOrExpiredVrfCodeMsg,
			})
		default:
			a.log.Error("verify email", zap.Error(err))
			return c.Status(http.StatusInternalServerError).JSON(web.ErrResp{
				Error: internalErrorMsg,
			})
		}
	}

	// create tokens
	tokens, err := a.userFacade.CreateTokens(ctx, verifiedUser)
	if err != nil {
		a.log.Error("creating tokens", zap.Error(err))
		return c.Status(http.StatusInternalServerError).JSON(web.ErrResp{
			Error: internalErrorMsg,
		})
	}

	// set refresh token as a cookie
	a.setRefreshTokenCookie(c, tokens.RefreshToken)

	return c.JSON(openapi.TokenResp{
		AccessToken: tokens.AccessToken,
	})
}
