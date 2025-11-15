package handlers

import (
	"net/http"
	"time"

	"github.com/OutOfStack/game-library-auth/internal/facade"
	"github.com/OutOfStack/game-library-auth/internal/web"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// LogoutHandler godoc
// @Summary      Logout user
// @Description  Revokes the refresh token and clears the refresh token cookie
// @Tags         auth
// @Success      204 "Successfully logged out"
// @Failure      500 {object} web.ErrResp
// @Router       /logout [post]
func (a *AuthAPI) LogoutHandler(c *fiber.Ctx) error {
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
