package handlers

import (
	"errors"
	"net/http"

	"github.com/OutOfStack/game-library-auth/internal/facade"
	"github.com/OutOfStack/game-library-auth/internal/web"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// SignUpHandler godoc
// @Summary		Register a new user
// @Description Create a new user account with the provided information
// @Tags  		auth
// @Accept 		json
// @Produce 	json
// @Param 		signup body SignUpReq true "User signup information"
// @Success		200 {object} TokenResp 	 "User credentials"
// @Failure 	400 {object} web.ErrResp "Invalid input data"
// @Failure 	409 {object} web.ErrResp "Username or publisher name already exists"
// @Failure 	500 {object} web.ErrResp "Internal server error"
// @Router		/signup [post]
func (a *AuthAPI) SignUpHandler(c *fiber.Ctx) error {
	ctx, span := tracer.Start(c.Context(), "signUp")
	defer span.End()

	var signUp SignUpReq
	if err := c.BodyParser(&signUp); err != nil {
		a.log.Error("parsing data", zap.Error(err))
		return c.Status(http.StatusBadRequest).JSON(web.ErrResp{
			Error: "Error parsing data",
		})
	}

	log := a.log.With(zap.String("username", signUp.Username))

	// validate
	if fields, err := web.Validate(signUp); err != nil {
		log.Info("validating sign up data", zap.Error(err))
		return c.Status(http.StatusBadRequest).JSON(web.ErrResp{
			Error:  validationErrorMsg,
			Fields: fields,
		})
	}

	// sign up
	user, err := a.userFacade.SignUp(ctx, signUp.Username, signUp.DisplayName, signUp.Email, signUp.Password, signUp.IsPublisher)
	if err != nil {
		switch {
		case errors.Is(err, facade.ErrSignUpUsernameExists):
			log.Info("username already exists")
			return c.Status(http.StatusConflict).JSON(web.ErrResp{
				Error: "This username is already taken",
			})
		case errors.Is(err, facade.ErrSignUpEmailExists):
			log.Info("email already exists")
			return c.Status(http.StatusConflict).JSON(web.ErrResp{
				Error: "This email is already taken",
			})
		case errors.Is(err, facade.ErrSignUpPublisherNameExists):
			log.Info("publisher already exists")
			return c.Status(http.StatusConflict).JSON(web.ErrResp{
				Error: "Publisher with this name already exists",
			})
		case errors.Is(err, facade.ErrSignUpEmailRequired):
			log.Info("email is required")
			return c.Status(http.StatusBadRequest).JSON(web.ErrResp{
				Error: "Email is required",
			})
		default:
			log.Error("sign up", zap.Error(err))
			return c.Status(http.StatusInternalServerError).JSON(web.ErrResp{
				Error: internalErrorMsg,
			})
		}
	}

	// create tokens
	tokens, err := a.userFacade.CreateTokens(ctx, user)
	if err != nil {
		log.Error("creating tokens", zap.Error(err))
		return c.Status(http.StatusInternalServerError).JSON(web.ErrResp{Error: internalErrorMsg})
	}

	// set refresh token as a cookie
	a.setRefreshTokenCookie(c, tokens.RefreshToken)

	return c.JSON(TokenResp{
		AccessToken: tokens.AccessToken,
	})
}
