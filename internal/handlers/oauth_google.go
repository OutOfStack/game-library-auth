package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/OutOfStack/game-library-auth/internal/facade"
	"github.com/OutOfStack/game-library-auth/internal/web"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// GoogleOAuthHandler godoc
// @Summary 		  Google OAuth sign in handler
// @Description 	  Handles Google OAuth 2.0 authentication
// @Tags 			  auth
// @Accept 			  json
// @Produce 		  json
// @Param 			  token body GoogleOAuthRequest true "Google OAuth token"
// @Success 		  200 {object} TokenResp "User credentials"
// @Failure 		  400 {object} web.ErrResp
// @Failure 		  401 {object} web.ErrResp
// @Router 			  /oauth/google [post]
func (a *AuthAPI) GoogleOAuthHandler(c *fiber.Ctx) error {
	ctx, span := tracer.Start(c.Context(), "googleOAuth")
	defer span.End()

	var req GoogleOAuthRequest
	if err := c.BodyParser(&req); err != nil {
		a.log.Error("parsing data", zap.Error(err))
		return c.Status(http.StatusBadRequest).JSON(web.ErrResp{
			Error: "Cannot parse request",
		})
	}

	// verify google token
	googleClaims, err := a.verifyGoogleIDToken(ctx, req.IDToken)
	if err != nil {
		a.log.Error("google token verify failed", zap.Error(err))
		return c.Status(http.StatusUnauthorized).JSON(web.ErrResp{
			Error: "Invalid token",
		})
	}

	// sign in or sign up
	user, err := a.userFacade.GoogleOAuth(ctx, googleClaims.Sub, googleClaims.Email)
	if err != nil {
		switch {
		case errors.Is(err, facade.ErrInvalidEmail):
			return c.Status(http.StatusBadRequest).JSON(web.ErrResp{
				Error: "Invalid email",
			})
		case errors.Is(err, facade.ErrOAuthSignInConflict):
			return c.Status(http.StatusConflict).JSON(web.ErrResp{
				Error: "Account setup incomplete. Please complete registration manually.",
			})
		default:
			return c.Status(http.StatusInternalServerError).JSON(web.ErrResp{
				Error: internalErrorMsg,
			})
		}
	}

	// issue jwt token
	jwtClaims := a.auth.CreateUserClaims(user)
	tokenStr, err := a.auth.GenerateToken(jwtClaims)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(web.ErrResp{
			Error: internalErrorMsg,
		})
	}

	return c.JSON(TokenResp{
		AccessToken: tokenStr,
	})
}

// verifyGoogleIDToken verifies Google ID token and returns claims
func (a *AuthAPI) verifyGoogleIDToken(ctx context.Context, token string) (*googleIDTokenClaims, error) {
	payload, err := a.googleTokenValidator.Validate(ctx, token, a.googleOAuthClientID)
	if err != nil {
		return nil, fmt.Errorf("validate google token: %w", err)
	}

	email, _ := payload.Claims["email"].(string)

	claims := &googleIDTokenClaims{
		Sub:   payload.Subject,
		Email: email,
	}

	if claims.Sub == "" || claims.Email == "" {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}
