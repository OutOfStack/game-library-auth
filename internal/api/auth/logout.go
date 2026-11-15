package auth

import (
	"net/http"
	"time"

	"github.com/OutOfStack/game-library-auth/internal/facade"
	"github.com/OutOfStack/game-library-auth/internal/web"
	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

// Logout revokes refresh token from cookie if present and clears refresh token cookie
func (a *API) Logout(c fiber.Ctx) error {
	ctx, span := tracer.Start(c.Context(), "logout")
	defer span.End()

	// read refresh token from cookie
	refreshToken := c.Cookies(refreshTokenCookieName)
	if refreshToken != "" {
		// revoke refresh token
		if err := a.userFacade.RevokeRefreshToken(ctx, refreshToken); err != nil {
			a.log.Error("revoke refresh token", zap.Error(err))
			return c.Status(http.StatusInternalServerError).JSON(web.ErrResp{
				Error: internalErrorMsg,
			})
		}
	}

	// clear refresh token cookie
	a.setRefreshTokenCookie(c, facade.RefreshToken{ExpiresAt: time.Unix(0, 0)})

	return c.SendStatus(http.StatusNoContent)
}
