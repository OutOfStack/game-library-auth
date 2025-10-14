package handlers

import (
	"errors"
	"net/http"

	"github.com/OutOfStack/game-library-auth/internal/facade"
	"github.com/OutOfStack/game-library-auth/internal/web"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// SignInHandler godoc
// @Summary      Sign in
// @Description  Authenticate a user and return an access token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        signin body SignInReq true "User credentials"
// @Success      200 {object} TokenResp
// @Failure      400 {object} web.ErrResp
// @Failure      401 {object} web.ErrResp
// @Failure      500 {object} web.ErrResp
// @Router       /signin [post]
func (a *AuthAPI) SignInHandler(c *fiber.Ctx) error {
	ctx, span := tracer.Start(c.Context(), "signIn")
	defer span.End()

	var signIn SignInReq
	if err := c.BodyParser(&signIn); err != nil {
		a.log.Error("parsing data", zap.Error(err))
		return c.Status(http.StatusBadRequest).JSON(web.ErrResp{
			Error: "Error parsing data",
		})
	}

	log := a.log.With(zap.String("username", signIn.Username))

	if fields, err := web.Validate(signIn); err != nil {
		log.Info("validating credentials", zap.Error(err))
		return c.Status(http.StatusBadRequest).JSON(web.ErrResp{
			Error:  validationErrorMsg,
			Fields: fields,
		})
	}

	// sign in
	user, err := a.userFacade.SignIn(ctx, signIn.Username, signIn.Password)
	if err != nil {
		switch {
		case errors.Is(err, facade.ErrSignInInvalidCredentials):
			log.Info("invalid username or password", zap.Error(err))
			return c.Status(http.StatusUnauthorized).JSON(web.ErrResp{
				Error: authErrorMsg,
			})
		default:
			log.Error("sign in", zap.Error(err))
			return c.Status(http.StatusInternalServerError).JSON(web.ErrResp{
				Error: internalErrorMsg,
			})
		}
	}

	// issue jwt token
	claims := a.auth.CreateUserClaims(user)
	tokenStr, err := a.auth.GenerateToken(claims)
	if err != nil {
		log.Error("generating jwt", zap.Error(err))
		return c.Status(http.StatusInternalServerError).JSON(web.ErrResp{
			Error: internalErrorMsg,
		})
	}

	return c.JSON(TokenResp{AccessToken: tokenStr})
}
