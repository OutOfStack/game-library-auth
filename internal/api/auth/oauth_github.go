package auth

import (
	"net/http"

	"github.com/OutOfStack/game-library-auth/internal/api/openapi"
	"github.com/OutOfStack/game-library-auth/internal/model"
	"github.com/OutOfStack/game-library-auth/internal/web"
	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

// GitHubOAuth signs in or signs up user with GitHub authorization code, returns access token and sets refresh token cookie
func (a *API) GitHubOAuth(c fiber.Ctx) error {
	ctx, span := tracer.Start(c.Context(), "gitHubOAuth")
	defer span.End()

	var req openapi.GitHubOAuthRequest
	if err := c.Bind().Body(&req); err != nil {
		a.log.Error("parsing data", zap.Error(err))
		return c.Status(http.StatusBadRequest).JSON(web.ErrResp{
			Error: invalidRequestMsg,
		})
	}

	// exchange code for user info
	githubUser, err := a.githubOAuthClient.ExchangeCodeForUser(ctx, req.Code)
	if err != nil {
		a.log.Error("github code exchange failed", zap.Error(err))
		return c.Status(http.StatusUnauthorized).JSON(web.ErrResp{
			Error: "Invalid authorization code",
		})
	}

	// sign in or sign up
	user, err := a.userFacade.GitHubOAuth(ctx, githubUser.ID, githubUser.Email, githubUser.Login, githubUser.EmailVerified)
	if err != nil {
		return handleOauthErrors(c, err, model.GitHubAuthTokenProvider)
	}

	// create tokens
	tokens, err := a.userFacade.CreateTokens(ctx, user)
	if err != nil {
		a.log.Error("creating tokens", zap.Error(err))
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
