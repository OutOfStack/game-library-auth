package handlers_test

import (
	"testing"

	"github.com/OutOfStack/game-library-auth/internal/appconf"
	"github.com/OutOfStack/game-library-auth/internal/handlers"
	mocks "github.com/OutOfStack/game-library-auth/internal/handlers/mocks"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

const (
	internalErrorMsg string = "Internal error"
	authErrorMsg     string = "Incorrect username or password"
)

func setupTest(t *testing.T, cfg *appconf.Cfg) (*mocks.MockAuth, *mocks.MockStorage, *mocks.MockGoogleTokenValidator, *handlers.AuthAPI, *fiber.App, *gomock.Controller) {
	t.Helper()

	ctrl := gomock.NewController(t)
	mockAuth := mocks.NewMockAuth(ctrl)
	mockStorage := mocks.NewMockStorage(ctrl)
	googleTokenValidator := mocks.NewMockGoogleTokenValidator(ctrl)

	logger := zap.NewNop()
	if cfg == nil {
		cfg = &appconf.Cfg{
			Auth: appconf.Auth{
				GoogleClientID: "test-client-id",
			},
		}
	}
	authAPI, err := handlers.NewAuthAPI(logger, cfg, mockAuth, mockStorage, googleTokenValidator)
	require.NoError(t, err)

	return mockAuth, mockStorage, googleTokenValidator, authAPI, fiber.New(), ctrl
}
