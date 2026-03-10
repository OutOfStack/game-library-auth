package auth_test

import (
	"testing"

	"github.com/OutOfStack/game-library-auth/internal/api/auth"
	mocks "github.com/OutOfStack/game-library-auth/internal/api/auth/mocks"
	"github.com/OutOfStack/game-library-auth/internal/appconf"
	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

const (
	internalErrorMsg string = "Internal error"
	authErrorMsg     string = "Incorrect username or password"
)

func setupTest(t *testing.T, cfg *appconf.Cfg) (
	*mocks.MockGoogleIDTokenClient, *mocks.MockGitHubOAuthClient, *auth.API, *mocks.MockUserFacade, *fiber.App, *gomock.Controller) {
	t.Helper()

	ctrl := gomock.NewController(t)
	mockGoogleTokenValidator := mocks.NewMockGoogleIDTokenClient(ctrl)
	mockGitHubOAuthClient := mocks.NewMockGitHubOAuthClient(ctrl)
	mockUserFacade := mocks.NewMockUserFacade(ctrl)

	logger := zap.NewNop()
	if cfg == nil {
		cfg = &appconf.Cfg{
			OAuth: appconf.OAuth{
				GoogleClientID:     "test-client-id",
				GitHubClientID:     "test-github-client-id",
				GitHubClientSecret: "test-github-client-secret",
			},
			EmailSender: appconf.EmailSender{
				ContactEmail: "contact@example.com",
			},
			Web: appconf.Web{
				RefreshCookieSameSite: "strict",
				RefreshCookieSecure:   true,
			},
		}
	}
	authAPICfg := auth.APICfg{
		ContactEmail:               cfg.EmailSender.ContactEmail,
		RefreshTokenCookieSameSite: cfg.Web.RefreshCookieSameSite,
		RefreshTokenCookieSecure:   cfg.Web.RefreshCookieSecure,
	}
	authAPI, err := auth.NewAPI(logger, mockGoogleTokenValidator, mockGitHubOAuthClient, mockUserFacade, authAPICfg)
	require.NoError(t, err)

	return mockGoogleTokenValidator, mockGitHubOAuthClient, authAPI, mockUserFacade, fiber.New(), ctrl
}
