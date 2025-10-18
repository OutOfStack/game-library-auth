package facade_test

import (
	"testing"

	"github.com/OutOfStack/game-library-auth/internal/auth"
	"github.com/OutOfStack/game-library-auth/internal/facade"
	mocks "github.com/OutOfStack/game-library-auth/internal/facade/mocks"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

func setupTest(t *testing.T) (*facade.Provider, *mocks.MockUserRepo, *mocks.MockEmailSender, *gomock.Controller) {
	t.Helper()

	ctrl := gomock.NewController(t)
	mockUserRepo := mocks.NewMockUserRepo(ctrl)
	mockEmailSender := mocks.NewMockEmailSender(ctrl)
	tokenGenerator := auth.NewUnsubscribeTokenGenerator([]byte("test-secret-key"))

	provider := facade.New(zap.NewNop(), mockUserRepo, mockEmailSender, tokenGenerator)

	return provider, mockUserRepo, mockEmailSender, ctrl
}
