package auth

import (
	"net/http"

	"github.com/OutOfStack/game-library-auth/internal/model"
	"github.com/OutOfStack/game-library-auth/internal/web"
	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

// GitHubOAuthHandler godoc
// @Summary 		  GitHub OAuth sign in handler
// @Description 	  Handles GitHub OAuth 2.0 authentication using authorization code
// @Tags 			  auth
// @Accept 			  json
// @Produce 		  json
// @Param 			  code body GitHubOAuthRequest true "GitHub OAuth authorization code"
// @Success 		  200 {object} TokenResp "User credentials"
// @Failure 		  400 {object} web.ErrResp
// @Failure 		  401 {object} web.ErrResp
// @Failure 		  403 {object} web.ErrResp
// @Failure 		  409 {object} web.ErrResp
// @Router 			  /oauth/github [post]
func (a *API) GitHubOAuthHandler(c fiber.Ctx) error {
	ctx, span := tracer.Start(c.Context(), "gitHubOAuth")
	defer span.End()

	var req GitHubOAuthRequest
	if err := c.Bind().Body(&req); err != nil {
		a.log.Error("parsing data", zap.Error(err))
		return c.Status(http.StatusBadRequest).JSON(web.ErrResp{
			Error: "Cannot parse request",
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

	return c.JSON(TokenResp{
		AccessToken: tokens.AccessToken,
	})
}
