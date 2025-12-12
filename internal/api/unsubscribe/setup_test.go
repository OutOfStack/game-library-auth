package unsubscribe_test

import (
	"testing"

	"github.com/OutOfStack/game-library-auth/internal/api/unsubscribe"
	mocks "github.com/OutOfStack/game-library-auth/internal/api/unsubscribe/mocks"
	"github.com/OutOfStack/game-library-auth/internal/auth"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

func setupTest(t *testing.T) (*unsubscribe.API, *mocks.MockUnsubscribeFacade, *gomock.Controller, *auth.UnsubscribeTokenGenerator, *mocks.MockViews) {
	t.Helper()

	ctrl := gomock.NewController(t)
	mockFacade := mocks.NewMockUnsubscribeFacade(ctrl)
	tokenGen := auth.NewUnsubscribeTokenGenerator([]byte("test-secret-key"))
	mockViews := mocks.NewMockViews(ctrl)

	api := unsubscribe.NewAPI(zap.NewNop(), tokenGen, mockFacade, "test@example.com")

	return api, mockFacade, ctrl, tokenGen, mockViews
}
