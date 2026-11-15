package auth

import (
	"net/http"
	"time"

	"github.com/OutOfStack/game-library-auth/internal/facade"
	"github.com/OutOfStack/game-library-auth/internal/web"
	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

// DeleteAccount deletes user account and clears refresh token cookie
func (a *API) DeleteAccount(c fiber.Ctx) error {
	ctx, span := tracer.Start(c.Context(), "deleteAccount")
	defer span.End()

	// get user ID from JWT
	userID, err := a.getUserIDFromJWT(c)
	if err != nil {
		a.log.Error("extracting user ID from JWT", zap.Error(err))
		return c.Status(http.StatusUnauthorized).JSON(web.ErrResp{
			Error: invalidAuthTokenMsg,
		})
	}

	log := a.log.With(zap.String("userId", userID))

	// delete user
	if err = a.userFacade.DeleteUser(ctx, userID); err != nil {
		log.Error("delete user", zap.String("userID", userID), zap.Error(err))
		return c.Status(http.StatusInternalServerError).JSON(web.ErrResp{
			Error: internalErrorMsg,
		})
	}

	// clear refresh token cookie
	a.setRefreshTokenCookie(c, facade.RefreshToken{ExpiresAt: time.Unix(0, 0)})

	return c.SendStatus(http.StatusNoContent)
}
