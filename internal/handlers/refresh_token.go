package handlers

import (
	"errors"
	"net/http"

	"github.com/OutOfStack/game-library-auth/internal/facade"
	"github.com/OutOfStack/game-library-auth/internal/web"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// RefreshTokenHandler godoc
// @Summary      Refresh access token
// @Description  Use a refresh token from httpOnly cookie to obtain a new access token
// @Tags         auth
// @Produce      json
// @Success      200 {object} TokenResp
// @Failure      401 {object} web.ErrResp "Invalid or expired refresh token"
// @Failure      500 {object} web.ErrResp
// @Router       /refresh [post]
func (a *AuthAPI) RefreshTokenHandler(c *fiber.Ctx) error {
	ctx, span := tracer.Start(c.Context(), "refreshToken")
	defer span.End()

	// read refresh token from cookie
	refreshToken := c.Cookies(refreshTokenCookieName)
	if refreshToken == "" {
		a.log.Info("refresh token cookie not found")
		return c.Status(http.StatusUnauthorized).JSON(web.ErrResp{
			Error: "Refresh token not found",
		})
	}

	// refresh access token
	accessToken, err := a.userFacade.RefreshAccessToken(ctx, refreshToken)
	if err != nil {
		switch {
		case errors.Is(err, facade.ErrRefreshTokenNotFound):
			a.log.Info("refresh token not found")
			return c.Status(http.StatusUnauthorized).JSON(web.ErrResp{
				Error: "Invalid refresh token",
			})
		case errors.Is(err, facade.ErrRefreshTokenExpired):
			a.log.Info("refresh token expired")
			return c.Status(http.StatusUnauthorized).JSON(web.ErrResp{
				Error: "Refresh token expired",
			})
		default:
			a.log.Error("refresh access token", zap.Error(err))
			return c.Status(http.StatusInternalServerError).JSON(web.ErrResp{
				Error: internalErrorMsg,
			})
		}
	}

	return c.JSON(TokenResp{
		AccessToken: accessToken,
	})
}
