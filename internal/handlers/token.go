package handlers

import (
	"log"
	"net/http"

	"github.com/OutOfStack/game-library-auth/internal/auth"
	"github.com/OutOfStack/game-library-auth/internal/web"
	"github.com/gofiber/fiber/v2"
)

// VerifyToken is a request type for JWT verification
type VerifyToken struct {
	Token string `json:"token" validate:"jwt"`
}

// VerifyTokenResp is a response type for JWT verification
type VerifyTokenResp struct {
	Valid bool `json:"valid"`
}

// TokenAPI describes dependencies for token endpoints
type TokenAPI struct {
	Auth *auth.Auth
	Log  *log.Logger
}

func (ta TokenAPI) verifyJWT(c *fiber.Ctx) error {
	var verifyToken VerifyToken
	if err := c.BodyParser(&verifyToken); err != nil {
		ta.Log.Printf("error parsing data: %v\n", err)
		return c.Status(http.StatusBadRequest).JSON(web.ErrResp{
			Error: "Error parsing data",
		})
	}

	// validate
	if fields, err := web.Validate(verifyToken); err != nil {
		ta.Log.Printf("validation error: %v\n", err)
		return c.Status(http.StatusBadRequest).JSON(web.ErrResp{
			Error:  validationErrorMsg,
			Fields: fields,
		})
	}

	tokenStr := verifyToken.Token

	_, err := ta.Auth.ValidateToken(tokenStr)
	if err != nil {
		ta.Log.Printf("token validation failed: %w", err)
		return c.JSON(VerifyTokenResp{
			Valid: false,
		})
	}

	return c.JSON(VerifyTokenResp{
		Valid: true,
	})
}
