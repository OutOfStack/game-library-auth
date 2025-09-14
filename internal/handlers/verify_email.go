package handlers

import (
	"errors"
	"net/http"

	"github.com/OutOfStack/game-library-auth/internal/facade"
	"github.com/OutOfStack/game-library-auth/internal/web"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
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

	// verify user email
	verifiedUser, err := a.userFacade.VerifyEmail(ctx, claims.UserID, req.Code)
	if err != nil {
		switch {
		case errors.Is(err, facade.VerifyEmailUserNotFoundErr):
			return c.Status(http.StatusNotFound).JSON(web.ErrResp{
				Error: "User not found",
			})
		case errors.Is(err, facade.VerifyEmailAlreadyVerifiedErr):
			return c.Status(http.StatusBadRequest).JSON(web.ErrResp{
				Error: "Email is already verified",
			})
		case errors.Is(err, facade.VerifyEmailInvalidOrExpiredErr):
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

	// issue jwt token
	jwtClaims := a.auth.CreateUserClaims(verifiedUser)
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
