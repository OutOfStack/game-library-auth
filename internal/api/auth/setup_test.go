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
	*mocks.MockGoogleTokenValidator, *auth.API, *mocks.MockUserFacade, *fiber.App, *gomock.Controller) {
	t.Helper()

	ctrl := gomock.NewController(t)
	mockGoogleTokenValidator := mocks.NewMockGoogleTokenValidator(ctrl)
	mockUserFacade := mocks.NewMockUserFacade(ctrl)

	logger := zap.NewNop()
	if cfg == nil {
		cfg = &appconf.Cfg{
			Auth: appconf.Auth{
				GoogleClientID: "test-client-id",
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
		GoogleOAuthClientID:        cfg.Auth.GoogleClientID,
		ContactEmail:               cfg.EmailSender.ContactEmail,
		RefreshTokenCookieSameSite: cfg.Web.RefreshCookieSameSite,
		RefreshTokenCookieSecure:   cfg.Web.RefreshCookieSecure,
	}
	authAPI, err := auth.NewAPI(logger, mockGoogleTokenValidator, mockUserFacade, authAPICfg)
	require.NoError(t, err)

	return mockGoogleTokenValidator, authAPI, mockUserFacade, fiber.New(), ctrl
}
