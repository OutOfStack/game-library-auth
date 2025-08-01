package handlers

import (
	"errors"
	"net/http"

	"github.com/OutOfStack/game-library-auth/internal/database"
	"github.com/OutOfStack/game-library-auth/internal/web"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// UpdateProfileHandler godoc
// @Summary 			Update user profile
// @Description 		Updates the profile information of a user
// @Tags 				auth
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
			Error: invalidAuthToken,
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

	// check if user exists
	usr, err := a.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return c.Status(http.StatusNotFound).JSON(web.ErrResp{
				Error: "User does not exist",
			})
		}
		log.Error("checking existence of user", zap.Error(err))
		return c.Status(http.StatusInternalServerError).JSON(web.ErrResp{
			Error: internalErrorMsg,
		})
	}

	if params.Password != nil && usr.OAuthProvider.Valid {
		return c.Status(http.StatusBadRequest).JSON(web.ErrResp{
			Error: "Cannot change password for OAuth provider users",
		})
	}

	if params.Password != nil && len(usr.PasswordHash) > 0 {
		// check password
		if err = bcrypt.CompareHashAndPassword(usr.PasswordHash, []byte(*params.Password)); err != nil {
			log.Info("invalid password", zap.Error(err))
			return c.Status(http.StatusUnauthorized).JSON(web.ErrResp{
				Error: "Invalid current password",
			})
		}

		// hash password
		passwordHash, gErr := bcrypt.GenerateFromPassword([]byte(*params.NewPassword), bcrypt.DefaultCost)
		if gErr != nil {
			log.Error("generating password hash", zap.Error(gErr))
			return c.Status(http.StatusInternalServerError).JSON(web.ErrResp{
				Error: internalErrorMsg,
			})
		}

		usr.PasswordHash = passwordHash
	}

	if params.Name != nil {
		usr.DisplayName = *params.Name
	}

	// update user
	if err = a.userRepo.UpdateUser(ctx, usr); err != nil {
		log.Error("update user", zap.Error(err))
		return c.Status(http.StatusInternalServerError).JSON(web.ErrResp{
			Error: internalErrorMsg,
		})
	}

	// create claims
	claims := a.auth.CreateClaims(usr)

	// generate jwt
	tokenStr, err := a.auth.GenerateToken(claims)
	if err != nil {
		log.Error("generating jwt", zap.Error(err))
		return c.Status(http.StatusInternalServerError).JSON(web.ErrResp{
			Error: internalErrorMsg,
		})
	}

	// return token
	return c.JSON(TokenResp{
		AccessToken: tokenStr,
	})
}
