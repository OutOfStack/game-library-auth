package handlers

import (
	"errors"
	"net/http"

	"github.com/OutOfStack/game-library-auth/internal/facade"
	"github.com/OutOfStack/game-library-auth/internal/model"
	"github.com/OutOfStack/game-library-auth/internal/web"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// UpdateProfileHandler godoc
// @Summary 			Update user profile
// @Description 		Updates the profile information of a user
// @Tags 				auth
// @Security     		Bearer
// @Accept 				json
// @Produce 			json
// @Param 				Authorization header string true "Bearer token"
// @Param 				profile body UpdateProfileReq true "Update profile parameters"
// @Success 			200 {object} TokenResp "Returns new access token"
// @Failure 			400 {object} web.ErrResp "Bad request"
// @Failure 			401 {object} web.ErrResp "Invalid password or token"
// @Failure 			404 {object} web.ErrResp "User not found"
// @Failure 			500 {object} web.ErrResp "Internal server error"
// @Router 				/account [patch]
func (a *AuthAPI) UpdateProfileHandler(c *fiber.Ctx) error {
	ctx, span := tracer.Start(c.Context(), "updateProfile")
	defer span.End()

	// get user ID from JWT
	userID, err := a.getUserIDFromJWT(c)
	if err != nil {
		a.log.Error("extracting user ID from JWT", zap.Error(err))
		return c.Status(http.StatusUnauthorized).JSON(web.ErrResp{
			Error: invalidAuthTokenMsg,
		})
	}

	var params UpdateProfileReq
	if err = c.BodyParser(&params); err != nil {
		a.log.Error("parsing data", zap.Error(err))
		return c.Status(http.StatusBadRequest).JSON(web.ErrResp{
			Error: "Error parsing data",
		})
	}

	log := a.log.With(zap.String("userId", userID))

	// validate
	if fields, validateErr := web.Validate(params); validateErr != nil {
		log.Info("validating update profile data", zap.Error(validateErr))
		return c.Status(http.StatusBadRequest).JSON(web.ErrResp{
			Error:  validationErrorMsg,
			Fields: fields,
		})
	}

	if params.Password == nil && params.NewPassword != nil {
		return c.Status(http.StatusBadRequest).JSON(web.ErrResp{
			Error: "Empty password",
		})
	}
	if params.Password != nil && params.NewPassword == nil {
		return c.Status(http.StatusBadRequest).JSON(web.ErrResp{
			Error: "Empty new password",
		})
	}
	if params.NewPassword != nil && params.ConfirmNewPassword != nil && *params.NewPassword != *params.ConfirmNewPassword {
		return c.Status(http.StatusBadRequest).JSON(web.ErrResp{
			Error: "Confirm password does not match",
		})
	}

	// update profile
	updatedUser, err := a.userFacade.UpdateUserProfile(ctx, userID, model.UpdateProfileParams{
		Name:        params.Name,
		Password:    params.Password,
		NewPassword: params.NewPassword,
	})
	if err != nil {
		switch {
		case errors.Is(err, facade.UpdateProfileUserNotFoundErr):
			return c.Status(http.StatusNotFound).JSON(web.ErrResp{
				Error: "User does not exist",
			})
		case errors.Is(err, facade.UpdateProfileNotAllowedErr):
			return c.Status(http.StatusBadRequest).JSON(web.ErrResp{
				Error: "Cannot change password for OAuth provider users",
			})
		case errors.Is(err, facade.UpdateProfileInvalidPasswordErr):
			return c.Status(http.StatusUnauthorized).JSON(web.ErrResp{
				Error: "Invalid current password",
			})
		default:
			log.Error("update profile", zap.Error(err))
			return c.Status(http.StatusInternalServerError).JSON(web.ErrResp{
				Error: internalErrorMsg,
			})
		}
	}

	// issue jwt token
	claims := a.auth.CreateUserClaims(updatedUser)
	tokenStr, err := a.auth.GenerateToken(claims)
	if err != nil {
		log.Error("generating jwt", zap.Error(err))
		return c.Status(http.StatusInternalServerError).JSON(web.ErrResp{
			Error: internalErrorMsg,
		})
	}

	return c.JSON(TokenResp{
		AccessToken: tokenStr,
	})
}
