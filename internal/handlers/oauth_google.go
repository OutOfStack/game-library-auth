package handlers

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/OutOfStack/game-library-auth/internal/auth"
	"github.com/OutOfStack/game-library-auth/internal/database"
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
// @Success 		  200 {object} TokenResp
// @Failure 		  400 {object} web.ErrResp
// @Failure 		  401 {object} web.ErrResp
// @Router 			  /oauth/google [post]
func (a *AuthAPI) GoogleOAuthHandler(c *fiber.Ctx) error {
	ctx, span := tracer.Start(c.Context(), "googleOAuth")
	defer span.End()

	var req GoogleOAuthRequest
	if err := c.BodyParser(&req); err != nil {
		a.log.Error("parsing data", zap.Error(err))
		return c.Status(http.StatusBadRequest).JSON(web.ErrResp{Error: "cannot parse request"})
	}

	claims, err := a.verifyGoogleIDToken(ctx, req.IDToken)
	if err != nil {
		a.log.Error("google token verify failed", zap.Error(err))
		return c.Status(http.StatusUnauthorized).JSON(web.ErrResp{Error: "invalid token"})
	}

	user, err := a.storage.GetUserByOAuth(ctx, "google", claims.Sub)
	if errors.Is(err, database.ErrNotFound) {
		user = database.NewUser(strings.SplitN(claims.Email, "@", 2)[0], claims.Name, nil, database.UserRoleName, claims.Picture)
		user.SetOAuthID(auth.GoogleAuthTokenProvider, claims.Sub)

		if err = a.storage.CreateUser(ctx, user); err != nil {
			a.log.Error("create user (google oauth)", zap.Error(err))
			return c.Status(http.StatusInternalServerError).JSON(web.ErrResp{
				Error: internalErrorMsg,
			})
		}
	} else if err != nil {
		a.log.Error("get user (google oauth)", zap.Error(err))
		return c.Status(http.StatusInternalServerError).JSON(web.ErrResp{
			Error: internalErrorMsg,
		})
	}

	// issue token
	jwtClaims := a.auth.CreateClaims(user)
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
		return nil, err
	}

	email, _ := payload.Claims["email"].(string)
	name, _ := payload.Claims["name"].(string)
	picture, _ := payload.Claims["picture"].(string)

	claims := &googleIDTokenClaims{
		Sub:     payload.Subject,
		Email:   email,
		Name:    name,
		Picture: picture,
	}

	if claims.Sub == "" || claims.Email == "" {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}
