package facade_test

import (
	"testing"

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

	logger := zap.NewNop()
	provider := facade.New(logger, mockUserRepo, mockEmailSender, false)

	return provider, mockUserRepo, mockEmailSender, ctrl
}
