package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/OutOfStack/game-library-auth/internal/facade"
	"github.com/OutOfStack/game-library-auth/internal/web"
	"github.com/gofiber/fiber/v3"
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
// @Failure 		  403 {object} web.ErrResp
// @Failure 		  409 {object} web.ErrResp
// @Router 			  /oauth/google [post]
func (a *API) GoogleOAuthHandler(c fiber.Ctx) error {
	ctx, span := tracer.Start(c.Context(), "googleOAuth")
	defer span.End()

	var req GoogleOAuthRequest
	if err := c.Bind().Body(&req); err != nil {
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
	user, err := a.userFacade.GoogleOAuth(ctx, googleClaims.Sub, googleClaims.Email, googleClaims.EmailVerified)
	if err != nil {
		switch {
		case errors.Is(err, facade.ErrOAuthEmailUnverified):
			return c.Status(http.StatusForbidden).JSON(web.ErrResp{
				Error: "Google email is not verified. Please verify your email and try again.",
			})
		case errors.Is(err, facade.ErrInvalidEmail):
			return c.Status(http.StatusBadRequest).JSON(web.ErrResp{
				Error: "Invalid email",
			})
		case errors.Is(err, facade.ErrOAuthPublisherConflict):
			return c.Status(http.StatusConflict).JSON(web.ErrResp{
				Error: "Publisher account found. Please sign in with your username and password.",
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

	// create tokens
	tokens, err := a.userFacade.CreateTokens(ctx, user)
	if err != nil {
		a.log.Error("creating tokens", zap.Error(err))
		return c.Status(http.StatusInternalServerError).JSON(web.ErrResp{
			Error: internalErrorMsg,
		})
	}

	// set refresh token as a cookie
	a.setRefreshTokenCookie(c, tokens.RefreshToken)

	return c.JSON(TokenResp{
		AccessToken: tokens.AccessToken,
	})
}

// verifyGoogleIDToken verifies Google ID token and returns claims
func (a *API) verifyGoogleIDToken(ctx context.Context, token string) (*googleIDTokenClaims, error) {
	payload, err := a.googleIDTokenClient.ValidateIDToken(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("validate google id token: %w", err)
	}

	email, _ := payload.Claims["email"].(string)
	emailVerified, _ := payload.Claims["email_verified"].(bool)

	claims := &googleIDTokenClaims{
		Sub:           payload.Subject,
		Email:         email,
		EmailVerified: emailVerified,
	}

	if claims.Sub == "" || claims.Email == "" {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}
