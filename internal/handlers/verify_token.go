package handlers

import (
	"net/http"

	"github.com/OutOfStack/game-library-auth/internal/web"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// VerifyToken verifies token
func (a *AuthAPI) VerifyToken(c *fiber.Ctx) error {
	_, span := tracer.Start(c.Context(), "handlers.verifyToken")
	defer span.End()

	var verifyToken VerifyToken
	if err := c.BodyParser(&verifyToken); err != nil {
		a.log.Error("parsing data", zap.Error(err))
		return c.Status(http.StatusBadRequest).JSON(web.ErrResp{
			Error: "Error parsing data",
		})
	}

	// validate
	if fields, err := web.Validate(verifyToken); err != nil {
		a.log.Error("validating token", zap.Error(err))
		return c.Status(http.StatusBadRequest).JSON(web.ErrResp{
			Error:  validationErrorMsg,
			Fields: fields,
		})
	}

	tokenStr := verifyToken.Token

	_, err := a.auth.ValidateToken(tokenStr)
	if err != nil {
		a.log.Error("token validation failed", zap.Error(err))
		return c.JSON(VerifyTokenResp{
			Valid: false,
		})
	}

	return c.JSON(VerifyTokenResp{
		Valid: true,
	})
}
