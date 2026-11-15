package auth

import (
	"errors"
	"net/http"

	"github.com/OutOfStack/game-library-auth/internal/api/openapi"
	"github.com/OutOfStack/game-library-auth/internal/facade"
	"github.com/OutOfStack/game-library-auth/internal/model"
	"github.com/OutOfStack/game-library-auth/internal/web"
	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

// UpdateProfile updates user name and/or password, returns new access token and sets refresh token cookie
func (a *API) UpdateProfile(c fiber.Ctx) error {
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

	var params openapi.UpdateProfileReq
	if err = c.Bind().Body(&params); err != nil {
		a.log.Error("parsing data", zap.Error(err))
		return c.Status(http.StatusBadRequest).JSON(web.ErrResp{
			Error: parsingDataErrorMsg,
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
		case errors.Is(err, facade.ErrUpdateProfileUserNotFound):
			return c.Status(http.StatusNotFound).JSON(web.ErrResp{
				Error: "User does not exist",
			})
		case errors.Is(err, facade.ErrUpdateProfileNotAllowed):
			return c.Status(http.StatusBadRequest).JSON(web.ErrResp{
				Error: "Cannot change password for OAuth provider users",
			})
		case errors.Is(err, facade.ErrUpdateProfileInvalidPassword):
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

	// create tokens
	tokens, err := a.userFacade.CreateTokens(ctx, updatedUser)
	if err != nil {
		log.Error("creating tokens", zap.Error(err))
		return c.Status(http.StatusInternalServerError).JSON(web.ErrResp{
			Error: internalErrorMsg,
		})
	}

	// set refresh token as a cookie
	a.setRefreshTokenCookie(c, tokens.RefreshToken)

	return c.JSON(openapi.TokenResp{
		AccessToken: tokens.AccessToken,
	})
}
