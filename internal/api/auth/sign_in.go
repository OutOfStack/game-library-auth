package auth

import (
	"errors"
	"net/http"

	"github.com/OutOfStack/game-library-auth/internal/api/openapi"
	"github.com/OutOfStack/game-library-auth/internal/facade"
	"github.com/OutOfStack/game-library-auth/internal/web"
	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

// SignIn authenticates user with username and password, returns access token and sets refresh token cookie
func (a *API) SignIn(c fiber.Ctx) error {
	ctx, span := tracer.Start(c.Context(), "signIn")
	defer span.End()

	var signIn openapi.SignInReq
	if err := c.Bind().Body(&signIn); err != nil {
		a.log.Error("parsing data", zap.Error(err))
		return c.Status(http.StatusBadRequest).JSON(web.ErrResp{
			Error: parsingDataErrorMsg,
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

	// create tokens
	tokens, err := a.userFacade.CreateTokens(ctx, user)
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
