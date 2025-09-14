package handlers

import (
	"net/http"

	"github.com/OutOfStack/game-library-auth/internal/web"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// VerifyTokenHandler 	godoc
// @Summary 			Verify JWT token
// @Description 		Validates a JWT token and returns if it's valid
// @Tags 				auth
// @Accept 				json
// @Produce 			json
// @Param 				token body VerifyTokenReq true "Token to verify"
// @Success 			200 {object} VerifyTokenResp
// @Failure 			400 {object} web.ErrResp
// @Router 				/token/verify [post]
func (a *AuthAPI) VerifyTokenHandler(c *fiber.Ctx) error {
	_, span := tracer.Start(c.Context(), "verifyToken")
	defer span.End()

	var verifyToken VerifyTokenReq
	if err := c.BodyParser(&verifyToken); err != nil {
		a.log.Error("parsing data", zap.Error(err))
		return c.Status(http.StatusBadRequest).JSON(web.ErrResp{
			Error: "Error parsing data",
		})
	}

	if fields, err := web.Validate(verifyToken); err != nil {
		a.log.Error("validating token", zap.Error(err))
		return c.Status(http.StatusBadRequest).JSON(web.ErrResp{
			Error:  validationErrorMsg,
			Fields: fields,
		})
	}

	// validate token
	_, err := a.auth.ValidateToken(verifyToken.Token)
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
