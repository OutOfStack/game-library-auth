package handlers_test

import (
	"testing"

	"github.com/OutOfStack/game-library-auth/internal/handlers"
	mocks "github.com/OutOfStack/game-library-auth/internal/handlers/mocks"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

const (
	internalErrorMsg string = "Internal error"
	authErrorMsg     string = "Incorrect username or password"
)

func setupTest(t *testing.T) (*mocks.MockAuth, *mocks.MockStorage, *handlers.AuthAPI, *fiber.App, *gomock.Controller) {
	t.Helper()

	ctrl := gomock.NewController(t)
	mockAuth := mocks.NewMockAuth(ctrl)
	mockStorage := mocks.NewMockStorage(ctrl)

	logger := zap.NewNop()
	authAPI := handlers.NewAuthAPI(logger, mockAuth, mockStorage)

	return mockAuth, mockStorage, authAPI, fiber.New(), ctrl
}
