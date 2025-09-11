package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"

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

	claims, err := a.verifyGoogleIDToken(ctx, req.IDToken)
	if err != nil {
		a.log.Error("google token verify failed", zap.Error(err))
		return c.Status(http.StatusUnauthorized).JSON(web.ErrResp{
			Error: "Invalid token",
		})
	}

	user, err := a.userRepo.GetUserByOAuth(ctx, "google", claims.Sub)
	if err != nil && !errors.Is(err, database.ErrNotFound) {
		a.log.Error("get user (google oauth)", zap.Error(err))
		return c.Status(http.StatusInternalServerError).JSON(web.ErrResp{
			Error: internalErrorMsg,
		})
	}

	// create user if not found
	if errors.Is(err, database.ErrNotFound) {
		username, uErr := extractUsernameFromEmail(claims.Email)
		if uErr != nil {
			a.log.Error("extract username from email", zap.String("email", claims.Email), zap.Error(uErr))
			return c.Status(http.StatusBadRequest).JSON(web.ErrResp{
				Error: "Invalid email",
			})
		}

		user = database.NewUser(username, "", nil, database.UserRoleName)
		user.SetOAuthID(auth.GoogleAuthTokenProvider, claims.Sub)
		user.SetEmail(claims.Email, true)

		// create user
		if err = a.userRepo.CreateUser(ctx, user); err != nil {
			if errors.Is(err, database.ErrUsernameExists) {
				a.log.Warn("username already exists during oauth", zap.String("username", user.Username))
				return c.Status(http.StatusConflict).JSON(web.ErrResp{
					Error: "Account setup incomplete. Please complete registration manually.",
				})
			}
			a.log.Error("create user (google oauth)", zap.String("username", user.Username), zap.Error(err))
			return c.Status(http.StatusInternalServerError).JSON(web.ErrResp{
				Error: internalErrorMsg,
			})
		}
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
