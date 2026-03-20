package auth

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/OutOfStack/game-library-auth/internal/auth"
	"github.com/OutOfStack/game-library-auth/internal/facade"
	"github.com/OutOfStack/game-library-auth/internal/model"
	"github.com/OutOfStack/game-library-auth/internal/web"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

var oauthProviderNames = map[string]string{
	model.GoogleAuthTokenProvider: "Google",
	model.GitHubAuthTokenProvider: "GitHub",
}

// getClaims extracts and validates JWT from Authorization header and returns the claims
func (a *API) getClaims(c fiber.Ctx) (auth.Claims, error) {
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return auth.Claims{}, fmt.Errorf("authorization header required")
	}

	// Expected format: "Bearer <token>"
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return auth.Claims{}, fmt.Errorf("invalid authorization header format")
	}

	tokenStr := parts[1]
	claims, err := a.userFacade.GetClaimsFromAccessToken(tokenStr)
	if err != nil {
		return auth.Claims{}, fmt.Errorf("invalid or expired token: %w", err)
	}

	if _, err = uuid.Parse(claims.UserID); err != nil {
		return auth.Claims{}, fmt.Errorf("invalid user ID")
	}

	return claims, nil
}

// getUserIDFromJWT extracts and validates JWT from Authorization header and returns the user ID
func (a *API) getUserIDFromJWT(c fiber.Ctx) (string, error) {
	claims, err := a.getClaims(c)
	if err != nil {
		return "", err
	}

	return claims.UserID, nil
}

// setRefreshTokenCookie sets the refresh token as an httpOnly cookie
func (a *API) setRefreshTokenCookie(c fiber.Ctx, refreshToken facade.RefreshToken) {
	c.Cookie(&fiber.Cookie{
		Name:     refreshTokenCookieName,
		Value:    refreshToken.Token,
		Path:     "/",
		HTTPOnly: true,
		Secure:   a.cfg.RefreshTokenCookieSecure,
		SameSite: a.cfg.RefreshTokenCookieSameSite,
		Expires:  refreshToken.ExpiresAt,
	})
}

func handleOauthErrors(c fiber.Ctx, err error, oauthProvider string) error {
	switch {
	case errors.Is(err, facade.ErrOAuthEmailUnverified):
		if provider, ok := oauthProviderNames[oauthProvider]; ok {
			oauthProvider = provider
		}
		return c.Status(http.StatusForbidden).JSON(web.ErrResp{
			Error: fmt.Sprintf("%s email is not verified. Please verify your email on %s and try again.", oauthProvider, oauthProvider),
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
