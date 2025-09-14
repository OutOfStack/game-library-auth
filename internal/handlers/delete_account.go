package handlers

import (
	"net/http"

	"github.com/OutOfStack/game-library-auth/internal/web"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// DeleteAccountHandler godoc
// @Summary 			Delete user account
// @Description 		Deletes a user account
// @Tags 				auth
// @Produce 			json
// @Param 				Authorization header string true "Bearer token"
// @Success 			204 "Successfully deleted account"
// @Failure 			401 {object} web.ErrResp "Unauthorized"
// @Failure 			500 {object} web.ErrResp "Internal server error"
// @Router 				/account [delete]
func (a *AuthAPI) DeleteAccountHandler(c *fiber.Ctx) error {
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

	return c.SendStatus(http.StatusNoContent)
}
